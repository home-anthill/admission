package customerrors

// Status represents the category of a gRPC send failure.
type Status int

// Possible gRPC send error statuses.
const (
	ConnectionError Status = iota + 1
	BadParams
)

// GrpcSendError represents a failure when sending a gRPC request to a downstream service.
type GrpcSendError struct {
	Status  Status
	Message string
}

// Error returns the human-readable error message.
func (se GrpcSendError) Error() string {
	return se.Message
}
