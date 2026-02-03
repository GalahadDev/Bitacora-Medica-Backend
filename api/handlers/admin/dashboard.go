package admin

import (
	"net/http"
	"time"

	"bitacora-medica-backend/api/database"
	"bitacora-medica-backend/api/domains"

	"github.com/gin-gonic/gin"
)

func GetDashboardStatsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := database.GetDB()
		stats := domains.AdminDashboardStats{}

		// 1. Contadores Globales
		db.Model(&domains.User{}).Count(&stats.TotalUsers)
		db.Model(&domains.User{}).Where("status = ?", "INACTIVE").Count(&stats.PendingUsers)
		db.Model(&domains.Patient{}).Count(&stats.ActivePatients)
		db.Model(&domains.Session{}).Count(&stats.TotalSessions)

		// 2. Gráfico de Crecimiento

		spanishMonths := map[string]string{
			"January": "Ene", "February": "Feb", "March": "Mar", "April": "Abr",
			"May": "May", "June": "Jun", "July": "Jul", "August": "Ago",
			"September": "Sep", "October": "Oct", "November": "Nov", "December": "Dic",
		}

		growthMap := make(map[string]int64)
		monthOrder := make([]string, 0)

		for i := 5; i >= 0; i-- {
			d := time.Now().AddDate(0, -i, 0)
			key := d.Format("2006-01") // YYYY-MM
			growthMap[key] = 0
			monthOrder = append(monthOrder, key)
		}

		type MonthlyResult struct {
			MonthStr string // YYYY-MM
			Count    int64
		}
		var results []MonthlyResult

		db.Raw(`
			SELECT to_char(created_at, 'YYYY-MM') as month_str, count(*) as count 
			FROM users 
			WHERE created_at >= ? 
			GROUP BY to_char(created_at, 'YYYY-MM')
		`, time.Now().AddDate(0, -6, 0)).Scan(&results)

		for _, r := range results {
			if _, ok := growthMap[r.MonthStr]; ok {
				growthMap[r.MonthStr] = r.Count
			}
		}
		var finalGrowth []domains.UserGrowthStats
		for _, key := range monthOrder {
			t, _ := time.Parse("2006-01", key)
			label := spanishMonths[t.Format("January")]

			finalGrowth = append(finalGrowth, domains.UserGrowthStats{
				Month: label,
				Count: growthMap[key],
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"stats":  stats,
			"growth": finalGrowth,
		})
	}
}
