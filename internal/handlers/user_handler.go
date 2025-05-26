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
			Error: "Failed to register user " + err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(models.RegisterResponse{
		Message: "User registered successfully",
		UserID:  user.ID,
	})
}

func (h *UserHandler) Login(c *fiber.Ctx) error {
	req := c.Locals("request").(models.LoginRequest)

	token, userId, err := h.userService.Login(req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to login")
		return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
			Error: err.Error(),
		})
	}

	return c.JSON(models.LoginResponse{
		UserID: userId,
		Token:  token,
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
				Error: err.Error(),
			})
		case "invalid user ID format":
			return c.Status(http.StatusBadRequest).JSON(models.ErrorResponse{
				Error: err.Error(),
			})
		default:
			fmt.Printf("Error fetching user profile: %v\n", err)
			return c.Status(http.StatusInternalServerError).JSON(models.ErrorResponse{
				Error: err.Error(),
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
	isAdmin, _ := c.Locals("isAdmin").(bool)
	if authUserID != requestedID && !isAdmin {
		return c.Status(http.StatusForbidden).JSON(models.ErrorResponse{
			Error: "Forbidden - you can only update your own profile",
		})
	}

	req := c.Locals("request").(models.UpdateProfileRequest)

	profile, err := h.userService.UpdateProfile(requestedID, &req)
	if err != nil {
		switch err.Error() {
		case "user not found":
			return c.Status(http.StatusNotFound).JSON(models.ErrorResponse{
				Error: err.Error(),
			})
		case "invalid user ID format":
			return c.Status(http.StatusBadRequest).JSON(models.ErrorResponse{
				Error: err.Error(),
			})
		default:
			fmt.Printf("Error updating user profile: %v\n", err)
			return c.Status(http.StatusInternalServerError).JSON(models.ErrorResponse{
				Error: err.Error(),
			})
		}
	}

	return c.Status(http.StatusOK).JSON(profile)
}

func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	// Get authenticated user ID from context
	requestedID := c.Params("id")

	err := h.userService.DeleteUser(requestedID)
	if err != nil {
		switch err.Error() {
		case "invalid user ID format":
			return c.Status(http.StatusBadRequest).JSON(models.ErrorResponse{
				Error: err.Error(),
			})
		default:
			fmt.Printf("Error deleting user: %v\n", err)
			return c.Status(http.StatusInternalServerError).JSON(models.ErrorResponse{
				Error: err.Error(),
			})
		}
	}

	return c.Status(http.StatusOK).JSON(models.SuccessResponse{
		Message: "User deleted successfully",
	})
}

func (h *UserHandler) GetAllUsers(c *fiber.Ctx) error {

	var filter models.UserFilter
	var pagination models.PaginationQuery

	// Selalu parse pagination dari query
	if err := c.QueryParser(&pagination); err != nil {
		return c.Status(http.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "invalid query params: " + err.Error(),
		})
	}

	// Jika POST, parse filter dari body
	if c.Method() == fiber.MethodPost {
		if err := c.BodyParser(&filter); err != nil {
			return c.Status(http.StatusBadRequest).JSON(models.ErrorResponse{
				Error: "invalid body: " + err.Error(),
			})
		}
	}

	users, err := h.userService.GetUsers(&filter, &pagination)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: err.Error(),
		})
	}

	return c.JSON(users)
}
