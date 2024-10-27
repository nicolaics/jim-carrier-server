package types

import "time"

type CurrencyStore interface {
	CreateCurrency(name string) error
	GetCurrencyByName(name string) (*Currency, error)
}

type Currency struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}