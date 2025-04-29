package handlers

import (
	"e_metting/internal/models"
	"e_metting/internal/services"
	"fmt"
	"net/http"

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
			Error: err.Error(),
		})
	}

	return c.JSON(models.LoginResponse{
		Token: token,
	})
}

func (h *UserHandler) GetProfile(c *fiber.Ctx) error {
	// Get authenticated user ID from context
	authUserID, _ := c.Locals("userID").(string)
	requestedID := c.Params("id")

	// Optional: Check if user is requesting their own profile or has admin rights
	isAdmin, _ := c.Locals("isAdmin").(bool)
	if authUserID != requestedID && !isAdmin {
		return c.Status(http.StatusForbidden).JSON(models.ErrorResponse{
			Error: "Forbidden, you can only access your own profile",
		})
	}

	profile, err := h.userService.GetProfile(requestedID)
	if err != nil {
		switch err.Error() {
		case "user not found":
			return c.Status(http.StatusNotFound).JSON(models.ErrorResponse{
				Error: "user not found",
			})
		case "invalid user ID format":
			return c.Status(http.StatusBadRequest).JSON(models.ErrorResponse{
				Error: "invalid user ID format",
			})
		default:
			fmt.Printf("Error fetching user profile: %v\n", err)
			return c.Status(http.StatusInternalServerError).JSON(models.ErrorResponse{
				Error: "Failed to fetch user profile",
			})
		}
	}

	return c.Status(http.StatusOK).JSON(profile)
}

func (h *UserHandler) UpdateProfile(c *fiber.Ctx) error {
	// Get authenticated user ID from context
	authUserID, _ := c.Locals("userID").(string)
	requestedID := c.Params("id")

	// Optional: Check if user is requesting their own profile or has admin rights
	if authUserID != requestedID {
		return c.Status(http.StatusForbidden).JSON(models.ErrorResponse{
			Error: "Forbidden",
		})
	}

	req := c.Locals("request").(models.UpdateProfileRequest)

	profile, err := h.userService.UpdateProfile(requestedID, &req)
	if err != nil {
		switch err.Error() {
		case "user not found":
			return c.Status(http.StatusNotFound).JSON(models.ErrorResponse{
				Error: "user not found",
			})
		case "invalid user ID format":
			return c.Status(http.StatusBadRequest).JSON(models.ErrorResponse{
				Error: "invalid user ID format",
			})
		default:
			fmt.Printf("Error updating user profile: %v\n", err)
			return c.Status(http.StatusInternalServerError).JSON(models.ErrorResponse{
				Error: "Failed to update user profile",
			})
		}
	}

	return c.Status(http.StatusOK).JSON(profile)
}
