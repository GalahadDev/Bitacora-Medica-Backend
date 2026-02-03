package patients

import (
	"net/http"

	"bitacora-medica-backend/api/database"
	"bitacora-medica-backend/api/domains"

	"github.com/gin-gonic/gin"
)

type UpdatePatientInput struct {
	DisabilityReport string `json:"disability_report"`
	CareNotes        string `json:"care_notes"`
}

// @Summary      Update patient
// @Description  Update patient details (disability report, care notes)
// @Tags         Patients
// @Accept       json
// @Produce      json
// @Param        id     path      string              true  "Patient ID"
// @Param        input  body      UpdatePatientInput  true  "Update Data"
// @Success      200    {object}  map[string]interface{}
// @Failure      400    {object}  map[string]string
// @Failure      404    {object}  map[string]string
// @Failure      500    {object}  map[string]string
// @Router       /patients/{id} [put]
// @Security     Bearer
func UpdatePatientHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var input UpdatePatientInput

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		db := database.GetDB()
		var patient domains.Patient

		if err := db.First(&patient, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Patient not found"})
			return
		}

		patient.DisabilityReport = input.DisabilityReport
		patient.CareNotes = input.CareNotes

		if err := db.Save(&patient).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update patient"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Patient updated successfully",
			"data":    patient,
		})
	}
}
