package bank

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/nicolaics/jim-carrier-server/logger"
	"github.com/nicolaics/jim-carrier-server/types"
	"github.com/nicolaics/jim-carrier-server/utils"
)

type Handler struct {
	bankDetailStore types.BankDetailStore
	userStore       types.UserStore
}

func NewHandler(bankDetailStore types.BankDetailStore, userStore types.UserStore) *Handler {
	return &Handler{bankDetailStore: bankDetailStore, userStore: userStore}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/bank-detail/update", h.handleUpdateBankDetail).Methods(http.MethodPost)
	router.HandleFunc("/bank-detail/update", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)

	router.HandleFunc("/bank-detail", h.handleGetBankDetail).Methods(http.MethodGet)
	router.HandleFunc("/bank-detail", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
}

func (h *Handler) handleUpdateBankDetail(w http.ResponseWriter, r *http.Request) {
	var payload types.UpdateBankDetailPayload

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

	err = h.bankDetailStore.UpdateBankDetails(user.ID, payload.BankName, payload.AccountNumber, payload.AccountHolder)
	if err != nil {
		log.Printf("error update bank detail: %v", err)
		logger.WriteServerLog(fmt.Sprintf("error update bank detail: %v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", time.Now().UTC()))
		return
	}

	utils.WriteJSON(w, http.StatusOK, "bank details updated")
}

func (h *Handler) handleGetBankDetail(w http.ResponseWriter, r *http.Request) {
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

	bankDetail, err := h.bankDetailStore.GetBankDataOfUser(user.ID)
	if err != nil {
		log.Printf("error get bank detail: %v", err)
		logger.WriteServerLog(fmt.Sprintf("error get bank detail: %v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", time.Now().UTC()))
		return
	}

	utils.WriteJSON(w, http.StatusOK, bankDetail)
}
