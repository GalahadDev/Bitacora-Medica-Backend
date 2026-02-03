package reports

import (
	"net/http"

	"bitacora-medica-backend/api/database"
	"bitacora-medica-backend/api/domains"

	"github.com/gin-gonic/gin"
)

// @Summary      List patient reports
// @Description  List reports for a specific patient
// @Tags         Reports
// @Produce      json
// @Param        patient_id query string true "Patient ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /reports/list [get]
// @Security     Bearer
func ListPatientReportsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		patientID := c.Query("patient_id")
		if patientID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "patient_id is required"})
			return
		}

		db := database.GetDB()
		var reports []domains.ProfessionalReport

		if err := db.Preload("Author").
			Where("patient_id = ?", patientID).
			Order("date_range_end DESC").
			Find(&reports).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reports"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": reports})
	}
}
