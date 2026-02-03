package domains

import (
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null"`
	Type      string     `gorm:"type:varchar(50);not null"`
	Message   string     `gorm:"type:text;not null"`
	IsRead    bool       `gorm:"default:false"`
	RelatedID *uuid.UUID `gorm:"type:uuid"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
}
