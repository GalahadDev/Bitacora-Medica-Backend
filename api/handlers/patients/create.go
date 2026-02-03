package patients

import (
	"encoding/json"
	"net/http"
	"time"

	"bitacora-medica-backend/api/config"
	"bitacora-medica-backend/api/database"
	"bitacora-medica-backend/api/domains"
	"bitacora-medica-backend/api/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
)

// Input Validado
type CreatePatientInput struct {
	FirstName      string `json:"first_name" binding:"required"`
	LastName       string `json:"last_name" binding:"required"`
	RUT            string `json:"rut" binding:"required"`
	BirthDate      string `json:"birth_date" binding:"required"` // YYYY-MM-DD
	Email          string `json:"email" binding:"required,email"`
	Phone          string `json:"phone"`
	Diagnosis      string `json:"diagnosis"`
	ConsentPDFUrl  string `json:"consent_pdf_url"`
	Sex            string `json:"sex" binding:"required"`
	EmergencyPhone string `json:"emergency_phone"`
}

func calculateAge(birthDateStr string) int {
	birthDate, err := time.Parse("2006-01-02", birthDateStr)
	if err != nil {
		return 0
	}
	now := time.Now()
	age := now.Year() - birthDate.Year()

	if now.YearDay() < birthDate.YearDay() {
		age--
	}
	return age
}

// @Summary      Create a new patient
// @Description  Create a new patient record
// @Tags         Patients
// @Accept       json
// @Produce      json
// @Param        input body CreatePatientInput true "Patient Creation Data"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /patients [post]
// @Security     Bearer
func CreatePatientHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		currentUser := c.MustGet("currentUser").(domains.User)

		var input CreatePatientInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if !utils.ValidateRUT(input.RUT) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid RUT format or verification digit"})
			return
		}

		if _, err := time.Parse("2006-01-02", input.BirthDate); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Birth date must be YYYY-MM-DD"})
			return
		}

		age := calculateAge(input.BirthDate)

		personalInfoMap := map[string]interface{}{
			"first_name":      input.FirstName,
			"last_name":       input.LastName,
			"rut":             input.RUT,
			"birth_date":      input.BirthDate,
			"email":           input.Email,
			"phone":           input.Phone,
			"diagnosis":       input.Diagnosis,
			"sex":             input.Sex,
			"age":             age,
			"emergency_phone": input.EmergencyPhone,
		}

		personalInfoBytes, err := json.Marshal(personalInfoMap)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process personal info"})
			return
		}

		patient := domains.Patient{
			CreatorID:     currentUser.ID,
			PersonalInfo:  datatypes.JSON(personalInfoBytes),
			ConsentPDFUrl: input.ConsentPDFUrl,
		}

		if err := database.GetDB().Create(&patient).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create patient"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Patient created successfully",
			"data":    patient,
		})
	}
}
