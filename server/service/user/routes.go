package user

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/nicolaics/jim-carrier/constants"
	"github.com/nicolaics/jim-carrier/service/auth"
	"github.com/nicolaics/jim-carrier/types"
	"github.com/nicolaics/jim-carrier/utils"
)

type Handler struct {
	store types.UserStore
}

func NewHandler(store types.UserStore) *Handler {
	return &Handler{store: store}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/user/current", h.handleGetCurrentUser).Methods(http.MethodGet)
	router.HandleFunc("/user/current", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)

	router.HandleFunc("/user", h.handleDelete).Methods(http.MethodDelete)
	router.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)

	router.HandleFunc("/user/modify", h.handleModify).Methods(http.MethodPatch)
	router.HandleFunc("/user/modify", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)

	router.HandleFunc("/user/update-password", h.handleUpdatePassword).Methods(http.MethodPatch)
	router.HandleFunc("/user/update-password", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)

	router.HandleFunc("/user/update-profile-picture", h.handleUpdateProfilePicture).Methods(http.MethodPatch)
	router.HandleFunc("/user/update-profile-picture", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)

	router.HandleFunc("/user/logout", h.handleLogout).Methods(http.MethodPost)
	router.HandleFunc("/user/logout", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
}

func (h *Handler) RegisterUnprotectedRoutes(router *mux.Router) {
	router.HandleFunc("/user/login", h.handleLogin).Methods(http.MethodPost)
	router.HandleFunc("/user/login", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)

	router.HandleFunc("/user/send-verification", h.handleSendVerification).Methods(http.MethodPost)
	router.HandleFunc("/user/send-verification", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)

	router.HandleFunc("/user/register", h.handleRegister).Methods(http.MethodPost)
	router.HandleFunc("/user/register", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)

	router.HandleFunc("/user/reset-password", h.handleResetPassword).Methods(http.MethodPatch)
	router.HandleFunc("/user/reset-password", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)

	// router.HandleFunc("/user/login/google", h.handleLogin).Methods(http.MethodPost)
	// router.HandleFunc("/user/login/google", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.LoginUserPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors))
		return
	}

	user, err := h.store.GetUserByEmail(payload.Email)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("not found, invalid email: %v", err))
		return
	}

	password, err := h.store.GetUserPasswordByEmail(payload.Email)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	// check password match
	if !(auth.ComparePassword(password, []byte(payload.Password))) {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("password incorrect"))
		return
	}

	tokenDetails, err := auth.CreateJWT(user.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	err = h.store.SaveToken(user.ID, tokenDetails)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	err = h.store.UpdateLastLoggedIn(user.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	tokens := map[string]string{
		"token": tokenDetails.Token,
	}

	utils.WriteJSON(w, http.StatusOK, tokens)
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.RegisterUserPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors))
		return
	}

	// verify the code within 5 minutes
	valid, err := h.store.ValidateLoginCodeWithinTime(payload.Email, payload.VerificationCode, 5)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("code validation error: %v", err))
		return
	}

	if !valid {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid verification code or code has expired"))
		return
	}

	err = h.store.UpdateVerificationCodeStatus(payload.Email, constants.VERIFY_CODE_COMPLETE)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error updating verification code: %v", err))
		return
	}

	// check if the newly created user exists
	_, err = h.store.GetUserByEmail(payload.Email)
	if err == nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("user with email %s already exists", payload.Email))
		return
	}

	hashedPassword, err := auth.HashPassword(payload.Password)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	err = h.store.CreateUser(types.User{
		Name:        payload.Name,
		Email:       payload.Email,
		Password:    hashedPassword,
		PhoneNumber: payload.PhoneNumber,
		Provider:    constants.PROVIDER_EMAIL,
		FCMToken:    payload.FCMToken,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
	}

	user, _ := h.store.GetUserByEmail(payload.Email)

	if len(payload.ProfilePicture) > 0 {
		var imageExtension string

		// check image type
		mimeType := http.DetectContentType(payload.ProfilePicture)
		switch mimeType {
		case "image/jpeg":
			imageExtension = ".jpg"
		case "image/png":
			imageExtension = ".png"
		default:
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("unsupported image type"))
			return
		}

		// save the image
		imageURL, err := utils.SaveProfilePicture(user.ID, payload.ProfilePicture, imageExtension)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to save image"))
			return
		}

		err = h.store.UpdateProfilePicture(user.ID, imageURL)
		if err != nil {
			utils.WriteError(w, http.StatusCreated, fmt.Errorf("user created but error update profile picture: %v", err))
			return
		}
	}

	utils.WriteJSON(w, http.StatusCreated, fmt.Sprintf("user %s successfully created", payload.Name))
}

func (h *Handler) handleGetCurrentUser(w http.ResponseWriter, r *http.Request) {
	// validate token
	user, err := h.store.ValidateUserToken(w, r)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("invalid token: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, user)
}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	var payload types.RemoveUserPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors))
		return
	}

	// validate token
	_, err := h.store.ValidateUserToken(w, r)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("invalid token: %v", err))
		return
	}

	user, err := h.store.GetUserByID(payload.ID)
	if user == nil || err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	err = h.store.DeleteUser(user)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, fmt.Sprintf("%s successfully deleted", user.Name))
}

func (h *Handler) handleModify(w http.ResponseWriter, r *http.Request) {
	var payload types.ModifyUserPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors))
		return
	}

	// validate token
	_, err := h.store.ValidateUserToken(w, r)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("invalid token: %v", err))
		return
	}

	user, err := h.store.GetUserByID(payload.ID)
	if user == nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	err = h.store.ModifyUser(user.ID, types.User{
		Name:        payload.Name,
		PhoneNumber: payload.PhoneNumber,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, fmt.Sprintf("%s updated into", payload.Name))
}

func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	accessDetails, err := auth.ExtractTokenFromClient(r)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("invalid token"))
		return
	}

	// check user exists or not
	_, err = h.store.GetUserByID(accessDetails.UserID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user id %d doesn't exists", accessDetails.UserID))
		return
	}

	err = h.store.UpdateLastLoggedIn(accessDetails.UserID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	err = h.store.DeleteToken(accessDetails.UserID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, "successfully logged out")
}

func (h *Handler) handleSendVerification(w http.ResponseWriter, r *http.Request) {
	var payload types.UserVerificationCodePayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors))
		return
	}

	isUserExist, err := h.store.IsUserExist(payload.Email)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user table error: %v", err))
		return
	}

	// check whether there is an active verification code that has been sent within 1 minute
	valid, err := h.store.DelayCodeWithinTime(payload.Email, 1)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("verify code table error: %v", err))
		return
	}

	if valid {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("verification code has already been sent within 1 minute"))
		return
	}

	code := utils.GenerateRandomCodeNumbers(6)

	var accountStatus string
	var requestType int

	// if email exist, send the message for forget password
	if isUserExist {
		accountStatus = "Password Reset"
		requestType = constants.FORGET_PASSWORD
	} else { // else signup
		accountStatus = "Signup"
		requestType = constants.SIGNUP
	}

	subject := fmt.Sprintf("Your Verification Code for %s", accountStatus)
	body := fmt.Sprintf("Your verification code for %s is: %s", strings.ToLower(accountStatus), code)
	err = utils.SendEmail(payload.Email, subject, body, "", "")
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to send email: %v", err))
		return
	}

	// if signup request type is 0, forget password 1
	err = h.store.SaveVerificationCode(payload.Email, code, requestType)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error saving verification code: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, fmt.Sprintf("Verification email for %s sent successfully!", strings.ToLower(accountStatus)))
}

func (h *Handler) handleResetPassword(w http.ResponseWriter, r *http.Request) {
	var payload types.ResetPasswordPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors))
		return
	}

	user, err := h.store.GetUserByEmail(payload.Email)
	if user == nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	hashedPassword, err := auth.HashPassword(payload.Password)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	err = h.store.UpdatePassword(user.ID, hashedPassword)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, "Password reset successfully")
}

func (h *Handler) handleUpdatePassword(w http.ResponseWriter, r *http.Request) {
	var payload types.UpdatePasswordPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors))
		return
	}

	// validate token
	user, err := h.store.ValidateUserToken(w, r)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("invalid token: %v", err))
		return
	}

	password, err := h.store.GetUserPasswordByEmail(user.Email)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	if !(auth.ComparePassword(password, []byte(payload.OldPassword))) {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("incorrect old password"))
		return
	}

	hashedPassword, err := auth.HashPassword(payload.NewPassword)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	err = h.store.UpdatePassword(user.ID, hashedPassword)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, "Password updated successfully")
}

func (h *Handler) handleUpdateProfilePicture(w http.ResponseWriter, r *http.Request) {
	var payload types.UpdateProfilePicturePayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors))
		return
	}

	// validate token
	user, err := h.store.ValidateUserToken(w, r)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("invalid token: %v", err))
		return
	}

	var imageExtension string

	if len(payload.ProfilePicture) > 0 {
		// check image type
		mimeType := http.DetectContentType(payload.ProfilePicture)
		switch mimeType {
		case "image/jpeg":
			imageExtension = ".jpg"
		case "image/png":
			imageExtension = ".png"
		default:
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("unsupported image type"))
			return
		}

		// save the image
		imageURL, err := utils.SaveProfilePicture(user.ID, payload.ProfilePicture, imageExtension)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to save image"))
			return
		}

		err = h.store.UpdateProfilePicture(user.ID, imageURL)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error update profile picture: %v", err))
			return
		}
	} else {
		err = h.store.UpdateProfilePicture(user.ID, "")
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error update profile picture: %v", err))
			return
		}
	}

	utils.WriteJSON(w, http.StatusOK, "profile picture updated successfully")
}

/*
func LoginGoogle(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var authData GoogleAuthDataLogin

	// 요청 바디에서 JSON 데이터 읽기
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	// JSON 데이터를 구조체로 언마샬
	if err := json.Unmarshal(body, &authData); err != nil {
		http.Error(w, "Error unmarshaling request body", http.StatusBadRequest)
		return
	}

	// 여기서 idToken과 serverAuthCode 처리 로직 구현
	tokenInfo, err := oauth.VerifyIdToken(authData.IDToken)
	if err != nil {
		http.Error(w, "Error verifying ID Token: "+err.Error(), http.StatusBadRequest)
		return
	}

	email, ok := tokenInfo.Claims["email"].(string)
	if !ok {
		// email 키가 없거나 값이 문자열이 아닌 경우 에러 처리
		http.Error(w, "Invalid email claim", http.StatusBadRequest)
		return
	}

	// users 테이블에서 이메일과 provider 확인
	exists, provider, err := mysql.CheckEmailAndProvider(email)
	if err != nil {
		// 데이터베이스 에러 처리
		util.SendErrorResponse(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if exists && provider != "google" {
		// provider가 "google"이 아닌 경우 에러 메시지를 반환
		util.SendErrorResponse(w, "This email is associated with a different login method.", http.StatusNotFound)
		return
	} else if !exists {
		// 이메일이 존재하지 않는 경우 로그인 페이지로 이동하도록 메시지를 반환
		util.SendErrorResponse(w, "toRegist", http.StatusNotFound)
		return
	}

	/////////////

	// 해당 사용자의 id 값을 조회합니다.
	id, err := mysql.GetUserIDByEmail(email)
	if err != nil {
		util.SendErrorResponse(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 해당 사용자의 모든 토큰 삭제
	err = mysql.RemoveAllTokensForUser(email)
	if err != nil {
		util.SendErrorResponse(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 새로운 JWT 토큰 생성
	token, err := jwt.GenerateToken(id, email)
	if err != nil {
		util.SendErrorResponse(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// 사용자의 fcm 기기 id 변경
	err = mysql.UpdateFCMid(authData.FCMid, id)
	log.Print("FCMid test : ", authData.FCMid)
	if err != nil {
		util.SERVERtxtLogging("'" + email + "' failed to update FCMid : " + authData.FCMid)
	}

	// 토큰을 데이터베이스에 저장
	err = mysql.StoreToken(email, token)
	if err != nil {
		util.SendErrorResponse(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 사용자 정보 조회
	userdata, err := mysql.GetUserByID(id)
	if err != nil {
		util.SendErrorResponse(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if userdata == nil {
		util.SendErrorResponse(w, "Email not found", http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"userdata": userdata,
		"email":    email,
		"token":    token,
	}

	w.Header().Set("Content-Type", "application/json")
	util.SendSuccessResponse(w, "Login Successful", http.StatusOK, response)
}
*/
