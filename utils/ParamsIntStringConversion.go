package utils

import "github.com/nicolaics/jim-carrier-server/constants"

// to set the order status from string into int
func OrderStatusStringToInt(orderStr string) int {
	var orderStatus int
	switch orderStr {
	case constants.WAITING_STATUS_STR:
		orderStatus = constants.ORDER_STATUS_WAITING
	case constants.COMPLETED_STATUS_STR:
		orderStatus = constants.ORDER_STATUS_COMPLETED
	case constants.CANCELLED_STATUS_STR:
		orderStatus = constants.ORDER_STATUS_CANCELLED
	case constants.CONFIRMED_STATUS_STR:
		orderStatus = constants.ORDER_STATUS_CONFIRMED
	case constants.EN_ROUTE_STATUS_STR:
		orderStatus = constants.ORDER_STATUS_EN_ROUTE
	default:
		orderStatus = -1
	}

	return orderStatus
}

// to get the order status string from int
func OrderStatusIntToString(orderStatus int) string {
	var orderStr string
	switch orderStatus {
	case constants.ORDER_STATUS_CANCELLED:
		orderStr = constants.CANCELLED_STATUS_STR
	case constants.ORDER_STATUS_WAITING:
		orderStr = constants.WAITING_STATUS_STR
	case constants.ORDER_STATUS_COMPLETED:
		orderStr = constants.COMPLETED_STATUS_STR
	case constants.ORDER_STATUS_EN_ROUTE:
		orderStr = constants.EN_ROUTE_STATUS_STR
	case constants.ORDER_STATUS_CONFIRMED:
		orderStr = constants.CONFIRMED_STATUS_STR
	}

	return orderStr
}

// to set the payment status from string into int
func PaymentStatusStringToInt(paymentStr string) int {
	var paymentStatus int
	switch paymentStr {
	case constants.PENDING_STATUS_STR:
		paymentStatus = constants.PAYMENT_STATUS_PENDING
	case constants.COMPLETED_STATUS_STR:
		paymentStatus = constants.PAYMENT_STATUS_COMPLETED
	case constants.CANCELLED_STATUS_STR:
		paymentStatus = constants.PAYMENT_STATUS_CANCELLED
	default:
		paymentStatus = -1
	}

	return paymentStatus
}

// to get the payment status string from int
func PaymentStatusIntToString(paymentStatus int) string {
	var paymentStr string
	switch paymentStatus {
	case constants.PAYMENT_STATUS_CANCELLED:
		paymentStr = constants.CANCELLED_STATUS_STR
	case constants.PAYMENT_STATUS_PENDING:
		paymentStr = constants.PENDING_STATUS_STR
	case constants.PAYMENT_STATUS_COMPLETED:
		paymentStr = constants.COMPLETED_STATUS_STR
	}

	return paymentStr
}

func ExpStatusIntToString(expStatus int) string {
	var expStatusStr string
	switch expStatus {
	case constants.EXP_STATUS_AVAILABLE:
		expStatusStr = constants.AVAILABLE_STATUS_STR
	case constants.EXP_STATUS_EXPIRED:
		expStatusStr = constants.EXPIRED_STATUS_STR
	}

	return expStatusStr
}