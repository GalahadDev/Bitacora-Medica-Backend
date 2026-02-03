package collaborations

import (
	"net/http"

	"bitacora-medica-backend/api/config"
	"bitacora-medica-backend/api/database"
	"bitacora-medica-backend/api/domains"
	"bitacora-medica-backend/api/services"

	"github.com/gin-gonic/gin"
)

// InviteCollabHandler recibe la configuración para enviar correos
// @Summary      Invite professional
// @Description  Invite another professional to collaborate on a patient
// @Tags         Collaborations
// @Accept       json
// @Produce      json
// @Param        input body domains.InviteInput true "Invitation Data"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /collaborations/invite [post]
// @Security     Bearer
func InviteCollabHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		currentUser := c.MustGet("currentUser").(domains.User)
		var input domains.InviteInput

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		db := database.GetDB()

		var patient domains.Patient
		if err := db.Where("id = ? AND creator_id = ?", input.PatientID, currentUser.ID).First(&patient).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Patient not found or you are not the creator"})
			return
		}

		var invitedUser domains.User
		if err := db.Where("email = ?", input.Email).First(&invitedUser).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User with this email not found in the platform"})
			return
		}

		if invitedUser.ID == currentUser.ID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "You cannot invite yourself"})
			return
		}

		collab := domains.Collaboration{
			PatientID:      patient.ID,
			ProfessionalID: invitedUser.ID,
			Status:         domains.CollabPending,
		}

		if err := db.Where("patient_id = ? AND professional_id = ?", patient.ID, invitedUser.ID).FirstOrCreate(&collab).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create invitation"})
			return
		}

		go func() {
			notifier := services.NewNotificationService(cfg)
			notifier.NotifyCollabInvite(invitedUser.ID, patient.ID, currentUser.Email)
		}()

		c.JSON(http.StatusCreated, gin.H{"message": "Invitation sent", "data": collab})
	}
}
