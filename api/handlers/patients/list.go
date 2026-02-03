package patients

import (
	"fmt"
	"net/http"

	"bitacora-medica-backend/api/database"
	"bitacora-medica-backend/api/domains"

	"github.com/gin-gonic/gin"
)

// ListPatientsHandler devuelve la lista de pacientes
// 1. Creados por el profesional actual
// 2. O compartidos con él mediante una colaboración ACEPTADA
// @Summary      List patients
// @Description  List patients created by or shared with the professional
// @Tags         Patients
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]string
// @Router       /patients [get]
// @Security     Bearer
func ListPatientsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		currentUser := c.MustGet("currentUser").(domains.User)
		var patients []domains.Patient

		db := database.GetDB()

		page := 1
		limit := 10
		if c.Query("page") != "" {
			fmt.Sscan(c.Query("page"), &page)
		}
		if c.Query("limit") != "" {
			fmt.Sscan(c.Query("limit"), &limit)
		}
		offset := (page - 1) * limit

		query := db.Where("creator_id = ?", currentUser.ID).
			Or("id IN (?)", db.Table("collaborations").
				Select("patient_id").
				Where("professional_id = ? AND status = ?", currentUser.ID, domains.CollabAccepted))

		var total int64
		query.Model(&domains.Patient{}).Count(&total)

		err := query.Limit(limit).Offset(offset).Find(&patients).Error

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch patients"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": patients,
			"meta": gin.H{
				"total":     total,
				"page":      page,
				"limit":     limit,
				"last_page": (int(total) + limit - 1) / limit,
			},
		})
	}
}
