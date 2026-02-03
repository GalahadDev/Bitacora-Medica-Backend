package domains

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DocumentCategory string

const (
	DocCategoryLab          DocumentCategory = "LAB"
	DocCategoryImaging      DocumentCategory = "IMAGING"
	DocCategoryPrescription DocumentCategory = "PRESCRIPTION"
	DocCategoryOther        DocumentCategory = "OTHER"
)

type PatientDocument struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	PatientID uuid.UUID `gorm:"type:uuid;not null;index"`
	Patient   Patient   `gorm:"foreignKey:PatientID"`

	Name        string           `gorm:"type:varchar(255);not null"`
	Category    DocumentCategory `gorm:"type:varchar(50);not null"`
	Date        time.Time        `gorm:"not null"`
	FileUrl     string           `gorm:"type:text;not null"`
	FileType    string           `gorm:"type:varchar(50)"` // pdf, jpg, png,
	Description string           `gorm:"type:text"`

	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (d *PatientDocument) BeforeCreate(tx *gorm.DB) (err error) {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return
}
