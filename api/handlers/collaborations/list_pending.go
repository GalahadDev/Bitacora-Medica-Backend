package collaborations

import (
	"net/http"

	"bitacora-medica-backend/api/database"
	"bitacora-medica-backend/api/domains"

	"github.com/gin-gonic/gin"
)

// GetPendingInvitationsHandler lista las invitaciones donde soy el profesional invitado
// @Summary      List pending invitations
// @Description  Get list of pending collaboration invitations for the current user
// @Tags         Collaborations
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]string
// @Router       /collaborations/pending [get]
// @Security     Bearer
func GetPendingInvitationsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		currentUser := c.MustGet("currentUser").(domains.User)

		var invitations []domains.Collaboration

		if err := database.GetDB().
			Preload("Patient").
			Where("professional_id = ? AND status = ?", currentUser.ID, domains.CollabPending).
			Find(&invitations).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch invitations"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": invitations})
	}
}
