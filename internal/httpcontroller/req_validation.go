package httpcontroller

import (
	"errors"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"log/slog"
)

const (
	maxNameLength = 200
	minNameLength = 10
)

func (request AuthRequest) Validate(log *slog.Logger) *RespErr {
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
			return Err500
		}
		var errValVal validation.Errors
		if errors.As(errValid, &errValVal) {
			log.Info("user sent invalid auth params",
				slog.Any("invalids params", errValVal),
			)
			return NewErrf(400, "invalid user auth parameters: %s", errValVal.Error())
		}
		return Err500
	}
	return nil
}
