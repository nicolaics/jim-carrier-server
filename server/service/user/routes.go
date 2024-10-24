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

	router.HandleFunc("/user/logout", h.handleLogout).Methods(http.MethodGet)
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
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
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
	err = utils.SendEmail(payload.Email, subject, body)
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

	// // to notify the front-end where to proceed, home screen or reset password screen
	// res := map[string]string{
	// 	"message": fmt.Sprintf("Verification email for %s sent successfully!", strings.ToLower(accountStatus)),
	// 	"accountStatus": strings.ToLower(accountStatus),
	// }

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
