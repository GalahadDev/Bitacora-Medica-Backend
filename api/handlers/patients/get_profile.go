package patients

import (
	"net/http"

	"bitacora-medica-backend/api/config"
	"bitacora-medica-backend/api/database"
	"bitacora-medica-backend/api/domains"
	"bitacora-medica-backend/api/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TeamMember struct {
	domains.User
	CollaborationID uuid.UUID `json:"collaboration_id"`
}

type PatientProfileResponse struct {
	Patient        domains.Patient   `json:"patient"`
	Team           []TeamMember      `json:"team"`
	RecentSessions []domains.Session `json:"recent_sessions"`
	IncidentCount  int64             `json:"incident_count"`
}

// @Summary      Get patient profile
// @Description  Get detailed patient profile including team and recent sessions
// @Tags         Patients
// @Produce      json
// @Param        id   path      string  true  "Patient ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /patients/{id} [get]
// @Security     Bearer
func GetPatientProfileHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		db := database.GetDB()

		var patient domains.Patient
		if err := db.First(&patient, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Patient not found"})
			return
		}

		storageSvc := services.NewStorageService(cfg)
		if patient.ConsentPDFUrl != "" {
			signed, err := storageSvc.GetPublicURL("patients-consent", patient.ConsentPDFUrl)
			if err == nil {
				patient.ConsentPDFUrl = signed
			}
		}

		var collaborators []TeamMember
		db.Table("users").
			Select("users.*, collaborations.id as collaboration_id").
			Joins("JOIN collaborations ON collaborations.professional_id = users.id").
			Where("collaborations.patient_id = ? AND collaborations.status = ?", id, domains.CollabAccepted).
			Scan(&collaborators)

		var creator domains.User
		if err := db.First(&creator, "id = ?", patient.CreatorID).Error; err == nil {
			creatorMember := TeamMember{
				User: creator,
			}
			collaborators = append(collaborators, creatorMember)
		}

		var sessions []domains.Session
		db.Where("patient_id = ?", id).Order("created_at DESC").Limit(5).Find(&sessions)

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

		var incidentCount int64
		db.Model(&domains.Session{}).Where("patient_id = ? AND has_incident = ?", id, true).Count(&incidentCount)

		response := PatientProfileResponse{
			Patient:        patient,
			Team:           collaborators,
			RecentSessions: sessions,
			IncidentCount:  incidentCount,
		}

		c.JSON(http.StatusOK, gin.H{"data": response})
	}
}
