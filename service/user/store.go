package user

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/nicolaics/jim-carrier-server/constants"
	"github.com/nicolaics/jim-carrier-server/service/auth/jwt"
	"github.com/nicolaics/jim-carrier-server/types"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) GetUserByEmail(email string) (*types.User, error) {
	query := `SELECT id, name, email, phone_number, provider, 
					profile_picture_url, fcm_token, 
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
				profile_picture_url, fcm_token, 
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
					profile_picture_url, fcm_token, 
					last_logged_in, created_at 
					FROM user WHERE name LIKE ?`
		searchVal := "%"

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
					profile_picture_url, fcm_token, 
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
					profile_picture_url, fcm_token, 
					last_logged_in, created_at 
					FROM user WHERE phone_number LIKE ?`
		searchVal := "%"

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
					profile_picture_url, fcm_token, 
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
				profile_picture_url, fcm_token, 
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
	query := `INSERT INTO user (name, email, password, 
								phone_number, provider, 
								fcm_token, profile_picture_url) 
				VALUES (?, ?, ?, ?, ?, ?, ?)`

	defaultProfilePicture := constants.PROFILE_IMG_DIR_PATH + "default.png"

	_, err := s.db.Exec(query, user.Name, user.Email, user.Password,
		user.PhoneNumber, user.Provider, user.FCMToken,
		defaultProfilePicture)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) DeleteUser(user *types.User) error {
	query := `SELECT COUNT(*) FROM order_list WHERE giver_id = ? AND (order_status != ? OR payment_status = ?)`
	row := s.db.QueryRow(query, user.ID, constants.ORDER_STATUS_COMPLETED, constants.PAYMENT_STATUS_PENDING)
	if row.Err() != nil {
		return row.Err()
	}

	var count int
	err := row.Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("there is still pending giving order")
	}

	query = `SELECT COUNT(*) 
			FROM listing 
			JOIN order_list ON order.listing_id = listing.id 
			WHERE listing.carrier_id = ? AND order.order_status != ?`
	row = s.db.QueryRow(query, user.ID, constants.ORDER_STATUS_COMPLETED)
	if row.Err() != nil {
		return row.Err()
	}

	count = 0
	err = row.Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("there is still pending carrying order")
	}

	_, err = s.db.Exec("DELETE FROM review WHERE reviewer_id = ?", user.ID)
	if err != nil {
		return err
	}

	_, err = s.db.Exec("DELETE FROM order_list WHERE giver_id = ?", user.ID)
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

func (s *Store) UpdateProfilePicture(id int, profPicUrl string) error {
	query := `UPDATE user SET profile_picture_url = ? WHERE id = ?`
	_, err := s.db.Exec(query, profPicUrl, id)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) SaveToken(userId int, accessTokenDetails *types.TokenDetails, refreshTokenDetails *types.TokenDetails) error {
	accessTokenExp := time.Unix(accessTokenDetails.TokenExp, 0)   //converting Unix to UTC(to Time object)
	refreshTokenExp := time.Unix(refreshTokenDetails.TokenExp, 0) //converting Unix to UTC(to Time object)

	query := "INSERT INTO verify_token(user_id, uuid, token_type, expired_at) VALUES (?, ?, ?, ?)"
	_, err := s.db.Exec(query, userId, accessTokenDetails.UUID, constants.ACCESS_TOKEN, accessTokenExp)
	if err != nil {
		return err
	}

	query = `SELECT COUNT(*) FROM verify_token WHERE user_id = ? AND token_type = ?`
	row := s.db.QueryRow(query, userId, constants.REFRESH_TOKEN)
	if row.Err() != nil {
		return row.Err()
	}

	var count int
	err = row.Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		query = "INSERT INTO verify_token(user_id, uuid, token_type, expired_at) VALUES (?, ?, ?, ?)"
		_, err = s.db.Exec(query, userId, refreshTokenDetails.UUID, constants.REFRESH_TOKEN, refreshTokenExp)
	} else {
		query = `UPDATE verify_token SET uuid = ?, expired_at = ? WHERE user_id = ? AND token_type = ?`
		_, err = s.db.Exec(query, refreshTokenDetails.UUID, refreshTokenExp, userId, constants.REFRESH_TOKEN)
	}
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

func (s *Store) ValidateUserAccessToken(w http.ResponseWriter, r *http.Request) (*types.User, error) {
	query := "DELETE FROM verify_token WHERE expired_at < ?"
	_, err := s.db.Exec(query, time.Now().UTC().Format("2006-01-02 15:04:05"))
	if err != nil {
		return nil, fmt.Errorf("error deleting expired token: %v", err)
	}

	accessDetails, err := jwt.ExtractAccessTokenFromClient(r)
	if err != nil {
		return nil, err
	}

	query = `SELECT COUNT(*) FROM verify_token WHERE user_id = ? AND expired_at >= ? AND token_type = ?`
	row := s.db.QueryRow(query, accessDetails.UserID, time.Now().UTC().Format("2006-01-02 15:04:05"), constants.ACCESS_TOKEN)
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
	} else if count == 0 {
		return nil, fmt.Errorf("access token expired")
	}

	query = `SELECT user_id FROM verify_token WHERE uuid = ? AND user_id = ? AND expired_at >= ? AND token_type = ?`
	row = s.db.QueryRow(query, accessDetails.UUID, accessDetails.UserID, time.Now().UTC().Format("2006-01-02 15:04:05"), constants.ACCESS_TOKEN)
	if row.Err() != nil {
		if row.Err() == sql.ErrNoRows {
			return nil, nil
		}

		return nil, row.Err()
	}

	var userId int
	err = row.Scan(&userId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// check if user exist
	user, err := s.GetUserByID(userId)
	if err != nil {
		delErr := s.DeleteToken(accessDetails.UserID)
		if delErr != nil {
			return nil, fmt.Errorf("delete error: %v", delErr)
		}

		return nil, fmt.Errorf("account not found")
	}

	return user, nil
}

func (s *Store) ValidateUserRefreshToken(refreshToken string) (*types.User, error) {
	query := "DELETE FROM verify_token WHERE expired_at < ?"
	_, err := s.db.Exec(query, time.Now().UTC().Format("2006-01-02 15:04:05"))
	if err != nil {
		return nil, fmt.Errorf("error deleting expired token: %v", err)
	}

	accessDetails, err := jwt.ExtractRefreshTokenFromClient(refreshToken)
	if err != nil {
		return nil, err
	}

	query = `SELECT COUNT(*) FROM verify_token WHERE user_id = ? AND expired_at >= ? AND token_type = ?`
	row := s.db.QueryRow(query, accessDetails.UserID, time.Now().UTC().Format("2006-01-02 15:04:05"), constants.REFRESH_TOKEN)
	if row.Err() != nil {
		return nil, row.Err()
	}

	var count int
	err = row.Scan(&count)
	if err != nil {
		return nil, err
	}

	if count == 0 {
		err = s.DeleteToken(accessDetails.UserID)
		if err != nil {
			return nil, fmt.Errorf("delete error: %v", err)
		}

		return nil, fmt.Errorf("refresh token expired")
	} else if count > 1 {
		return nil, fmt.Errorf("contact admin")
	}

	query = `SELECT user_id FROM verify_token WHERE uuid = ? AND user_id = ? AND expired_at >= ? AND token_type = ?`
	row = s.db.QueryRow(query, accessDetails.UUID, accessDetails.UserID, time.Now().UTC().Format("2006-01-02 15:04:05"), constants.REFRESH_TOKEN)
	if row.Err() != nil {
		if row.Err() == sql.ErrNoRows {
			return nil, nil
		}

		return nil, row.Err()
	}

	var userId int
	err = row.Scan(&userId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		
		return nil, err
	}

	user, err := s.GetUserByID(userId)
	if err != nil {
		delErr := s.DeleteToken(accessDetails.UserID)
		if delErr != nil {
			return nil, fmt.Errorf("delete error: %v", delErr)
		}

		return nil, fmt.Errorf("account not found")
	}

	return user, nil
}

func (s *Store) UpdateAccessToken(userId int, accessTokenDetails *types.TokenDetails) error {
	accessTokenExp := time.Unix(accessTokenDetails.TokenExp, 0) //converting Unix to UTC(to Time object)

	query := `INSERT INTO verify_token(user_id, uuid, token_type, expired_at) VALUES (?, ?, ?, ?)`
	_, err := s.db.Exec(query, userId, accessTokenDetails.UUID, constants.ACCESS_TOKEN, accessTokenExp)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) IsAccessTokenExist(userId int) (bool, error) {
	query := `SELECT COUNT(*) FROM verify_token WHERE user_id = ? AND expired_at >= ? AND token_type = ?`
	row := s.db.QueryRow(query, userId, time.Now().UTC().Format("2006-01-02 15:04:05"), constants.ACCESS_TOKEN)
	if row.Err() != nil {
		return true, row.Err()
	}

	var count int
	err := row.Scan(&count)
	if err != nil {
		return true, err
	}

	return (count > 0), nil
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

func (s *Store) CheckProvider(email string) (bool, string, error) {
	query := `SELECT provider FROM user WHERE email = ?`
	row := s.db.QueryRow(query, email)
	if row.Err() != nil {
		return false, "", row.Err()
	}

	var provider string
	err := row.Scan(&provider)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, "", nil
		}

		return false, "", err
	}

	return true, provider, nil
}

func (s *Store) UpdateFCMToken(id int, fcmToken string) error {
	query := `UPDATE user SET fcm_token = ? WHERE id = ?`
	_, err := s.db.Exec(query, fcmToken, id)
	if err != nil {
		return err
	}

	return nil
}

func scanRowIntoUser(rows *sql.Rows) (*types.User, error) {
	temp := new(struct {
		ID                int
		Name              string
		Email             string
		PhoneNumber       string
		Provider          string
		ProfilePictureURL string
		FCMToken          sql.NullString
		LastLoggedIn      time.Time `json:"lastLoggedIn"`
		CreatedAt         time.Time `json:"createdAt"`
	})

	err := rows.Scan(
		&temp.ID,
		&temp.Name,
		&temp.Email,
		&temp.PhoneNumber,
		&temp.Provider,
		&temp.ProfilePictureURL,
		&temp.FCMToken,
		&temp.LastLoggedIn,
		&temp.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	user := &types.User{
		ID:                temp.ID,
		Name:              temp.Name,
		Email:             temp.Email,
		PhoneNumber:       temp.PhoneNumber,
		Provider:          temp.Provider,
		ProfilePictureURL: temp.ProfilePictureURL,
		FCMToken:          temp.FCMToken.String,
		LastLoggedIn:      temp.LastLoggedIn,
		CreatedAt:         temp.CreatedAt,
	}

	user.CreatedAt = user.CreatedAt.Local()
	user.LastLoggedIn = user.LastLoggedIn.Local()

	return user, nil
}
