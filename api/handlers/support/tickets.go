package support

import (
	"net/http"

	"bitacora-medica-backend/api/config"
	"bitacora-medica-backend/api/database"
	"bitacora-medica-backend/api/domains"
	"bitacora-medica-backend/api/services"

	"github.com/gin-gonic/gin"
)

// CreateTicketHandler (Para Usuarios) - Ahora recibe config
// @Summary      Create support ticket
// @Tags         Support
// @Accept       json
// @Produce      json
// @Param        input body domains.CreateTicketInput true "Ticket Data"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /support [post]
// @Security     Bearer
func CreateTicketHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		currentUser := c.MustGet("currentUser").(domains.User)
		var input domains.CreateTicketInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ticket := domains.SupportTicket{
			UserID:  currentUser.ID,
			Subject: input.Subject,
			Message: input.Message,
			Status:  domains.TicketOpen,
		}

		if err := database.GetDB().Create(&ticket).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create ticket"})
			return
		}

		go func() {
			svc := services.NewNotificationService(cfg)
			svc.NotifyTicketCreated(currentUser.Email, ticket.Subject)
		}()

		c.JSON(http.StatusCreated, gin.H{"message": "Ticket created", "id": ticket.ID})
	}
}

// ListTicketsHandler (Dual: Admin ve todo, Usuario ve lo suyo)
// Este NO requiere cambios porque no envía correos
// @Summary      List support tickets
// @Description  List tickets (Admin sees all, User sees own)
// @Tags         Support
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /support [get]
// @Security     Bearer
func ListTicketsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		currentUser := c.MustGet("currentUser").(domains.User)
		db := database.GetDB()
		var tickets []domains.SupportTicket

		if currentUser.Role == domains.RoleAdmin {
			db.Preload("User").Order("status DESC, created_at ASC").Find(&tickets)
		} else {
			db.Where("user_id = ?", currentUser.ID).Order("created_at DESC").Find(&tickets)
		}

		c.JSON(http.StatusOK, gin.H{"data": tickets})
	}
}

// ReplyTicketHandler (Solo Admin) - Ahora recibe config
// @Summary      Reply to ticket
// @Description  Admin reply to a support ticket
// @Tags         Support
// @Accept       json
// @Produce      json
// @Param        id     path      string                  true  "Ticket ID"
// @Param        input  body      domains.ReplyTicketInput true  "Reply Data"
// @Success      200    {object}  map[string]interface{}
// @Failure      400    {object}  map[string]string
// @Failure      404    {object}  map[string]string
// @Failure      500    {object}  map[string]string
// @Router       /support/{id}/reply [put]
// @Security     Bearer
func ReplyTicketHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		ticketID := c.Param("id")
		var input domains.ReplyTicketInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		db := database.GetDB()
		var ticket domains.SupportTicket
		if err := db.First(&ticket, "id = ?", ticketID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Ticket not found"})
			return
		}

		ticket.AdminResponse = input.Response
		ticket.Status = domains.TicketClosed

		if err := db.Save(&ticket).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reply ticket"})
			return
		}

		go func() {
			svc := services.NewNotificationService(cfg)
			svc.NotifyTicketReply(ticket.UserID, ticket.Subject, input.Response)
		}()

		c.JSON(http.StatusOK, gin.H{"message": "Ticket replied and closed"})
	}
}
