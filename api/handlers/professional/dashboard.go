package professional

import (
	"net/http"
	"time"

	"bitacora-medica-backend/api/database"
	"bitacora-medica-backend/api/domains"

	"github.com/gin-gonic/gin"
)

// @Summary      Get dashboard summary
// @Description  Get dashboard statistics for the professional
// @Tags         Professional
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /dashboard/summary [get]
// @Security     Bearer
func GetMyDashboardHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		currentUser := c.MustGet("currentUser").(domains.User)
		db := database.GetDB()

		stats := domains.ProfessionalDashboardStats{}

		db.Model(&domains.Patient{}).Count(&stats.ActivePatients)

		now := time.Now()
		startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

		db.Model(&domains.Session{}).
			Where("professional_id = ? AND created_at >= ?", currentUser.ID, startOfMonth).
			Count(&stats.MonthlySessions)

		db.Model(&domains.Session{}).
			Where("professional_id = ? AND has_incident = ?", currentUser.ID, true).
			Count(&stats.ReportedIncidents)

		spanishDays := map[string]string{
			"Monday":    "Lun",
			"Tuesday":   "Mar",
			"Wednesday": "Mié",
			"Thursday":  "Jue",
			"Friday":    "Vie",
			"Saturday":  "Sáb",
			"Sunday":    "Dom",
		}

		activityMap := make(map[string]int64)
		daysOrder := make([]string, 0) // Para recordar el orden cronológico

		for i := 6; i >= 0; i-- {
			date := time.Now().AddDate(0, 0, -i)

			dateKey := date.Format("2006-01-02")

			activityMap[dateKey] = 0
			daysOrder = append(daysOrder, dateKey)
		}

		type DailyResult struct {
			Date  string // PostgreSQL devuelve YYYY-MM-DD
			Count int64
		}
		var results []DailyResult

		db.Raw(`
			SELECT to_char(created_at, 'YYYY-MM-DD') as date, count(*) as count
			FROM sessions 
			WHERE professional_id = ? AND created_at >= ?
			GROUP BY to_char(created_at, 'YYYY-MM-DD')
		`, currentUser.ID, time.Now().AddDate(0, 0, -7)).Scan(&results)

		for _, r := range results {
			if _, exists := activityMap[r.Date]; exists {
				activityMap[r.Date] = r.Count
			}
		}

		var finalActivity []domains.ActivityStats
		for _, dateKey := range daysOrder {
			t, _ := time.Parse("2006-01-02", dateKey)
			dayLabel := spanishDays[t.Format("Monday")]

			finalActivity = append(finalActivity, domains.ActivityStats{
				Day:   dayLabel,
				Count: activityMap[dateKey],
			})
		}

		var recentPatients []domains.Patient
		db.Raw(`
			SELECT DISTINCT ON (p.id) p.* FROM patients p
			JOIN sessions s ON s.patient_id = p.id
			WHERE s.professional_id = ?
			ORDER BY p.id, s.created_at DESC
			LIMIT 5
		`, currentUser.ID).Scan(&recentPatients)

		c.JSON(http.StatusOK, gin.H{
			"stats":           stats,
			"activity":        finalActivity,
			"recent_patients": recentPatients,
		})
	}
}
