package publickey

import (
	"database/sql"
	"time"

	"github.com/nicolaics/jim-carrier-server/types"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) UpdatePublicKey(id int, e []byte, m []byte) error {
	publicKey, err := s.GetPublicKeyByUserID(id)
	if err != nil {
		return err
	}

	if publicKey == nil {
		query := `INSERT INTO public_key (user_id, e, m) VALUES (?, ?, ?)`
		_, err = s.db.Exec(query, id, string(e), string(m))
		if err != nil {
			return err
		}
	} else {
		if (publicKey.E == string(e)) && (publicKey.M == string(m)) {
			return nil
		}

		query := `UPDATE public_key SET e = ?, m = ?, last_modified_at = ? WHERE user_id = ?`
		_, err = s.db.Exec(query, string(e), string(m), time.Now(), id)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) GetPublicKeyByUserID(userId int) (*types.PublicKey, error) {
	query := `SELECT * FROM public_key WHERE user_id = ?`
	row := s.db.QueryRow(query, userId)
	if row.Err() != nil {
		if row.Err() == sql.ErrNoRows {
			return nil, nil
		}

		return nil, row.Err()
	}

	publicKey := new(types.PublicKey)

	err := row.Scan(&publicKey.ID, &publicKey.UserID, &publicKey.E, &publicKey.M, &publicKey.LastModifiedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	publicKey.LastModifiedAt = publicKey.LastModifiedAt.Local()

	return publicKey, nil
}
