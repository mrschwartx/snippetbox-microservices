package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"

	"snippetbox/backend-snippet/internal/apperror"
)

const (
	APIPrefix = "/api/v1"
	idBitSize = min(63, strconv.IntSize)
)

func bindJSON(c *echo.Context, dst any) error {
	decoder := json.NewDecoder(c.Request().Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return bodyError(err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return bodyError(err)
	}
	return nil
}

func bodyError(err error) error {
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		return errBodyTooLarge
	}
	return errMalformedBody
}

func parseID(raw string) (uint, error) {
	id, err := strconv.ParseUint(raw, 10, idBitSize)
	if err != nil {
		return 0, apperror.ErrNotFound
	}
	return uint(id), nil
}
