package types

import (
	"database/sql"
	"time"
)

type OrderStore interface {
	CreateOrder(Order) error
	GetOrderByID(int) (*Order, error)
	// GetOrderByPaymentProofURL(paymentProofUrl string) (*Order, error)

	GetOrderByCarrierID(int) ([]OrderCarrierReturnFromDB, error)
	GetOrderByGiverID(int) ([]OrderGiverReturnFromDB, error)

	GetCarrierOrderByID(orderId int, userId int) (*OrderCarrierReturnFromDB, error)
	GetGiverOrderByID(orderId int, userId int) (*OrderGiverReturnFromDB, error)

	DeleteOrder(orderId int, userId int) error

	ModifyOrder(int, Order) error
	UpdatePackageLocation(id int, orderStatus int, packageLocation string) error
	UpdatePaymentStatus(id int, paymentStatus int, paymentProofUrl string) error
	UpdateOrderStatus(id int, orderStatus int, packageLocation string) error

	IsOrderDuplicate(userId int, listingId int) (bool, error)
	IsPaymentProofURLExist(string) bool
	IsPackageImageURLExist(packageImgUrl string) bool
}

type RegisterOrderPayload struct {
	ListingID      int     `json:"listingId" validate:"required"`
	Weight         float64 `json:"weight" validate:"required"`
	Price          float64 `json:"price" validate:"required"`
	Currency       string  `json:"currency" validate:"required"`
	PackageContent string  `json:"packageContent" validate:"required"`
	PackageImage   []byte  `json:"packageImage" validate:"required"`
	Notes          string  `json:"notes"`
}

type ViewOrderDetailPayload struct {
	ID int `json:"id" validate:"required"`
}

type DeleteOrderPayload ViewOrderDetailPayload

type ModifyOrderPayload struct {
	ID              int     `json:"id" validate:"required"`
	ListingID       int     `json:"listingId" validate:"required"`
	Weight          float64 `json:"weight" validate:"required"`
	Price           float64 `json:"price" validate:"required"`
	Currency        string  `json:"currency" validate:"required"`
	PackageContent  string  `json:"packageContent" validate:"required"`
	PackageImage    []byte  `json:"packageImage"`
	PaymentStatus   string  `json:"paymentStatus" validate:"required"`
	PackageLocation string  `json:"packageLocation" validate:"required"`
	Notes           string  `json:"notes"`
}

type UpdatePackageLocationPayload struct {
	ID              int    `json:"id" validate:"required"`
	PackageLocation string `json:"packageLocation" validate:"required"`
	OrderStatus     string `json:"orderStatus"`
}

type UpdatePaymentStatusPayload struct {
	ID            int    `json:"id" validate:"required"`
	PaymentStatus string `json:"paymentStatus" validate:"required"`
	PaymentProof  []byte `json:"paymentProof"`
}

type UpdateOrderStatusPayload struct {
	ID              int    `json:"id" validate:"required"`
	OrderStatus     string `json:"orderStatus" validate:"required"`
	PackageLocation string `json:"packageLocation"`
}

type GetPaymentProofImagePayload struct {
	PaymentProofURL string `json:"paymentProofUrl" validate:"required"`
}

type ReturnPaymentProofImagePayload struct {
	PaymentProof []byte `json:"paymentProof" validate:"required"`
}

type OrderGiverReturnFromDB struct {
	ID              int            `json:"id"`
	Weight          float64        `json:"weight"`
	Price           float64        `json:"price"`
	Currency        string         `json:"currency"`
	PackageContent  string         `json:"packageContent"`
	PackageImageURL string         `json:"packageImageUrl"`
	PaymentStatus   int            `json:"paymentStatus"`
	PaidAt          sql.NullTime   `json:"paidAt"`
	PaymentProofURL sql.NullString `json:"paymentProofUrl"`
	OrderStatus     int            `json:"orderStatus"`
	PackageLocation string         `json:"packageLocation"`
	Notes           sql.NullString `json:"notes"`
	CreatedAt       time.Time      `json:"createdAt"`
	LastModifiedAt  time.Time      `json:"lastModifiedAt"`

	Listing struct {
		ID            int       `json:"id"`
		CarrierName   string    `json:"carrierName"`
		Destination   string    `json:"destination"`
		DepartureDate time.Time `json:"departureDate"`
	} `json:"listing"`
}

type OrderGiverReturnPayload struct {
	ID              int       `json:"id"`
	Weight          float64   `json:"weight"`
	Price           float64   `json:"price"`
	Currency        string    `json:"currency"`
	PackageContent  string    `json:"packageContent"`
	PackageImageURL string    `json:"packageImageUrl"`
	PaymentStatus   string    `json:"paymentStatus"`
	PaidAt          time.Time `json:"paidAt"`
	PaymentProofURL string    `json:"paymentProofUrl"`
	OrderStatus     string    `json:"orderStatus"`
	PackageLocation string    `json:"packageLocation"`
	Notes           string    `json:"notes"`
	CreatedAt       time.Time `json:"createdAt"`
	LastModifiedAt  time.Time `json:"lastModifiedAt"`

	Listing struct {
		ID            int       `json:"id"`
		CarrierName   string    `json:"carrierName"`
		Destination   string    `json:"destination"`
		DepartureDate time.Time `json:"departureDate"`
	} `json:"listing"`
}

type OrderCarrierReturnFromDB struct {
	Listing struct {
		ID            int       `json:"id"`
		Destination   string    `json:"destination"`
		DepartureDate time.Time `json:"departureDate"`
	} `json:"listing"`

	ID               int            `json:"id"`
	GiverName        string         `json:"giverName"`
	GiverPhoneNumber string         `json:"giverPhoneNumber"`
	Weight           float64        `json:"weight"`
	Price            float64        `json:"price"`
	Currency         string         `json:"currency"`
	PackageContent   string         `json:"packageContent"`
	PackageImageURL  string         `json:"packageImageUrl"`
	PaymentStatus    int            `json:"paymentStatus"`
	PaidAt           sql.NullTime   `json:"paidAt"`
	PaymentProofURL  sql.NullString `json:"paymentProofUrl"`
	OrderStatus      int            `json:"orderStatus"`
	PackageLocation  string         `json:"packageLocation"`
	Notes            sql.NullString `json:"notes"`
	CreatedAt        time.Time      `json:"createdAt"`
	LastModifiedAt   time.Time      `json:"lastModifiedAt"`
}

type OrderCarrierReturnPayload struct {
	Listing struct {
		ID            int       `json:"id"`
		Destination   string    `json:"destination"`
		DepartureDate time.Time `json:"departureDate"`
	} `json:"listing"`

	ID               int       `json:"id"`
	GiverName        string    `json:"giverName"`
	GiverPhoneNumber string    `json:"giverPhoneNumber"`
	Weight           float64   `json:"weight"`
	Price            float64   `json:"price"`
	Currency         string    `json:"currency"`
	PackageContent   string    `json:"packageContent"`
	PackageImageURL  string    `json:"packageImageUrl"`
	PaymentStatus    string    `json:"paymentStatus"`
	PaidAt           time.Time `json:"paidAt"`
	PaymentProofURL  string    `json:"paymentProofUrl"`
	OrderStatus      string    `json:"orderStatus"`
	PackageLocation  string    `json:"packageLocation"`
	Notes            string    `json:"notes"`
	CreatedAt        time.Time `json:"createdAt"`
	LastModifiedAt   time.Time `json:"lastModifiedAt"`
}

type Order struct {
	ID              int            `json:"id"`
	ListingID       int            `json:"listingId"`
	GiverID         int            `json:"giverId"`
	Weight          float64        `json:"weight"`
	Price           float64        `json:"price"`
	CurrencyID      int            `json:"currencyId"`
	PackageContent  string         `json:"packageContent"`
	PackageImageURL string         `json:"packageImageUrl"`
	PaymentStatus   int            `json:"paymentStatus"`
	PaidAt          sql.NullTime   `json:"paidAt"`
	PaymentProofURL sql.NullString `json:"paymentProofUrl"`
	OrderStatus     int            `json:"orderStatus"`
	PackageLocation string         `json:"packageLocation"`
	Notes           string         `json:"notes"`
	CreatedAt       time.Time      `json:"createdAt"`
	LastModifiedAt  time.Time      `json:"lastModifiedAt"`
	DeletedAt       sql.NullTime   `json:"deletedAt"`
}
