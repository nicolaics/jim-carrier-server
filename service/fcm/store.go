package fcm

import (
	"database/sql"

	"github.com/nicolaics/jim-carrier/types"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) CreateFCMHistory(fcmHistory types.FCMHistory) error {
	query := `INSERT INTO fcm_history (to_user_id, title, body, image_url, link, response) 
				VALUES (?, ?, ?, ? ,? , ?)`
	_, err := s.db.Exec(query, fcmHistory.ToUserID, fcmHistory.Title, fcmHistory.Body,
		fcmHistory.ImageURL, fcmHistory.Link, fcmHistory.Response)
	if err != nil {
		return err
	}

	return nil
}
