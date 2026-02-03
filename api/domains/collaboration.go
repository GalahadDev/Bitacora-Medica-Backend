package domains

import (
	"time"

	"github.com/google/uuid"
)

type CollabStatus string

const (
	CollabPending  CollabStatus = "PENDING"
	CollabAccepted CollabStatus = "ACCEPTED"
	CollabRejected CollabStatus = "REJECTED"
	CollabRevoked  CollabStatus = "REVOKED"
)

type Collaboration struct {
	ID             uuid.UUID    `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	PatientID      uuid.UUID    `gorm:"type:uuid;not null"`
	ProfessionalID uuid.UUID    `gorm:"type:uuid;not null"`
	Status         CollabStatus `gorm:"type:varchar(20);default:'PENDING';not null"`
	InvitedAt      time.Time    `gorm:"autoCreateTime"`
	UpdatedAt      time.Time    `gorm:"autoUpdateTime"`
	Professional   User         `gorm:"foreignKey:ProfessionalID"`
	Patient        Patient      `gorm:"foreignKey:PatientID"`
}

type InviteInput struct {
	PatientID string `json:"patient_id" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
}
