package patients

import (
	"encoding/json"
	"net/http"
	"time"

	"bitacora-medica-backend/api/database"
	"bitacora-medica-backend/api/domains"

	"github.com/gin-gonic/gin"
)

type AIContextResponse struct {
	Patient     PatientSummary      `json:"patient"`
	Team        []ContextTeamMember `json:"team"`
	FullHistory []SessionDetailed   `json:"full_history_sessions"`
	Reports     []ReportSummary     `json:"reports"`
	ConsentInfo string              `json:"consent_info"`
}

type PatientSummary struct {
	ID               string      `json:"id"`
	Name             string      `json:"name,omitempty"`
	PersonalInfo     interface{} `json:"personal_info"`
	DisabilityReport string      `json:"disability_report"`
	CareNotes        string      `json:"care_notes"`
	IncidentCount    int64       `json:"incident_count"`
}

type SessionDetailed struct {
	Date               time.Time              `json:"date"`
	ProfessionalName   string                 `json:"professional_name"`
	Description        string                 `json:"description"`
	Plan               string                 `json:"plan"`
	Achievements       string                 `json:"achievements"`
	PatientPerformance string                 `json:"patient_performance"`
	Vitals             map[string]interface{} `json:"vitals,omitempty"`
	HasIncident        bool                   `json:"has_incident"`
	IncidentDetails    string                 `json:"incident_details,omitempty"`
	NextSessionNotes   string                 `json:"next_session_notes,omitempty"`
}

type ContextTeamMember struct {
	Name      string `json:"name"`
	Role      string `json:"role"`
	Email     string `json:"email"`
	Specialty string `json:"specialty"`
}

type ReportSummary struct {
	Title     string    `json:"title"`
	Author    string    `json:"author"`
	DateRange string    `json:"date_range"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// GetPatientAIContextHandler devuelve todo el contexto disponible del paciente
// @Summary      Get FULL patient context for AI
// @Description  Get comprehensive patient data (profile, team, history, reports) for RAG
// @Tags         Patients
// @Produce      json
// @Param        id   path      string  true  "Patient ID"
// @Success      200  {object}  AIContextResponse
// @Failure      403  {object}  map[string]string
// @Router       /patients/{id}/ai-context [get]
// @Security     Bearer
func GetPatientAIContextHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		patientID := c.Param("id")
		db := database.GetDB()

		var patient domains.Patient
		if err := db.First(&patient, "id = ?", patientID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Patient not found"})
			return
		}

		type DBTeamMember struct {
			domains.User
			CollaborationID string `json:"collaboration_id"`
		}
		var dbCollaborators []DBTeamMember
		db.Table("users").
			Select("users.*, collaborations.id as collaboration_id").
			Joins("JOIN collaborations ON collaborations.professional_id = users.id").
			Where("collaborations.patient_id = ? AND collaborations.status = ?", patientID, domains.CollabAccepted).
			Scan(&dbCollaborators)

		var team []ContextTeamMember
		for _, m := range dbCollaborators {
			name := m.Email
			specialty := "Unknown"

			var info map[string]interface{}
			if err := json.Unmarshal(m.ProfileData, &info); err == nil {
				if val, ok := info["full_name"].(string); ok && val != "" {
					name = val
				}
				if val, ok := info["specialty"].(string); ok && val != "" {
					specialty = val
				}
			}

			team = append(team, ContextTeamMember{
				Name:      name,
				Role:      "Colaborador",
				Email:     m.Email,
				Specialty: specialty,
			})
		}

		var sessions []domains.Session
		db.Preload("Creator").
			Where("patient_id = ?", patientID).
			Order("created_at desc").
			Find(&sessions)

		var sessionHistory []SessionDetailed
		for _, s := range sessions {
			profName := s.Creator.Email

			var creatorInfo map[string]interface{}
			if err := json.Unmarshal(s.Creator.ProfileData, &creatorInfo); err == nil {
				if val, ok := creatorInfo["full_name"].(string); ok && val != "" {
					profName = val
				}
			}

			var vitals map[string]interface{}
			if s.Vitals != nil {
				_ = json.Unmarshal(s.Vitals, &vitals)
			}

			sessionHistory = append(sessionHistory, SessionDetailed{
				Date:               s.CreatedAt,
				ProfessionalName:   profName,
				Description:        s.Description,
				Plan:               s.InterventionPlan,
				Achievements:       s.Achievements,
				PatientPerformance: s.PatientPerformance,
				Vitals:             vitals,
				HasIncident:        s.HasIncident,
				IncidentDetails:    s.IncidentDetails,
				NextSessionNotes:   s.NextSessionNotes,
			})
		}

		var reports []domains.ProfessionalReport
		db.Preload("Author").
			Where("patient_id = ?", patientID).
			Order("created_at desc").
			Find(&reports)

		var reportSummaries []ReportSummary
		for _, r := range reports {
			authorName := r.Author.Email
			reportSummaries = append(reportSummaries, ReportSummary{
				Title:     "Reporte Mensual",
				Author:    authorName,
				DateRange: r.DateRangeStart.Format("2006-01-02") + " - " + r.DateRangeEnd.Format("2006-01-02"),
				Content:   r.Content + "\nObjetivos: " + r.ObjectivesAchieved,
				CreatedAt: r.CreatedAt,
			})
		}

		var incidentCount int64
		db.Model(&domains.Session{}).Where("patient_id = ? AND has_incident = ?", patientID, true).Count(&incidentCount)

		response := AIContextResponse{
			Patient: PatientSummary{
				ID:               patient.ID.String(),
				PersonalInfo:     patient.PersonalInfo,
				DisabilityReport: patient.DisabilityReport,
				CareNotes:        patient.CareNotes,
				IncidentCount:    incidentCount,
			},
			Team:        team,
			FullHistory: sessionHistory,
			Reports:     reportSummaries,
			ConsentInfo: patient.ConsentPDFUrl,
		}

		c.JSON(http.StatusOK, response)
	}
}
