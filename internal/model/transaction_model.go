package model

type CreateTransactionRequest struct {
	CustomerID        string  `json:"customer_id" validate:"required,uuid"`
	ContractNumber    string  `json:"contract_number" validate:"required,max=100"`
	TenorMonths       int     `json:"tenor_months" validate:"required,oneof=1 2 3 6"`
	OTRAmount         float64 `json:"otr_amount" validate:"required,gt=0"`
	AdminFee          float64 `json:"admin_fee" validate:"required,gte=0"`
	InstallmentAmount float64 `json:"installment_amount" validate:"required,gt=0"`
	InterestAmount    float64 `json:"interest_amount" validate:"required,gte=0"`
	AssetName         string  `json:"asset_name" validate:"required,max=255"`
}

type TransactionResponse struct {
	ID                string  `json:"id"`
	CustomerID        string  `json:"customer_id"`
	ContractNumber    string  `json:"contract_number"`
	OTRAmount         float64 `json:"otr_amount"`
	AdminFee          float64 `json:"admin_fee"`
	InstallmentAmount float64 `json:"installment_amount"`
	InterestAmount    float64 `json:"interest_amount"`
	AssetName         string  `json:"asset_name"`
}
