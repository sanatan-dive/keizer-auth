package controllers

import (
	"errors"
	"fmt"

	"keizer-auth-api/internal/services"
	"keizer-auth-api/internal/utils"
	"keizer-auth-api/internal/validators"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthController struct {
	authService    *services.AuthService
	sessionService *services.SessionService
}

func NewAuthController(as *services.AuthService, ss *services.SessionService) *AuthController {
	return &AuthController{authService: as, sessionService: ss}
}

func (ac *AuthController) SignIn(c *fiber.Ctx) error {
	return c.SendString("Login successful!")
}

func (ac *AuthController) SignUp(c *fiber.Ctx) error {
	user := new(validators.SignUpUser)

	if err := c.BodyParser(user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if valid, errors := user.Validate(); !valid {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"errors": errors})
	}

	if err := ac.authService.RegisterUser(user); err != nil {
		if errors.Is(err, gorm.ErrCheckConstraintViolated) {
			return c.
				Status(fiber.StatusBadRequest).
				JSON(fiber.Map{
					"error": "Input validation error, please check your details",
				})
		}

		return c.
			Status(fiber.StatusInternalServerError).
			JSON(fiber.Map{"error": "Failed to sign up user"})
	}

	return c.JSON(fiber.Map{"message": "User Signed Up!"})
}

func (ac *AuthController) VerifyOTP(c *fiber.Ctx) error {
	verifyOtpBody := new(validators.VerifyOTP)

	if err := c.BodyParser(verifyOtpBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	isOtpValid, err := ac.authService.VerifyOTP(verifyOtpBody)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "OTP not found"})
		}
		if err.Error() == "otp expired" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "OTP expired"})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to verify OTP"})
	}
	if !isOtpValid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "OTP not valid"})
	}

	parsedUuid, err := uuid.Parse(verifyOtpBody.Id)
	if err != nil {
		return fmt.Errorf("error parsing uuid %w", err)
	}
	sessionId, err := ac.sessionService.CreateSession(parsedUuid)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create session"})
	}
	utils.SetSessionCookie(c, sessionId)

	return c.JSON(fiber.Map{"message": "OTP Verified!"})
}
