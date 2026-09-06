package errcode

// Stable machine identifiers for structured logs (and optional HTTP/RPC).
// Prefer these string values when emitting applog.Event.Code.
// RPC API keeps aliases in sumeru/core/server/api for the same literals.
const (
	Unauthorized        = "UNAUTHORIZED"
	SessionExpired      = "SESSION_EXPIRED"
	InvalidCredentials  = "INVALID_CREDENTIALS"
	ValidationError     = "VALIDATION_ERROR"
	EmailAlreadyExists  = "EMAIL_ALREADY_EXISTS"
	AccessDenied        = "ACCESS_DENIED"
	RecordNotFound      = "RECORD_NOT_FOUND"
	NotFound            = "NOT_FOUND"
	InternalError       = "INTERNAL_ERROR"
	SyncPartial         = "SYNC_PARTIAL"
	XMLLinkFailed       = "XML_LINK_FAILED"
	CronUpdateFailed    = "CRON_UPDATE_FAILED"
	CronCommitFailed    = "CRON_COMMIT_FAILED"
	CronHandlerFailed   = "CRON_HANDLER_FAILED"
)
