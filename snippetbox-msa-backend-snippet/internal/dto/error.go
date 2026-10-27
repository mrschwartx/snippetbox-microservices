package dto

const (
	CodeValidation       = "VALIDATION_ERROR"
	CodeNotFound         = "NOT_FOUND"
	CodeUnauthenticated  = "UNAUTHENTICATED"
	CodeBadRequest       = "BAD_REQUEST"
	CodePayloadTooLarge  = "PAYLOAD_TOO_LARGE"
	CodeMethodNotAllowed = "METHOD_NOT_ALLOWED"
	CodeHTTPError        = "HTTP_ERROR"
	CodeInternal         = "INTERNAL_ERROR"
)

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code           string            `json:"code"`
	Message        string            `json:"message"`
	FieldErrors    map[string]string `json:"fields,omitempty"`
	NonFieldErrors []string          `json:"non_field,omitempty"`
}

func NewErrorResponse(code, message string) ErrorResponse {
	return ErrorResponse{Error: ErrorBody{Code: code, Message: message}}
}
