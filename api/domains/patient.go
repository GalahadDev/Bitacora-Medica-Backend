package domains

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Patient struct {
	ID               uuid.UUID      `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	CreatorID        uuid.UUID      `gorm:"type:uuid;not null;index"`
	PersonalInfo     datatypes.JSON `gorm:"type:jsonb;not null;column:personal_info"`
	DisabilityReport string         `gorm:"type:text"`
	CareNotes        string         `gorm:"type:text"`
	ConsentPDFUrl    string         `gorm:"type:text;not null"`
	CreatedAt        time.Time      `gorm:"autoCreateTime"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime"`
	DeletedAt        gorm.DeletedAt `gorm:"index"`
}

type CreatePatientInput struct {
	FirstName        string         `form:"first_name" binding:"required"`
	LastName         string         `form:"last_name" binding:"required"`
	RUT              string         `form:"rut" binding:"required"`
	Phone            string         `form:"phone"`
	DisabilityReport string         `form:"disability_report"`
	CareNotes        string         `form:"care_notes"`
	PersonalInfo     datatypes.JSON `gorm:"type:jsonb;not null;column:personal_info"`
}
