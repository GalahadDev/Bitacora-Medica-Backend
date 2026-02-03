package domains

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type UserRole string

const (
	RoleAdmin        UserRole = "ADMIN"
	RoleProfessional UserRole = "PROFESSIONAL"
	RoleBusiness     UserRole = "BUSINESS"
)

type UserStatus string

const (
	StatusInactive UserStatus = "INACTIVE"
	StatusActive   UserStatus = "ACTIVE"
	StatusRejected UserStatus = "REJECTED"
)

type User struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key"`
	Email        string         `gorm:"uniqueIndex;not null"`
	Role         UserRole       `gorm:"type:user_role;default:'PROFESSIONAL'"`
	Status       UserStatus     `gorm:"type:user_status;default:'INACTIVE'"`
	AvatarURL    string         `gorm:"type:text"`
	ProfileData  datatypes.JSON `gorm:"type:jsonb"`
	RejectReason string         `gorm:"type:text"`
	CreatedAt    time.Time      `gorm:"autoCreateTime"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return
}
