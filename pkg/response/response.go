package response

import (
	"github.com/gofiber/fiber/v2"
)

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
}

type ErrorResponse struct {
	Success bool `json:"success"`
	Error   struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func Success(c *fiber.Ctx, data interface{}) error {
	return c.JSON(Response{
		Success: true,
		Data:    data,
	})
}

func SuccessWithMeta(c *fiber.Ctx, data, meta interface{}) error {
	return c.JSON(Response{
		Success: true,
		Data:    data,
		Meta:    meta,
	})
}

func Created(c *fiber.Ctx, data interface{}) error {
	return c.Status(fiber.StatusCreated).JSON(Response{
		Success: true,
		Data:    data,
	})
}

func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

func Error(c *fiber.Ctx, status int, code int, message string) error {
	return c.Status(status).JSON(ErrorResponse{
		Success: false,
		Error: struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}{
			Code:    code,
			Message: message,
		},
	})
}

func BadRequest(c *fiber.Ctx, code int, message string) error {
	return Error(c, fiber.StatusBadRequest, code, message)
}

func Unauthorized(c *fiber.Ctx, message string) error {
	return Error(c, fiber.StatusUnauthorized, 401, message)
}

func Forbidden(c *fiber.Ctx, message string) error {
	return Error(c, fiber.StatusForbidden, 403, message)
}

func NotFound(c *fiber.Ctx, message string) error {
	return Error(c, fiber.StatusNotFound, 404, message)
}

func Conflict(c *fiber.Ctx, code int, message string) error {
	return Error(c, fiber.StatusConflict, code, message)
}

func InternalError(c *fiber.Ctx, code int, message string) error {
	return Error(c, fiber.StatusInternalServerError, code, message)
}
