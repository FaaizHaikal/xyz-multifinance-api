package model

import "time"

type CustomerResponse struct {
	ID             string    `json:"id"`
	NIK            string    `json:"nik"`
	FullName       string    `json:"full_name"`
	LegalName      string    `json:"legal_name"`
	BirthPlace     string    `json:"birth_place"`
	BirthDate      time.Time `json:"birth_date"`
	Salary         float64   `json:"salary"`
	KTPPhotoURL    string    `json:"ktp_photo_url"`
	SelfiePhotoURL string    `json:"selfie_photo_url"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type RegisterCustomerRequest struct {
	NIK            string  `json:"nik" validate:"required,len=16"`
	FullName       string  `json:"full_name" validate:"required,max=100"`
	LegalName      string  `json:"legal_name" validate:"required,max=100"`
	BirthPlace     string  `json:"birth_place" validate:"required,max=100"`
	BirthDate      string  `json:"birth_date" validate:"required,datetime=2006-01-02"` // Date format YYYY-MM-DD
	Salary         float64 `json:"salary" validate:"required,gt=0"`
	KTPPhotoURL    string  `json:"ktp_photo_url" validate:"omitempty,url"`
	SelfiePhotoURL string  `json:"selfie_photo_url" validate:"omitempty,url"`
}
