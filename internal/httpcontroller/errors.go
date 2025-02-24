package httpcontroller

import (
	"errors"
	"github.com/gofiber/fiber/v2"
)

func ErrorHandler(c *fiber.Ctx, err error) error {

	if err != nil {
		fiberErr := new(fiber.Error)
		if errors.As(err, &fiberErr) {
			return ResErr(c, fiberErr.Code, fiberErr.Message)
		}
		return ResErr(c, 500, "Internal error")
	}

	return nil
}

func ResErr(c *fiber.Ctx, code int, msg string) error {
	c.Response().Header.Set("Content-Type", "application/json")
	c.Status(code)

	return c.JSON(&ErrorResponse{Errors: &msg})
}
