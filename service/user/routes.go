package user

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/nicolaics/jim-carrier-server/constants"
	"github.com/nicolaics/jim-carrier-server/logger"
	"github.com/nicolaics/jim-carrier-server/service/auth"
	"github.com/nicolaics/jim-carrier-server/service/auth/jwt"
	"github.com/nicolaics/jim-carrier-server/service/auth/oauth"
	"github.com/nicolaics/jim-carrier-server/types"
	"github.com/nicolaics/jim-carrier-server/utils"
)

type Handler struct {
	userStore types.UserStore
	bucket    *s3.S3
}

func NewHandler(userStore types.UserStore, bucket *s3.S3) *Handler {
	return &Handler{userStore: userStore, bucket: bucket}
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

	router.HandleFunc("/user/verify-verification", h.handleVerifyVerification).Methods(http.MethodPost)
	router.HandleFunc("/user/verify-verification", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)

	router.HandleFunc("/user/register", h.handleRegister).Methods(http.MethodPost)
	router.HandleFunc("/user/register", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)

	router.HandleFunc("/user/reset-password", h.handleResetPassword).Methods(http.MethodPatch)
	router.HandleFunc("/user/reset-password", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)

	router.HandleFunc("/user/login/google", h.handleLoginGoogle).Methods(http.MethodPost)
	router.HandleFunc("/user/login/google", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)

	router.HandleFunc("/user/register/google", h.handleRegisterGoogle).Methods(http.MethodPost)
	router.HandleFunc("/user/register/google", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)

	router.HandleFunc("/user/refresh", h.handleRefreshToken).Methods(http.MethodPost)
	router.HandleFunc("/user/refresh", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)

	router.HandleFunc("/user/login/auto", h.handleAutoLogin).Methods(http.MethodPost)
	router.HandleFunc("/user/login/auto", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.LoginUserPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		log.Printf("payload error: %v \n", err)
		logger.WriteServerLog(fmt.Sprintf("payload error: %v", err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("payload error"))
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		log.Printf("invalid payload: %v", errors)
		logger.WriteServerLog(fmt.Sprintf("payload error: %v", errors))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload"))
		return
	}

	// check whether the provider is email or not
	exists, provider, err := h.userStore.CheckProvider(payload.Email)
	if err != nil {
		log.Println(err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("%v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	if exists && provider != "email" {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("this email is associated with a different login method"))
		return
	} else if !exists {
		utils.WriteError(w, http.StatusNotFound, fmt.Errorf("account not found, go to registration"))
		return
	}

	user, err := h.userStore.GetUserByEmail(payload.Email)
	if err != nil {
		log.Printf("user not found: %v", err)
		logger.WriteServerLog(fmt.Sprintf("user not found: %v", err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user not found"))
		return
	}

	password, err := h.userStore.GetUserPasswordByEmail(payload.Email)
	if err != nil {
		log.Println(err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("%v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	// check password match
	if !(auth.ComparePassword(password, []byte(payload.Password))) {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("invalid password"))
		return
	}

	isAccessTokenExist, err := h.userStore.IsAccessTokenExist(user.ID)
	if err != nil {
		log.Println(err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("%v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	if isAccessTokenExist {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("logged in from other device"))
		return
	}

	accessTokenDetails, err := jwt.CreateAccessToken(user.ID)
	if err != nil {
		log.Printf("failed to generate access token: %v", err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("failed to generate access token: %v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	refreshTokenDetails, err := jwt.CreateRefreshToken(user.ID)
	if err != nil {
		log.Printf("failed to generate refresh token: %v", err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("failed to generate refresh token: %v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	err = h.userStore.SaveToken(user.ID, accessTokenDetails, refreshTokenDetails)
	if err != nil {
		log.Printf("error saving token: %v", err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("error saving token: %v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	err = h.userStore.UpdateLastLoggedIn(user.ID)
	if err != nil {
		log.Println(err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("%v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	if user.FCMToken != payload.FCMToken {
		err = h.userStore.UpdateFCMToken(user.ID, payload.FCMToken)
		if err != nil {
			logger.WriteServerLog(fmt.Sprintf("error update FCM token for user %s: %v", user.Email, err))
		}
	}

	tokens := map[string]string{
		"access_token":  accessTokenDetails.Token,
		"refresh_token": refreshTokenDetails.Token,
	}

	utils.WriteJSON(w, http.StatusOK, tokens)
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.RegisterUserPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		log.Printf("payload error: %v \n", err)
		logger.WriteServerLog(fmt.Sprintf("payload error: %v", err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("payload error"))
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		log.Printf("invalid payload: %v", errors)
		logger.WriteServerLog(fmt.Sprintf("payload error: %v", errors))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload"))
		return
	}

	// verify the code within 5 minutes
	valid, err := h.userStore.ValidateLoginCodeWithinTime(payload.Email, payload.VerificationCode, 5, constants.SIGNUP)
	if err != nil {
		log.Printf("code validation error: %v", err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("code validation error: %v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	if !valid {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("wrong verification code or code has expired"))
		return
	}

	err = h.userStore.UpdateVerificationCodeStatus(payload.Email, constants.VERIFY_CODE_COMPLETE, constants.SIGNUP)
	if err != nil {
		log.Printf("error update verification code: %v", err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("error update verification code: %v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	// check if the newly created user exists
	exist, _, err := h.userStore.CheckProvider(payload.Email)
	if err != nil {
		log.Println(err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("%v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}
	if exist {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("user already exists, please proceed to Login"))
		return
	}

	hashedPassword, err := auth.HashPassword(payload.Password)
	if err != nil {
		log.Println(err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("%v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	err = h.userStore.CreateUser(types.User{
		Name:        payload.Name,
		Email:       payload.Email,
		Password:    hashedPassword,
		PhoneNumber: payload.PhoneNumber,
		Provider:    constants.PROVIDER_EMAIL,
		FCMToken:    payload.FCMToken,
	})
	if err != nil {
		log.Printf("error create user: %v", err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("error create user: %v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
	}

	user, _ := h.userStore.GetUserByEmail(payload.Email)

	if len(payload.ProfilePicture) > constants.PROFILE_IMG_MAX_BYTES {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("image size exceeds the limit of 5MB"))
		return
	} else if len(payload.ProfilePicture) > 0 {
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
		imageURL, err := utils.SaveProfilePicture(user.ID, payload.ProfilePicture, user.ProfilePictureURL, imageExtension, h.bucket)
		if err != nil {
			logger.WriteServerLog(fmt.Sprintf("user %s created but error saving profile picture: %v", payload.Email, err))
		}

		if imageURL != "" {
			err = h.userStore.UpdateProfilePicture(user.ID, imageURL)
			if err != nil {
				logger.WriteServerLog(fmt.Sprintf("user %s created but error update profile picture: %v", payload.Email, err))
			}
		}
	}

	utils.WriteJSON(w, http.StatusCreated, fmt.Sprintf("user %s successfully created", payload.Name))
}

func (h *Handler) handleGetCurrentUser(w http.ResponseWriter, r *http.Request) {
	// validate token
	user, err := h.userStore.ValidateUserAccessToken(w, r)
	if err != nil {
		log.Printf("token invalid: %v", err)
		logger.WriteServerLog(fmt.Sprintf("token invalid: %v", err))
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("token invalid"))
		return
	}

	imageBytes, err := utils.GetImage(user.ProfilePictureURL, h.bucket)
	if err != nil {
		log.Printf("error reading the picture: %v", err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("error reading the picture: %v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	response := types.ReturnUserPayload{
		ID:             user.ID,
		Name:           user.Name,
		Email:          user.Email,
		PhoneNumber:    user.PhoneNumber,
		Provider:       user.Provider,
		ProfilePicture: imageBytes,
		FCMToken:       user.FCMToken,
		LastLoggedIn:   user.LastLoggedIn,
		CreatedAt:      user.CreatedAt,
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	// validate token
	user, err := h.userStore.ValidateUserAccessToken(w, r)
	if err != nil {
		log.Printf("token invalid: %v", err)
		logger.WriteServerLog(fmt.Sprintf("token invalid: %v", err))
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("token invalid"))
		return
	}

	user, err = h.userStore.GetUserByID(user.ID)
	if user == nil || err != nil {
		log.Printf("payload error: %v \n", err)
		logger.WriteServerLog(fmt.Sprintf("payload error: %v", err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("payload error"))
		return
	}

	isAllowed, err := h.userStore.IsDeleteUserAllowed(user.ID)
	if err != nil {
		log.Printf("is delete user error: %v \n", err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("is delete user error: %v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	if !isAllowed {
		utils.WriteError(w, http.StatusForbidden, fmt.Errorf("there are still ongoing listings/orders"))
		return
	}

	err = h.userStore.DeleteUser(user)
	if err != nil {
		log.Printf("error delete user: %v", err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("error delete user: %v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	utils.WriteJSON(w, http.StatusOK, fmt.Sprintf("%s successfully deleted", user.Name))
}

func (h *Handler) handleModify(w http.ResponseWriter, r *http.Request) {
	var payload types.ModifyUserPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		log.Printf("payload error: %v \n", err)
		logger.WriteServerLog(fmt.Sprintf("payload error: %v", err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("payload error"))
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		log.Printf("invalid payload: %v", errors)
		logger.WriteServerLog(fmt.Sprintf("payload error: %v", errors))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload"))
		return
	}

	// validate token
	user, err := h.userStore.ValidateUserAccessToken(w, r)
	if err != nil {
		log.Printf("token invalid: %v", err)
		logger.WriteServerLog(fmt.Sprintf("token invalid: %v", err))
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("token invalid"))
		return
	}

	user, err = h.userStore.GetUserByID(user.ID)
	if user == nil {
		log.Printf("payload error: %v \n", err)
		logger.WriteServerLog(fmt.Sprintf("payload error: %v", err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("payload error"))
		return
	}

	err = h.userStore.ModifyUser(user.ID, types.User{
		Name:        payload.Name,
		PhoneNumber: payload.PhoneNumber,
	})
	if err != nil {
		log.Printf("error modify user: %v", err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("error modify user: %v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	utils.WriteJSON(w, http.StatusOK, fmt.Sprintf("%s updated into", payload.Name))
}

func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	accessDetails, err := jwt.ExtractAccessTokenFromClient(r)
	if err != nil {
		log.Printf("token invalid: %v", err)
		logger.WriteServerLog(fmt.Sprintf("token invalid: %v", err))
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("token invalid"))
		return
	}

	// check user exists or not
	_, err = h.userStore.GetUserByID(accessDetails.UserID)
	if err != nil {
		log.Printf("user id %d doesn't exists: %v", accessDetails.UserID, err)
		logger.WriteServerLog(fmt.Sprintf("user id %d doesn't exists: %v", accessDetails.UserID, err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user doesn't exists"))
		return
	}

	err = h.userStore.UpdateLastLoggedIn(accessDetails.UserID)
	if err != nil {
		log.Println(err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("%v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	err = h.userStore.DeleteToken(accessDetails.UserID)
	if err != nil {
		log.Println(err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("%v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	utils.WriteJSON(w, http.StatusOK, "successfully logged out")
}

func (h *Handler) handleSendVerification(w http.ResponseWriter, r *http.Request) {
	var payload types.UserVerificationCodePayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		log.Printf("payload error: %v \n", err)
		logger.WriteServerLog(fmt.Sprintf("payload error: %v", err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("payload error"))
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		log.Printf("invalid payload: %v", errors)
		logger.WriteServerLog(fmt.Sprintf("payload error: %v", errors))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload"))
		return
	}

	isUserExist, err := h.userStore.IsUserExist(payload.Email)
	if err != nil {
		log.Printf("user table error: %v", err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("user table error: %v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	// check whether there is an active verification code that has been sent within 1 minute
	valid, err := h.userStore.DelayCodeWithinTime(payload.Email, 1)
	if err != nil {
		log.Printf("verify code table error: %v", err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("verify code table error: %v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
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
	body := fmt.Sprintf("<p>Your verification code for %s is:</p><br><h2>%s</h2>", strings.ToLower(accountStatus), code)
	err = utils.SendEmail(payload.Email, subject, body, "", "")
	if err != nil {
		log.Printf("failed to send email: %v", err)
		logger.WriteServerLog(fmt.Sprintf("failed to send email: %v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to send verification code email"))
		return
	}

	// if signup request type is 0, forget password 1
	err = h.userStore.SaveVerificationCode(payload.Email, code, requestType)
	if err != nil {
		log.Printf("error saving verification code: %v", err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("error saving verification code: %v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	utils.WriteJSON(w, http.StatusOK, fmt.Sprintf("Verification email for %s sent successfully!", strings.ToLower(accountStatus)))
}

func (h *Handler) handleResetPassword(w http.ResponseWriter, r *http.Request) {
	var payload types.ResetPasswordPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		log.Printf("payload error: %v \n", err)
		logger.WriteServerLog(fmt.Sprintf("payload error: %v", err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("payload error"))
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		log.Printf("invalid payload: %v", errors)
		logger.WriteServerLog(fmt.Sprintf("payload error: %v", errors))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload"))
		return
	}

	user, err := h.userStore.GetUserByEmail(payload.Email)
	if user == nil {
		log.Printf("payload error: %v \n", err)
		logger.WriteServerLog(fmt.Sprintf("payload error: %v", err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("payload error"))
		return
	}

	hashedPassword, err := auth.HashPassword(payload.NewPassword)
	if err != nil {
		log.Println(err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("%v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	err = h.userStore.UpdatePassword(user.ID, hashedPassword)
	if err != nil {
		log.Println(err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("%v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	utils.WriteJSON(w, http.StatusOK, "Password reset successfully")
}

func (h *Handler) handleUpdatePassword(w http.ResponseWriter, r *http.Request) {
	var payload types.UpdatePasswordPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		log.Printf("payload error: %v \n", err)
		logger.WriteServerLog(fmt.Sprintf("payload error: %v", err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("payload error"))
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		log.Printf("invalid payload: %v", errors)
		logger.WriteServerLog(fmt.Sprintf("payload error: %v", errors))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload"))
		return
	}

	// validate token
	user, err := h.userStore.ValidateUserAccessToken(w, r)
	if err != nil {
		log.Printf("token invalid: %v", err)
		logger.WriteServerLog(fmt.Sprintf("token invalid: %v", err))
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("token invalid"))
		return
	}

	password, err := h.userStore.GetUserPasswordByEmail(user.Email)
	if err != nil {
		log.Println(err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("%v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	if !(auth.ComparePassword(password, []byte(payload.OldPassword))) {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("incorrect old password"))
		return
	}

	hashedPassword, err := auth.HashPassword(payload.NewPassword)
	if err != nil {
		log.Println(err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("%v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	err = h.userStore.UpdatePassword(user.ID, hashedPassword)
	if err != nil {
		log.Println(err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("%v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	utils.WriteJSON(w, http.StatusOK, "Password updated successfully")
}

func (h *Handler) handleUpdateProfilePicture(w http.ResponseWriter, r *http.Request) {
	var payload types.UpdateProfilePicturePayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		log.Printf("payload error: %v \n", err)
		logger.WriteServerLog(fmt.Sprintf("payload error: %v", err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("payload error"))
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		log.Printf("invalid payload: %v", errors)
		logger.WriteServerLog(fmt.Sprintf("payload error: %v", errors))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload"))
		return
	}

	// validate token
	user, err := h.userStore.ValidateUserAccessToken(w, r)
	if err != nil {
		log.Printf("token invalid: %v", err)
		logger.WriteServerLog(fmt.Sprintf("token invalid: %v", err))
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("token invalid"))
		return
	}

	var imageExtension string

	if len(payload.ProfilePicture) > constants.PROFILE_IMG_MAX_BYTES {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("image size exceeds the limit of 5MB"))
		return
	} else if len(payload.ProfilePicture) > 0 {
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
		imageURL, err := utils.SaveProfilePicture(user.ID, payload.ProfilePicture, imageExtension, user.ProfilePictureURL, h.bucket)
		if err != nil {
			log.Printf("failed to save image: %v", err)
			logFile, _ := logger.WriteServerLog(fmt.Sprintf("failed to save image: %v", err))
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
			return
		}

		err = h.userStore.UpdateProfilePicture(user.ID, imageURL)
		if err != nil {
			log.Printf("error update profile picture: %v", err)
			logFile, _ := logger.WriteServerLog(fmt.Sprintf("error update profile picture: %v", err))
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
			return
		}
	} else {
		err = h.userStore.UpdateProfilePicture(user.ID, "")
		if err != nil {
			log.Printf("error update profile picture: %v", err)
			logFile, _ := logger.WriteServerLog(fmt.Sprintf("error update profile picture: %v", err))
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
			return
		}
	}

	utils.WriteJSON(w, http.StatusOK, "profile picture updated successfully")
}

func (h *Handler) handleLoginGoogle(w http.ResponseWriter, r *http.Request) {
	var payload types.LoginGooglePayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		log.Printf("payload error: %v \n", err)
		logger.WriteServerLog(fmt.Sprintf("payload error: %v", err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("payload error"))
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		log.Printf("invalid payload: %v", errors)
		logger.WriteServerLog(fmt.Sprintf("payload error: %v", errors))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload"))
		return
	}

	// Verify the token received
	tokenInfo, err := oauth.VerifyIDToken(payload.IDToken)
	if err != nil {
		log.Printf("error verifying id token: %v", err)
		logger.WriteServerLog(fmt.Sprintf("error verifying id token: %v", err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error verifying id token"))
		return
	}

	email, ok := tokenInfo.Claims["email"].(string)
	if !ok {
		// there's no email key or something happened
		log.Printf("invalid email claim: %v", err)
		logger.WriteServerLog(fmt.Sprintf("invalid email claim: %v", err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid email claim"))
		return
	}

	// check whether the provider is google or not
	exists, provider, err := h.userStore.CheckProvider(email)
	if err != nil {
		log.Println(err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("%v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	if exists && provider != "google" {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("this email is associated with a different login method"))
		return
	} else if !exists {
		utils.WriteError(w, http.StatusNotFound, fmt.Errorf("account not found, go to registration"))
		return
	}

	// get the user
	user, err := h.userStore.GetUserByEmail(email)
	if err != nil || user == nil {
		log.Println(err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("%v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	if user.FCMToken != payload.FCMToken {
		err = h.userStore.UpdateFCMToken(user.ID, payload.FCMToken)
		if err != nil {
			logger.WriteServerLog(fmt.Sprintf("error update FCM token for user %s: %v", user.Email, err))
		}
	}

	isAccessTokenExist, err := h.userStore.IsAccessTokenExist(user.ID)
	if err != nil {
		log.Println(err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("%v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	if isAccessTokenExist {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("logged in from other device"))
		return
	}

	accessTokenDetails, err := jwt.CreateAccessToken(user.ID)
	if err != nil {
		log.Printf("failed to generate access token: %v", err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("failed to generate access token: %v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	refreshTokenDetails, err := jwt.CreateRefreshToken(user.ID)
	if err != nil {
		log.Printf("failed to generate refresh token: %v", err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("failed to generate refresh token: %v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	log.Print("FCM token: ", payload.FCMToken)

	err = h.userStore.UpdateFCMToken(user.ID, payload.FCMToken)
	if err != nil {
		logger.WriteServerLog(fmt.Sprintf("%s failed to update FCM Token: %v", user.Email, err))
	}

	err = h.userStore.SaveToken(user.ID, accessTokenDetails, refreshTokenDetails)
	if err != nil {
		log.Printf("error saving token: %v", err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("error saving token: %v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	err = h.userStore.UpdateLastLoggedIn(user.ID)
	if err != nil {
		log.Println(err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("%v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	response := map[string]interface{}{
		"user":          user,
		"access_token":  accessTokenDetails.Token,
		"refresh_token": refreshTokenDetails.Token,
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) handleRegisterGoogle(w http.ResponseWriter, r *http.Request) {
	var payload types.RegisterGooglePayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		log.Printf("payload error: %v \n", err)
		logger.WriteServerLog(fmt.Sprintf("payload error: %v", err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("payload error"))
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		log.Printf("invalid payload: %v", errors)
		logger.WriteServerLog(fmt.Sprintf("payload error: %v", errors))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload"))
		return
	}

	// Verify the token received
	tokenInfo, err := oauth.VerifyIDToken(payload.IDToken)
	if err != nil {
		log.Printf("error verifying id token: %v", err)
		logger.WriteServerLog(fmt.Sprintf("error verifying id token: %v", err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error verifying id token"))
		return
	}

	email, ok := tokenInfo.Claims["email"].(string)
	if !ok {
		// there's no email key or something happened
		log.Printf("invalid email claim: %v", err)
		logger.WriteServerLog(fmt.Sprintf("invalid email claim: %v", err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid email claim"))
		return
	}

	exist, _, err := h.userStore.CheckProvider(email)
	if err != nil {
		log.Println(err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("%v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}
	if exist {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user exist already, proceed to login"))
		return
	}

	log.Print("Regist google FCM token: ", payload.FCMToken)

	err = h.userStore.CreateUser(types.User{
		Name:        payload.Name,
		Email:       email,
		PhoneNumber: payload.PhoneNumber,
		Provider:    "google",
		FCMToken:    payload.FCMToken,
	})
	if err != nil {
		log.Printf("error create user: %v", err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("error create user: %v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	user, err := h.userStore.GetUserByEmail(email)
	if err != nil {
		log.Println(err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("%v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	if payload.ProfilePictureURL != "" {
		imageData, imageExtension, err := utils.DownloadImage(payload.ProfilePictureURL)
		if err != nil {
			logger.WriteServerLog(fmt.Sprintf("user %s created but error download profile picture: %v", email, err))
		} else {
			// save the image
			imageURL, err := utils.SaveProfilePicture(user.ID, imageData, imageExtension, user.ProfilePictureURL, h.bucket)
			if err != nil {
				logger.WriteServerLog(fmt.Sprintf("user %s created but error saving profile picture: %v", email, err))
			}

			if imageURL != "" {
				err = h.userStore.UpdateProfilePicture(user.ID, imageURL)
				if err != nil {
					logger.WriteServerLog(fmt.Sprintf("user %s created but error update profile picture: %v", email, err))
				}
			}
		}
	}

	accessTokenDetails, err := jwt.CreateAccessToken(user.ID)
	if err != nil {
		log.Printf("failed to generate access token: %v", err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("failed to generate access token: %v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	refreshTokenDetails, err := jwt.CreateRefreshToken(user.ID)
	if err != nil {
		log.Printf("failed to generate refresh token: %v", err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("failed to generate refresh token: %v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	err = h.userStore.SaveToken(user.ID, accessTokenDetails, refreshTokenDetails)
	if err != nil {
		log.Printf("error saving token: %v", err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("error saving token: %v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	tokens := map[string]string{
		"access_token":  accessTokenDetails.Token,
		"refresh_token": refreshTokenDetails.Token,
	}

	utils.WriteJSON(w, http.StatusCreated, tokens)
}

func (h *Handler) handleRefreshToken(w http.ResponseWriter, r *http.Request) {
	var payload types.RefreshTokenPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		log.Printf("payload error: %v \n", err)
		logger.WriteServerLog(fmt.Sprintf("payload error: %v", err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("payload error"))
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		log.Printf("invalid payload: %v", errors)
		logger.WriteServerLog(fmt.Sprintf("payload error: %v", errors))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload"))
		return
	}

	user, err := h.userStore.ValidateUserRefreshToken(payload.RefreshToken)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, err)
		return
	}

	accessTokenDetails, err := jwt.CreateAccessToken(user.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to generate access token: %v", err))
		return
	}

	err = h.userStore.UpdateAccessToken(user.ID, accessTokenDetails)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error update access token: %v", err))
		return
	}

	tokens := map[string]string{
		"access_token": accessTokenDetails.Token,
	}

	utils.WriteJSON(w, http.StatusOK, tokens)
}

func (h *Handler) handleAutoLogin(w http.ResponseWriter, r *http.Request) {
	var payload types.AutoLoginPayload

	// validate token
	if err := utils.ParseJSON(r, &payload); err != nil {
		log.Printf("payload error: %v \n", err)
		logger.WriteServerLog(fmt.Sprintf("payload error: %v", err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("payload error"))
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		log.Printf("invalid payload: %v", errors)
		logger.WriteServerLog(fmt.Sprintf("payload error: %v", errors))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload"))
		return
	}

	user, err := h.userStore.ValidateUserRefreshToken(payload.RefreshToken)
	if err != nil {
		log.Printf("error validate user refresh token: %v", err)
		logger.WriteServerLog(fmt.Sprintf("error validate user refresh token: %v", err))
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
		return
	}

	err = h.userStore.DeleteToken(user.ID)
	if err != nil {
		log.Printf("error delete token: %v", err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("error delete token: %v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	accessTokenDetails, err := jwt.CreateAccessToken(user.ID)
	if err != nil {
		log.Printf("failed to generate access token: %v", err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("failed to generate access token: %v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	err = h.userStore.UpdateAccessToken(user.ID, accessTokenDetails)
	if err != nil {
		log.Printf("error saving token: %v", err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("error saving token: %v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	if user.FCMToken != payload.FCMToken {
		err = h.userStore.UpdateFCMToken(user.ID, payload.FCMToken)
		if err != nil {
			logger.WriteServerLog(fmt.Sprintf("error update FCM token for user %s: %v", user.Email, err))
		}
	}

	err = h.userStore.UpdateLastLoggedIn(user.ID)
	if err != nil {
		log.Println(err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("%v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	tokens := map[string]string{
		"access_token": accessTokenDetails.Token,
	}

	utils.WriteJSON(w, http.StatusOK, tokens)
}

func (h *Handler) handleVerifyVerification(w http.ResponseWriter, r *http.Request) {
	var payload types.VerifyVerificationCodePayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		log.Printf("payload error: %v \n", err)
		logger.WriteServerLog(fmt.Sprintf("payload error: %v", err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("payload error"))
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		log.Printf("invalid payload: %v", errors)
		logger.WriteServerLog(fmt.Sprintf("payload error: %v", errors))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload"))
		return
	}

	// verify the code within 5 minutes
	valid, err := h.userStore.ValidateLoginCodeWithinTime(payload.Email, payload.VerificationCode, 5, constants.FORGET_PASSWORD)
	if err != nil {
		log.Printf("code validation error: %v", err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("code validation error: %v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	if !valid {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("wrong verification code or code has expired"))
		return
	}

	err = h.userStore.UpdateVerificationCodeStatus(payload.Email, constants.VERIFY_CODE_COMPLETE, constants.FORGET_PASSWORD)
	if err != nil {
		log.Printf("error updating verification code: %v", err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("error updating verification code: %v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	utils.WriteJSON(w, http.StatusOK, "Verification successful!")
}
