package order

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/nicolaics/jim-carrier-server/constants"
	"github.com/nicolaics/jim-carrier-server/logger"
	"github.com/nicolaics/jim-carrier-server/types"
	"github.com/nicolaics/jim-carrier-server/utils"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Handler struct {
	orderStore      types.OrderStore
	userStore       types.UserStore
	listingStore    types.ListingStore
	currencyStore   types.CurrencyStore
	fcmHistoryStore types.FCMHistoryStore
	bankDetailStore types.BankDetailStore
	bucket          *s3.S3
}

func NewHandler(orderStore types.OrderStore, userStore types.UserStore,
	listingStore types.ListingStore, currencyStore types.CurrencyStore,
	fcmHistoryStore types.FCMHistoryStore, bankDetailStore types.BankDetailStore, bucket *s3.S3) *Handler {
	return &Handler{
		orderStore:      orderStore,
		userStore:       userStore,
		listingStore:    listingStore,
		currencyStore:   currencyStore,
		fcmHistoryStore: fcmHistoryStore,
		bankDetailStore: bankDetailStore,
		bucket:          bucket,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/order", h.handleRegister).Methods(http.MethodPost)
	router.HandleFunc("/order", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)

	router.HandleFunc("/order/{reqType}", h.handleGetAll).Methods(http.MethodGet)
	router.HandleFunc("/order/{reqType}", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)

	router.HandleFunc("/order/{reqType}/detail", h.handleGetDetail).Methods(http.MethodPost)
	router.HandleFunc("/order/{reqType}/detail", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)

	router.HandleFunc("/order", h.handleDelete).Methods(http.MethodDelete)

	router.HandleFunc("/order/{reqType}", h.handleModify).Methods(http.MethodPatch)

	router.HandleFunc("/order/get-payment-details", h.handleGetPaymentDetails).Methods(http.MethodPost)
	router.HandleFunc("/order/get-payment-details", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var payload types.RegisterOrderPayload

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
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("account not found"))
		return
	}
	if err != nil {
		log.Println(err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("%v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	listing, err := h.listingStore.GetListingByID(payload.ListingID)
	if err != nil {
		log.Printf("listing id %d not found: %v", payload.ListingID, err)
		logger.WriteServerLog(fmt.Sprintf("listing id %d not found: %v", payload.ListingID, err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("listing not found"))
		return
	}

	carrier, err := h.userStore.GetUserByID(listing.CarrierID)
	if carrier == nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("carrier not found"))
		return
	}
	if err != nil {
		log.Println(err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("%v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	isDuplicate, err := h.orderStore.IsOrderDuplicate(user.ID, listing.ID)
	if isDuplicate {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("duplicate order"))
		return
	}
	if err != nil {
		log.Println(err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("%v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	if payload.Weight > listing.WeightAvailable {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("ordered weight is greater than available weight"))
		return
	}

	currency, err := h.currencyStore.GetCurrencyByName(payload.Currency)
	if err != nil {
		log.Println(err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("%v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	if currency == nil {
		err = h.currencyStore.CreateCurrency(payload.Currency)
		if err != nil {
			log.Printf("error create currency: %v", err)
			logFile, _ := logger.WriteServerLog(fmt.Sprintf("error create currency: %v", err))
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
			return
		}

		currency, err = h.currencyStore.GetCurrencyByName(payload.Currency)
		if err != nil {
			log.Println(err)
			logFile, _ := logger.WriteServerLog(fmt.Sprintf("%v", err))
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
			return
		}
	}

	var packageImgURL string

	if len(payload.PackageImage) > constants.PACKAGE_IMG_MAX_BYTES {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("image size exceeds the limit of 10MB"))
		return
	} else if len(payload.PackageImage) > 0 {
		var imageExtension string

		// check image type
		mimeType := http.DetectContentType(payload.PackageImage)
		switch mimeType {
		case "image/jpeg":
			imageExtension = ".jpg"
		case "image/png":
			imageExtension = ".png"
		default:
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("unsupported image type"))
			return
		}

		filePath := constants.PACKAGE_IMG_DIR_PATH + utils.GeneratePictureFileName(imageExtension)

		isPackageImageURLExist := h.orderStore.IsPackageImageURLExist(filePath)

		for isPackageImageURLExist {
			filePath = constants.PACKAGE_IMG_DIR_PATH + utils.GeneratePictureFileName(imageExtension)
			isPackageImageURLExist = h.orderStore.IsPackageImageURLExist(filePath)
		}

		// save the image
		err := utils.SavePackageImage(payload.PackageImage, filePath)
		if err != nil {
			logger.WriteServerLog(fmt.Sprintf("error saving package image: %v", err))
		}

		packageImgURL = filePath
	}

	err = h.orderStore.CreateOrder(types.Order{
		ListingID:       listing.ID,
		GiverID:         user.ID,
		Weight:          payload.Weight,
		Price:           payload.Price,
		CurrencyID:      currency.ID,
		PackageContent:  payload.PackageContent,
		PackageImageURL: packageImgURL,
		Notes:           payload.Notes,
	})
	if err != nil {
		log.Printf("error create order: %v", err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("error create order: %v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	orderId, err := h.orderStore.GetOrderID(types.Order{
		ListingID:       listing.ID,
		GiverID:         user.ID,
		Weight:          payload.Weight,
		Price:           payload.Price,
		CurrencyID:      currency.ID,
		PackageContent:  payload.PackageContent,
		PackageImageURL: packageImgURL,
		Notes:           payload.Notes,
	})
	if err != nil {
		log.Printf("error finding order ID: %v", err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("error finding order ID: %v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	subject := "New Order Arrived!"

	body := utils.CreateEmailBodyOfOrder(subject, user.Name, listing.Destination, currency.Name, payload.Notes, payload.PackageContent, payload.Weight, payload.Price)

	err = utils.SendEmail(carrier.Email, subject, body, packageImgURL, "Package Image")
	if err != nil {
		logger.WriteServerLog(fmt.Sprintf("error send email to carrier: %v", err))
	}

	fcmHistory := types.FCMHistory{
		ToUserID: carrier.ID,
		ToToken:  carrier.FCMToken,
		Data: types.FCMData{
			Type:    "confirm_order",
			OrderID: fmt.Sprintf("%d", orderId),
		},
		Title: subject,
		Body:  fmt.Sprintf("Confirm order no. %d before %s 23:59 KST (GMT +09)", orderId, time.Now().Local().AddDate(0, 0, 2).Format("02 Jan 2006")),
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

	utils.WriteJSON(w, http.StatusCreated, "order created")
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
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("%v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	vars := mux.Vars(r)
	reqType := vars["reqType"]

	var ordersReturn interface{}
	if reqType == "carrier" {
		orders, err := h.orderStore.GetOrdersByCarrierID(user.ID)
		if err != nil {
			log.Println(err)
			logFile, _ := logger.WriteServerLog(fmt.Sprintf("%v", err))
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
			return
		}

		ordersReturnTemp := make([]types.OrderCarrierReturnPayload, 0)

		for _, order := range orders {
			paymentStatus := utils.PaymentStatusIntToString(order.PaymentStatus)
			orderStatus := utils.OrderStatusIntToString(order.OrderStatus)

			packageImage, err := utils.GetImage(order.PackageImageURL, h.bucket)
			if err != nil {
				log.Printf("error reading the picture: %v", err)
				logFile, _ := logger.WriteServerLog(fmt.Sprintf("error reading the picture: %v", err))
				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
				return
			}

			temp := types.OrderCarrierReturnPayload{
				Listing:          order.Listing,
				ID:               order.ID,
				GiverName:        order.GiverName,
				GiverPhoneNumber: order.GiverPhoneNumber,
				GiverEmail:       order.GiverEmail,
				Weight:           order.Weight,
				Price:            order.Price,
				Currency:         order.Currency,
				PackageContent:   order.PackageContent,
				PackageImage:     packageImage,
				PaymentStatus:    paymentStatus,
				PaidAt:           order.PaidAt.Time,
				PaymentProofURL:  order.PaymentProofURL.String,
				OrderStatus:      orderStatus,
				PackageLocation:  order.PackageLocation,
				Notes:            order.Notes.String,
				CreatedAt:        order.CreatedAt,
				LastModifiedAt:   order.LastModifiedAt,
			}
			ordersReturnTemp = append(ordersReturnTemp, temp)
		}

		ordersReturn = ordersReturnTemp
	} else if reqType == "giver" {
		orders, err := h.orderStore.GetOrdersByGiverID(user.ID)
		if err != nil {
			log.Println(err)
			logFile, _ := logger.WriteServerLog(fmt.Sprintf("%v", err))
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
			return
		}

		ordersReturnTemp := make([]types.OrderGiverReturnPayload, 0)
		for _, order := range orders {
			paymentStatus := utils.PaymentStatusIntToString(order.PaymentStatus)

			orderStatus := utils.OrderStatusIntToString(order.OrderStatus)

			// packageImage, err := utils.GetImage(order.PackageImageURL)
			// if err != nil {
			// 	log.Printf("error reading the picture: %v", err)
			// 	logFile, _ := logger.WriteServerLog(fmt.Sprintf("error reading the picture: %v", err))
			// 	utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
			// 	return
			// }

			temp := types.OrderGiverReturnPayload{
				Listing:        order.Listing,
				ID:             order.ID,
				Weight:         order.Weight,
				Price:          order.Price,
				Currency:       order.Currency,
				PackageContent: order.PackageContent,
				// PackageImage:    packageImage,
				PaymentStatus:   paymentStatus,
				PaidAt:          order.PaidAt.Time,
				PaymentProofURL: order.PaymentProofURL.String,
				OrderStatus:     orderStatus,
				PackageLocation: order.PackageLocation,
				Notes:           order.Notes.String,
				CreatedAt:       order.CreatedAt,
				LastModifiedAt:  order.LastModifiedAt,
			}
			ordersReturnTemp = append(ordersReturnTemp, temp)
		}

		ordersReturn = ordersReturnTemp

	} else {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("unknown request parameter"))
		return
	}

	utils.WriteJSON(w, http.StatusOK, ordersReturn)
}

func (h *Handler) handleGetDetail(w http.ResponseWriter, r *http.Request) {
	var payload types.ViewOrderDetailPayload

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
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("account not found"))
		return
	}
	if err != nil {
		log.Println(err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("%v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	vars := mux.Vars(r)
	reqType := vars["reqType"]

	var returnOrder interface{}

	if reqType == "carrier" {
		order, err := h.orderStore.GetCarrierOrderByID(payload.ID, user.ID)
		if err != nil {
			log.Println(err)
			logFile, _ := logger.WriteServerLog(fmt.Sprintf("%v", err))
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
			return
		}

		paymentStatus := utils.PaymentStatusIntToString(order.PaymentStatus)

		orderStatus := utils.OrderStatusIntToString(order.OrderStatus)

		packageImage, err := utils.GetImage(order.PackageImageURL, h.bucket)
		if err != nil {
			log.Printf("error reading the picture: %v", err)
			logFile, _ := logger.WriteServerLog(fmt.Sprintf("error reading the picture: %v", err))
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
			return
		}

		returnOrder = types.OrderCarrierReturnPayload{
			Listing:          order.Listing,
			ID:               order.ID,
			GiverName:        order.GiverName,
			GiverPhoneNumber: order.GiverPhoneNumber,
			Weight:           order.Weight,
			Price:            order.Price,
			Currency:         order.Currency,
			PackageContent:   order.PackageContent,
			PackageImage:     packageImage,
			PaymentStatus:    paymentStatus,
			PaidAt:           order.PaidAt.Time,
			PaymentProofURL:  order.PaymentProofURL.String,
			OrderStatus:      orderStatus,
			PackageLocation:  order.PackageLocation,
			Notes:            order.Notes.String,
			CreatedAt:        order.CreatedAt,
			LastModifiedAt:   order.LastModifiedAt,
		}
	} else if reqType == "giver" {
		order, err := h.orderStore.GetGiverOrderByID(payload.ID, user.ID)
		if err != nil {
			log.Println(err)
			logFile, _ := logger.WriteServerLog(fmt.Sprintf("%v", err))
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
			return
		}

		paymentStatus := utils.PaymentStatusIntToString(order.PaymentStatus)

		orderStatus := utils.OrderStatusIntToString(order.OrderStatus)

		packageImage, err := utils.GetImage(order.PackageImageURL, h.bucket)
		if err != nil {
			log.Printf("error reading the picture: %v", err)
			logFile, _ := logger.WriteServerLog(fmt.Sprintf("error reading the picture: %v", err))
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
			return
		}

		returnOrder = types.OrderGiverReturnPayload{
			Listing:         order.Listing,
			ID:              order.ID,
			Weight:          order.Weight,
			Price:           order.Price,
			Currency:        order.Currency,
			PackageContent:  order.PackageContent,
			PackageImage:    packageImage,
			PaymentStatus:   paymentStatus,
			PaidAt:          order.PaidAt.Time,
			PaymentProofURL: order.PaymentProofURL.String,
			OrderStatus:     orderStatus,
			PackageLocation: order.PackageLocation,
			Notes:           order.Notes.String,
			CreatedAt:       order.CreatedAt,
			LastModifiedAt:  order.LastModifiedAt,
		}
	} else {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("unknown request parameter"))
		return
	}

	utils.WriteJSON(w, http.StatusOK, returnOrder)
}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	var payload types.DeleteOrderPayload

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
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("account not found"))
		return
	}
	if err != nil {
		log.Println(err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("%v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	err = h.orderStore.DeleteOrder(payload.ID, user.ID)
	if err != nil {
		log.Printf("error deleting order: %v", err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("error deleting order: %v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	utils.WriteJSON(w, http.StatusOK, "delete order success")
}

func (h *Handler) handleModify(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	reqType := vars["reqType"]

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

	returnMsg := ""

	if reqType == "all" {
		var payload types.ModifyOrderPayload

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

		order, err := h.orderStore.GetOrderByID(payload.ID)
		if err != nil {
			log.Printf("order not found: %v", err)
			logger.WriteServerLog(fmt.Sprintf("order not found: %v", err))
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("order not found"))
			return
		}

		listing, err := h.listingStore.GetListingByID(payload.ListingID)
		if err != nil {
			log.Printf("listing id %d not found: %v", payload.ListingID, err)
			logger.WriteServerLog(fmt.Sprintf("listing id %d not found: %v", payload.ListingID, err))
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("listing not found"))
			return
		}

		carrier, err := h.userStore.GetUserByID(listing.CarrierID)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("carrier id %d not found: %v", listing.CarrierID, err))
			return
		}

		paymentStatus := utils.PaymentStatusStringToInt(payload.PaymentStatus)
		if paymentStatus == -1 {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("unknown payment status"))
			return
		}

		currency, err := h.currencyStore.GetCurrencyByName(payload.Currency)
		if err != nil {
			log.Println(err)
			logFile, _ := logger.WriteServerLog(fmt.Sprintf("%v", err))
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
			return
		}

		if currency == nil {
			err = h.currencyStore.CreateCurrency(payload.Currency)
			if err != nil {
				log.Printf("error create currency: %v", err)
				logFile, _ := logger.WriteServerLog(fmt.Sprintf("error create currency: %v", err))
				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
				return
			}

			currency, err = h.currencyStore.GetCurrencyByName(payload.Currency)
			if err != nil {
				log.Println(err)
				logFile, _ := logger.WriteServerLog(fmt.Sprintf("%v", err))
				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
				return
			}
		}

		packageImgURL := order.PackageImageURL

		if order.PackageContent != payload.PackageContent {
			if len(payload.PackageImage) > constants.PACKAGE_IMG_MAX_BYTES {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("image size exceeds the limit of 10MB"))
				return
			} else if len(payload.PackageImage) > 0 {
				var imageExtension string

				// check image type
				mimeType := http.DetectContentType(payload.PackageImage)
				switch mimeType {
				case "image/jpeg":
					imageExtension = ".jpg"
				case "image/png":
					imageExtension = ".png"
				default:
					utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("unsupported image type"))
					return
				}

				filePath := constants.PACKAGE_IMG_DIR_PATH + utils.GeneratePictureFileName(imageExtension)

				isPackageImageURLExist := h.orderStore.IsPackageImageURLExist(filePath)

				for isPackageImageURLExist {
					filePath = constants.PACKAGE_IMG_DIR_PATH + utils.GeneratePictureFileName(imageExtension)
					isPackageImageURLExist = h.orderStore.IsPackageImageURLExist(filePath)
				}

				// save the image
				err := utils.SavePackageImage(payload.PackageImage, filePath)
				if err != nil {
					logger.WriteServerLog(fmt.Sprintf("error saving package image: %v", err))
				}

				packageImgURL = filePath
			}
		}

		err = h.listingStore.AddWeightAvailable(listing.ID, order.Weight)
		if err != nil {
			log.Printf("error reset weight available: %v", err)
			logFile, _ := logger.WriteServerLog(fmt.Sprintf("error reset weight available: %v", err))
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
			return
		}

		err = h.orderStore.ModifyOrder(order.ID, types.Order{
			Weight:          payload.Weight,
			Price:           payload.Price,
			CurrencyID:      currency.ID,
			PackageContent:  payload.PackageContent,
			PackageImageURL: packageImgURL,
			PaymentStatus:   paymentStatus,
			OrderStatus:     constants.ORDER_STATUS_WAITING,
			PackageLocation: payload.PackageLocation,
			Notes:           payload.Notes,
		})
		if err != nil {
			log.Printf("error modify order: %v", err)
			logger.WriteServerLog(fmt.Sprintf("error modify order: %v", err))
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("modify order failed"))
			return
		}

		subject := "Re-confirm Needed!"

		body := utils.CreateEmailBodyOfOrder(subject, user.Name, listing.Destination, currency.Name, payload.Notes, payload.PackageContent, payload.Weight, payload.Price)

		err = utils.SendEmail(carrier.Email, subject, body, "", "")
		if err != nil {
			logger.WriteServerLog(fmt.Sprintf("error send email to carrier: %v", err))
		}

		fcmHistory := types.FCMHistory{
			ToUserID: carrier.ID,
			ToToken:  carrier.FCMToken,
			Data: types.FCMData{
				Type:    "confirm_order",
				OrderID: fmt.Sprintf("%d", order.ID),
			},
			Title: subject,
			Body:  fmt.Sprintf("Confirm order no. %d before %s 23:59 KST (GMT +09)", order.ID, time.Now().Local().AddDate(0, 0, 2).Format("02 Jan 2006")),
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

		returnMsg = "order modified"
	} else if reqType == "package-location" {
		var payload types.UpdatePackageLocationPayload

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

		order, err := h.orderStore.GetOrderByID(payload.ID)
		if err != nil {
			log.Printf("order not found: %v", err)
			logger.WriteServerLog(fmt.Sprintf("order not found: %v", err))
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("order not found"))
			return
		}

		orderStatus := utils.OrderStatusStringToInt(payload.OrderStatus)

		listing, err := h.listingStore.GetListingByID(order.ListingID)
		if err != nil {
			log.Printf("listing id %d not found: %v", order.ListingID, err)
			logger.WriteServerLog(fmt.Sprintf("listing id %d not found: %v", order.ListingID, err))
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("listing not found"))
			return
		}

		carrier, err := h.userStore.GetUserByID(listing.CarrierID)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("carrier not found"))
			return
		}

		if carrier.ID != user.ID {
			utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
			return
		}

		giver, err := h.userStore.GetUserByID(order.GiverID)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("carrier not found"))
			return
		}

		err = h.orderStore.UpdatePackageLocation(order.ID, orderStatus, payload.PackageLocation)
		if err != nil {
			log.Printf("error update package location: %v", err)
			logFile, _ := logger.WriteServerLog(fmt.Sprintf("error update package location: %v", err))
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
			return
		}

		subject := "Your Package Location is Updated"
		body := fmt.Sprintf("<h4>Your package for order no. %d has an update!</h4><h4>It status now is</h4><br><h2>%s</h2>",
			order.ID, payload.PackageLocation)
		err = utils.SendEmail(giver.Email, subject, body, "", "")
		if err != nil {
			logger.WriteServerLog(fmt.Sprintf("error sending email to %s for updating package location: %v", giver.Email, err))
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

		returnMsg = "package location updated"
	} else if reqType == "payment-status" {
		var payload types.UpdatePaymentStatusPayload

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

		order, err := h.orderStore.GetOrderByID(payload.ID)
		if err != nil {
			log.Printf("order not found: %v", err)
			logger.WriteServerLog(fmt.Sprintf("order not found: %v", err))
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("order not found"))
			return
		}

		paymentStatus := utils.PaymentStatusStringToInt(payload.PaymentStatus)
		if paymentStatus == -1 {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("unknown payment status"))
			return
		}

		var paymentProofUrl string
		if paymentStatus == constants.PAYMENT_STATUS_COMPLETED {
			if len(payload.PaymentProof) < 1 {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("no payment proof"))
				return
			}

			if len(payload.PaymentProof) > constants.PAYMENT_PROOF_MAX_BYTES {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("the image size exceeds the limit of 10MB"))
				return
			}

			var imageExtension string

			mimeType := http.DetectContentType(payload.PaymentProof)
			switch mimeType {
			case "image/jpeg":
				imageExtension = ".jpg"
			case "image/png":
				imageExtension = ".png"
			default:
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("unsupported image type"))
				return
			}

			filePath := constants.PAYMENT_PROOF_DIR_PATH + utils.GeneratePictureFileName(imageExtension)

			isPaymentProofUrlExist := h.orderStore.IsPaymentProofURLExist(filePath)

			for isPaymentProofUrlExist {
				filePath = constants.PAYMENT_PROOF_DIR_PATH + utils.GeneratePictureFileName(imageExtension)
				isPaymentProofUrlExist = h.orderStore.IsPaymentProofURLExist(filePath)
			}

			err = utils.SavePaymentProof(payload.PaymentProof, filePath)
			if err != nil {
				log.Printf("error saving payment proof: %v", err)
				logFile, _ := logger.WriteServerLog(fmt.Sprintf("error saving payment proof: %v", err))
				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
				return
			}

			listing, err := h.listingStore.GetListingByID(order.ListingID)
			if err != nil {
				log.Printf("error get listing: %v", err)
				logFile, _ := logger.WriteServerLog(fmt.Sprintf("error get listing: %v", err))
				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
				return
			}

			carrier, err := h.userStore.GetUserByID(listing.CarrierID)
			if err != nil {
				log.Printf("error get carrier: %v", err)
				logFile, _ := logger.WriteServerLog(fmt.Sprintf("error get carrier: %v", err))
				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
				return
			}

			subject := fmt.Sprintf("Payment Completed for Order No. %d", order.ID)
			body := fmt.Sprintf("<h4>Payment has been</h4><br><h2>completed</h2><br><h4>by %s for order no. %d!</h4><p>Below is the payment proof!<p>",
				user.Name, order.ID)

			err = utils.SendEmail(carrier.Email, subject, body, filePath, "Payment Proof")
			if err != nil {
				log.Printf("error sending payment completion email to carrier: %v", err)
				logFile, _ := logger.WriteServerLog(fmt.Sprintf("error sending payment completion email to carrier: %v", err))
				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
				return
			}

			fcmHistory := types.FCMHistory{
				ToUserID: carrier.ID,
				ToToken:  carrier.FCMToken,
				Data: types.FCMData{
					Type:    "payment_updated",
					OrderID: fmt.Sprintf("%d", order.ID),
				},
				Title: subject,
				Body:  fmt.Sprintf("Payment has been completed by %s for order no. %d!", user.Name, order.ID),
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

			paymentProofUrl = filePath
		}

		err = h.orderStore.UpdatePaymentStatus(order.ID, paymentStatus, paymentProofUrl)
		if err != nil {
			log.Printf("error update payment status: %v", err)
			logFile, _ := logger.WriteServerLog(fmt.Sprintf("error update payment status: %v", err))
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
			return
		}

		returnMsg = "payment status updated"
	} else if reqType == "order-status" {
		var payload types.UpdateOrderStatusPayload

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

		order, err := h.orderStore.GetOrderByID(payload.ID)
		if err != nil {
			log.Printf("order not found: %v", err)
			logger.WriteServerLog(fmt.Sprintf("order not found: %v", err))
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("order not found"))
			return
		}

		orderStatus := utils.OrderStatusStringToInt(payload.OrderStatus)
		if orderStatus == -1 {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("unknown order status"))
			return
		}

		if orderStatus == constants.ORDER_STATUS_CONFIRMED {
			listing, err := h.listingStore.GetListingByID(order.ListingID)
			if listing == nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("listing not found"))
				return
			}
			if err != nil {
				log.Println(err)
				logFile, _ := logger.WriteServerLog(fmt.Sprintf("%v", err))
				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
				return
			}

			if listing.CarrierID != user.ID {
				utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("you are not the carrier"))
				return
			}

			if order.OrderStatus != constants.ORDER_STATUS_WAITING {
				utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("order is not in the waiting status"))
				return
			}

			deadline := time.Date(time.Now().Local().Year(), time.Now().Local().Month(), time.Now().Local().Day(), 0, 0, 0, 0, time.Now().Local().Location())
			deadline = deadline.AddDate(0, 0, 2)

			if order.OrderConfirmationDeadline.After(deadline) {
				err = h.orderStore.UpdateOrderStatus(order.ID, constants.ORDER_STATUS_CANCELLED, "")
				if err != nil {
					log.Printf("error update order status: %v", err)
					logFile, _ := logger.WriteServerLog(fmt.Sprintf("error update order status: %v", err))
					utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
					return
				}

				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("order has been automatically canceled due to the deadline has passed"))
				return
			}

			if (listing.WeightAvailable - order.Weight) < 0.0 {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("weight available is not enough"))
				return
			}

			err = h.listingStore.SubtractWeightAvailable(listing.ID, order.Weight)
			if err != nil {
				log.Printf("error update weight available: %v", err)
				logFile, _ := logger.WriteServerLog(fmt.Sprintf("error update weight available: %v", err))
				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
				return
			}
		}

		err = h.orderStore.UpdateOrderStatus(order.ID, orderStatus, payload.PackageLocation)
		if err != nil {
			log.Printf("error update order status: %v", err)
			logFile, _ := logger.WriteServerLog(fmt.Sprintf("error update order status: %v", err))
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
			return
		}

		listing, err := h.listingStore.GetListingByID(order.ListingID)
		if err != nil {
			log.Printf("error get listing: %v", err)
			logFile, _ := logger.WriteServerLog(fmt.Sprintf("error get listing: %v", err))
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
			return
		}

		giver, err := h.userStore.GetUserByID(order.GiverID)
		if err != nil {
			log.Printf("error get giver: %v", err)
			logFile, _ := logger.WriteServerLog(fmt.Sprintf("error get giver: %v", err))
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
			return
		}

		var caser = cases.Title(language.English)
		subject := fmt.Sprintf("Order %s for Order No. %d", caser.String(payload.OrderStatus), order.ID)

		var emailBody string
		var fcmBody string

		if orderStatus == constants.ORDER_STATUS_COMPLETED {
			emailBody = fmt.Sprintf("<h4>Package has been delivered to</h4><br><h2>%s at %s!</h2>", listing.Destination, time.Now().Format("2006-01-02 15:04"))
			fcmBody = fmt.Sprintf("Package has been delivered to %s at %s!", listing.Destination, time.Now().Format("2006-01-02 15:04"))
		} else {
			emailBody = fmt.Sprintf("<h4>Order number %d has been updated into:</h4><br><h2>%s</h2>", order.ID, strings.ToUpper(payload.OrderStatus))
			fcmBody = fmt.Sprintf("Order number %d has been updated into %s", order.ID, strings.ToUpper(payload.OrderStatus))
		}

		err = utils.SendEmail(giver.Email, subject, emailBody, "", "")
		if err != nil {
			log.Printf("error sending update order status email to carrier: %v", err)
			logFile, _ := logger.WriteServerLog(fmt.Sprintf("error sending update order status email to carrier: %v", err))
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
			return
		}

		fcmHistory := types.FCMHistory{
			ToUserID: giver.ID,
			ToToken:  giver.FCMToken,
			Data: types.FCMData{
				Type:    "order_updated",
				OrderID: fmt.Sprintf("%d", order.ID),
			},
			Title: subject,
			Body:  fcmBody,
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

		returnMsg = "order status updated"
	}

	utils.WriteJSON(w, http.StatusCreated, returnMsg)
}

func (h *Handler) handleGetPaymentDetails(w http.ResponseWriter, r *http.Request) {
	var payload types.GetPaymentDetailsPayload

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

	carrier, err := h.userStore.GetUserByID(payload.CarrierID)
	if carrier == nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("account not found"))
		return
	}
	if err != nil {
		log.Println(err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("%v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	bankDetail, err := h.bankDetailStore.GetBankDetailByUserID(carrier.ID)
	if err != nil {
		log.Printf("error fetching bank data: %v", err)
		logFile, _ := logger.WriteServerLog(fmt.Sprintf("error fetching bank data: %v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	var returnMsg interface{}

	if bankDetail == nil {
		returnMsg = map[string]string{
			"status":  "not exist",
			"message": "Carrier hasn't updated his/her bank account! Please contact him/her directly using email!",
		}
	} else {
		returnMsg = map[string]string{
			"status":         "exist",
			"bank_name":      bankDetail.BankName,
			"account_number": bankDetail.AccountNumber,
			"account_holder": bankDetail.AccountHolder,
		}
	}

	utils.WriteJSON(w, http.StatusOK, returnMsg)
}
