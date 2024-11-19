package order

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/nicolaics/jim-carrier/constants"
	"github.com/nicolaics/jim-carrier/logger"
	"github.com/nicolaics/jim-carrier/types"
	"github.com/nicolaics/jim-carrier/utils"
)

type Handler struct {
	orderStore    types.OrderStore
	userStore     types.UserStore
	listingStore  types.ListingStore
	currencyStore types.CurrencyStore
}

func NewHandler(orderStore types.OrderStore, userStore types.UserStore,
	listingStore types.ListingStore, currencyStore types.CurrencyStore) *Handler {
	return &Handler{
		orderStore:    orderStore,
		userStore:     userStore,
		listingStore:  listingStore,
		currencyStore: currencyStore,
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

	// router.HandleFunc("/order/get-payment-proof", h.handleGetPaymentProofImage).Methods(http.MethodPost)
	// router.HandleFunc("/order/get-payment-proof", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var payload types.RegisterOrderPayload

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

	listing, err := h.listingStore.GetListingByID(payload.ListingID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("listing id %d not found: %v", payload.ListingID, err))
		return
	}

	carrier, err := h.userStore.GetUserByID(listing.CarrierID)
	if carrier == nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("carrier not found"))
		return
	}
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	isDuplicate, err := h.orderStore.IsOrderDuplicate(user.ID, listing.ID)
	if isDuplicate {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("duplicate order"))
		return
	}
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
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

	err = h.listingStore.SubtractWeightAvailable(listing.ID, payload.Weight)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error update weight available: %v", err))
		return
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
		errTemp := h.listingStore.AddWeightAvailable(listing.ID, payload.Weight)
		var errorMsg error

		if errTemp != nil {
			errorMsg = fmt.Errorf("error reset weight: %v\nerror create order: %v", errTemp, err)
		} else {
			errorMsg = fmt.Errorf("error create order: %v", err)
		}

		utils.WriteError(w, http.StatusInternalServerError, errorMsg)
		return
	}

	packageContents := utils.WrapText(payload.PackageContent, 100)

	subject := "New Order Arrived!"

	body := "New order just arrived to your listing!\n\n"
	body += "Here are the details:\n"
	body += fmt.Sprintf("\t%-15s: %s\n", "Name", user.Name)
	body += fmt.Sprintf("\t%-15s: %s\n", "Destination", listing.Destination)
	body += fmt.Sprintf("\t%-15s: %.1f\n", "Weight", payload.Weight)
	body += fmt.Sprintf("\t%-15s: %.1f\n", "Total Price", payload.Price)
	body += fmt.Sprintf("\t%-15s: %s\n", "Package Content", packageContents[0])

	for _, line := range packageContents[1:] {
		fmt.Printf("\t%-15s  %s\n", "", line)
	}

	if payload.Notes != "" {
		body += fmt.Sprintf("\t%-15s: %s\n", "Notes", payload.Notes)
	}

	body += "\nAttached is the image of the package!\n\n"
	body += "Confirm the order before\n"
	body += fmt.Sprintf("\t\t%s 23:59 KST (GMT +09)\n", time.Now().Local().AddDate(0, 0, 2).Format("02 JAN 2006"))
	body += "If not confirmed by then, the order will automatically be cancelled!"

	err = utils.SendEmail(carrier.Email, subject, body, "", "")
	if err != nil {
		logger.WriteServerLog(fmt.Sprintf("error send email to carrier: %v", err))
	}

	utils.WriteJSON(w, http.StatusCreated, "order created")
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

	var ordersReturn interface{}
	if reqType == "carrier" {
		orders, err := h.orderStore.GetOrderByCarrierID(user.ID)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		ordersReturnTemp := make([]types.OrderCarrierReturnPayload, 0)
		for _, order := range orders {
			paymentStatus := utils.PaymentStatusIntToString(order.PaymentStatus)

			orderStatus := utils.OrderStatusIntToString(order.OrderStatus)

			temp := types.OrderCarrierReturnPayload{
				Listing:          order.Listing,
				ID:               order.ID,
				GiverName:        order.GiverName,
				GiverPhoneNumber: order.GiverPhoneNumber,
				Weight:           order.Weight,
				Price:            order.Price,
				Currency:         order.Currency,
				PackageContent:   order.PackageContent,
				PackageImageURL:  order.PackageImageURL,
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
		orders, err := h.orderStore.GetOrderByGiverID(user.ID)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		ordersReturnTemp := make([]types.OrderGiverReturnPayload, 0)
		for _, order := range orders {
			paymentStatus := utils.PaymentStatusIntToString(order.PaymentStatus)

			orderStatus := utils.OrderStatusIntToString(order.OrderStatus)

			temp := types.OrderGiverReturnPayload{
				Listing:         order.Listing,
				ID:              order.ID,
				Weight:          order.Weight,
				Price:           order.Price,
				Currency:        order.Currency,
				PackageContent:  order.PackageContent,
				PackageImageURL: order.PackageImageURL,
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

	vars := mux.Vars(r)
	reqType := vars["reqType"]

	var returnOrder interface{}

	if reqType == "carrier" {
		order, err := h.orderStore.GetCarrierOrderByID(payload.ID, user.ID)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		paymentStatus := utils.PaymentStatusIntToString(order.PaymentStatus)

		orderStatus := utils.OrderStatusIntToString(order.OrderStatus)

		returnOrder = types.OrderCarrierReturnPayload{
			Listing:          order.Listing,
			ID:               order.ID,
			GiverName:        order.GiverName,
			GiverPhoneNumber: order.GiverPhoneNumber,
			Weight:           order.Weight,
			Price:            order.Price,
			Currency:         order.Currency,
			PackageContent:   order.PackageContent,
			PackageImageURL:  order.PackageImageURL,
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
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		paymentStatus := utils.PaymentStatusIntToString(order.PaymentStatus)

		orderStatus := utils.OrderStatusIntToString(order.OrderStatus)

		returnOrder = types.OrderGiverReturnPayload{
			Listing:         order.Listing,
			ID:              order.ID,
			Weight:          order.Weight,
			Price:           order.Price,
			Currency:        order.Currency,
			PackageContent:  order.PackageContent,
			PackageImageURL: order.PackageImageURL,
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

	err = h.orderStore.DeleteOrder(payload.ID, user.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error deleting order: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, "delete order success")
}

func (h *Handler) handleModify(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	reqType := vars["reqType"]

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

	returnMsg := ""

	if reqType == "all" {
		var payload types.ModifyOrderPayload

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

		order, err := h.orderStore.GetOrderByID(payload.ID)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("order not found: %v", err))
			return
		}

		listing, err := h.listingStore.GetListingByID(payload.ListingID)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("listing id %d not found: %v", payload.ListingID, err))
			return
		}

		paymentStatus := utils.PaymentStatusStringToInt(payload.PaymentStatus)
		if paymentStatus == -1 {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("unknown payment status"))
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
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error reset weight available: %v", err))
			return
		}

		err = h.listingStore.SubtractWeightAvailable(listing.ID, payload.Weight)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error update weight available: %v", err))
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
			errTemp := h.listingStore.AddWeightAvailable(listing.ID, payload.Weight)
			var errorMsg error

			if errTemp != nil {
				errorMsg = fmt.Errorf("error reset weight: %v\nerror modify order: %v", errTemp, err)
			} else {
				errorMsg = fmt.Errorf("error modify order: %v", err)
			}

			utils.WriteError(w, http.StatusInternalServerError, errorMsg)
			return
		}

		returnMsg = "order modified"
	} else if reqType == "package-location" {
		var payload types.UpdatePackageLocationPayload

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

		order, err := h.orderStore.GetOrderByID(payload.ID)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("order not found: %v", err))
			return
		}

		orderStatus := utils.OrderStatusStringToInt(payload.OrderStatus)

		err = h.orderStore.UpdatePackageLocation(order.ID, orderStatus, payload.PackageLocation)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error update package location: %v", err))
			return
		}

		returnMsg = "package location updated"
	} else if reqType == "payment-status" {
		var payload types.UpdatePaymentStatusPayload

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

		order, err := h.orderStore.GetOrderByID(payload.ID)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("order not found: %v", err))
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
				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error saving payment proof: %v", err))
				return
			}

			listing, err := h.listingStore.GetListingByID(order.ListingID)
			if err != nil {
				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error get listing: %v", err))
				return
			}

			carrier, err := h.userStore.GetUserByID(listing.CarrierID)
			if err != nil {
				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error get carrier: %v", err))
				return
			}

			subject := fmt.Sprintf("Payment Completed for Order No. %d", order.ID)
			body := fmt.Sprintf("Payment has been completed by %s for package to %s at %s!\nBelow is the payment proof!",
				user.Name, listing.Destination, listing.DepartureDate.Format("2006-01-02"))

			err = utils.SendEmail(carrier.Email, subject, body, filePath, "Payment Proof")
			if err != nil {
				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error sending payment completion email to carrier: %v", err))
				return
			}

			paymentProofUrl = filePath
		}

		err = h.orderStore.UpdatePaymentStatus(order.ID, paymentStatus, paymentProofUrl)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error update payment status: %v", err))
			return
		}

		returnMsg = "payment status updated"
	} else if reqType == "order-status" {
		var payload types.UpdateOrderStatusPayload

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

		order, err := h.orderStore.GetOrderByID(payload.ID)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("order not found: %v", err))
			return
		}

		orderStatus := utils.OrderStatusStringToInt(payload.OrderStatus)
		if orderStatus == -1 {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("unknown order status"))
			return
		}

		err = h.orderStore.UpdateOrderStatus(order.ID, orderStatus, payload.PackageLocation)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error update order status: %v", err))
			return
		}

		if orderStatus == constants.ORDER_STATUS_COMPLETED {
			listing, err := h.listingStore.GetListingByID(order.ListingID)
			if err != nil {
				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error get listing: %v", err))
				return
			}

			giver, err := h.userStore.GetUserByID(order.GiverID)
			if err != nil {
				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error get giver: %v", err))
				return
			}

			subject := fmt.Sprintf("Order Completed for Order No. %d", order.ID)
			body := fmt.Sprintf("Package has been delivered to %s at %s!", listing.Destination, time.Now().Format("2006-01-02 15:04"))

			err = utils.SendEmail(giver.Email, subject, body, "", "")
			if err != nil {
				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error sending payment completion email to carrier: %v", err))
				return
			}
		}

		returnMsg = "order status updated"
	}

	utils.WriteJSON(w, http.StatusCreated, returnMsg)
}

/*
func (h *Handler) handleGetPaymentProofImage(w http.ResponseWriter, r *http.Request) {
	var payload types.GetPaymentProofImagePayload

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

	order, err := h.orderStore.GetOrderByPaymentProofURL(payload.PaymentProofURL)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error getting order: %v", err))
		return
	}

	listing, err := h.listingStore.GetListingByID(order.ListingID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error getting listing: %v", err))
		return
	}

	if user.ID != order.GiverID && user.ID != listing.CarrierID {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("you are not related to this order"))
		return
	}

	utils.WriteJSON(w, http.StatusOK, )
}
*/
