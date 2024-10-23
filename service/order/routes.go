package order

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
	orderStore types.OrderStore
	userStore  types.UserStore
	listingStore types.ListingStore
}

func NewHandler(orderStore types.OrderStore, userStore types.UserStore, listingStore types.ListingStore) *Handler {
	return &Handler{
		orderStore: orderStore,
		userStore:  userStore,
		listingStore: listingStore,
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

	router.HandleFunc("/order", h.handleModify).Methods(http.MethodPatch)
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

	isDuplicate, err := h.orderStore.IsOrderDuplicate(user.ID, listing.ID)
	if err != nil || isDuplicate {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("duplicate order"))
		return
	}

	var paymentStatus int

	if payload.PaymentStatus == "pending" {
		paymentStatus = constants.PAYMENT_STATUS_PENDING
	} else if payload.PaymentStatus == "completed" {
		paymentStatus = constants.PAYMENT_STATUS_COMPLETED
	} else if payload.PaymentStatus == "cancelled" {
		paymentStatus = constants.PAYMENT_STATUS_CANCELLED
	} else {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("unknown payment status"))
		return
	}

	var orderStatus int

	if payload.OrderStatus == "waiting" {
		orderStatus = constants.ORDER_STATUS_WAITING
	} else if payload.OrderStatus == "completed" {
		orderStatus = constants.ORDER_STATUS_COMPLETED
	} else if payload.OrderStatus == "cancelled" {
		orderStatus = constants.ORDER_STATUS_CANCELLED
	} else if payload.OrderStatus == "verifying" {
		orderStatus = constants.ORDER_STATUS_VERIFYING
	} else {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("unknown payment status"))
		return
	}

	err = h.orderStore.CreateOrder(types.Order{
		ListingID: listing.ID,
		GiverID: user.ID,
		Weight: payload.Weight,
		Price: payload.Price,
		PaymentStatus: paymentStatus,
		OrderStatus: orderStatus,
		PackageLocation: payload.PackageLocation,
		Notes: payload.Notes,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error create order: %v", err))
		return
	}

	err = h.listingStore.SubtractWeightAvailable(listing.ID, payload.Weight)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error update weight available: %v", err))
		return
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

	var orders interface{}
	if reqType == "carrier" {
		orders, err = h.orderStore.GetOrderByCarrierID(user.ID)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}
	} else if reqType == "giver" {
		orders, err = h.orderStore.GetOrderByGiverID(user.ID)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}
	} else {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("unknown request parameter"))
		return
	}

	if orders == nil {
		utils.WriteJSON(w, http.StatusOK, "no order found")
		return
	}
	
	utils.WriteJSON(w, http.StatusOK, orders)
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

	var order interface{}

	if reqType == "carrier" {
		order, err = h.orderStore.GetCarrierOrderByID(payload.ID, user.ID)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}
	} else if reqType == "giver" {
		order, err = h.orderStore.GetGiverOrderByID(payload.ID, user.ID)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}
	} else {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("unknown request parameter"))
		return
	}

	if order == nil {
		utils.WriteJSON(w, http.StatusOK, "no order found")
		return
	}
	
	utils.WriteJSON(w, http.StatusOK, order)
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

	order, err := h.orderStore.GetOrderByID(payload.ID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("order not found: %v", err))
		return
	}

	listing, err := h.listingStore.GetListingByID(payload.NewData.ListingID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("listing id %d not found: %v", payload.NewData.ListingID, err))
		return
	}

	var paymentStatus int

	if payload.NewData.PaymentStatus == "pending" {
		paymentStatus = constants.PAYMENT_STATUS_PENDING
	} else if payload.NewData.PaymentStatus == "completed" {
		paymentStatus = constants.PAYMENT_STATUS_COMPLETED
	} else if payload.NewData.PaymentStatus == "cancelled" {
		paymentStatus = constants.PAYMENT_STATUS_CANCELLED
	} else {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("unknown payment status"))
		return
	}

	var orderStatus int

	if payload.NewData.OrderStatus == "waiting" {
		orderStatus = constants.ORDER_STATUS_WAITING
	} else if payload.NewData.OrderStatus == "completed" {
		orderStatus = constants.ORDER_STATUS_COMPLETED
	} else if payload.NewData.OrderStatus == "cancelled" {
		orderStatus = constants.ORDER_STATUS_CANCELLED
	} else if payload.NewData.OrderStatus == "verifying" {
		orderStatus = constants.ORDER_STATUS_VERIFYING
	} else {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("unknown payment status"))
		return
	}

	err = h.orderStore.ModifyOrder(order.ID, types.Order{
		Weight: payload.NewData.Weight,
		Price: payload.NewData.Price,
		PaymentStatus: paymentStatus,
		OrderStatus: orderStatus,
		PackageLocation: payload.NewData.PackageLocation,
		Notes: payload.NewData.Notes,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error modify order: %v", err))
		return
	}

	err = h.listingStore.AddWeightAvailable(listing.ID, order.Weight)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error reset weight available: %v", err))
		return
	}

	err = h.listingStore.SubtractWeightAvailable(listing.ID, payload.NewData.Weight)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error update weight available: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusCreated, "order created")
}
