package user

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/nicolaics/jim-carrier/constants"
	"github.com/nicolaics/jim-carrier/service/auth"
	"github.com/nicolaics/jim-carrier/types"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) GetUserByEmail(email string) (*types.User, error) {
	query := `SELECT id, name, email, phone_number, provider, 
					last_logged_in, created_at 
				FROM user WHERE email = ?`
	rows, err := s.db.Query(query, email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	user := new(types.User)

	for rows.Next() {
		user, err = scanRowIntoUser(rows)

		if err != nil {
			return nil, err
		}
	}

	if user.ID == 0 {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

func (s *Store) GetUserByName(name string) (*types.User, error) {
	query := `SELECT id, name, email, phone_number, provider, 
					last_logged_in, created_at 
				FROM user WHERE name = ?`
	rows, err := s.db.Query(query, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	user := new(types.User)

	for rows.Next() {
		user, err = scanRowIntoUser(rows)

		if err != nil {
			return nil, err
		}
	}

	if user.ID == 0 {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

func (s *Store) GetUserPasswordByEmail(email string) (string, error) {
	query := `SELECT password FROM user WHERE email = ?`
	row := s.db.QueryRow(query, email)
	if row.Err() != nil {
		return "", row.Err()
	}

	var password string
	err := row.Scan(&password)
	if err != nil {
		return "", err
	}

	return password, nil
}

func (s *Store) GetUserBySearchName(name string) ([]types.User, error) {
	query := "SELECT COUNT(*) FROM user WHERE name = ?"
	row := s.db.QueryRow(query, name)
	if row.Err() != nil {
		return nil, row.Err()
	}

	var count int

	err := row.Scan(&count)
	if err != nil {
		return nil, err
	}

	users := make([]types.User, 0)

	if count == 0 {
		query = `SELECT id, name, email, phone_number, provider, 
						last_logged_in, created_at 
					FROM user WHERE name LIKE ?`
		searchVal := "%"

		log.Println("search val user: ", searchVal)

		for _, val := range name {
			if string(val) != " " {
				searchVal += (string(val) + "%")
			}
		}

		rows, err := s.db.Query(query, searchVal)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			user, err := scanRowIntoUser(rows)

			if err != nil {
				return nil, err
			}

			users = append(users, *user)
		}

		return users, nil
	}
	query = `SELECT id, name, email, phone_number, provider, 
						last_logged_in, created_at 
					FROM user WHERE name = ?`
	rows, err := s.db.Query(query, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		user, err := scanRowIntoUser(rows)

		if err != nil {
			return nil, err
		}

		users = append(users, *user)
	}

	return users, nil
}

func (s *Store) GetUserBySearchPhoneNumber(phoneNumber string) ([]types.User, error) {
	query := "SELECT COUNT(*) FROM user WHERE phone_number = ?"
	row := s.db.QueryRow(query, phoneNumber)
	if row.Err() != nil {
		return nil, row.Err()
	}

	var count int

	err := row.Scan(&count)
	if err != nil {
		return nil, err
	}

	users := make([]types.User, 0)

	if count == 0 {
		query = `SELECT id, name, email, phone_number, provider, 
						last_logged_in, created_at 
					FROM user WHERE phone_number LIKE ?`
		searchVal := "%"

		log.Println("search val user: ", searchVal)

		for _, val := range phoneNumber {
			if string(val) != " " {
				searchVal += (string(val) + "%")
			}
		}

		rows, err := s.db.Query(query, searchVal)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			user, err := scanRowIntoUser(rows)

			if err != nil {
				return nil, err
			}

			users = append(users, *user)
		}

		return users, nil
	}

	query = `SELECT id, name, email, phone_number, provider, 
						last_logged_in, created_at 
					FROM user WHERE phone_number = ?`
	rows, err := s.db.Query(query, phoneNumber)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		user, err := scanRowIntoUser(rows)

		if err != nil {
			return nil, err
		}

		users = append(users, *user)
	}

	return users, nil
}

func (s *Store) GetUserByID(id int) (*types.User, error) {
	query := `SELECT id, name, email, phone_number, provider, 
					last_logged_in, created_at 
				FROM user WHERE id = ?`
	rows, err := s.db.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	user := new(types.User)

	for rows.Next() {
		user, err = scanRowIntoUser(rows)

		if err != nil {
			return nil, err
		}
	}

	if user.ID == 0 {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

func (s *Store) CreateUser(user types.User) error {
	query := `INSERT INTO user (name, email, password, phone_number, provider) VALUES (?, ?, ?, ?, ?)`
	_, err := s.db.Exec(query, user.Name, user.Email, user.Password, user.PhoneNumber, user.Provider)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) DeleteUser(user *types.User) error {
	_, err := s.db.Exec("DELETE FROM review WHERE reviewer_id = ?", user.ID)
	if err != nil {
		return err
	}

	_, err = s.db.Exec("DELETE FROM order WHERE giver_id = ?", user.ID)
	if err != nil {
		return err
	}

	_, err = s.db.Exec("DELETE FROM listing WHERE carrier_id = ?", user.ID)
	if err != nil {
		return err
	}

	_, err = s.db.Exec("DELETE FROM user WHERE id = ?", user.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) UpdateLastLoggedIn(id int) error {
	_, err := s.db.Exec("UPDATE user SET last_logged_in = ? WHERE id = ? ",
		time.Now(), id)

	if err != nil {
		return err
	}

	return nil
}

func (s *Store) ModifyUser(id int, user types.User) error {
	query := `UPDATE user SET name = ?, phone_number = ? WHERE id = ?`
	_, err := s.db.Exec(query, user.Name, user.PhoneNumber, id)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) UpdatePassword(id int, password string) error {
	query := `UPDATE user SET password = ? WHERE id = ?`
	_, err := s.db.Exec(query, password, id)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) SaveToken(userId int, tokenDetails *types.TokenDetails) error {
	tokenExp := time.Unix(tokenDetails.TokenExp, 0) //converting Unix to UTC(to Time object)

	query := "INSERT INTO verify_token(user_id, uuid, expired_at) VALUES (?, ?, ?)"
	_, err := s.db.Exec(query, userId, tokenDetails.UUID, tokenExp)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) DeleteToken(userId int) error {
	query := "DELETE FROM verify_token WHERE user_id = ?"
	_, err := s.db.Exec(query, userId)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) ValidateUserToken(w http.ResponseWriter, r *http.Request) (*types.User, error) {
	query := "DELETE FROM verify_token WHERE expired_at < ?"
	_, err := s.db.Exec(query, time.Now().UTC().Format("2006-01-02 15:04:05"))
	if err != nil {
		return nil, fmt.Errorf("error deleting expired token: %v", err)
	}

	accessDetails, err := auth.ExtractTokenFromClient(r)
	if err != nil {
		return nil, err
	}

	query = "SELECT COUNT(*) FROM verify_token WHERE user_id = ? AND expired_at >= ?"
	row := s.db.QueryRow(query, accessDetails.UserID, time.Now().UTC().Format("2006-01-02 15:04:05"))
	if row.Err() != nil {
		return nil, row.Err()
	}

	var count int
	err = row.Scan(&count)
	if err != nil {
		return nil, err
	}

	if count > 1 {
		return nil, fmt.Errorf("logged in from other device")
	}

	query = "SELECT user_id FROM verify_token WHERE uuid = ? AND user_id = ? AND expired_at >= ?"
	rows, err := s.db.Query(query, accessDetails.UUID, accessDetails.UserID, time.Now().UTC().Format("2006-01-02 15:04:05"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userId int

	for rows.Next() {
		err = rows.Scan(&userId)
		if err != nil {
			return nil, err
		}
	}

	// check if user exist
	user, err := s.GetUserByID(userId)
	if err != nil {
		delErr := s.DeleteToken(accessDetails.UserID)
		if delErr != nil {
			return nil, fmt.Errorf("delete error: %v", delErr)
		}

		return nil, fmt.Errorf("token expired, log in again")
	}

	return user, nil
}

func (s *Store) DelayCodeWithinTime(email string, minutes int) (bool, error) {
	query := `SELECT COUNT(*) FROM verify_code 
			  WHERE email = ? AND status = ?
			  AND TIMESTAMPDIFF(MINUTE, created_at, UTC_TIMESTAMP) <= ?`

	var count int
	err := s.db.QueryRow(query, email, constants.VERIFY_CODE_WAITING, minutes).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to validate login code: %v", err)
	}
	
	return count > 0, nil
}

func (s *Store) ValidateLoginCodeWithinTime(email, code string, minutes int) (bool, error) {
	query := `DELETE FROM verify_code 
			  WHERE email = ? AND status = ? 
			  AND TIMESTAMPDIFF(MINUTE, created_at, UTC_TIMESTAMP) > ?`
	_, err := s.db.Exec(query, email, constants.VERIFY_CODE_COMPLETE, minutes)
	if err != nil {
		return false, err
	}
	
	query = `SELECT COUNT(*) FROM verify_code 
			  WHERE email = ? AND code = ? AND status = ? 
			  AND TIMESTAMPDIFF(MINUTE, created_at, UTC_TIMESTAMP) <= ?`
	
	var count int
	
	err = s.db.QueryRow(query, email, code, constants.VERIFY_CODE_WAITING, minutes).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to validate login code: %v", err)
	}

	return count > 0, nil
}

func (s *Store) SaveVerificationCode(email, code string, requestType int) error {
	query := `INSERT INTO verify_code(email, code, request_type) VALUES(?, ?, ?)`
	_, err := s.db.Exec(query, email, code, requestType)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) UpdateVerificationCodeStatus(email string, status int) error {
	query := "UPDATE verify_code SET status = ? WHERE email = ?"
	_, err := s.db.Exec(query, status, email)
	if err != nil {
		return err
	}
	
	return nil
}

func (s *Store) IsUserExist(email string) (bool, error) {
	var count int

	err := s.db.QueryRow("SELECT COUNT(*) FROM user WHERE email = ? ", email).Scan(&count)
	if err != nil {
		return false, err
	}

	return (count > 0), nil
}

func scanRowIntoUser(rows *sql.Rows) (*types.User, error) {
	user := new(types.User)

	err := rows.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PhoneNumber,
		&user.Provider,
		&user.LastLoggedIn,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	user.CreatedAt = user.CreatedAt.Local()
	user.LastLoggedIn = user.LastLoggedIn.Local()

	return user, nil
}
