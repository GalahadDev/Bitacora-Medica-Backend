package collaborations

import (
	"net/http"

	"bitacora-medica-backend/api/config"
	"bitacora-medica-backend/api/database"
	"bitacora-medica-backend/api/domains"
	"bitacora-medica-backend/api/services"

	"github.com/gin-gonic/gin"
)

type RespondInvitationInput struct {
	Status string `json:"status" binding:"required,oneof=ACCEPTED REJECTED"`
}

// RespondInvitationHandler recibe la configuración para enviar correos
// @Summary      Respond to invitation
// @Description  Accept or Reject a collaboration invitation
// @Tags         Collaborations
// @Accept       json
// @Produce      json
// @Param        id     path      string                  true  "Collaboration ID"
// @Param        input  body      RespondInvitationInput  true  "Response Data"
// @Success      200    {object}  map[string]interface{}
// @Failure      400    {object}  map[string]string
// @Failure      403    {object}  map[string]string
// @Failure      404    {object}  map[string]string
// @Failure      500    {object}  map[string]string
// @Router       /collaborations/{id}/respond [put]
// @Security     Bearer
func RespondInvitationHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		collabID := c.Param("id")
		currentUser := c.MustGet("currentUser").(domains.User)

		var input RespondInvitationInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Status must be ACCEPTED or REJECTED"})
			return
		}

		db := database.GetDB()
		var collab domains.Collaboration

		if err := db.Preload("Patient").First(&collab, "id = ?", collabID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invitation not found"})
			return
		}

		if collab.ProfessionalID != currentUser.ID {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not the recipient of this invitation"})
			return
		}

		if collab.Status != domains.CollabPending {
			c.JSON(http.StatusBadRequest, gin.H{"error": "This invitation has already been processed"})
			return
		}

		newStatus := domains.CollabStatus(input.Status)
		collab.Status = newStatus

		if err := db.Save(&collab).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update invitation status"})
			return
		}

		go func() {
			var creator domains.User
			// Buscamos al creador usando el CreatorID que viene en collab.Patient
			if err := db.First(&creator, "id = ?", collab.Patient.CreatorID).Error; err == nil {
				notifier := services.NewNotificationService(cfg)
				notifier.NotifyInviteResponse(creator.ID, currentUser.Email, newStatus)
			}
		}()

		c.JSON(http.StatusOK, gin.H{
			"message": "Invitation updated successfully",
			"status":  newStatus,
		})
	}
}
