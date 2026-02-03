package admin

import (
	"net/http"

	"bitacora-medica-backend/api/config"
	"bitacora-medica-backend/api/database"
	"bitacora-medica-backend/api/domains"
	"bitacora-medica-backend/api/services"

	"github.com/gin-gonic/gin"
)

func ListPendingUsersHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var users []domains.User
		database.GetDB().Where("status = ?", domains.StatusInactive).Find(&users)
		c.JSON(http.StatusOK, gin.H{"data": users})
	}
}

type ReviewUserInput struct {
	Action       string `json:"action" binding:"required,oneof=APPROVE REJECT"`
	RejectReason string `json:"reject_reason"` // Obligatorio si es REJECT
}

func ReviewUserHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		targetUserID := c.Param("id")

		var input ReviewUserInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		db := database.GetDB()
		var user domains.User
		if err := db.First(&user, "id = ?", targetUserID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		switch input.Action {
		case "APPROVE":
			user.Status = domains.StatusActive
			user.RejectReason = ""
		case "REJECT":
			if input.RejectReason == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Reject reason is required"})
				return
			}
			user.Status = domains.StatusRejected
			user.RejectReason = input.RejectReason
		}

		if err := db.Save(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user status"})
			return
		}

		go func() {
			notifier := services.NewNotificationService(cfg)
			notifier.NotifyAccountStatus(user.ID, user.Status, user.RejectReason)
		}()

		c.JSON(http.StatusOK, gin.H{
			"message":    "User status updated successfully",
			"new_status": user.Status,
		})
	}
}
