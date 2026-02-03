package sessions

import (
	"net/http"

	"bitacora-medica-backend/api/config"
	"bitacora-medica-backend/api/database"
	"bitacora-medica-backend/api/domains"
	"bitacora-medica-backend/api/services"

	"github.com/gin-gonic/gin"
)

// GetSessionHandler busca una sesión por su UUID
// @Summary      Get session details
// @Tags         Sessions
// @Produce      json
// @Param        id   path      string  true  "Session ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]string
// @Router       /sessions/{id} [get]
// @Security     Bearer
func GetSessionHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var session domains.Session
		if err := database.GetDB().First(&session, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
			return
		}

		storageSvc := services.NewStorageService(cfg)
		if len(session.Photos) > 0 {
			var signedPhotos []string
			for _, photoPath := range session.Photos {
				if len(photoPath) > 4 && photoPath[:4] == "http" {
					signedPhotos = append(signedPhotos, photoPath)
					continue
				}
				signed, err := storageSvc.GetPublicURL("session-evidence", photoPath)
				if err == nil {
					signedPhotos = append(signedPhotos, signed)
				} else {
					signedPhotos = append(signedPhotos, photoPath)
				}
			}
			session.Photos = signedPhotos
		}

		if session.IncidentPhoto != "" {
			if len(session.IncidentPhoto) <= 4 || session.IncidentPhoto[:4] != "http" {
				signed, err := storageSvc.GetPublicURL("session-evidence", session.IncidentPhoto)
				if err == nil {
					session.IncidentPhoto = signed
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{"data": session})
	}
}
