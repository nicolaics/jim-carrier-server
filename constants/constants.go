package constants

const EXP_STATUS_AVAILABLE = 0
const EXP_STATUS_EXPIRED = 1

const PAYMENT_STATUS_PENDING = 0
const PAYMENT_STATUS_CANCELLED = 1
const PAYMENT_STATUS_REFUNDED = 2
const PAYMENT_STATUS_COMPLETED = 3

const ORDER_STATUS_EN_ROUTE = 0
const ORDER_STATUS_CONFIRMED = 1
const ORDER_STATUS_WAITING = 2
const ORDER_STATUS_COMPLETED = 3
const ORDER_STATUS_CANCELLED = 4

const WAITING_STATUS_STR = "waiting"
const COMPLETED_STATUS_STR = "completed"
const CANCELLED_STATUS_STR = "cancelled"
const CONFIRMED_STATUS_STR = "confirmed"
const EN_ROUTE_STATUS_STR = "en-route"
const PENDING_STATUS_STR = "pending"
const REFUNDED_STATUS_STR = "refunded"
const EXPIRED_STATUS_STR = "expired"
const AVAILABLE_STATUS_STR = "available"

const VERIFY_CODE_WAITING = 0
const VERIFY_CODE_COMPLETE = 1

const SIGNUP = 0
const FORGET_PASSWORD = 1

const PROVIDER_EMAIL = "email"
const PROVIDER_GMAIL = "gmail"

const REVIEW_GIVER_TO_CARRIER = 0
const REVIEW_CARRIER_TO_GIVER = 1

const PAYMENT_PROOF_DIR_PATH = "static/img/payment_proof/"
const PROFILE_IMG_DIR_PATH = "static/img/profile_img/"
const PACKAGE_IMG_DIR_PATH = "static/img/package/"

const PAYMENT_PROOF_MAX_BYTES = 10 << 20 // 10MB in bytes
const PROFILE_IMG_MAX_BYTES = 5 << 20 // 5MB in bytes
const PACKAGE_IMG_MAX_BYTES = 10 << 20 // 10MB in bytes

const ACCESS_TOKEN = 0
const REFRESH_TOKEN = 1
