package review

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	// "github.com/nicolaics/jim-carrier/constants"
	"github.com/nicolaics/jim-carrier/types"
	"github.com/nicolaics/jim-carrier/utils"
)

type Handler struct {
	reviewStore types.ReviewStore
	orderStore types.OrderStore
	userStore  types.UserStore
}

func NewHandler(reviewStore types.ReviewStore, orderStore types.OrderStore, userStore types.UserStore) *Handler {
	return &Handler{
		reviewStore: reviewStore,
		orderStore: orderStore,
		userStore:  userStore,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/review", h.handleRegister).Methods(http.MethodPost)
	router.HandleFunc("/review", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)

	// router.HandleFunc("/review/{reqType}", h.handleGetAll).Methods(http.MethodGet)
	// router.HandleFunc("/review/{reqType}", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)

	// router.HandleFunc("/review/{reqType}/detail", h.handleGetDetail).Methods(http.MethodPost)
	// router.HandleFunc("/review/{reqType}/detail", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)

	// router.HandleFunc("/review", h.handleDelete).Methods(http.MethodDelete)

	// router.HandleFunc("/review", h.handleModify).Methods(http.MethodPatch)
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var payload types.RegisterReviewPayload

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
	user, err := h.userStore.ValidateUserToken(w, r)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("invalid token: %v", err))
		return
	}

	user, err = h.userStore.GetUserByID(user.ID)
	if user == nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("account not found"))
		return
	}
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	order, err := h.orderStore.GetOrderByID(payload.OrderID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("order id %d not found: %v", payload.OrderID, err))
		return
	}

	isDuplicate, err := h.reviewStore.IsReviewDuplicate(user.ID, order.ListingID)
	if err != nil || isDuplicate {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("duplicate review"))
		return
	}

	err = h.reviewStore.CreateReview(types.Review{
		ListingID: order.ListingID,
		ReviewerID: user.ID,
		Content: payload.Content,
		Rating: payload.Rating,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error create review: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusCreated, "review created")
}

