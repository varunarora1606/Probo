package models

import (
	"time"

	"github.com/varunarora1606/Probo/internal/types"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name     string
	Email    string
	Password string
}

type Order struct {
	BetId    string 	`gorm:"type:text;primaryKey"` //BetId
	EventId  string		`gorm:"type:text;not null;unique"`   //Lets see if I use it
    UserID   string		`gorm:"type:text;not null"`
    MarketID string		`gorm:"type:text;not null"`
    Side     types.Side	`gorm:"type:text;check:side IN ('yes','no')"`
	TransactionType types.TransactionType `gorm:"column:transaction_type;type:text;check:transaction_type In ('buy','sell')"`
    Price    int		`gorm:"type:numeric"`
    Quantity int		
    // Status   string		`gorm:"type:text;check:status IN ('open','matched', 'cancelled')"`
	CreatedAt time.Time
}

// type Order struct {
// 	types.Order
// }