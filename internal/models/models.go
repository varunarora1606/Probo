package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name     string 
	Email    string 
	Password string 
}

type Order struct {
	BetId    uuid.UUID 	`gorm:"type:uuid;primaryKey"` //BetId
	EventId  uuid.UUID	`gorm:"type:uuid;not null"`   //Lets see if I use it
    UserID   uuid.UUID	`gorm:"type:uuid;not null"`
    MarketID uuid.UUID	`gorm:"type:uuid;not null"`
    Side     string		`gorm:"type:text;check:side IN ('yes','no')"`
	TransactionType string `gorm:"type:text;check:transactionType In ('buy','sell')"`
    Price    int		`gorm:"type:numeric"`
    Quantity int		
    // Status   string		`gorm:"type:text;check:status IN ('open','matched', 'cancelled')"`
	CreatedAt time.Time
}