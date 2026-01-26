package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Payslip struct {
	ID          string         `gorm:"type:uuid;primaryKey" json:"id"`
	EmployeeID  string         `gorm:"type:uuid;not null;index" json:"employeeId"`
	Year        int            `gorm:"not null" json:"year"`
	Month       int            `gorm:"not null" json:"month"`
	BasicSalary float64        `gorm:"type:decimal(15,2);not null;default:0" json:"basicSalary"`
	Allowances  float64        `gorm:"type:decimal(15,2);not null;default:0" json:"allowances"`
	Deductions  float64        `gorm:"type:decimal(15,2);not null;default:0" json:"deductions"`
	GrossSalary float64        `gorm:"type:decimal(15,2);not null;default:0" json:"grossSalary"`
	NetSalary   float64        `gorm:"type:decimal(15,2);not null;default:0" json:"netSalary"`
	Status      string         `gorm:"type:varchar(20);not null;default:'draft'" json:"status"`
	PaidAt      *time.Time     `json:"paidAt,omitempty"`
	Notes       string         `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Employee *User `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
}

func (p *Payslip) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}

func (p *Payslip) TableName() string {
	return "payslips"
}
