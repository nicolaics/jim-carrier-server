package types

import (
	"database/sql"
	"time"
)

type ListingStore interface {
	CreateListing(Listing) error
	GetAllListings() ([]ListingReturnFromDB, error)
	GetListingsByCarrierID(carrierId int) ([]ListingReturnFromDB, error)

	UpdateListingExpStatus() error
	IsListingDuplicate(carrierId int, destination string, weightAvailable float64, departureDate time.Time) (bool, error)

	GetListingByPayload(carrierName string, destination string, weightAvailable float64, pricePerKg float64, departureDate time.Time) (*ListingReturnFromDB, error)
	GetListingByID(id int) (*ListingReturnFromDB, error)

	DeleteListing(id int) error

	ModifyListing(int, Listing) error

	SubtractWeightAvailable(listindId int, minusValue float64) error
	AddWeightAvailable(listindId int, addValue float64) error
}

type PostListingPayload struct {
	Destination       string  `json:"destination" validate:"required"`
	WeightAvailable   float64 `json:"weightAvailable" validate:"required"`
	PricePerKg        float64 `json:"pricePerKg" validate:"required"`
	Currency          string  `json:"currency" validate:"required"`
	DepartureDate     string  `json:"departureDate" validate:"required"`
	LastReceivedDate  string  `json:"lastReceivedDate" validate:"required"`
	Description       string  `json:"description"`
	BankName          string  `json:"bankName" validate:"required"`
	BankAccountNumber string  `json:"bankAccountNumber" validate:"required"`
}

type GetListingDetailPayload struct {
	ID int `json:"id"`
}

type DeleteListingPayload GetListingDetailPayload

type ModifyListingPayload struct {
	ID      int                `json:"id"`
	NewData PostListingPayload `json:"newData" validate:"required"`
}

type ListingReturnPayload struct {
	ID                    int       `json:"id"`
	CarrierID             int       `json:"carrierId"`
	CarrierName           string    `json:"carrierName"`
	CarrierProfilePicture []byte    `json:"carrierProfilePicture"`
	Destination           string    `json:"destination"`
	WeightAvailable       float64   `json:"weightAvailable"`
	PricePerKg            float64   `json:"pricePerKg"`
	Currency              string    `json:"currency"`
	DepartureDate         time.Time `json:"departureDate"`
	LastReceivedDate      time.Time `json:"lastReceivedDate"`
	Description           string    `json:"description"`
	CarrierRating         float64   `json:"carrierRating"`
	LastModifiedAt        time.Time `json:"lastModifiedAt"`
}

type ListingReturnFromDB struct {
	ID               int            `json:"id"`
	CarrierID        int            `json:"carrierId"`
	CarrierName      string         `json:"carrierName"`
	Destination      string         `json:"destination"`
	WeightAvailable  float64        `json:"weightAvailable"`
	PricePerKg       float64        `json:"pricePerKg"`
	Currency         string         `json:"currency"`
	DepartureDate    time.Time      `json:"departureDate"`
	LastReceivedDate time.Time      `json:"lastReceivedDate"`
	Description      sql.NullString `json:"description"`
	CarrierRating    float64        `json:"carrierRating"`
	LastModifiedAt   time.Time      `json:"lastModifiedAt"`
}

type Listing struct {
	ID               int          `json:"id"`
	CarrierID        int          `json:"carrierId"`
	Destination      string       `json:"destination"`
	WeightAvailable  float64      `json:"weightAvailable"`
	PricePerKg       float64      `json:"pricePerKg"`
	CurrencyID       int          `json:"currencyId"`
	DepartureDate    time.Time    `json:"departureDate"`
	LastReceivedDate time.Time    `json:"lastReceivedDate"`
	ExpStatus        int          `json:"expStatus"`
	Description      string       `json:"description"`
	CreatedAt        time.Time    `json:"createdAt"`
	LastModifiedAt   time.Time    `json:"lastModifiedAt"`
	DeletedAt        sql.NullTime `json:"deletedAt"`
}
