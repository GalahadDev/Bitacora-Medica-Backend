package domains

import (
	"time"

	"github.com/google/uuid"
)

type TicketStatus string

const (
	TicketOpen   TicketStatus = "OPEN"
	TicketClosed TicketStatus = "CLOSED"
)

type SupportTicket struct {
	ID            uuid.UUID    `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID        uuid.UUID    `gorm:"type:uuid;not null"`
	Subject       string       `gorm:"type:varchar(255);not null"`
	Message       string       `gorm:"type:text;not null"`
	AdminResponse string       `gorm:"type:text"`
	Status        TicketStatus `gorm:"type:varchar(20);default:'OPEN';not null"`
	CreatedAt     time.Time    `gorm:"autoCreateTime"`
	UpdatedAt     time.Time    `gorm:"autoUpdateTime"`
	User          User         `gorm:"foreignKey:UserID"`
}

type CreateTicketInput struct {
	Subject string `json:"subject" binding:"required"`
	Message string `json:"message" binding:"required"`
}

type ReplyTicketInput struct {
	Response string `json:"response" binding:"required"`
}
