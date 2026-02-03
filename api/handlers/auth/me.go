package auth

import (
	"bitacora-medica-backend/api/domains"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetMeHandler devuelve los datos del usuario autenticado
// @Summary      Get current user
// @Description  Get the currently authenticated user's profile
// @Tags         Auth
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]string
// @Router       /auth/me [get]
// @Security     Bearer
func GetMeHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Obtener usuario autenticado (del contexto del middleware)
		currentUser := c.MustGet("currentUser").(domains.User)

		// 2. Devolver los datos
		c.JSON(http.StatusOK, gin.H{
			"user": currentUser,
		})
	}
}
