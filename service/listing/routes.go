package listing

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/nicolaics/jim-carrier/constants"
	"github.com/nicolaics/jim-carrier/types"
	"github.com/nicolaics/jim-carrier/utils"
)

type Handler struct {
	listingStore types.ListingStore
	userStore    types.UserStore
}

func NewHandler(listingStore types.ListingStore, userStore types.UserStore) *Handler {
	return &Handler{
		listingStore: listingStore,
		userStore:    userStore,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/listing", h.handlePost).Methods(http.MethodPost)
	router.HandleFunc("/listing", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)

	router.HandleFunc("/listing", h.handleGetAll).Methods(http.MethodGet)

	router.HandleFunc("/listing/detail", h.handleGetDetail).Methods(http.MethodPost)
	router.HandleFunc("/listing/detail", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)

	router.HandleFunc("/listing", h.handleDelete).Methods(http.MethodDelete)

	router.HandleFunc("/listing", h.handleModify).Methods(http.MethodPatch)
}

func (h *Handler) handlePost(w http.ResponseWriter, r *http.Request) {
	var payload types.PostListingPayload

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
	carrier, err := h.userStore.ValidateUserToken(w, r)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("invalid token: %v", err))
		return
	}

	carrier, err = h.userStore.GetUserByID(carrier.ID)
	if carrier == nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("account not found"))
		return
	}
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	departureDate, err := utils.ParseDate(payload.DepartureDate)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error parsing date"))
		return
	}

	isDuplicate, err := h.listingStore.IsListingDuplicate(carrier.ID, payload.Destination, payload.WeightAvailable, *departureDate)
	if err != nil || isDuplicate {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("duplicate listing"))
		return
	}

	err = h.listingStore.CreateListing(types.Listing{
		CarrierID:       carrier.ID,
		Destination:     payload.Destination,
		WeightAvailable: payload.WeightAvailable,
		PricePerKg:      payload.PricePerKg,
		DepartureDate:   *departureDate,
		ExpStatus:       constants.EXP_STATUS_AVAILABLE,
		Description:     payload.Description,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error create listing: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusCreated, "listing created")
}

func (h *Handler) handleGetAll(w http.ResponseWriter, r *http.Request) {
	// validate token
	_, err := h.userStore.ValidateUserToken(w, r)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("invalid token: %v", err))
		return
	}

	listings, err := h.listingStore.GetAllListings()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, listings)
}

func (h *Handler) handleGetDetail(w http.ResponseWriter, r *http.Request) {
	var payload types.GetListingDetailPayload

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
	_, err := h.userStore.ValidateUserToken(w, r)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("invalid token: %v", err))
		return
	}

	listing, err := h.listingStore.GetListingByID(payload.ID)
	if listing == nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("listing not found"))
		return
	}
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, listing)
}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	var payload types.DeleteListingPayload

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

	listing, err := h.listingStore.GetListingByID(payload.ID)
	if listing == nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("listing not found"))
		return
	}
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	if listing.CarrierID != user.ID {
		utils.WriteError(w, http.StatusForbidden, fmt.Errorf("you are not the owner of the listing"))
		return
	}

	err = h.listingStore.DeleteListing(payload.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error delete listing: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, "delete listing success")
}

func (h *Handler) handleModify(w http.ResponseWriter, r *http.Request) {
	var payload types.ModifyListingPayload

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

	listing, err := h.listingStore.GetListingByID(payload.ID)
	if listing == nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("listing not found"))
		return
	}
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	if listing.CarrierID != user.ID {
		utils.WriteError(w, http.StatusForbidden, fmt.Errorf("you are not the owner of the listing"))
		return
	}

	newDepartureDate, err := utils.ParseDate(payload.NewData.DepartureDate)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error parsing date"))
		return
	}

	err = h.listingStore.ModifyListing(listing.ID, types.Listing{
		Destination:     payload.NewData.Destination,
		WeightAvailable: payload.NewData.WeightAvailable,
		PricePerKg:      payload.NewData.PricePerKg,
		DepartureDate:   *newDepartureDate,
		ExpStatus:       constants.EXP_STATUS_AVAILABLE,
		Description:     payload.NewData.Description,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error modify listing: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, "modify success")
}
