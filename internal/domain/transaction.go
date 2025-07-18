package domain

import "time"

type Transaction struct {
	ID                string    `gorm:"primaryKey;type:char(36)" json:"id"`
	CustomerID        string    `gorm:"type:char(36)" json:"customer_id"`
	ContractNumber    string    `gorm:"unique;type:varchar(100)" json:"contract_number"`
	OTRAmount         float64   `gorm:"type:decimal(15,2)" json:"otr_amount"`
	AdminFee          float64   `gorm:"type:decimal(15,2)" json:"admin_fee"`
	InstallmentAmount float64   `gorm:"type:decimal(15,2)" json:"installment_amount"`
	InterestAmount    float64   `gorm:"type:decimal(15,2)" json:"interest_amount"`
	AssetName         string    `gorm:"type:varchar(100)" json:"asset_name"`
	CreatedAt         time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
