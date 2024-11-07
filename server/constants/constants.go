package constants

const EXP_STATUS_EXPIRED = 1
const EXP_STATUS_AVAILABLE = 0

const PAYMENT_STATUS_PENDING = 0
const PAYMENT_STATUS_COMPLETED = 1
const PAYMENT_STATUS_CANCELLED = 2

const ORDER_STATUS_WAITING = 0
const ORDER_STATUS_COMPLETED = 1
const ORDER_STATUS_CANCELLED = 2
const ORDER_STATUS_VERIFYING = 3
const ORDER_STATUS_EN_ROUTE = 4

const WAITING_STATUS_STR = "waiting"
const COMPLETED_STATUS_STR = "completed"
const CANCELLED_STATUS_STR = "cancelled"
const VERIFYING_STATUS_STR = "verifying"
const EN_ROUTE_STATUS_STR = "en-route"
const PENDING_STATUS_STR = "pending"

const VERIFY_CODE_WAITING = 0
const VERIFY_CODE_COMPLETE = 1

const SIGNUP = 0
const FORGET_PASSWORD = 1

const PROVIDER_EMAIL = "email"
const PROVIDER_GMAIL = "gmail"

const REVIEW_GIVER_TO_CARRIER = 0
const REVIEW_CARRIER_TO_GIVER = 1

const PAYMENT_PROOF_DIR_PATH = "./static/payment_proof/"
const PROFILE_IMG_DIR_PATH = "./static/profile_img/"

const PAYMENT_PROOF_MAX_BYTES = 10 << 20 // 10MB in bytes
const PROFILE_IMG_MAX_BYTES = 5 << 20 // 5MB in bytes