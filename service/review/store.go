package review

import (
	"database/sql"
	// "fmt"
	// "time"

	"github.com/nicolaics/jim-carrier/types"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) CreateReview(review types.Review) error {
	query := `INSERT INTO review 
				(listing_id, reviewer_id, content, rating) 
				VALUES (?, ?, ?, ?)`
	_, err := s.db.Exec(query, review.ListingID, review.ReviewerID, 
						review.Content, review.Rating)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) IsReviewDuplicate(uid int, listingId int) (bool, error) {
	query := `SELECT COUNT(*) FROM review WHERE reviewer_id = ? AND listing_id = ?`
	row := s.db.QueryRow(query, uid, listingId)
	if row.Err() != nil {
		return true, row.Err()
	}

	var count int
	err := row.Scan(&count)
	if err != nil {
		return true, err
	}

	return count > 0, nil
}