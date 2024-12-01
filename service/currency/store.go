package currency

import (
	"database/sql"
	"strings"

	"github.com/nicolaics/jim-carrier-server/types"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) CreateCurrency(name string) error {
	query := `INSERT INTO currency (name) VALUES (?)`
	_, err := s.db.Exec(query, strings.ToUpper(name))
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetCurrencyByName(name string) (*types.Currency, error) {
	query := `SELECT * FROM currency WHERE name = ?`
	rows, err := s.db.Query(query, strings.ToUpper(name))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	currency := new(types.Currency)

	for rows.Next() {
		currency, err = scanRowIntoCurrency(rows)
		if err != nil {
			return nil, err
		}
	}

	if currency.ID == 0 {
		return nil, nil
	}
	

	return currency, nil
}

func scanRowIntoCurrency(rows *sql.Rows) (*types.Currency, error) {
	currency := new(types.Currency)

	err := rows.Scan(
		&currency.ID,
		&currency.Name,
		&currency.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	currency.CreatedAt = currency.CreatedAt.Local()

	return currency, nil
}