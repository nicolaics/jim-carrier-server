package listing

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/nicolaics/jim-carrier-server/constants"
	"github.com/nicolaics/jim-carrier-server/logger"
	"github.com/nicolaics/jim-carrier-server/types"
	"github.com/nicolaics/jim-carrier-server/utils"
)

type Handler struct {
	listingStore    types.ListingStore
	userStore       types.UserStore
	currencyStore   types.CurrencyStore
	reviewStore     types.ReviewStore
	bankDetailStore types.BankDetailStore
	orderStore      types.OrderStore
	fcmHistoryStore types.FCMHistoryStore
}

func NewHandler(listingStore types.ListingStore, userStore types.UserStore,
	currencyStore types.CurrencyStore, reviewStore types.ReviewStore,
	bankDetailStore types.BankDetailStore, orderStore types.OrderStore,
	fcmHistoryStore types.FCMHistoryStore) *Handler {
	return &Handler{
		listingStore:    listingStore,
		userStore:       userStore,
		currencyStore:   currencyStore,
		reviewStore:     reviewStore,
		bankDetailStore: bankDetailStore,
		orderStore:      orderStore,
		fcmHistoryStore: fcmHistoryStore,
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

	router.HandleFunc("/listing/count-orders", h.handleCountOrdersForOneListing).Methods(http.MethodPost)
	router.HandleFunc("/listing/count-orders", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
}

func (h *Handler) handlePost(w http.ResponseWriter, r *http.Request) {
	var payload types.PostListingPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		log.Printf("payload error: %v", err)
		logger.WriteServerLog(fmt.Sprintf("post listing payload error: %v", err))
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
	carrier, err := h.userStore.ValidateUserAccessToken(w, r)
	if err != nil {
		log.Printf("invalid token: %v", err)
		logger.WriteServerLog(fmt.Sprintf("post listing invalid token: %v", err))
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("token invalid"))
		return
	}

	carrier, err = h.userStore.GetUserByID(carrier.ID)
	if carrier == nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user not found"))
		return
	}
	if err != nil {
		log.Printf("get user at listing: %v", err)
		logger.WriteServerLog(fmt.Sprintf("get user at listing: %v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", time.Now().UTC()))
		return
	}

	departureDate, err := utils.ParseDate(payload.DepartureDate)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error parsing date"))
		return
	}

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
		log.Println(err)
		logger.WriteServerLog(fmt.Sprintf("%v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", time.Now().UTC()))
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
			log.Printf("get currency at listing: %v", err)
			logger.WriteServerLog(fmt.Sprintf("get currency at listing: %v", err))
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", time.Now().UTC()))
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
		log.Printf("error create listing: %v", err)
		logger.WriteServerLog(fmt.Sprintf("error create listing: %v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to create new listing"))
		return
	}

	err = h.bankDetailStore.UpdateBankDetails(carrier.ID, payload.BankName, payload.AccountNumber, payload.AccountHolder)
	if err != nil {
		logger.WriteServerLog(fmt.Sprintf("failed to update bank details: %v", err))
	}

	utils.WriteJSON(w, http.StatusCreated, "listing successfully created")
}

func (h *Handler) handleGetAll(w http.ResponseWriter, r *http.Request) {
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
		logger.WriteServerLog(fmt.Sprintf("%v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", time.Now().UTC()))
		return
	}

	vars := mux.Vars(r)
	reqType := vars["reqType"]

	var listings []types.ListingReturnFromDB

	if reqType == "all" {
		listings, err = h.listingStore.GetAllListings(user.ID)
		if err != nil {
			log.Println(err)
			logger.WriteServerLog(fmt.Sprintf("%v", err))
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", time.Now().UTC()))
			return
		}
	} else if reqType == "carrier" {
		listings, err = h.listingStore.GetListingsByCarrierID(user.ID)
		if err != nil {
			log.Println(err)
			logger.WriteServerLog(fmt.Sprintf("%v", err))
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", time.Now().UTC()))
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
			log.Println(err)
			logger.WriteServerLog(fmt.Sprintf("%v", err))
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", time.Now().UTC()))
			return
		}

		carrier, err := h.userStore.GetUserByID(listing.CarrierID)
		if err != nil {
			log.Printf("carrier %d not found: %v", listing.CarrierID, err)
			logger.WriteServerLog(fmt.Sprintf("carrier %d not found: %v", listing.CarrierID, err))
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", time.Now().UTC()))
			return
		}

		imageBytes, err := utils.GetImage(carrier.ProfilePictureURL)
		if err != nil {
			log.Printf("error fetching profile picture for %d: %v", listing.CarrierID, err)
			logger.WriteServerLog(fmt.Sprintf("error fetching profile picture for %d: %v", listing.CarrierID, err))
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", time.Now().UTC()))
			return
		}

		bankDetail, err := h.bankDetailStore.GetBankDataOfUser(carrier.ID)
		if err != nil {
			log.Printf("error fetching bank data for %d: %v", listing.CarrierID, err)
			logger.WriteServerLog(fmt.Sprintf("error fetching bank data for %d: %v", listing.CarrierID, err))
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", time.Now().UTC()))
			return
		}

		response = append(response, types.ListingReturnPayload{
			ID:                    listing.ID,
			CarrierID:             listing.CarrierID,
			CarrierName:           listing.CarrierName,
			CarrierEmail:          listing.CarrierEmail,
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
			BankDetail:            *bankDetail,
		})
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) handleGetDetail(w http.ResponseWriter, r *http.Request) {
	var payload types.GetListingDetailPayload

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

	listing, err := h.listingStore.GetListingByID(payload.ID)
	if listing == nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("listing not found"))
		return
	}
	if err != nil {
		log.Println(err)
		logger.WriteServerLog(fmt.Sprintf("%v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", time.Now().UTC()))
		return
	}

	avgRating, err := h.reviewStore.GetAverageRating(listing.CarrierID, constants.REVIEW_GIVER_TO_CARRIER)
	if err != nil {
		log.Println(err)
		logger.WriteServerLog(fmt.Sprintf("%v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", time.Now().UTC()))
		return
	}

	listing.CarrierRating = avgRating

	utils.WriteJSON(w, http.StatusOK, listing)
}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	var payload types.DeleteListingPayload

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

	listing, err := h.listingStore.GetListingByID(payload.ID)
	if listing == nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("listing not found"))
		return
	}
	if err != nil {
		log.Println(err)
		logger.WriteServerLog(fmt.Sprintf("%v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", time.Now().UTC()))
		return
	}

	if listing.CarrierID != user.ID {
		utils.WriteError(w, http.StatusForbidden, fmt.Errorf("you are not the owner of the listing"))
		return
	}

	orderCount, err := h.orderStore.GetOrderCountByListingID(listing.ID)
	if err != nil {
		log.Println(err)
		logger.WriteServerLog(fmt.Sprintf("%v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", time.Now().UTC()))
		return
	}

	if orderCount > 0 {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("there are active orders"))
		return
	}

	err = h.listingStore.DeleteListing(payload.ID)
	if err != nil {
		log.Printf("error delete listing: %v", err)
		logger.WriteServerLog(fmt.Sprintf("error delete listing: %v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error delete listing: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, "delete listing success")
}

func (h *Handler) handleModify(w http.ResponseWriter, r *http.Request) {
	var payload types.ModifyListingPayload

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

	listing, err := h.listingStore.GetListingByID(payload.ID)
	if listing == nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("listing not found"))
		return
	}
	if err != nil {
		log.Println(err)
		logger.WriteServerLog(fmt.Sprintf("%v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", time.Now().UTC()))
		return
	}

	if listing.CarrierID != user.ID {
		utils.WriteError(w, http.StatusForbidden, fmt.Errorf("you are not the owner of the listing"))
		return
	}

	newDepartureDate, err := utils.ParseDate(payload.DepartureDate)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error parsing date"))
		return
	}

	newLastReceivedDate, err := utils.ParseDate(payload.LastReceivedDate)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error parsing date"))
		return
	}

	currency, err := h.currencyStore.GetCurrencyByName(payload.Currency)
	if err != nil {
		log.Println(err)
		logger.WriteServerLog(fmt.Sprintf("%v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", time.Now().UTC()))
		return
	}

	if currency == nil {
		err = h.currencyStore.CreateCurrency(payload.Currency)
		if err != nil {
			log.Printf("error create currency: %v", err)
			logger.WriteServerLog(fmt.Sprintf("error create currency: %v", err))
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", time.Now().UTC()))
			return
		}

		currency, err = h.currencyStore.GetCurrencyByName(payload.Currency)
		if err != nil {
			log.Println(err)
			logger.WriteServerLog(fmt.Sprintf("%v", err))
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", time.Now().UTC()))
			return
		}
	}

	err = h.bankDetailStore.UpdateBankDetails(user.ID, payload.BankName, payload.AccountNumber, payload.AccountHolder)
	if err != nil {
		logger.WriteServerLog(fmt.Sprintf("error update bank details for user %s:\n%v", user.Email, err))
	}

	err = h.listingStore.ModifyListing(listing.ID, types.Listing{
		Destination:      payload.Destination,
		WeightAvailable:  payload.WeightAvailable,
		PricePerKg:       payload.PricePerKg,
		CurrencyID:       currency.ID,
		DepartureDate:    *newDepartureDate,
		LastReceivedDate: *newLastReceivedDate,
		ExpStatus:        constants.EXP_STATUS_AVAILABLE,
		Description:      payload.Description,
	})
	if err != nil {
		log.Printf("error modify listing: %v", err)
		logger.WriteServerLog(fmt.Sprintf("error modify listing: %v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("modify listing failed"))
		return
	}

	utils.WriteJSON(w, http.StatusOK, "modify success")
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
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("account not found"))
		return
	}
	if err != nil {
		log.Println(err)
		logger.WriteServerLog(fmt.Sprintf("%v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", time.Now().UTC()))
		return
	}

	bankDetail, err := h.bankDetailStore.GetBankDataOfUser(user.ID)
	if err != nil {
		log.Printf("error fetching bank data: %v", err)
		logger.WriteServerLog(fmt.Sprintf("error fetching bank data: %v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", time.Now().UTC()))
		return
	}

	utils.WriteJSON(w, http.StatusOK, bankDetail)
}

func (h *Handler) handleUpdatePackageLocation(w http.ResponseWriter, r *http.Request) {
	var payload types.UpdateBulkPackageLocationPayload

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

	listing, err := h.listingStore.GetListingByID(payload.ID)
	if listing == nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("listing not found"))
		return
	}
	if err != nil {
		log.Println(err)
		logger.WriteServerLog(fmt.Sprintf("%v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", time.Now().UTC()))
		return
	}

	if listing.CarrierID != user.ID {
		utils.WriteError(w, http.StatusForbidden, fmt.Errorf("you are not the owner of the listing"))
		return
	}

	orders, err := h.orderStore.GetOrdersByListingID(listing.ID)
	if err != nil {
		log.Println(err)
		logger.WriteServerLog(fmt.Sprintf("%v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", time.Now().UTC()))
		return
	}

	subject := "Your Package Location is Updated"

	for _, order := range orders {
		giver, err := h.userStore.GetUserByEmail(order.GiverEmail)
		if err != nil {
			logger.WriteServerLog(fmt.Sprintf("error finding account of %s for updating package location: %v", order.GiverEmail, err))
			continue
		}

		body := fmt.Sprintf("<h4>Your package for order no. %d has an update!</h4><h4>It status now is</h4><br><h2>%s</h2>",
			order.ID, payload.PackageLocation)
		err = utils.SendEmail(giver.Email, subject, body, "", "")
		if err != nil {
			logger.WriteServerLog(fmt.Sprintf("error sending email to %s for updating package location: %v", order.GiverEmail, err))
		}

		fcmHistory := types.FCMHistory{
			ToUserID: giver.ID,
			ToToken:  giver.FCMToken,
			Data: types.FCMData{
				Type:    "order_updated",
				OrderID: fmt.Sprintf("%d", order.ID),
			},
			Title: subject,
			Body:  fmt.Sprintf("Package location for order no. %d is %s", order.ID, payload.PackageLocation),
		}

		fcmHistory.Response, err = utils.SendFCMToOne(fcmHistory)
		if err != nil {
			logger.WriteServerLog(fmt.Sprintf("error sending notification to carrier: %v", err))
		} else {
			err = h.fcmHistoryStore.CreateFCMHistory(fcmHistory)
			if err != nil {
				logger.WriteServerLog(fmt.Sprintf("error update fcm history: %v", err))
			}
		}
	}

	utils.WriteJSON(w, http.StatusOK, "package location updated")
}

func (h *Handler) handleCountOrdersForOneListing(w http.ResponseWriter, r *http.Request) {
	var payload types.GetListingDetailPayload

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

	listing, err := h.listingStore.GetListingByID(payload.ID)
	if err != nil {
		log.Println(err)
		logger.WriteServerLog(fmt.Sprintf("%v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", time.Now().UTC()))
		return
	}

	if listing.CarrierID != user.ID {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("you are not the owner of the listing"))
		return
	}

	orderCount, err := h.orderStore.GetOrderCountByListingID(listing.ID)
	if err != nil {
		log.Println(err)
		logger.WriteServerLog(fmt.Sprintf("%v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", time.Now().UTC()))
		return
	}

	if orderCount == 0 {
		utils.WriteJSON(w, http.StatusOK, "modify allowed")
	} else {
		utils.WriteError(w, http.StatusForbidden, fmt.Errorf("there are orders for this listing"))
	}
}
