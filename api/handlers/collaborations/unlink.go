package collaborations

import (
	"bitacora-medica-backend/api/database"
	"bitacora-medica-backend/api/domains"
	"net/http"

	"github.com/gin-gonic/gin"
)

func UnlinkProfessionalHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		currentUser := c.MustGet("currentUser").(domains.User)
		collabID := c.Param("id")

		db := database.GetDB()

		var collab domains.Collaboration
		if err := db.Preload("Patient").First(&collab, "id = ?", collabID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Collaboration not found"})
			return
		}

		if collab.Patient.CreatorID != currentUser.ID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only the Patient Creator can unlink professionals"})
			return
		}

		if collab.Status == domains.CollabRevoked {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Professional is already unlinked"})
			return
		}

		collab.Status = domains.CollabRevoked
		if err := db.Save(&collab).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unlink professional"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Professional unlinked successfully. History preserved.",
			"data":    collab,
		})
	}
}
