package types

import (
	"time"
)

type ReviewStore interface {
	CreateReview(Review) error

	IsReviewDuplicate(uid int, listingId int) (bool, error)
}

type RegisterReviewPayload struct {
	OrderID int    `json:"orderId" validate:"required"`
	Content string `json:"content"`
	Rating  int    `json:"rating" validate:"required"`
}

type Review struct {
	ID             int       `json:"id"`
	ListingID      int       `json:"listingId"`
	ReviewerID     int       `json:"reviewerId"`
	Content        string    `json:"content"`
	Rating         int       `json:"rating"`
	CreatedAt      time.Time `json:"createdAt"`
	LastModifiedAt time.Time `json:"lastModifiedAt"`
}
