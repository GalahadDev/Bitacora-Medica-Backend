package domains

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Session struct {
	ID                 uuid.UUID      `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	PatientID          uuid.UUID      `gorm:"type:uuid;not null;index"`
	ProfessionalID     uuid.UUID      `gorm:"type:uuid;not null;index"`
	Creator            User           `gorm:"foreignKey:ProfessionalID" json:"Creator"`
	InterventionPlan   string         `gorm:"type:text;not null"`
	Vitals             datatypes.JSON `gorm:"type:jsonb"`
	Description        string         `gorm:"type:text;not null"`
	Achievements       string         `gorm:"type:text"`
	PatientPerformance string         `gorm:"type:text"`
	Photos             pq.StringArray `gorm:"type:text[]"`
	HasIncident        bool           `gorm:"not null;default:false"`
	IncidentDetails    string         `gorm:"type:text"`
	IncidentPhoto      string         `gorm:"type:text"`
	NextSessionNotes   string         `gorm:"type:text"`

	CreatedAt time.Time `gorm:"index"`
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type CreateSessionInput struct {
	PatientID          string                 `json:"patient_id" binding:"required"`
	InterventionPlan   string                 `json:"intervention_plan" binding:"required"`
	Vitals             map[string]interface{} `json:"vitals"`
	Description        string                 `json:"description" binding:"required"`
	Achievements       string                 `json:"achievements"`
	PatientPerformance string                 `json:"patient_performance"`
	Photos             []string               `json:"photos"`
	HasIncident        bool                   `json:"has_incident"`
	IncidentDetails    string                 `json:"incident_details"`
	IncidentPhoto      string                 `json:"incident_photo"`
	NextSessionNotes   string                 `json:"next_session_notes"`
}
