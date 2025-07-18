package domain

import "time"

type Customer struct {
	ID             string    `gorm:"primaryKey;type:char(36)" json:"id"`
	NIK            string    `gorm:"unique;type:varchar(16)" json:"nik"`
	FullName       string    `gorm:"type:varchar(100)" json:"full_name"`
	Password       string    `gorm:"type:varchar(255)" json:"-"`
	LegalName      string    `gorm:"type:varchar(100)" json:"legal_name"`
	BirthPlace     string    `gorm:"type:varchar(100)" json:"birth_place"`
	BirthDate      time.Time `gorm:"type:date" json:"birth_date"`
	Salary         float64   `gorm:"type:decimal(15,2)" json:"salary"`
	KTPPhotoURL    string    `gorm:"type:text" json:"ktp_photo_url"`
	SelfiePhotoURL string    `gorm:"type:text" json:"selfie_photo_url"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
