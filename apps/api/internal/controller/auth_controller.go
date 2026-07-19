package controller

import (
	"github.com/gofiber/fiber/v2"

	"retailpulse/apps/api/internal/domain"
	"retailpulse/apps/api/internal/platform/validation"
	"retailpulse/apps/api/internal/service"
)

type authController struct {
	service *service.AuthService
}

func RegisterAuthRoutes(router fiber.Router, authService *service.AuthService) {
	controller := authController{service: authService}
	router.Post("/register", controller.register)
	router.Post("/login", controller.login)
	router.Post("/refresh", controller.refresh)
	router.Post("/logout", controller.logout)
}

type registerRequest struct {
	OrganizationName string `json:"organizationName"`
	Name             string `json:"name"`
	Email            string `json:"email"`
	Password         string `json:"password"`
	AccountType      string `json:"accountType"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

func (c authController) register(ctx *fiber.Ctx) error {
	var request registerRequest
	if err := ctx.BodyParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid JSON body")
	}
	if err := validation.RegisterRequest(request.OrganizationName, request.Name, request.Email, request.Password, request.AccountType); err != nil {
		return err
	}
	role := domain.RoleOwner
	if request.AccountType == "seller" {
		role = domain.RoleSeller
	}
	response, err := c.service.Register(ctx.Context(), service.RegisterInput{
		OrganizationName: request.OrganizationName,
		Name:             request.Name,
		Email:            request.Email,
		Password:         request.Password,
		Role:             role,
		IPAddress:        ctx.IP(),
		UserAgent:        ctx.Get(fiber.HeaderUserAgent),
	})
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusCreated).JSON(response)
}

func (c authController) login(ctx *fiber.Ctx) error {
	var request loginRequest
	if err := ctx.BodyParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid JSON body")
	}
	if err := validation.LoginRequest(request.Email, request.Password); err != nil {
		return err
	}
	response, err := c.service.Login(ctx.Context(), service.LoginInput{
		Email:     request.Email,
		Password:  request.Password,
		IPAddress: ctx.IP(),
		UserAgent: ctx.Get(fiber.HeaderUserAgent),
	})
	if err != nil {
		return err
	}
	return ctx.JSON(response)
}

func (c authController) logout(ctx *fiber.Ctx) error {
	var request refreshRequest
	if err := ctx.BodyParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid JSON body")
	}
	if err := c.service.Logout(ctx.Context(), request.RefreshToken); err != nil {
		return err
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

func (c authController) refresh(ctx *fiber.Ctx) error {
	var request refreshRequest
	if err := ctx.BodyParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid JSON body")
	}
	response, err := c.service.Refresh(ctx.Context(), request.RefreshToken, ctx.IP(), ctx.Get(fiber.HeaderUserAgent))
	if err != nil {
		return err
	}
	return ctx.JSON(response)
}
