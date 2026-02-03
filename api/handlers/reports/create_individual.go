package reports

import (
	"net/http"
	"time"

	"bitacora-medica-backend/api/database"
	"bitacora-medica-backend/api/domains"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CreateReportInput struct {
	PatientID          string `json:"patient_id" binding:"required"`
	DateRangeStart     string `json:"start_date" binding:"required"` // YYYY-MM-DD
	DateRangeEnd       string `json:"end_date" binding:"required"`
	Content            string `json:"content" binding:"required"`
	ObjectivesAchieved string `json:"objectives"`
}

// @Summary      Create individual report
// @Description  Create a professional report for a patient
// @Tags         Reports
// @Accept       json
// @Produce      json
// @Param        input body CreateReportInput true "Report Data"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /reports [post]
// @Security     Bearer
func CreateIndividualReportHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		currentUser := c.MustGet("currentUser").(domains.User)
		var input CreateReportInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Parsing fechas
		start, _ := time.Parse("2006-01-02", input.DateRangeStart)
		end, _ := time.Parse("2006-01-02", input.DateRangeEnd)

		report := domains.ProfessionalReport{
			PatientID:          uuid.MustParse(input.PatientID),
			AuthorID:           currentUser.ID,
			DateRangeStart:     start,
			DateRangeEnd:       end,
			Content:            input.Content,
			ObjectivesAchieved: input.ObjectivesAchieved,
		}

		if err := database.GetDB().Create(&report).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create report"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "Report submitted", "id": report.ID})
	}
}
