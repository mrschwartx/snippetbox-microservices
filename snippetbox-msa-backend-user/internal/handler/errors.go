package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v5"

	"snippetbox/backend-user/internal/apperror"
	"snippetbox/backend-user/internal/dto"
)

var (
	errMalformedBody = errors.New("handler: malformed request body")
	errBodyTooLarge  = errors.New("handler: request body too large")
)

func NewErrorHandler(log *slog.Logger) echo.HTTPErrorHandler {
	return func(c *echo.Context, err error) {
		if err == nil {
			return
		}

		status, body := describe(err)
		if status >= http.StatusInternalServerError {
			log.Error("request failed",
				"error", err,
				"method", c.Request().Method,
				"path", c.Request().URL.Path,
			)
		}
		if responseCommitted(c) {
			return
		}

		_ = c.JSON(status, body)
	}
}

func responseCommitted(c *echo.Context) bool {
	res, err := echo.UnwrapResponse(c.Response())
	return err == nil && res.Committed
}

func describe(err error) (int, dto.ErrorResponse) {
	var validation *apperror.ValidationError
	var coded echo.HTTPStatusCoder

	switch {
	case errors.As(err, &validation):
		return http.StatusUnprocessableEntity, dto.ErrorResponse{Error: dto.ErrorBody{
			Code:           dto.CodeValidation,
			Message:        "Validation failed",
			FieldErrors:    validation.FieldErrors,
			NonFieldErrors: validation.NonFieldErrors,
		}}
	case errors.Is(err, apperror.ErrNotFound):
		return http.StatusNotFound, dto.NewErrorResponse(dto.CodeNotFound, "Resource not found")
	case errors.Is(err, apperror.ErrInvalidCredentials):
		return http.StatusUnauthorized, dto.NewErrorResponse(dto.CodeInvalidCredentials, "Email or password is incorrect")
	case errors.Is(err, apperror.ErrUnauthenticated):
		return http.StatusUnauthorized, dto.NewErrorResponse(dto.CodeUnauthenticated, "Authentication required")
	case errors.Is(err, errMalformedBody):
		return http.StatusBadRequest, dto.NewErrorResponse(dto.CodeBadRequest, "Request body must be valid JSON with known fields")
	case errors.Is(err, errBodyTooLarge):
		return http.StatusRequestEntityTooLarge, dto.NewErrorResponse(dto.CodePayloadTooLarge, "Request body is too large")
	case errors.As(err, &coded):
		status := coded.StatusCode()
		return status, dto.NewErrorResponse(codeForStatus(status), http.StatusText(status))
	default:
		return http.StatusInternalServerError, dto.NewErrorResponse(dto.CodeInternal, "Internal server error")
	}
}

func codeForStatus(status int) string {
	switch status {
	case http.StatusNotFound:
		return dto.CodeNotFound
	case http.StatusMethodNotAllowed:
		return dto.CodeMethodNotAllowed
	case http.StatusBadRequest:
		return dto.CodeBadRequest
	default:
		return dto.CodeHTTPError
	}
}
