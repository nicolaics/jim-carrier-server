package review

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"

	"github.com/nicolaics/jim-carrier-server/constants"
	"github.com/nicolaics/jim-carrier-server/logger"
	"github.com/nicolaics/jim-carrier-server/types"
	"github.com/nicolaics/jim-carrier-server/utils"
)

type Handler struct {
	reviewStore  types.ReviewStore
	orderStore   types.OrderStore
	listingStore types.ListingStore
	userStore    types.UserStore
}

func NewHandler(reviewStore types.ReviewStore, orderStore types.OrderStore,
	listingStore types.ListingStore, userStore types.UserStore) *Handler {
	return &Handler{
		reviewStore:  reviewStore,
		orderStore:   orderStore,
		listingStore: listingStore,
		userStore:    userStore,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/review", h.handleRegister).Methods(http.MethodPost)
	router.HandleFunc("/review", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)

	// router.HandleFunc("/review/sent", h.handleGetAllSent).Methods(http.MethodGet)
	// router.HandleFunc("/review/sent", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)

	router.HandleFunc("/review/received", h.handleGetAllReceived).Methods(http.MethodPost)
	router.HandleFunc("/review/received", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)

	router.HandleFunc("/review", h.handleDelete).Methods(http.MethodDelete)

	router.HandleFunc("/review", h.handleModify).Methods(http.MethodPatch)
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var payload types.RegisterReviewPayload

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
	reviewer, err := h.userStore.ValidateUserAccessToken(w, r)
	if err != nil {
		log.Printf("token invalid: %v", err)
		logger.WriteServerLog(fmt.Sprintf("token invalid: %v", err))
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("token invalid"))
		return
	}

	reviewer, err = h.userStore.GetUserByID(reviewer.ID)
	if reviewer == nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("account not found"))
		return
	}
	if err != nil {
		log.Println(err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("%v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	order, err := h.orderStore.GetOrderByID(payload.OrderID)
	if err != nil {
		log.Printf("order id %d not found: %v", payload.OrderID, err)
		logger.WriteServerLog(fmt.Sprintf("order id %d not found: %v", payload.OrderID, err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("order not found"))
		return
	}

	reviewee, err := h.userStore.GetUserByName(payload.RevieweeName)
	if err != nil {
		log.Printf("reviewee account not found: %v", err)
		logger.WriteServerLog(fmt.Sprintf("reviewee account not found: %v", err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("reviewee account not found"))
		return
	}

	isDuplicate, err := h.reviewStore.IsReviewDuplicate(reviewer.ID, reviewee.ID, order.ID)
	if err != nil || isDuplicate {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("duplicate review"))
		return
	}

	listing, err := h.listingStore.GetListingByID(order.ListingID)
	if err != nil {
		log.Printf("listing id %d not found: %v", order.ListingID, err)
		logger.WriteServerLog(fmt.Sprintf("listing id %d not found: %v", order.ListingID, err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("listing not found"))
		return
	}

	var reviewType int
	switch reviewer.ID {
	case order.GiverID:
		reviewType = constants.REVIEW_GIVER_TO_CARRIER
	case listing.CarrierID:
		reviewType = constants.REVIEW_CARRIER_TO_GIVER
	default:
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("reviewer id error"))
		return
	}

	err = h.reviewStore.CreateReview(types.Review{
		OrderID:    order.ID,
		ReviewerID: reviewer.ID,
		RevieweeID: reviewee.ID,
		Content:    payload.Content,
		Rating:     payload.Rating,
		ReviewType: reviewType,
	})
	if err != nil {
		log.Printf("error create review: %v", err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("error create review: %v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	utils.WriteJSON(w, http.StatusCreated, "review created")
}

/*
func (h *Handler) handleGetAllSent(w http.ResponseWriter, r *http.Request) {
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
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("account not found"))
		return
	}
	if err != nil {
		log.Println(err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("%v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	reviews, err := h.reviewStore.GetSentReviewsByUserID(user.ID)
	if err != nil {
		log.Println(err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("%v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	utils.WriteJSON(w, http.StatusOK, reviews)
}
*/

func (h *Handler) handleGetAllReceived(w http.ResponseWriter, r *http.Request) {
	var payload types.ReceivedReviewPayload

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
	_, err := h.userStore.ValidateUserAccessToken(w, r)
	if err != nil {
		log.Printf("token invalid: %v", err)
		logger.WriteServerLog(fmt.Sprintf("token invalid: %v", err))
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("token invalid"))
		return
	}

	reviews, err := h.reviewStore.GetReceivedReviewsByUserID(payload.CarrierID)
	if err != nil {
		log.Println(err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("%v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	utils.WriteJSON(w, http.StatusOK, reviews)
}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	var payload types.DeleteReviewPayload

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

	review, err := h.reviewStore.GetReviewByID(payload.ID)
	if review == nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("review not found"))
		return
	}
	if err != nil {
		log.Println(err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("%v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	if user.ID != review.ReviewerID {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("you aren't the reviewer"))
		return
	}

	err = h.reviewStore.DeleteReview(payload.ID)
	if err != nil {
		log.Printf("error delete review: %v", err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("error delete review: %v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	utils.WriteJSON(w, http.StatusOK, "delete review success")
}

func (h *Handler) handleModify(w http.ResponseWriter, r *http.Request) {
	var payload types.ModifyReviewPayload

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

	review, err := h.reviewStore.GetReviewByID(payload.ID)
	if review == nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("review not found"))
		return
	}
	if err != nil {
		log.Println(err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("%v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	if review.ReviewerID != user.ID {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("you are not the reviewier"))
		return
	}

	err = h.reviewStore.ModifyReview(review.ID, payload.Content, payload.Rating)
	if err != nil {
		log.Printf("error modify review: %v", err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("error modify review: %v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	utils.WriteJSON(w, http.StatusOK, "modify success")
}
