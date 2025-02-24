package httpcontroller

import (
	"errors"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/gofiber/fiber/v2"
	"log/slog"
)

const (
	maxNameLength = 200
	minNameLength = 10
)

func (request AuthRequest) Validate(log *slog.Logger) *fiber.Error {
	rb := request
	errValid := validation.ValidateStruct(&rb,
		validation.Field(&rb.Username,
			validation.Required, validation.RuneLength(minNameLength, maxNameLength),
		),
		validation.Field(&rb.Password,
			validation.Required, is.Alphanumeric,
		),
	)
	if errValid != nil {
		var errValInternal validation.InternalError
		if errors.As(errValid, &errValInternal) {
			log.Error("fail to validate input: " + errValInternal.Error())
			return fiber.NewError(500)
		}
		var errValVal validation.Errors
		if errors.As(errValid, &errValVal) {
			log.Info("user sent invalid auth params",
				slog.Any("invalids params", errValVal),
			)
			return fiber.NewError(400, "invalid user auth parameters: %s", errValVal.Error())
		}
		return fiber.NewError(500)
	}
	return nil
}
