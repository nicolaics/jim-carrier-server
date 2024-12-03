package bank

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

func (s *Store) UpdateBankDetails(uid int, bankName, accountNumber, accountHolder string) error {
	query := `SELECT COUNT(*) FROM bank_detail WHERE user_id = ?`
	row := s.db.QueryRow(query, uid)
	if row.Err() != nil {
		return row.Err()
	}

	var count int
	err := row.Scan(&count)
	if err != nil {
		return err
	}

	if count == 1 {
		query = `UPDATE bank_detail SET bank_name = ?, account_number = ?, account_holder = ?, last_modified_at = ?  
					WHERE user_id = ?`
		_, err = s.db.Exec(query, bankName, accountNumber, accountHolder, time.Now(), uid)
		if err != nil {
			return err
		}

		return nil
	} else if count > 1 {
		query = `DELETE FROM bank_detail WHERE user_id = ?`
		_, err = s.db.Exec(query, uid)
		if err != nil {
			return err
		}
	}

	query = `INSERT INTO bank_detail (user_id, bank_name, account_number, account_holder) 
				VALUES (?, ?, ?, ?)`
	_, err = s.db.Exec(query, uid, bankName, accountNumber, accountHolder)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetBankDetailByUserID(uid int) (*types.BankDetail, error) {
	query := `SELECT * FROM bank_detail WHERE user_id = ?`
	row := s.db.QueryRow(query, uid)
	if row.Err() != nil {
		if row.Err() == sql.ErrNoRows {
			return nil, nil
		}

		return nil, row.Err()
	}

	bankDetail := new(types.BankDetail)

	err := row.Scan(
		&bankDetail.ID,
		&bankDetail.UserID,
		&bankDetail.BankName,
		&bankDetail.AccountNumber,
		&bankDetail.AccountHolder,
		&bankDetail.LastModifiedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return bankDetail, nil
}

func (s *Store) GetBankDataOfUser(uid int) (*types.BankDetailReturn, error) {
	query := `SELECT bank_name, account_number, account_holder FROM bank_detail WHERE user_id = ?`
	row := s.db.QueryRow(query, uid)
	if row.Err() != nil {
		if row.Err() == sql.ErrNoRows {
			return nil, nil
		}

		return nil, row.Err()
	}

	bankDetail := new(types.BankDetailReturn)

	err := row.Scan(
		&bankDetail.BankName,
		&bankDetail.AccountNumber,
		&bankDetail.AccountHolder,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return bankDetail, nil
}
