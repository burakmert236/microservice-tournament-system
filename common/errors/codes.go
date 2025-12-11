package errors

const (
	// Generic codes
	CodeNotFound               = "NOT_FOUND"
	CodeAlreadyExists          = "ALREADY_EXISTS"
	CodeInvalidInput           = "INVALID_INPUT"
	CodeUnauthorized           = "UNAUTHORIZED"
	CodeForbidden              = "FORBIDDEN"
	CodeConflict               = "CONFLICT"
	CodeInternalServer         = "INTERNAL_SERVER"
	CodeServiceUnavailable     = "SERVICE_UNAVAILABLE"
	CodeEventPublishError      = "EVENT_PUBLISH_ERROR"
	CodeEventSubscribtionError = "EVENT_SUBSCRIPTION_ERROR"
	CodeObjectMarshalError     = "OBJECT_MARSHALL_ERROR"
	CodeObjectUnmarshalError   = "OBJECT_UNMARSHALL_ERROR"
	CodeDatabaseError          = "DATABASE_ERROR"
	CodeTransactionError       = "TRANSACTION_ERROR"
	CodeGrpcCallError          = "GRPC_CALL_ERROR"
	CodeRedisOperationError    = "REDIS_ERROR"
)
