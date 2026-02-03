package sessions

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"bitacora-medica-backend/api/config"
	"bitacora-medica-backend/api/database"
	"bitacora-medica-backend/api/domains"
	"bitacora-medica-backend/api/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/datatypes"
)

// CreateSessionHandler ahora requiere la configuración para enviar correos
// @Summary      Create session
// @Description  Record a new therapy session
// @Tags         Sessions
// @Accept       json
// @Produce      json
// @Param        input body domains.CreateSessionInput true "Session Data"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /sessions [post]
// @Security     Bearer
func CreateSessionHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {

		currentUserInterface, _ := c.Get("currentUser")
		currentUser := currentUserInterface.(domains.User)

		var input domains.CreateSessionInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if input.HasIncident && input.IncidentDetails == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Incident details are mandatory when an incident is reported.",
			})
			return
		}

		patientID, err := uuid.Parse(input.PatientID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Patient ID"})
			return
		}

		vitalsJSON, _ := json.Marshal(input.Vitals)

		session := domains.Session{
			PatientID:          patientID,
			ProfessionalID:     currentUser.ID,
			InterventionPlan:   input.InterventionPlan,
			Vitals:             datatypes.JSON(vitalsJSON),
			Description:        input.Description,
			Achievements:       input.Achievements,
			PatientPerformance: input.PatientPerformance,
			Photos:             pq.StringArray(input.Photos),
			HasIncident:        input.HasIncident,
			IncidentDetails:    input.IncidentDetails,
			IncidentPhoto:      input.IncidentPhoto,
			NextSessionNotes:   input.NextSessionNotes,
		}

		if err := database.DB.Create(&session).Error; err != nil {
			slog.Error("Failed to create session", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
			return
		}

		if session.HasIncident {

			go func() {
				notifier := services.NewNotificationService(cfg)
				notifier.NotifyIncident(session.PatientID, session.IncidentDetails)

				slog.Warn("INCIDENT REPORTED - Notifications triggered",
					"patient_id", session.PatientID,
					"professional", currentUser.Email)
			}()
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Session recorded successfully",
			"data":    session,
		})
	}
}
