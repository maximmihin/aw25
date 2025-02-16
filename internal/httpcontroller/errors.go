package httpcontroller

import (
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
)

type RespErr struct {
	Code int
	Msg  string
}

func (r RespErr) Error() string {
	return r.Msg
}

var Err500 = &RespErr{
	Code: 500,
	Msg:  "Internal error",
}

func NewErrf(code int, msg string, a ...any) *RespErr {

	if len(a) > 0 {
		msg = fmt.Sprintf(msg, a)
	}
	return &RespErr{
		Code: code,
		Msg:  msg,
	}
}

func ErrorHandler(c *fiber.Ctx, err error) error {

	if err != nil {
		ftErr := new(RespErr)
		if errors.As(err, &ftErr) {
			return SendErr(c, ftErr.Code, ftErr.Msg)
		}
		fiberErr := new(fiber.Error)
		if errors.As(err, &fiberErr) {
			return SendErr(c, fiberErr.Code, fiberErr.Message)
		}
		return SendErr(c, 500, "Internal error")
	}

	return nil
}

func SendErr(c *fiber.Ctx, code int, msg string) error {
	c.Response().Header.Set("Content-Type", "application/json")
	c.Status(code)

	return c.JSON(&ErrorResponse{Errors: &msg})
}
