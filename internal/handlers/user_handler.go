package handlers

import (
	"e_metting/internal/models"
	"e_metting/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

type UserHandler struct {
	userService services.UserService
}

func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) Register(c *fiber.Ctx) error {
	req := c.Locals("request").(models.RegisterRequest)

	user, err := h.userService.Register(req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to register user")
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to register user",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(models.RegisterResponse{
		Message: "User registered successfully",
		UserID:  user.ID,
	})
}

func (h *UserHandler) Login(c *fiber.Ctx) error {
	req := c.Locals("request").(models.LoginRequest)

	token, err := h.userService.Login(req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to login")
		return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
			Error: "Invalid credentials",
		})
	}

	return c.JSON(models.LoginResponse{
		Token: token,
	})
}
