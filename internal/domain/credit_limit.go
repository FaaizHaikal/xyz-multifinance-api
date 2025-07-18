package domain

import "time"

type CreditLimit struct {
	ID          string    `gorm:"primaryKey;type:char(36)" json:"id"`                              // UUID CHAR(36)
	CustomerID  string    `gorm:"type:char(36);uniqueIndex:idx_customer_tenor" json:"customer_id"` // Foreign key to Customer.ID
	TenorMonths int       `gorm:"type:int;uniqueIndex:idx_customer_tenor" json:"tenor_months"`     // Tenor in months (e.g., 1, 2, 3, 6)
	LimitAmount float64   `gorm:"type:decimal(15,2)" json:"limit_amount"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type CreditLimitRepository interface {
	CreateCreditLimit(creditLimit *CreditLimit) error
	GetCreditLimitByCustomerAndTenor(customerID string, tenorMonths int) (*CreditLimit, error)
	UpdateCreditLimit(creditLimit *CreditLimit) error
	GetCreditLimitsByCustomerID(customerID string) ([]CreditLimit, error)
}
