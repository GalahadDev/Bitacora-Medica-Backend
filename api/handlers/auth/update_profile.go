package auth

import (
	"encoding/json"
	"net/http"

	"bitacora-medica-backend/api/database"
	"bitacora-medica-backend/api/domains"
	"bitacora-medica-backend/api/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
)

type UpdateProfileInput struct {
	FullName         string `json:"full_name"`
	Specialty        string `json:"specialty"`
	Phone            string `json:"phone"`
	Gender           string `json:"gender"`
	Bio              string `json:"bio"`
	BirthDate        string `json:"birth_date"`
	Rut              string `json:"rut"`
	Nationality      string `json:"nationality"`
	ResidenceCountry string `json:"residence_country"`
	University       string `json:"university"`
}

// @Summary      Update user profile
// @Description  Update the authenticated user's profile information
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        input body UpdateProfileInput true "Profile Update Data"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /auth/profile [put]
// @Security     Bearer
func UpdateProfileHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		currentUser := c.MustGet("currentUser").(domains.User)

		var input UpdateProfileInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if input.Rut != "" {
			if !utils.ValidateRUT(input.Rut) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid RUT format or verification digit"})
				return
			}
		}

		db := database.GetDB()

		var currentProfile map[string]interface{}
		if len(currentUser.ProfileData) == 0 {
			currentProfile = make(map[string]interface{})
		} else {
			if err := json.Unmarshal(currentUser.ProfileData, &currentProfile); err != nil {
				currentProfile = make(map[string]interface{})
			}
		}

		updates := map[string]string{
			"full_name":         input.FullName,
			"specialty":         input.Specialty,
			"phone":             input.Phone,
			"gender":            input.Gender,
			"bio":               input.Bio,
			"birth_date":        input.BirthDate,
			"rut":               input.Rut,
			"nationality":       input.Nationality,
			"residence_country": input.ResidenceCountry,
			"university":        input.University,
		}

		for key, val := range updates {
			if val != "" {
				currentProfile[key] = val
			}
		}

		newProfileJSON, err := json.Marshal(currentProfile)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process profile data"})
			return
		}

		// Actualizamos profile_data
		if err := db.Model(&currentUser).Update("profile_data", datatypes.JSON(newProfileJSON)).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Profile updated successfully",
			"data":    currentProfile,
		})
	}
}
