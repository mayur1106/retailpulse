package controller

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"retailpulse/apps/api/internal/middleware"
	"retailpulse/apps/api/internal/repository"
)

type commerceController struct {
	repo *repository.CommerceRepository
}

func RegisterCommerceRoutes(router fiber.Router, repo *repository.CommerceRepository) {
	c := commerceController{repo: repo}
	router.Post("/products", c.createProduct)
	router.Get("/products/:id", c.product)
	router.Put("/products/:id", c.updateProduct)
	router.Patch("/products/:id", c.updateProduct)
	router.Delete("/products/:id", c.deleteProduct)
	router.Post("/inventory", c.createInventoryItem)
	router.Get("/inventory/:id", c.inventoryItem)
	router.Put("/inventory/:id", c.updateInventoryItem)
	router.Patch("/inventory/:id", c.updateInventoryItem)
	router.Delete("/inventory/:id", c.deleteInventoryItem)
	router.Get("/cms", c.cmsConfig)
	router.Put("/cms", c.saveCMSConfig)
	router.Patch("/cms", c.saveCMSConfig)
	router.Post("/demo/generate", c.generateDemo)
	router.Get("/:resource", c.list)
}

func RegisterStorefrontRoutes(router fiber.Router, repo *repository.CommerceRepository) {
	c := commerceController{repo: repo}
	router.Get("/:storeSlug/config", c.storefrontConfig)
	router.Get("/:storeSlug/settings", c.storefrontSettings)
	router.Get("/:storeSlug/categories", c.storefrontCategories)
	router.Get("/:storeSlug/products", c.storefrontProducts)
	router.Get("/:storeSlug/products/:productSlug", c.storefrontProduct)
	router.Get("/:storeSlug/cart", c.getCart)
	router.Post("/:storeSlug/cart", c.createCart)
	router.Delete("/:storeSlug/cart", c.clearCart)
	router.Post("/:storeSlug/cart/items", c.addCartItem)
	router.Patch("/:storeSlug/cart/items/:itemID", c.updateCartItem)
	router.Delete("/:storeSlug/cart/items/:itemID", c.removeCartItem)
	router.Post("/:storeSlug/cart/checkout-started", c.cartCheckoutStarted)
	router.Post("/:storeSlug/checkout", c.checkout)
}

func (c commerceController) list(ctx *fiber.Ctx) error {
	principal, err := middleware.Principal(ctx)
	if err != nil {
		return err
	}
	items, err := c.repo.List(ctx.Context(), principal.OrganizationID, ctx.Params("resource"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return ctx.JSON(fiber.Map{"items": items})
}

func (c commerceController) createProduct(ctx *fiber.Ctx) error {
	principal, err := middleware.Principal(ctx)
	if err != nil {
		return err
	}
	var input repository.CommerceProductInput
	if err := ctx.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid JSON body")
	}
	item, err := c.repo.CreateProduct(ctx.Context(), principal.OrganizationID, input)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return ctx.Status(fiber.StatusCreated).JSON(item)
}

func (c commerceController) product(ctx *fiber.Ctx) error {
	principal, err := middleware.Principal(ctx)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid product id")
	}
	item, err := c.repo.Product(ctx.Context(), principal.OrganizationID, id)
	if err == pgx.ErrNoRows {
		return fiber.NewError(fiber.StatusNotFound, "product not found")
	}
	if err != nil {
		return err
	}
	return ctx.JSON(item)
}

func (c commerceController) updateProduct(ctx *fiber.Ctx) error {
	principal, err := middleware.Principal(ctx)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid product id")
	}
	var input repository.CommerceProductInput
	if err := ctx.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid JSON body")
	}
	item, err := c.repo.UpdateProduct(ctx.Context(), principal.OrganizationID, id, input)
	if err == pgx.ErrNoRows {
		return fiber.NewError(fiber.StatusNotFound, "product not found")
	}
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return ctx.JSON(item)
}

func (c commerceController) deleteProduct(ctx *fiber.Ctx) error {
	principal, err := middleware.Principal(ctx)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid product id")
	}
	err = c.repo.DeleteProduct(ctx.Context(), principal.OrganizationID, id)
	if err == pgx.ErrNoRows {
		return fiber.NewError(fiber.StatusNotFound, "product not found")
	}
	if err != nil {
		return err
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

func (c commerceController) cmsConfig(ctx *fiber.Ctx) error {
	principal, err := middleware.Principal(ctx)
	if err != nil {
		return err
	}
	item, err := c.repo.CMSConfig(ctx.Context(), principal.OrganizationID)
	if err != nil {
		return err
	}
	return ctx.JSON(item)
}

func (c commerceController) saveCMSConfig(ctx *fiber.Ctx) error {
	principal, err := middleware.Principal(ctx)
	if err != nil {
		return err
	}
	var input repository.CommerceCMSConfigInput
	if err := ctx.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid JSON body")
	}
	item, err := c.repo.SaveCMSConfig(ctx.Context(), principal.OrganizationID, input)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return ctx.JSON(item)
}

func (c commerceController) createInventoryItem(ctx *fiber.Ctx) error {
	principal, err := middleware.Principal(ctx)
	if err != nil {
		return err
	}
	var input repository.CommerceInventoryInput
	if err := ctx.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid JSON body")
	}
	item, err := c.repo.CreateInventoryItem(ctx.Context(), principal.OrganizationID, input)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return ctx.Status(fiber.StatusCreated).JSON(item)
}

func (c commerceController) inventoryItem(ctx *fiber.Ctx) error {
	principal, err := middleware.Principal(ctx)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid inventory item id")
	}
	item, err := c.repo.InventoryItem(ctx.Context(), principal.OrganizationID, id)
	if err == pgx.ErrNoRows {
		return fiber.NewError(fiber.StatusNotFound, "inventory item not found")
	}
	if err != nil {
		return err
	}
	return ctx.JSON(item)
}

func (c commerceController) updateInventoryItem(ctx *fiber.Ctx) error {
	principal, err := middleware.Principal(ctx)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid inventory item id")
	}
	var input repository.CommerceInventoryInput
	if err := ctx.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid JSON body")
	}
	item, err := c.repo.UpdateInventoryItem(ctx.Context(), principal.OrganizationID, id, input)
	if err == pgx.ErrNoRows {
		return fiber.NewError(fiber.StatusNotFound, "inventory item not found")
	}
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return ctx.JSON(item)
}

func (c commerceController) deleteInventoryItem(ctx *fiber.Ctx) error {
	principal, err := middleware.Principal(ctx)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid inventory item id")
	}
	err = c.repo.DeleteInventoryItem(ctx.Context(), principal.OrganizationID, id)
	if err == pgx.ErrNoRows {
		return fiber.NewError(fiber.StatusNotFound, "inventory item not found")
	}
	if err != nil {
		return err
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

func (c commerceController) generateDemo(ctx *fiber.Ctx) error {
	principal, err := middleware.Principal(ctx)
	if err != nil {
		return err
	}
	var input struct {
		Months int `json:"months"`
	}
	_ = ctx.BodyParser(&input)
	result, err := c.repo.GenerateDemo(ctx.Context(), principal.OrganizationID, input.Months)
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusCreated).JSON(result)
}

func (c commerceController) storefrontSettings(ctx *fiber.Ctx) error {
	item, err := c.repo.StorefrontSettings(ctx.Context(), ctx.Params("storeSlug"))
	if err == pgx.ErrNoRows {
		return fiber.NewError(fiber.StatusNotFound, "store not found")
	}
	if err != nil {
		return err
	}
	return ctx.JSON(item)
}

func (c commerceController) storefrontConfig(ctx *fiber.Ctx) error {
	item, err := c.repo.StorefrontConfig(ctx.Context(), ctx.Params("storeSlug"))
	if err == pgx.ErrNoRows {
		return fiber.NewError(fiber.StatusNotFound, "store not found")
	}
	if err != nil {
		return err
	}
	return ctx.JSON(item)
}

func (c commerceController) storefrontCategories(ctx *fiber.Ctx) error {
	items, err := c.repo.StorefrontCategories(ctx.Context(), ctx.Params("storeSlug"))
	if err != nil {
		return err
	}
	return ctx.JSON(fiber.Map{"items": items})
}

func (c commerceController) storefrontProducts(ctx *fiber.Ctx) error {
	limit, _ := strconv.Atoi(ctx.Query("limit", "48"))
	items, err := c.repo.StorefrontProducts(ctx.Context(), ctx.Params("storeSlug"), ctx.Query("category"), limit)
	if err != nil {
		return err
	}
	return ctx.JSON(fiber.Map{"items": items})
}

func (c commerceController) storefrontProduct(ctx *fiber.Ctx) error {
	item, err := c.repo.StorefrontProduct(ctx.Context(), ctx.Params("storeSlug"), ctx.Params("productSlug"))
	if err == pgx.ErrNoRows {
		return fiber.NewError(fiber.StatusNotFound, "product not found")
	}
	if err != nil {
		return err
	}
	return ctx.JSON(item)
}

func (c commerceController) getCart(ctx *fiber.Ctx) error {
	item, err := c.repo.GetCart(ctx.Context(), ctx.Params("storeSlug"), repository.CartRequest{
		CartToken: ctx.Query("cartToken"),
		VisitorID: ctx.Query("visitorId"),
	})
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return ctx.JSON(item)
}

func (c commerceController) createCart(ctx *fiber.Ctx) error {
	var input repository.CartRequest
	_ = ctx.BodyParser(&input)
	item, err := c.repo.GetCart(ctx.Context(), ctx.Params("storeSlug"), input)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return ctx.JSON(item)
}

func (c commerceController) addCartItem(ctx *fiber.Ctx) error {
	var input repository.CartItemInput
	if err := ctx.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid JSON body")
	}
	item, err := c.repo.AddCartItem(ctx.Context(), ctx.Params("storeSlug"), input)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return ctx.Status(fiber.StatusCreated).JSON(item)
}

func (c commerceController) updateCartItem(ctx *fiber.Ctx) error {
	itemID, err := uuid.Parse(ctx.Params("itemID"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid cart item id")
	}
	var input struct {
		CartToken string `json:"cartToken"`
		Quantity  int    `json:"quantity"`
	}
	if err := ctx.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid JSON body")
	}
	item, err := c.repo.UpdateCartItem(ctx.Context(), ctx.Params("storeSlug"), input.CartToken, itemID, input.Quantity)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return ctx.JSON(item)
}

func (c commerceController) removeCartItem(ctx *fiber.Ctx) error {
	itemID, err := uuid.Parse(ctx.Params("itemID"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid cart item id")
	}
	item, err := c.repo.RemoveCartItem(ctx.Context(), ctx.Params("storeSlug"), ctx.Query("cartToken"), itemID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return ctx.JSON(item)
}

func (c commerceController) clearCart(ctx *fiber.Ctx) error {
	item, err := c.repo.ClearCart(ctx.Context(), ctx.Params("storeSlug"), ctx.Query("cartToken"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return ctx.JSON(item)
}

func (c commerceController) cartCheckoutStarted(ctx *fiber.Ctx) error {
	var input repository.CartRequest
	if err := ctx.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid JSON body")
	}
	item, err := c.repo.MarkCartCheckoutStarted(ctx.Context(), ctx.Params("storeSlug"), input)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return ctx.JSON(item)
}

func (c commerceController) checkout(ctx *fiber.Ctx) error {
	var input repository.CheckoutRequest
	if err := ctx.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid JSON body")
	}
	result, err := c.repo.Checkout(ctx.Context(), ctx.Params("storeSlug"), input)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return ctx.Status(fiber.StatusCreated).JSON(result)
}
