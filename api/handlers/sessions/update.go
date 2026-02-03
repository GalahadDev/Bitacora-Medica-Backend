package sessions

import (
	"encoding/json"
	"net/http"

	"bitacora-medica-backend/api/database"
	"bitacora-medica-backend/api/domains"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"gorm.io/datatypes"
)

// UpdateSessionHandler permite editar una sesión (Solo el autor)
// @Summary      Update session
// @Description  Update specific session fields (Only Author or Admin)
// @Tags         Sessions
// @Accept       json
// @Produce      json
// @Param        id     path      string                    true  "Session ID"
// @Param        input  body      domains.CreateSessionInput true  "Update Data"
// @Success      200    {object}  map[string]interface{}
// @Failure      400    {object}  map[string]string
// @Failure      403    {object}  map[string]string
// @Failure      404    {object}  map[string]string
// @Failure      500    {object}  map[string]string
// @Router       /sessions/{id} [put]
// @Security     Bearer
func UpdateSessionHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		currentUser := c.MustGet("currentUser").(domains.User)

		var session domains.Session
		db := database.GetDB()
		if err := db.First(&session, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
			return
		}

		if session.ProfessionalID != currentUser.ID && currentUser.Role != domains.RoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "You can only edit your own sessions"})
			return
		}

		var input domains.CreateSessionInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if input.Vitals != nil {
			vitalsJSON, _ := json.Marshal(input.Vitals)
			session.Vitals = datatypes.JSON(vitalsJSON)
		}

		session.InterventionPlan = input.InterventionPlan
		session.Description = input.Description
		session.Achievements = input.Achievements
		session.PatientPerformance = input.PatientPerformance
		session.NextSessionNotes = input.NextSessionNotes
		session.HasIncident = input.HasIncident
		session.IncidentDetails = input.IncidentDetails
		session.IncidentPhoto = input.IncidentPhoto
		session.Photos = pq.StringArray(input.Photos)

		if err := db.Save(&session).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update session"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Session updated", "data": session})
	}
}
