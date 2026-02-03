package sessions

import (
	"fmt"
	"net/http"

	"bitacora-medica-backend/api/config"
	"bitacora-medica-backend/api/database"
	"bitacora-medica-backend/api/domains"
	"bitacora-medica-backend/api/services"

	"github.com/gin-gonic/gin"
)

// ListSessionsHandler obtiene sesiones con filtros
// @Summary      List sessions
// @Description  List sessions with optional filters
// @Tags         Sessions
// @Produce      json
// @Param        patient_id       query     string  false  "Filter by Patient ID"
// @Param        professional_id  query     string  false  "Filter by Professional ID"
// @Param        has_incident     query     boolean false  "Filter by Incident presence"
// @Success      200              {object}  map[string]interface{}
// @Failure      500              {object}  map[string]string
// @Router       /sessions [get]
// @Security     Bearer
func ListSessionsHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		db := database.GetDB()
		var sessions []domains.Session

		query := db.Model(&domains.Session{}).Preload("Creator")

		patientID := c.Query("patient_id")
		if patientID != "" {
			query = query.Where("patient_id = ?", patientID)
		}

		profID := c.Query("professional_id")
		if profID != "" {
			query = query.Where("professional_id = ?", profID)
		}

		incident := c.Query("has_incident")
		if incident == "true" {
			query = query.Where("has_incident = ?", true)
		}

		page := 1
		limit := 10
		if c.Query("page") != "" {
			fmt.Sscan(c.Query("page"), &page)
		}
		if c.Query("limit") != "" {
			fmt.Sscan(c.Query("limit"), &limit)
		}
		offset := (page - 1) * limit

		var total int64
		query.Count(&total)

		if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&sessions).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch sessions"})
			return
		}

		storageSvc := services.NewStorageService(cfg)
		for i := range sessions {

			if sessions[i].IncidentPhoto != "" {
				if len(sessions[i].IncidentPhoto) > 4 && sessions[i].IncidentPhoto[:4] == "http" {
				} else {
					signed, err := storageSvc.GetPublicURL("session-evidence", sessions[i].IncidentPhoto)
					if err == nil {
						sessions[i].IncidentPhoto = signed
					}
				}
			}

			if len(sessions[i].Photos) > 0 {
				var signedPhotos []string
				for _, p := range sessions[i].Photos {
					if len(p) > 4 && p[:4] == "http" {
						signedPhotos = append(signedPhotos, p)
						continue
					}
					signed, err := storageSvc.GetPublicURL("session-evidence", p)
					if err == nil {
						signedPhotos = append(signedPhotos, signed)
					} else {
						signedPhotos = append(signedPhotos, p)
					}
				}
				sessions[i].Photos = signedPhotos
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"data": sessions,
			"meta": gin.H{
				"total":     total,
				"page":      page,
				"limit":     limit,
				"last_page": (int(total) + limit - 1) / limit,
			},
		})
	}
}
