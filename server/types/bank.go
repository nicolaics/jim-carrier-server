package types

import "time"

type BankDetailStore interface {
	UpdateBankDetails(uid int, bankName, accountNumber, accountHolder string) error
	GetBankDetailByUserID(uid int) (*BankDetail, error)
}

type BankDetail struct {
	ID             int       `json:"id"`
	UserID         int       `json:"userId"`
	BankName       string    `json:"bankName"`
	AccountNumber  string    `json:"accountNumber"`
	AccountHolder  string    `json:"accountHolder"`
	LastModifiedAt time.Time `json:"lastModifiedAt"`
}
