package model

type SetCreditLimitRequest struct {
	CustomerID  string  `json:"customer_id" validate:"required,uuid"`
	TenorMonths int     `json:"tenor_months" validate:"required,oneof=1 2 3 6"` // Valid tenors as per case
	LimitAmount float64 `json:"limit_amount" validate:"required,gt=0"`
}

type CreditLimitResponse struct {
	ID          string  `json:"id"`
	CustomerID  string  `json:"customer_id"`
	TenorMonths int     `json:"tenor_months"`
	LimitAmount float64 `json:"limit_amount"`
}
