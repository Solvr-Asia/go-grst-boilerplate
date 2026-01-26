package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
	UserStatusPending  UserStatus = "pending"
)

type User struct {
	ID          string         `gorm:"type:uuid;primaryKey" json:"id"`
	Email       string         `gorm:"uniqueIndex;not null" json:"email"`
	Password    string         `gorm:"not null" json:"-"`
	Name        string         `gorm:"not null" json:"name"`
	Phone       string         `json:"phone"`
	Status      UserStatus     `gorm:"type:varchar(20);default:active" json:"status"`
	Roles       []string       `gorm:"type:text[];default:ARRAY['user']::TEXT[]" json:"roles"`
	CompanyCode string         `gorm:"type:varchar(50)" json:"companyCode"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return nil
}

func (u *User) TableName() string {
	return "users"
}
