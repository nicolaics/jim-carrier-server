package listing

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/nicolaics/jim-carrier/constants"
	"github.com/nicolaics/jim-carrier/logger"
	"github.com/nicolaics/jim-carrier/types"
	"github.com/nicolaics/jim-carrier/utils"
)

type Handler struct {
	listingStore    types.ListingStore
	userStore       types.UserStore
	currencyStore   types.CurrencyStore
	reviewStore     types.ReviewStore
	bankDetailStore types.BankDetailStore
	orderStore types.OrderStore
}

func NewHandler(listingStore types.ListingStore, userStore types.UserStore,
	currencyStore types.CurrencyStore, reviewStore types.ReviewStore,
	bankDetailStore types.BankDetailStore, orderStore types.OrderStore) *Handler {
	return &Handler{
		listingStore:    listingStore,
		userStore:       userStore,
		currencyStore:   currencyStore,
		reviewStore:     reviewStore,
		bankDetailStore: bankDetailStore,
		orderStore: orderStore,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/listing", h.handlePost).Methods(http.MethodPost)
	router.HandleFunc("/listing", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)

	router.HandleFunc("/listing/{reqType}", h.handleGetAll).Methods(http.MethodGet)
	router.HandleFunc("/listing/{reqType}", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)

	router.HandleFunc("/listing/detail", h.handleGetDetail).Methods(http.MethodPost)
	router.HandleFunc("/listing/detail", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)

	router.HandleFunc("/listing", h.handleDelete).Methods(http.MethodDelete)

	router.HandleFunc("/listing", h.handleModify).Methods(http.MethodPatch)

	router.HandleFunc("/listing/package-location", h.handleUpdatePackageLocation).Methods(http.MethodPatch)
	router.HandleFunc("/listing/package-location", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)

	router.HandleFunc("/listing/bank-detail", h.handleGetBankDetail).Methods(http.MethodGet)
	router.HandleFunc("/listing/bank-detail", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
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
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user not found"))
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

	log.Println("original departure date: ", payload.DepartureDate)
	log.Println("parsed departure date: ", departureDate)

	lastReceivedDate, err := utils.ParseDate(payload.LastReceivedDate)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error parsing date"))
		return
	}

	isDuplicate, err := h.listingStore.IsListingDuplicate(carrier.ID, payload.Destination, payload.WeightAvailable, *departureDate)
	if err != nil || isDuplicate {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("listing with the same information exists already"))
		return
	}

	currency, err := h.currencyStore.GetCurrencyByName(payload.Currency)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	if currency == nil {
		err = h.currencyStore.CreateCurrency(payload.Currency)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error create currency: %v", err))
			return
		}

		currency, err = h.currencyStore.GetCurrencyByName(payload.Currency)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}
	}

	err = h.listingStore.CreateListing(types.Listing{
		CarrierID:        carrier.ID,
		Destination:      payload.Destination,
		WeightAvailable:  payload.WeightAvailable,
		PricePerKg:       payload.PricePerKg,
		CurrencyID:       currency.ID,
		DepartureDate:    *departureDate,
		LastReceivedDate: *lastReceivedDate,
		ExpStatus:        constants.EXP_STATUS_AVAILABLE,
		Description:      payload.Description,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error create listing: %v", err))
		return
	}

	err = h.bankDetailStore.UpdateBankDetails(carrier.ID, payload.BankName, payload.AccountNumber, payload.AccountHolder)
	if err != nil {
		logger.WriteServerLog(fmt.Errorf("failed to update bank details: %v", err))
	}

	utils.WriteJSON(w, http.StatusCreated, "listing created")
}

func (h *Handler) handleGetAll(w http.ResponseWriter, r *http.Request) {
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

	vars := mux.Vars(r)
	reqType := vars["reqType"]

	var listings []types.ListingReturnFromDB

	if reqType == "all" {
		listings, err = h.listingStore.GetAllListings()
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}
	} else if reqType == "carrier" {
		listings, err = h.listingStore.GetListingsByCarrierID(user.ID)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}
	} else {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("unknown request parameter"))
		return
	}

	response := make([]types.ListingReturnPayload, 0)

	for _, listing := range listings {
		avgRating, err := h.reviewStore.GetAverageRating(listing.CarrierID, constants.REVIEW_GIVER_TO_CARRIER)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		carrier, err := h.userStore.GetUserByID(listing.CarrierID)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("carrier %d not found", listing.CarrierID))
			return
		}

		imageBytes, err := utils.GetImage(carrier.ProfilePictureURL)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error fetching profile picture for %d: %v", listing.CarrierID, err))
			return
		}

		response = append(response, types.ListingReturnPayload{
			ID:                    listing.ID,
			CarrierID:             listing.CarrierID,
			CarrierName:           listing.CarrierName,
			CarrierProfilePicture: imageBytes,
			Destination:           listing.Destination,
			WeightAvailable:       listing.WeightAvailable,
			PricePerKg:            listing.PricePerKg,
			Currency:              listing.Currency,
			DepartureDate:         listing.DepartureDate,
			LastReceivedDate:      listing.LastReceivedDate,
			Description:           listing.Description.String,
			CarrierRating:         avgRating,
			LastModifiedAt:        listing.LastModifiedAt,
		})
	}

	utils.WriteJSON(w, http.StatusOK, response)
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

	avgRating, err := h.reviewStore.GetAverageRating(listing.CarrierID, constants.REVIEW_GIVER_TO_CARRIER)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	listing.CarrierRating = avgRating

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

	newLastReceivedDate, err := utils.ParseDate(payload.NewData.LastReceivedDate)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error parsing date"))
		return
	}

	currency, err := h.currencyStore.GetCurrencyByName(payload.NewData.Currency)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	if currency == nil {
		err = h.currencyStore.CreateCurrency(payload.NewData.Currency)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error create currency: %v", err))
			return
		}

		currency, err = h.currencyStore.GetCurrencyByName(payload.NewData.Currency)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}
	}

	err = h.listingStore.ModifyListing(listing.ID, types.Listing{
		Destination:      payload.NewData.Destination,
		WeightAvailable:  payload.NewData.WeightAvailable,
		PricePerKg:       payload.NewData.PricePerKg,
		CurrencyID:       currency.ID,
		DepartureDate:    *newDepartureDate,
		LastReceivedDate: *newLastReceivedDate,
		ExpStatus:        constants.EXP_STATUS_AVAILABLE,
		Description:      payload.NewData.Description,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error modify listing: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, "modify success")
}

func (h *Handler) handleGetBankDetail(w http.ResponseWriter, r *http.Request) {
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

	bankDetail, err := h.bankDetailStore.GetBankDataOfUser(user.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, bankDetail)
}

func (h *Handler) handleUpdatePackageLocation(w http.ResponseWriter, r *http.Request) {
	var payload types.UpdateBulkPackageLocationPayload

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

	orders, err := h.orderStore.GetOrdersByListingID(listing.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	subject := "Your Package Location is Updated"

	for _, order := range(orders) {
		body := fmt.Sprintf("<h4>Your package for order number %d has an update!</h4><h4>It status now is</h4><br><h2>%s</h2>", 
								order.ID, payload.PackageLocation)
		err = utils.SendEmail(order.GiverEmail, subject, body, "", "")
		if err != nil {
			logger.WriteServerLog(fmt.Errorf("error sending email to %s for updating package location: %v", order.GiverEmail, err))
		}
	}

	utils.WriteJSON(w, http.StatusOK, "package location updated")
}
