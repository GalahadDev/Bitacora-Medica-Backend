package reports

import (
	"net/http"
	"time"

	"bitacora-medica-backend/api/database"
	"bitacora-medica-backend/api/domains"

	"github.com/gin-gonic/gin"
)

type MasterReportResponse struct {
	GeneratedAt time.Time `json:"generated_at"`
	DateRange   string    `json:"date_range"`

	TotalSessions  int64 `json:"total_sessions"`
	TotalIncidents int64 `json:"total_incidents"`

	ProfessionalSummaries []ProfessionalSummary `json:"professional_summaries"`
}

type ProfessionalSummary struct {
	ProfessionalName string `json:"professional_name"`
	Role             string `json:"role"` // Ej: Fonoaudiólogo
	Summary          string `json:"summary"`
	Objectives       string `json:"objectives"`
}

// @Summary      Generate master report
// @Description  Generate a consolidated master report for date range
// @Tags         Reports
// @Produce      json
// @Param        patient_id query string true "Patient ID"
// @Param        start_date query string true "Start Date (YYYY-MM-DD)"
// @Param        end_date   query string true "End Date (YYYY-MM-DD)"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /reports/master [get]
// @Security     Bearer
func GenerateMasterReportHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req domains.MasterReportRequest
		if err := c.ShouldBindQuery(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		db := database.GetDB()

		var reports []domains.ProfessionalReport
		if err := db.Preload("Author").
			Where("patient_id = ? AND date_range_start >= ? AND date_range_end <= ?",
				req.PatientID, req.StartDate, req.EndDate).
			Find(&reports).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reports"})
			return
		}

		var totalSessions int64
		var totalIncidents int64

		db.Model(&domains.Session{}).
			Where("patient_id = ? AND created_at BETWEEN ? AND ?", req.PatientID, req.StartDate, req.EndDate).
			Count(&totalSessions)

		db.Model(&domains.Session{}).
			Where("patient_id = ? AND created_at BETWEEN ? AND ? AND has_incident = ?", req.PatientID, req.StartDate, req.EndDate, true).
			Count(&totalIncidents)

		var summaries []ProfessionalSummary

		for _, r := range reports {

			summaries = append(summaries, ProfessionalSummary{
				ProfessionalName: r.Author.Email, // Idealmente usar Name del ProfileData
				Role:             string(r.Author.Role),
				Summary:          r.Content,
				Objectives:       r.ObjectivesAchieved,
			})
		}

		response := MasterReportResponse{
			GeneratedAt:           time.Now(),
			DateRange:             req.StartDate + " to " + req.EndDate,
			TotalSessions:         totalSessions,
			TotalIncidents:        totalIncidents,
			ProfessionalSummaries: summaries,
		}

		c.JSON(http.StatusOK, gin.H{"data": response})
	}
}
