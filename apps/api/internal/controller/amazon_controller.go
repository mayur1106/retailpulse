package controller

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"retailpulse/apps/api/internal/middleware"
	"retailpulse/apps/api/internal/service"
)

type amazonController struct {
	service   *service.AmazonService
	webAppURL string
}

func RegisterAmazonRoutes(router fiber.Router, amazonService *service.AmazonService, webAppURL string) {
	controller := amazonController{service: amazonService, webAppURL: webAppURL}
	router.Get("/stores", controller.listStores)
	router.Get("/oauth/status", controller.oauthStatus)
	router.Post("/oauth/start", controller.startOAuth)
	router.Post("/sandbox/connect", controller.connectSandbox)
	router.Post("/stores/:storeId/import/orders", controller.importOrders)
	router.Post("/stores/:storeId/import/:dataset", controller.importDataset)
}

func (c amazonController) importDataset(ctx *fiber.Ctx) error {
	principal, err := middleware.Principal(ctx)
	if err != nil {
		return err
	}
	storeID, err := uuid.Parse(ctx.Params("storeId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid store id")
	}
	var request importOrdersRequest
	if len(ctx.Body()) > 0 {
		if err := ctx.BodyParser(&request); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid JSON body")
		}
	}
	result, err := c.service.ImportDataset(ctx.Context(), principal, storeID, ctx.Params("dataset"), request.MarketplaceID)
	if err != nil {
		return err
	}
	return ctx.JSON(result)
}

func (c amazonController) connectSandbox(ctx *fiber.Ctx) error {
	principal, err := middleware.Principal(ctx)
	if err != nil {
		return err
	}
	var request startOAuthRequest
	if len(ctx.Body()) > 0 {
		if err := ctx.BodyParser(&request); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid JSON body")
		}
	}
	store, err := c.service.ConnectSandbox(ctx.Context(), principal, request.Region)
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusCreated).JSON(store)
}

func RegisterAmazonOAuthCallback(router fiber.Router, amazonService *service.AmazonService, webAppURL string) {
	controller := amazonController{service: amazonService, webAppURL: webAppURL}
	router.Get("/callback", controller.completeOAuth)
}

type startOAuthRequest struct {
	Region        string `json:"region"`
	MarketplaceID string `json:"marketplaceId"`
}

type importOrdersRequest struct {
	MarketplaceID string `json:"marketplaceId"`
	CreatedAfter  string `json:"createdAfter"`
}

func (c amazonController) listStores(ctx *fiber.Ctx) error {
	principal, err := middleware.Principal(ctx)
	if err != nil {
		return err
	}
	stores, err := c.service.ListStores(ctx.Context(), principal)
	if err != nil {
		return err
	}
	return ctx.JSON(fiber.Map{"stores": stores})
}

func (c amazonController) startOAuth(ctx *fiber.Ctx) error {
	principal, err := middleware.Principal(ctx)
	if err != nil {
		return err
	}
	var request startOAuthRequest
	if len(ctx.Body()) > 0 {
		if err := ctx.BodyParser(&request); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid JSON body")
		}
	}
	response, err := c.service.StartOAuth(ctx.Context(), principal, request.Region, request.MarketplaceID)
	if err != nil {
		return err
	}
	return ctx.JSON(response)
}

func (c amazonController) oauthStatus(ctx *fiber.Ctx) error {
	if _, err := middleware.Principal(ctx); err != nil {
		return err
	}
	return ctx.JSON(c.service.ConnectionStatus())
}

func (c amazonController) completeOAuth(ctx *fiber.Ctx) error {
	store, err := c.service.CompleteOAuth(ctx.Context(), service.AmazonOAuthCallbackInput{
		State:            ctx.Query("state"),
		OAuthCode:        ctx.Query("spapi_oauth_code"),
		SellingPartnerID: ctx.Query("selling_partner_id"),
	})
	if err != nil {
		return err
	}
	return ctx.Redirect(c.webAppURL+"/dashboard?amazonConnected="+store.ID.String(), fiber.StatusFound)
}

func (c amazonController) importOrders(ctx *fiber.Ctx) error {
	principal, err := middleware.Principal(ctx)
	if err != nil {
		return err
	}
	storeID, err := uuid.Parse(ctx.Params("storeId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid store id")
	}
	var request importOrdersRequest
	if len(ctx.Body()) > 0 {
		if err := ctx.BodyParser(&request); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid JSON body")
		}
	}
	createdAfter := time.Now().UTC().AddDate(0, 0, -7)
	if request.CreatedAfter != "" {
		parsed, err := time.Parse(time.RFC3339, request.CreatedAfter)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "createdAfter must be RFC3339")
		}
		createdAfter = parsed
	}
	result, err := c.service.ImportOrders(ctx.Context(), principal, storeID, request.MarketplaceID, createdAfter)
	if err != nil {
		return err
	}
	return ctx.JSON(result)
}
