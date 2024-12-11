package types

import "time"

type BankDetailStore interface {
	UpdateBankDetails(uid int, bankName, accountNumber, accountHolder string) error
	GetBankDetailByUserID(uid int) (*BankDetail, error)
	GetBankDataOfUser(uid int) (*BankDetailReturn, error)
}

type UpdateBankDetailPayload struct {
	BankName       string    `json:"bankName" validate:"required"`
	AccountNumber  string    `json:"accountNumber" validate:"required"`
	AccountHolder  string    `json:"accountHolder" validate:"required"`
}

type BankDetailReturn struct {
	BankName       string    `json:"bankName"`
	AccountNumber  string    `json:"accountNumber"`
	AccountHolder  string    `json:"accountHolder"`
}

type BankDetail struct {
	ID             int       `json:"id"`
	UserID         int       `json:"userId"`
	BankName       string    `json:"bankName"`
	AccountNumber  string    `json:"accountNumber"`
	AccountHolder  string    `json:"accountHolder"`
	LastModifiedAt time.Time `json:"lastModifiedAt"`
}
