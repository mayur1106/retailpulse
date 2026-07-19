package controller

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"retailpulse/apps/api/internal/middleware"
	"retailpulse/apps/api/internal/repository"
)

type analyticsController struct {
	repo *repository.AnalyticsRepository
}

func RegisterAnalyticsRoutes(router fiber.Router, repo *repository.AnalyticsRepository) {
	c := analyticsController{repo: repo}
	router.Get("/dashboard", c.dashboard)
	router.Get("/growth", c.growth)
	router.Get("/health", c.health)
	router.Get("/data/:resource", c.resourceList)
	router.Post("/demo/generate", c.generateDemo)
}

func (c analyticsController) resourceList(ctx *fiber.Ctx) error {
	principal, err := middleware.Principal(ctx)
	if err != nil {
		return err
	}
	storeID, err := parseScopedStoreID(ctx)
	if err != nil {
		return err
	}
	if exists, err := c.repo.StoreExists(ctx.Context(), principal.OrganizationID, storeID); err != nil {
		return err
	} else if !exists {
		return fiber.NewError(fiber.StatusNotFound, "store not found")
	}
	items, err := c.repo.ResourceList(ctx.Context(), principal.OrganizationID, storeID, ctx.Params("resource"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "unsupported analytics resource")
	}
	return ctx.JSON(fiber.Map{"items": items})
}

func (c analyticsController) dashboard(ctx *fiber.Ctx) error {
	principal, err := middleware.Principal(ctx)
	if err != nil {
		return err
	}
	days, _ := strconv.Atoi(ctx.Query("days", "90"))
	if days < 7 {
		days = 7
	}
	if days > 730 {
		days = 730
	}
	storeID, err := parseScopedStoreID(ctx)
	if err != nil {
		return err
	}
	if exists, err := c.repo.StoreExists(ctx.Context(), principal.OrganizationID, storeID); err != nil {
		return err
	} else if !exists {
		return fiber.NewError(fiber.StatusNotFound, "store not found")
	}
	result, err := c.repo.Dashboard(ctx.Context(), principal.OrganizationID, storeID, days)
	if err != nil {
		return err
	}
	return ctx.JSON(result)
}

func (c analyticsController) growth(ctx *fiber.Ctx) error {
	principal, err := middleware.Principal(ctx)
	if err != nil {
		return err
	}
	days, _ := strconv.Atoi(ctx.Query("days", "90"))
	storeID, err := parseScopedStoreID(ctx)
	if err != nil {
		return err
	}
	if exists, err := c.repo.StoreExists(ctx.Context(), principal.OrganizationID, storeID); err != nil {
		return err
	} else if !exists {
		return fiber.NewError(fiber.StatusNotFound, "store not found")
	}
	result, err := c.repo.GrowthIntelligence(ctx.Context(), principal.OrganizationID, storeID, days)
	if err != nil {
		return err
	}
	return ctx.JSON(result)
}

func (c analyticsController) health(ctx *fiber.Ctx) error {
	principal, err := middleware.Principal(ctx)
	if err != nil {
		return err
	}
	days, _ := strconv.Atoi(ctx.Query("days", "90"))
	storeID, err := parseScopedStoreID(ctx)
	if err != nil {
		return err
	}
	if exists, err := c.repo.StoreExists(ctx.Context(), principal.OrganizationID, storeID); err != nil {
		return err
	} else if !exists {
		return fiber.NewError(fiber.StatusNotFound, "store not found")
	}
	result, err := c.repo.SellerHealth(ctx.Context(), principal.OrganizationID, storeID, days)
	if err != nil {
		return err
	}
	return ctx.JSON(result)
}

type demoRequest struct {
	StoreID string `json:"storeId"`
	Months  int    `json:"months"`
}

func (c analyticsController) generateDemo(ctx *fiber.Ctx) error {
	principal, err := middleware.Principal(ctx)
	if err != nil {
		return err
	}
	var request demoRequest
	if err := ctx.BodyParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid JSON body")
	}
	storeID, err := uuid.Parse(request.StoreID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid store id")
	}
	// Tenant ownership is checked before generation; no cross-organization store can be seeded.
	var exists bool
	stores, err := c.repoStoreCheck(ctx, principal.OrganizationID, storeID)
	if err != nil {
		return err
	}
	exists = stores
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "store not found")
	}
	result, err := c.repo.GenerateDemo(ctx.Context(), principal.OrganizationID, storeID, request.Months)
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusCreated).JSON(result)
}

func (c analyticsController) repoStoreCheck(ctx *fiber.Ctx, organizationID, storeID uuid.UUID) (bool, error) {
	// Dashboard repository intentionally exposes this narrow ownership check through its aggregate query method.
	return c.repo.StoreExists(ctx.Context(), organizationID, storeID)
}

func parseScopedStoreID(ctx *fiber.Ctx) (uuid.UUID, error) {
	raw := ctx.Query("storeId")
	if raw == "" {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "storeId is required to keep sandbox and production data separate")
	}
	storeID, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "invalid store id")
	}
	return storeID, nil
}
