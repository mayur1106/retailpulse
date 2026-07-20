package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CommerceRepository struct{ pool *pgxpool.Pool }

func NewCommerceRepository(pool *pgxpool.Pool) *CommerceRepository {
	return &CommerceRepository{pool: pool}
}

type CommerceProductInput struct {
	Title          string                        `json:"title"`
	Slug           string                        `json:"slug"`
	Description    string                        `json:"description"`
	CategorySlug   string                        `json:"categorySlug"`
	SKU            string                        `json:"sku"`
	Brand          string                        `json:"brand"`
	Status         string                        `json:"status"`
	Price          float64                       `json:"price"`
	CompareAtPrice float64                       `json:"compareAtPrice"`
	CostPrice      float64                       `json:"costPrice"`
	Images         []string                      `json:"images"`
	Colors         []string                      `json:"colors"`
	Sizes          []string                      `json:"sizes"`
	StockQuantity  int                           `json:"stockQuantity"`
	IsFeatured     bool                          `json:"isFeatured"`
	ChannelCodes   []string                      `json:"channelCodes"`
	Variants       []CommerceProductVariantInput `json:"variants"`
}

type CommerceProductVariantInput struct {
	ID             string   `json:"id"`
	SKU            string   `json:"sku"`
	Color          string   `json:"color"`
	Size           string   `json:"size"`
	Price          *float64 `json:"price"`
	CompareAtPrice *float64 `json:"compareAtPrice"`
	CostPrice      *float64 `json:"costPrice"`
	StockQuantity  *int     `json:"stockQuantity"`
}

type CommerceInventoryInput struct {
	ProductID         string   `json:"productId"`
	Title             string   `json:"title"`
	SKU               string   `json:"sku"`
	Color             string   `json:"color"`
	Size              string   `json:"size"`
	Price             float64  `json:"price"`
	CompareAtPrice    float64  `json:"compareAtPrice"`
	CostPrice         float64  `json:"costPrice"`
	StockQuantity     int      `json:"stockQuantity"`
	ReservedQuantity  int      `json:"reservedQuantity"`
	LowStockThreshold int      `json:"lowStockThreshold"`
	Status            string   `json:"status"`
	ChannelCodes      []string `json:"channelCodes"`
}

type CommerceCMSConfigInput struct {
	Store           CommerceStoreInput             `json:"store"`
	Categories      []CommerceCategoryInput        `json:"categories"`
	Sections        []CommerceHomepageSectionInput `json:"sections"`
	PaymentMethods  []CommercePaymentMethodInput   `json:"paymentMethods"`
	ShippingOptions []CommerceShippingZoneInput    `json:"shippingOptions"`
}

type CommerceStoreInput struct {
	Name         string         `json:"name"`
	Slug         string         `json:"slug"`
	Domain       string         `json:"domain"`
	LogoURL      string         `json:"logoUrl"`
	CurrencyCode string         `json:"currencyCode"`
	CountryCode  string         `json:"countryCode"`
	Timezone     string         `json:"timezone"`
	Status       string         `json:"status"`
	Settings     map[string]any `json:"settings"`
}

type CommerceCategoryInput struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	ImageURL    string `json:"imageUrl"`
	SortOrder   int    `json:"sortOrder"`
	Status      string `json:"status"`
}

type CommerceHomepageSectionInput struct {
	ID            string         `json:"id"`
	SectionKey    string         `json:"sectionKey"`
	SectionType   string         `json:"sectionType"`
	Title         string         `json:"title"`
	Subtitle      string         `json:"subtitle"`
	Layout        string         `json:"layout"`
	ImageURL      string         `json:"imageUrl"`
	CTALabel      string         `json:"ctaLabel"`
	CTAHref       string         `json:"ctaHref"`
	CategorySlug  string         `json:"categorySlug"`
	ProductSource string         `json:"productSource"`
	MaxItems      int            `json:"maxItems"`
	Content       map[string]any `json:"content"`
	SortOrder     int            `json:"sortOrder"`
	Status        string         `json:"status"`
}

type CommercePaymentMethodInput struct {
	ID           string         `json:"id"`
	Code         string         `json:"code"`
	Name         string         `json:"name"`
	Provider     string         `json:"provider"`
	Instructions string         `json:"instructions"`
	SortOrder    int            `json:"sortOrder"`
	Status       string         `json:"status"`
	Settings     map[string]any `json:"settings"`
}

type CommerceShippingZoneInput struct {
	ID                    string   `json:"id"`
	Name                  string   `json:"name"`
	CountryCode           string   `json:"countryCode"`
	RegionCodes           []string `json:"regionCodes"`
	RateType              string   `json:"rateType"`
	Rate                  float64  `json:"rate"`
	FreeShippingThreshold float64  `json:"freeShippingThreshold"`
	EstimatedDaysMin      int      `json:"estimatedDaysMin"`
	EstimatedDaysMax      int      `json:"estimatedDaysMax"`
	CODEnabled            bool     `json:"codEnabled"`
	Status                string   `json:"status"`
}

type CartRequest struct {
	CartToken string         `json:"cartToken"`
	VisitorID string         `json:"visitorId"`
	Email     string         `json:"email"`
	Name      string         `json:"name"`
	Phone     string         `json:"phone"`
	Metadata  map[string]any `json:"metadata"`
}

type CartItemInput struct {
	CartToken string `json:"cartToken"`
	VisitorID string `json:"visitorId"`
	ProductID string `json:"productId"`
	VariantID string `json:"variantId"`
	Quantity  int    `json:"quantity"`
}

type CheckoutItem struct {
	ProductID string `json:"productId"`
	VariantID string `json:"variantId"`
	Quantity  int    `json:"quantity"`
}

type CheckoutRequest struct {
	CartToken       string         `json:"cartToken"`
	Email           string         `json:"email"`
	Name            string         `json:"name"`
	Phone           string         `json:"phone"`
	CouponCode      string         `json:"couponCode"`
	ShippingAddress map[string]any `json:"shippingAddress"`
	Items           []CheckoutItem `json:"items"`
}

func (r *CommerceRepository) DefaultStore(ctx context.Context, organizationID uuid.UUID) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `select id from commerce_stores where organization_id=$1 order by created_at limit 1`, organizationID).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != pgx.ErrNoRows {
		return uuid.Nil, err
	}
	err = r.pool.QueryRow(ctx, `
		insert into commerce_stores (organization_id,name,slug,status,settings)
		values ($1,'Rangavali','rangavali','live',$2)
		returning id
	`, organizationID, map[string]any{
		"announcement": "10% off first order · Free shipping above ₹1,999",
		"supportEmail": "support@rangavali.test",
		"returnPolicy": "Easy 7-day returns on unworn styles.",
	}).Scan(&id)
	return id, err
}

func (r *CommerceRepository) List(ctx context.Context, organizationID uuid.UUID, resource string) ([]map[string]any, error) {
	storeID, err := r.DefaultStore(ctx, organizationID)
	if err != nil {
		return nil, err
	}
	if err := r.ensureDefaultChannels(ctx, organizationID, storeID); err != nil {
		return nil, err
	}
	if err := r.ensureDefaultCMS(ctx, organizationID, storeID); err != nil {
		return nil, err
	}
	queries := map[string]string{
		"store":             `select id::text,name,slug,domain,logo_url,currency_code,country_code,timezone,status,settings,created_at,updated_at from commerce_stores where organization_id=$1 and id=$2`,
		"categories":        `select id::text,name,slug,description,image_url,sort_order,status,created_at,updated_at from commerce_categories where organization_id=$1 and store_id=$2 order by sort_order,name`,
		"homepage-sections": `select id::text,section_key,section_type,title,subtitle,layout,image_url,cta_label,cta_href,category_slug,product_source,max_items,content,sort_order,status,created_at,updated_at from commerce_homepage_sections where organization_id=$1 and store_id=$2 order by sort_order,title`,
		"payments":          `select id::text,code,name,provider,instructions,sort_order,status,settings,created_at,updated_at from commerce_payment_methods where organization_id=$1 and store_id=$2 order by sort_order,name`,
		"products":          `select p.id::text,p.title,p.slug,coalesce(c.name,'') category,coalesce(c.slug,'') category_slug,p.sku,p.brand,p.status,p.price,p.compare_at_price,p.cost_price,p.currency_code,p.images,p.options,p.tags,p.is_featured,coalesce(sum(v.stock_quantity-v.reserved_quantity),0)::int available_stock,coalesce(max(case when sc.code='website' then cl.listing_status end),'not_listed') website_status,coalesce(max(case when sc.code='amazon' then cl.listing_status end),'not_listed') amazon_status,coalesce(max(case when sc.code='google' then cl.listing_status end),'not_listed') google_status,coalesce(max(case when sc.code='meta' then cl.listing_status end),'not_listed') meta_status,coalesce(string_agg(distinct sc.code, ',') filter (where cl.listing_status='active'),'') listed_channel_codes,p.updated_at from commerce_products p left join commerce_categories c on c.id=p.category_id left join commerce_product_variants v on v.product_id=p.id left join commerce_channel_listings cl on cl.product_id=p.id left join commerce_sales_channels sc on sc.id=cl.channel_id where p.organization_id=$1 and p.store_id=$2 and p.status <> 'deleted' group by p.id,c.name,c.slug order by p.updated_at desc limit 250`,
		"inventory":         `select p.id::text product_id,p.title,p.slug,v.id::text variant_id,v.title variant,v.sku,v.color,v.size,v.price,v.compare_at_price,v.cost_price,v.stock_quantity,v.reserved_quantity,(v.stock_quantity-v.reserved_quantity) available,v.low_stock_threshold,v.status,coalesce(string_agg(distinct sc.code, ',') filter (where cl.listing_status='active'),'') listed_channel_codes from commerce_product_variants v join commerce_products p on p.id=v.product_id left join commerce_channel_listings cl on cl.variant_id=v.id left join commerce_sales_channels sc on sc.id=cl.channel_id where v.organization_id=$1 and v.store_id=$2 group by p.id,v.id order by available asc,p.title limit 300`,
		"orders":            `select o.id::text,o.order_number,o.channel_code,o.external_order_id,o.status,o.payment_status,o.fulfillment_status,o.total,o.currency_code,o.coupon_code,o.tracking_number,o.placed_at,coalesce(c.name,'Guest') customer,coalesce(c.email,'') email,coalesce(c.city,'') city,coalesce(c.region_code,'') region_code,count(oi.id)::int items from commerce_orders o left join commerce_customers c on c.id=o.customer_id left join commerce_order_items oi on oi.order_id=o.id where o.organization_id=$1 and o.store_id=$2 group by o.id,c.id order by o.placed_at desc limit 250`,
		"customers":         `select id::text,email,name,phone,country_code,region_code,city,total_spent,order_count,tags,created_at,updated_at from commerce_customers where organization_id=$1 and store_id=$2 order by total_spent desc limit 250`,
		"coupons":           `select id::text,code,name,discount_type,discount_value,minimum_order_value,usage_limit,used_count,starts_at,expires_at,status,created_at,updated_at from commerce_coupons where organization_id=$1 and store_id=$2 order by created_at desc`,
		"shipping":          `select id::text,name,country_code,region_codes,rate_type,rate,free_shipping_threshold,estimated_days_min,estimated_days_max,cod_enabled,status,created_at,updated_at from commerce_shipping_zones where organization_id=$1 and store_id=$2 order by name`,
		"returns":           `select r.id::text,r.reason,r.status,r.refund_amount,r.requested_at,coalesce(p.title,'') product,coalesce(o.order_number,'') order_number from commerce_returns r left join commerce_products p on p.id=r.product_id left join commerce_orders o on o.id=r.order_id where r.organization_id=$1 and r.store_id=$2 order by r.requested_at desc limit 200`,
		"channels":          `select sc.id::text,sc.code,sc.name,sc.channel_type,sc.status,count(distinct cl.product_id)::int listed_products,count(*) filter (where cl.sync_status='pending')::int pending_syncs,count(*) filter (where cl.sync_status='failed')::int failed_syncs,max(cl.last_synced_at) last_synced_at,sc.updated_at from commerce_sales_channels sc left join commerce_channel_listings cl on cl.channel_id=sc.id where sc.organization_id=$1 and sc.store_id=$2 group by sc.id order by case sc.code when 'website' then 1 when 'amazon' then 2 when 'google' then 3 when 'meta' then 4 when 'myntra' then 5 else 9 end`,
		"analytics":         `with sales as (select coalesce(sum(total),0) revenue,count(*)::int orders from commerce_orders where organization_id=$1 and store_id=$2 and placed_at>=current_date-89), product_sales as (select p.id,p.title,coalesce(sum(oi.quantity),0)::int units,coalesce(sum(oi.total_price),0) revenue from commerce_products p left join commerce_order_items oi on oi.product_id=p.id left join commerce_orders o on o.id=oi.order_id and o.placed_at>=current_date-89 where p.organization_id=$1 and p.store_id=$2 group by p.id), region_sales as (select coalesce(c.region_code,'Unknown') region,coalesce(c.city,'Unknown') city,count(o.id)::int orders,coalesce(sum(o.total),0) revenue from commerce_orders o left join commerce_customers c on c.id=o.customer_id where o.organization_id=$1 and o.store_id=$2 and o.placed_at>=current_date-89 group by c.region_code,c.city) select 'summary' kind,'Revenue' label,(select revenue from sales)::text value,'' secondary union all select 'summary','Orders',(select orders from sales)::text,'' union all select 'product',title,revenue::text,units::text from product_sales order by 1,3 desc limit 50`,
	}
	query, ok := queries[resource]
	if !ok {
		return nil, fmt.Errorf("unsupported commerce resource")
	}
	return queryMaps(ctx, r.pool, query, organizationID, storeID)
}

func (r *CommerceRepository) StorefrontProducts(ctx context.Context, slug, category string, limit int) ([]map[string]any, error) {
	if limit <= 0 || limit > 100 {
		limit = 48
	}
	args := []any{slug, limit}
	filter := ""
	if category != "" && category != "all" && category != "all-styles" {
		args = append(args, category)
		filter = " and c.slug=$3"
	}
	return queryMaps(ctx, r.pool, `
		select p.id::text,p.title name,p.slug,coalesce(c.name,'') category,
			coalesce((select min(v.price) from commerce_product_variants v where v.product_id=p.id and v.status='active'),p.price) price,
			coalesce((select min(v.compare_at_price) from commerce_product_variants v where v.product_id=p.id and v.status='active'),p.compare_at_price) original,
			p.currency_code,p.images,p.options,
			coalesce((p.images->>0),'/images/catalog/look-1.jpg') image,
			case when p.is_featured then 'BESTSELLER' when p.created_at > now()-interval '21 days' then 'NEW' else '' end badge,
			4.2 + (abs(hashtext(p.id::text)) % 8)::numeric/10 rating,
			38 + (abs(hashtext(p.slug)) % 680) reviews,
			coalesce((select color from commerce_product_variants v where v.product_id=p.id and v.color<>'' limit 1),'Classic') color,
			coalesce((select min(id::text) from commerce_product_variants v where v.product_id=p.id and v.status='active'),'') variant_id,
			coalesce((select sum(stock_quantity-reserved_quantity)::int from commerce_product_variants v where v.product_id=p.id),0) available_stock
		from commerce_products p
		left join commerce_categories c on c.id=p.category_id
		join commerce_stores s on s.id=p.store_id
		join commerce_sales_channels website_channel on website_channel.store_id=s.id and website_channel.code='website'
		join commerce_channel_listings website_listing on website_listing.channel_id=website_channel.id and website_listing.product_id=p.id and website_listing.listing_status='active'
		where s.slug=$1 and s.status in ('live','draft') and p.status='active'`+filter+`
		order by p.is_featured desc,p.updated_at desc limit $2
	`, args...)
}

func (r *CommerceRepository) StorefrontProduct(ctx context.Context, storeSlug, productSlug string) (map[string]any, error) {
	items, err := queryMaps(ctx, r.pool, `
		select p.id::text,p.title name,p.slug,p.description,coalesce(c.name,'') category,p.price,p.compare_at_price original,p.currency_code,p.images,p.options,
			coalesce((p.images->>0),'/images/catalog/look-1.jpg') image,
			case when p.is_featured then 'BESTSELLER' else '' end badge,
			4.2 + (abs(hashtext(p.id::text)) % 8)::numeric/10 rating,
			38 + (abs(hashtext(p.slug)) % 680) reviews,
			coalesce((select color from commerce_product_variants v where v.product_id=p.id and v.color<>'' limit 1),'Classic') color,
			coalesce((select jsonb_agg(jsonb_build_object('id',v.id::text,'title',v.title,'sku',v.sku,'color',v.color,'size',v.size,'price',v.price,'compareAtPrice',v.compare_at_price,'costPrice',v.cost_price,'stockQuantity',v.stock_quantity-v.reserved_quantity) order by v.color,v.size) from commerce_product_variants v where v.product_id=p.id),'[]'::jsonb) variants
		from commerce_products p
		left join commerce_categories c on c.id=p.category_id
		join commerce_stores s on s.id=p.store_id
		join commerce_sales_channels website_channel on website_channel.store_id=s.id and website_channel.code='website'
		join commerce_channel_listings website_listing on website_listing.channel_id=website_channel.id and website_listing.product_id=p.id and website_listing.listing_status='active'
		where s.slug=$1 and (p.slug=$2 or p.id::text=$2) and p.status='active'
	`, storeSlug, productSlug)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, pgx.ErrNoRows
	}
	return items[0], nil
}

func (r *CommerceRepository) StorefrontSettings(ctx context.Context, slug string) (map[string]any, error) {
	items, err := queryMaps(ctx, r.pool, `select id::text,name,slug,domain,logo_url,currency_code,country_code,settings from commerce_stores where slug=$1`, slug)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, pgx.ErrNoRows
	}
	return items[0], nil
}

func (r *CommerceRepository) CMSConfig(ctx context.Context, organizationID uuid.UUID) (map[string]any, error) {
	storeID, err := r.DefaultStore(ctx, organizationID)
	if err != nil {
		return nil, err
	}
	if err := r.ensureDefaultChannels(ctx, organizationID, storeID); err != nil {
		return nil, err
	}
	if err := r.ensureDefaultCMS(ctx, organizationID, storeID); err != nil {
		return nil, err
	}
	store, err := queryMaps(ctx, r.pool, `select id::text,name,slug,domain,logo_url,currency_code,country_code,timezone,status,settings,updated_at from commerce_stores where organization_id=$1 and id=$2`, organizationID, storeID)
	if err != nil {
		return nil, err
	}
	categories, err := r.List(ctx, organizationID, "categories")
	if err != nil {
		return nil, err
	}
	sections, err := r.List(ctx, organizationID, "homepage-sections")
	if err != nil {
		return nil, err
	}
	payments, err := r.List(ctx, organizationID, "payments")
	if err != nil {
		return nil, err
	}
	shipping, err := r.List(ctx, organizationID, "shipping")
	if err != nil {
		return nil, err
	}
	channels, err := r.List(ctx, organizationID, "channels")
	if err != nil {
		return nil, err
	}
	return map[string]any{"store": store[0], "categories": categories, "sections": sections, "paymentMethods": payments, "shippingOptions": shipping, "channels": channels}, nil
}

func (r *CommerceRepository) SaveCMSConfig(ctx context.Context, organizationID uuid.UUID, input CommerceCMSConfigInput) (map[string]any, error) {
	storeID, err := r.DefaultStore(ctx, organizationID)
	if err != nil {
		return nil, err
	}
	if err := r.ensureDefaultCMS(ctx, organizationID, storeID); err != nil {
		return nil, err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	store := input.Store
	if store.Name == "" {
		store.Name = "Rangavali"
	}
	if store.Slug == "" {
		store.Slug = "rangavali"
	}
	if store.CurrencyCode == "" {
		store.CurrencyCode = "INR"
	}
	if store.CountryCode == "" {
		store.CountryCode = "IN"
	}
	if store.Timezone == "" {
		store.Timezone = "Asia/Kolkata"
	}
	if store.Status == "" {
		store.Status = "live"
	}
	if store.Settings == nil {
		store.Settings = map[string]any{}
	}
	settingsJSON, _ := json.Marshal(store.Settings)
	if _, err := tx.Exec(ctx, `
		update commerce_stores
		set name=$3,slug=$4,domain=$5,logo_url=$6,currency_code=$7,country_code=$8,timezone=$9,status=$10,settings=$11,updated_at=now()
		where organization_id=$1 and id=$2
	`, organizationID, storeID, store.Name, slugify(store.Slug), store.Domain, store.LogoURL, store.CurrencyCode, store.CountryCode, store.Timezone, store.Status, settingsJSON); err != nil {
		return nil, err
	}
	for _, query := range []string{
		`update commerce_categories set status='inactive',updated_at=now() where organization_id=$1 and store_id=$2`,
		`update commerce_homepage_sections set status='inactive',updated_at=now() where organization_id=$1 and store_id=$2`,
		`update commerce_payment_methods set status='inactive',updated_at=now() where organization_id=$1 and store_id=$2`,
		`update commerce_shipping_zones set status='inactive',updated_at=now() where organization_id=$1 and store_id=$2`,
	} {
		if _, err := tx.Exec(ctx, query, organizationID, storeID); err != nil {
			return nil, err
		}
	}
	for i, category := range input.Categories {
		if strings.TrimSpace(category.Name) == "" {
			continue
		}
		if category.Slug == "" {
			category.Slug = slugify(category.Name)
		}
		if category.Status == "" {
			category.Status = "active"
		}
		if category.SortOrder == 0 {
			category.SortOrder = i + 1
		}
		if _, err := tx.Exec(ctx, `
			insert into commerce_categories (organization_id,store_id,name,slug,description,image_url,sort_order,status)
			values ($1,$2,$3,$4,$5,$6,$7,$8)
			on conflict (store_id,slug) do update set name=excluded.name,description=excluded.description,image_url=excluded.image_url,sort_order=excluded.sort_order,status=excluded.status,updated_at=now()
		`, organizationID, storeID, category.Name, slugify(category.Slug), category.Description, category.ImageURL, category.SortOrder, category.Status); err != nil {
			return nil, err
		}
	}
	for i, section := range input.Sections {
		if strings.TrimSpace(section.SectionKey) == "" {
			section.SectionKey = slugify(section.Title)
		}
		if section.SectionKey == "" {
			continue
		}
		if section.SectionType == "" {
			section.SectionType = "product_grid"
		}
		if section.Layout == "" {
			section.Layout = "grid"
		}
		if section.ProductSource == "" {
			section.ProductSource = "all"
		}
		if section.MaxItems <= 0 {
			section.MaxItems = 12
		}
		if section.Status == "" {
			section.Status = "active"
		}
		if section.SortOrder == 0 {
			section.SortOrder = i + 1
		}
		if section.Content == nil {
			section.Content = map[string]any{}
		}
		contentJSON, _ := json.Marshal(section.Content)
		if _, err := tx.Exec(ctx, `
			insert into commerce_homepage_sections (organization_id,store_id,section_key,section_type,title,subtitle,layout,image_url,cta_label,cta_href,category_slug,product_source,max_items,content,sort_order,status)
			values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
			on conflict (store_id,section_key) do update set section_type=excluded.section_type,title=excluded.title,subtitle=excluded.subtitle,layout=excluded.layout,image_url=excluded.image_url,cta_label=excluded.cta_label,cta_href=excluded.cta_href,category_slug=excluded.category_slug,product_source=excluded.product_source,max_items=excluded.max_items,content=excluded.content,sort_order=excluded.sort_order,status=excluded.status,updated_at=now()
		`, organizationID, storeID, section.SectionKey, section.SectionType, section.Title, section.Subtitle, section.Layout, section.ImageURL, section.CTALabel, section.CTAHref, section.CategorySlug, section.ProductSource, section.MaxItems, contentJSON, section.SortOrder, section.Status); err != nil {
			return nil, err
		}
	}
	for i, payment := range input.PaymentMethods {
		if strings.TrimSpace(payment.Name) == "" {
			continue
		}
		if payment.Code == "" {
			payment.Code = slugify(payment.Name)
		}
		if payment.Provider == "" {
			payment.Provider = payment.Code
		}
		if payment.Status == "" {
			payment.Status = "inactive"
		}
		if payment.SortOrder == 0 {
			payment.SortOrder = i + 1
		}
		if payment.Settings == nil {
			payment.Settings = map[string]any{}
		}
		settingsJSON, _ := json.Marshal(payment.Settings)
		if _, err := tx.Exec(ctx, `
			insert into commerce_payment_methods (organization_id,store_id,code,name,provider,instructions,sort_order,status,settings)
			values ($1,$2,$3,$4,$5,$6,$7,$8,$9)
			on conflict (store_id,code) do update set name=excluded.name,provider=excluded.provider,instructions=excluded.instructions,sort_order=excluded.sort_order,status=excluded.status,settings=excluded.settings,updated_at=now()
		`, organizationID, storeID, strings.ToLower(payment.Code), payment.Name, payment.Provider, payment.Instructions, payment.SortOrder, payment.Status, settingsJSON); err != nil {
			return nil, err
		}
	}
	for _, shipping := range input.ShippingOptions {
		if strings.TrimSpace(shipping.Name) == "" {
			continue
		}
		if shipping.CountryCode == "" {
			shipping.CountryCode = "IN"
		}
		if shipping.RateType == "" {
			shipping.RateType = "flat"
		}
		if shipping.EstimatedDaysMin <= 0 {
			shipping.EstimatedDaysMin = 3
		}
		if shipping.EstimatedDaysMax <= 0 {
			shipping.EstimatedDaysMax = 7
		}
		if shipping.Status == "" {
			shipping.Status = "active"
		}
		if shipping.ID != "" {
			id, parseErr := uuid.Parse(shipping.ID)
			if parseErr != nil {
				return nil, fmt.Errorf("invalid shipping option id")
			}
			if _, err := tx.Exec(ctx, `
				update commerce_shipping_zones set name=$4,country_code=$5,region_codes=$6,rate_type=$7,rate=$8,free_shipping_threshold=$9,estimated_days_min=$10,estimated_days_max=$11,cod_enabled=$12,status=$13,updated_at=now()
				where organization_id=$1 and store_id=$2 and id=$3
			`, organizationID, storeID, id, shipping.Name, shipping.CountryCode, shipping.RegionCodes, shipping.RateType, shipping.Rate, shipping.FreeShippingThreshold, shipping.EstimatedDaysMin, shipping.EstimatedDaysMax, shipping.CODEnabled, shipping.Status); err != nil {
				return nil, err
			}
		} else if _, err := tx.Exec(ctx, `
			insert into commerce_shipping_zones (organization_id,store_id,name,country_code,region_codes,rate_type,rate,free_shipping_threshold,estimated_days_min,estimated_days_max,cod_enabled,status)
			values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		`, organizationID, storeID, shipping.Name, shipping.CountryCode, shipping.RegionCodes, shipping.RateType, shipping.Rate, shipping.FreeShippingThreshold, shipping.EstimatedDaysMin, shipping.EstimatedDaysMax, shipping.CODEnabled, shipping.Status); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.CMSConfig(ctx, organizationID)
}

func (r *CommerceRepository) StorefrontConfig(ctx context.Context, slug string) (map[string]any, error) {
	store, err := r.StorefrontSettings(ctx, slug)
	if err != nil {
		return nil, err
	}
	storeID, err := uuid.Parse(stringValue(store["id"]))
	if err != nil {
		return nil, err
	}
	var orgID uuid.UUID
	if err := r.pool.QueryRow(ctx, `select organization_id from commerce_stores where id=$1`, storeID).Scan(&orgID); err != nil {
		return nil, err
	}
	if err := r.ensureDefaultCMS(ctx, orgID, storeID); err != nil {
		return nil, err
	}
	categories, err := r.StorefrontCategories(ctx, slug)
	if err != nil {
		return nil, err
	}
	sections, err := queryMaps(ctx, r.pool, `select id::text,section_key,section_type,title,subtitle,layout,image_url,cta_label,cta_href,category_slug,product_source,max_items,content,sort_order from commerce_homepage_sections where store_id=$1 and status='active' order by sort_order,title`, storeID)
	if err != nil {
		return nil, err
	}
	payments, err := queryMaps(ctx, r.pool, `select id::text,code,name,provider,instructions,sort_order,settings from commerce_payment_methods where store_id=$1 and status='active' order by sort_order,name`, storeID)
	if err != nil {
		return nil, err
	}
	shipping, err := queryMaps(ctx, r.pool, `select id::text,name,country_code,region_codes,rate_type,rate,free_shipping_threshold,estimated_days_min,estimated_days_max,cod_enabled from commerce_shipping_zones where store_id=$1 and status='active' order by name`, storeID)
	if err != nil {
		return nil, err
	}
	return map[string]any{"store": store, "categories": categories, "sections": sections, "paymentMethods": payments, "shippingOptions": shipping}, nil
}

func (r *CommerceRepository) StorefrontCategories(ctx context.Context, slug string) ([]map[string]any, error) {
	return queryMaps(ctx, r.pool, `select c.id::text,c.name,c.slug,c.description,c.image_url image,c.sort_order from commerce_categories c join commerce_stores s on s.id=c.store_id where s.slug=$1 and c.status='active' order by c.sort_order,c.name`, slug)
}

func (r *CommerceRepository) GetCart(ctx context.Context, storeSlug string, request CartRequest) (map[string]any, error) {
	cartID, _, _, err := r.ensureCart(ctx, storeSlug, request)
	if err != nil {
		return nil, err
	}
	if err := r.refreshCartTotals(ctx, cartID); err != nil {
		return nil, err
	}
	return r.cart(ctx, cartID)
}

func (r *CommerceRepository) AddCartItem(ctx context.Context, storeSlug string, input CartItemInput) (map[string]any, error) {
	if input.Quantity <= 0 {
		input.Quantity = 1
	}
	cartID, orgID, storeID, err := r.ensureCart(ctx, storeSlug, CartRequest{CartToken: input.CartToken, VisitorID: input.VisitorID})
	if err != nil {
		return nil, err
	}
	productID, err := uuid.Parse(input.ProductID)
	if err != nil {
		return nil, fmt.Errorf("valid productId is required")
	}
	var variantID uuid.UUID
	var title, variantTitle, sku, imageURL string
	var price float64
	if strings.TrimSpace(input.VariantID) != "" {
		variantID, err = uuid.Parse(input.VariantID)
		if err != nil {
			return nil, fmt.Errorf("valid variantId is required")
		}
		err = r.pool.QueryRow(ctx, `
			select p.title,v.title,v.sku,coalesce(p.images->>0,'/images/catalog/look-1.jpg'),v.price
			from commerce_products p
			join commerce_product_variants v on v.product_id=p.id
			join commerce_sales_channels sc on sc.store_id=p.store_id and sc.code='website'
			join commerce_channel_listings cl on cl.channel_id=sc.id and cl.product_id=p.id and cl.listing_status='active'
			where p.organization_id=$1 and p.store_id=$2 and p.id=$3 and v.id=$4 and p.status='active' and v.status='active'
		`, orgID, storeID, productID, variantID).Scan(&title, &variantTitle, &sku, &imageURL, &price)
	} else {
		err = r.pool.QueryRow(ctx, `
			select p.title,v.id,v.title,v.sku,coalesce(p.images->>0,'/images/catalog/look-1.jpg'),v.price
			from commerce_products p
			join commerce_product_variants v on v.product_id=p.id and v.status='active'
			join commerce_sales_channels sc on sc.store_id=p.store_id and sc.code='website'
			join commerce_channel_listings cl on cl.channel_id=sc.id and cl.product_id=p.id and cl.listing_status='active'
			where p.organization_id=$1 and p.store_id=$2 and p.id=$3 and p.status='active'
			order by v.stock_quantity-v.reserved_quantity desc limit 1
		`, orgID, storeID, productID).Scan(&title, &variantID, &variantTitle, &sku, &imageURL, &price)
	}
	if err != nil {
		return nil, fmt.Errorf("item unavailable")
	}
	var existingID uuid.UUID
	err = r.pool.QueryRow(ctx, `
		select id from commerce_cart_items
		where cart_id=$1 and product_id=$2 and variant_id=$3 and removed_at is null
	`, cartID, productID, variantID).Scan(&existingID)
	if err == nil {
		_, err = r.pool.Exec(ctx, `
			update commerce_cart_items
			set quantity=quantity+$2,unit_price=$3,product_title=$4,variant_title=$5,sku=$6,image_url=$7,updated_at=now()
			where id=$1
		`, existingID, input.Quantity, price, title, variantTitle, sku, imageURL)
	} else if err == pgx.ErrNoRows {
		_, err = r.pool.Exec(ctx, `
			insert into commerce_cart_items (organization_id,store_id,cart_id,product_id,variant_id,product_title,variant_title,sku,image_url,unit_price,quantity)
			values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		`, orgID, storeID, cartID, productID, variantID, title, variantTitle, sku, imageURL, price, input.Quantity)
	}
	if err != nil {
		return nil, err
	}
	if err := r.refreshCartTotals(ctx, cartID); err != nil {
		return nil, err
	}
	return r.cart(ctx, cartID)
}

func (r *CommerceRepository) UpdateCartItem(ctx context.Context, storeSlug, cartToken string, itemID uuid.UUID, quantity int) (map[string]any, error) {
	cartID, _, _, err := r.ensureCart(ctx, storeSlug, CartRequest{CartToken: cartToken})
	if err != nil {
		return nil, err
	}
	if quantity <= 0 {
		_, err = r.pool.Exec(ctx, `update commerce_cart_items set removed_at=now(),updated_at=now() where cart_id=$1 and id=$2 and removed_at is null`, cartID, itemID)
	} else {
		_, err = r.pool.Exec(ctx, `update commerce_cart_items set quantity=$3,updated_at=now() where cart_id=$1 and id=$2 and removed_at is null`, cartID, itemID, quantity)
	}
	if err != nil {
		return nil, err
	}
	if err := r.refreshCartTotals(ctx, cartID); err != nil {
		return nil, err
	}
	return r.cart(ctx, cartID)
}

func (r *CommerceRepository) RemoveCartItem(ctx context.Context, storeSlug, cartToken string, itemID uuid.UUID) (map[string]any, error) {
	return r.UpdateCartItem(ctx, storeSlug, cartToken, itemID, 0)
}

func (r *CommerceRepository) ClearCart(ctx context.Context, storeSlug, cartToken string) (map[string]any, error) {
	cartID, _, _, err := r.ensureCart(ctx, storeSlug, CartRequest{CartToken: cartToken})
	if err != nil {
		return nil, err
	}
	if _, err := r.pool.Exec(ctx, `update commerce_cart_items set removed_at=now(),updated_at=now() where cart_id=$1 and removed_at is null`, cartID); err != nil {
		return nil, err
	}
	if err := r.refreshCartTotals(ctx, cartID); err != nil {
		return nil, err
	}
	return r.cart(ctx, cartID)
}

func (r *CommerceRepository) MarkCartCheckoutStarted(ctx context.Context, storeSlug string, request CartRequest) (map[string]any, error) {
	cartID, _, _, err := r.ensureCart(ctx, storeSlug, request)
	if err != nil {
		return nil, err
	}
	_, err = r.pool.Exec(ctx, `
		update commerce_carts
		set checkout_started_at=coalesce(checkout_started_at,now()),email=coalesce(nullif($2,''),email),name=coalesce(nullif($3,''),name),phone=coalesce(nullif($4,''),phone),last_activity_at=now(),updated_at=now()
		where id=$1
	`, cartID, request.Email, request.Name, request.Phone)
	if err != nil {
		return nil, err
	}
	return r.cart(ctx, cartID)
}

func (r *CommerceRepository) CreateProduct(ctx context.Context, organizationID uuid.UUID, input CommerceProductInput) (map[string]any, error) {
	storeID, err := r.DefaultStore(ctx, organizationID)
	if err != nil {
		return nil, err
	}
	if input.Title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if input.Slug == "" {
		input.Slug = slugify(input.Title)
	}
	if input.Status == "" {
		input.Status = "active"
	}
	if input.Images == nil || len(input.Images) == 0 {
		input.Images = []string{"/images/catalog/look-1.jpg"}
	}
	if input.Colors == nil || len(input.Colors) == 0 {
		input.Colors = []string{"Ivory"}
	}
	if input.Sizes == nil || len(input.Sizes) == 0 {
		input.Sizes = []string{"S", "M", "L", "XL"}
	}
	if input.StockQuantity <= 0 {
		input.StockQuantity = 24
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	categoryID, err := r.ensureCategoryTx(ctx, tx, organizationID, storeID, input.CategorySlug)
	if err != nil {
		return nil, err
	}
	images, _ := json.Marshal(input.Images)
	options, _ := json.Marshal(map[string]any{"colors": input.Colors, "sizes": input.Sizes})
	var productID uuid.UUID
	err = tx.QueryRow(ctx, `
		insert into commerce_products (organization_id,store_id,category_id,title,slug,description,sku,brand,status,price,compare_at_price,cost_price,images,options,is_featured)
		values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		on conflict (store_id,slug) do update set title=excluded.title,description=excluded.description,sku=excluded.sku,brand=excluded.brand,status=excluded.status,price=excluded.price,compare_at_price=excluded.compare_at_price,cost_price=excluded.cost_price,images=excluded.images,options=excluded.options,is_featured=excluded.is_featured,updated_at=now()
		returning id
	`, organizationID, storeID, categoryID, input.Title, input.Slug, input.Description, input.SKU, input.Brand, input.Status, input.Price, input.CompareAtPrice, input.CostPrice, images, options, input.IsFeatured).Scan(&productID)
	if err != nil {
		return nil, err
	}
	variantOverrides := productVariantOverrides(input.Variants)
	for _, color := range input.Colors {
		for _, size := range input.Sizes {
			sku := strings.ToUpper(strings.ReplaceAll(input.SKU, " ", "-"))
			if sku == "" {
				sku = strings.ToUpper(strings.ReplaceAll(input.Slug, "-", ""))
			}
			colorCode := "NA"
			if strings.TrimSpace(color) != "" {
				colorCode = strings.ToUpper(color[:min(2, len(color))])
			}
			sku = fmt.Sprintf("%s-%s-%s", sku, colorCode, size)
			override := variantOverrides[variantKey(color, size)]
			variantPrice := input.Price
			variantCompareAtPrice := input.CompareAtPrice
			variantCostPrice := input.CostPrice
			variantStockQuantity := input.StockQuantity
			if override.Price != nil {
				variantPrice = *override.Price
			}
			if override.CompareAtPrice != nil {
				variantCompareAtPrice = *override.CompareAtPrice
			}
			if override.CostPrice != nil {
				variantCostPrice = *override.CostPrice
			}
			if override.StockQuantity != nil {
				variantStockQuantity = *override.StockQuantity
			}
			_, err = tx.Exec(ctx, `
				insert into commerce_product_variants (organization_id,store_id,product_id,title,sku,color,size,price,compare_at_price,cost_price,stock_quantity)
				values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
				on conflict (store_id,sku) do update set price=excluded.price,compare_at_price=excluded.compare_at_price,cost_price=excluded.cost_price,stock_quantity=excluded.stock_quantity,updated_at=now()
			`, organizationID, storeID, productID, strings.TrimSpace(color+" / "+size), sku, color, size, variantPrice, variantCompareAtPrice, variantCostPrice, variantStockQuantity)
			if err != nil {
				return nil, err
			}
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	_ = r.ensureDefaultChannels(ctx, organizationID, storeID)
	_ = r.syncChannelListings(ctx, organizationID, storeID, productID, nil, input.ChannelCodes)
	return r.Product(ctx, organizationID, productID)
}

func (r *CommerceRepository) Product(ctx context.Context, organizationID, productID uuid.UUID) (map[string]any, error) {
	storeID, err := r.DefaultStore(ctx, organizationID)
	if err != nil {
		return nil, err
	}
	items, err := queryMaps(ctx, r.pool, `
		select p.id::text,p.title,p.slug,p.description,coalesce(c.name,'') category,coalesce(c.slug,'') category_slug,p.sku,p.brand,p.status,p.price,p.compare_at_price,p.cost_price,p.currency_code,p.images,p.options,p.tags,p.is_featured,
			coalesce((select string_agg(distinct sc.code, ',') from commerce_channel_listings cl join commerce_sales_channels sc on sc.id=cl.channel_id where cl.product_id=p.id and cl.variant_id is null and cl.listing_status='active'),'') listed_channel_codes,
			coalesce((select jsonb_agg(jsonb_build_object('id',v.id::text,'title',v.title,'sku',v.sku,'color',v.color,'size',v.size,'price',v.price,'compareAtPrice',v.compare_at_price,'costPrice',v.cost_price,'stockQuantity',v.stock_quantity,'reservedQuantity',v.reserved_quantity,'lowStockThreshold',v.low_stock_threshold,'status',v.status) order by v.color,v.size) from commerce_product_variants v where v.product_id=p.id),'[]'::jsonb) variants
		from commerce_products p
		left join commerce_categories c on c.id=p.category_id
		where p.organization_id=$1 and p.store_id=$2 and p.id=$3
	`, organizationID, storeID, productID)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, pgx.ErrNoRows
	}
	return items[0], nil
}

func (r *CommerceRepository) UpdateProduct(ctx context.Context, organizationID, productID uuid.UUID, input CommerceProductInput) (map[string]any, error) {
	storeID, err := r.DefaultStore(ctx, organizationID)
	if err != nil {
		return nil, err
	}
	if input.Title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if input.Slug == "" {
		input.Slug = slugify(input.Title)
	}
	if input.Status == "" {
		input.Status = "active"
	}
	if len(input.Images) == 0 {
		input.Images = []string{"/images/catalog/look-1.jpg"}
	}
	var oldPrice, oldCompareAtPrice, oldCostPrice float64
	_ = r.pool.QueryRow(ctx, `select price,compare_at_price,cost_price from commerce_products where organization_id=$1 and store_id=$2 and id=$3`, organizationID, storeID, productID).Scan(&oldPrice, &oldCompareAtPrice, &oldCostPrice)
	categoryID, err := r.ensureCategory(ctx, organizationID, storeID, input.CategorySlug)
	if err != nil {
		return nil, err
	}
	images, _ := json.Marshal(input.Images)
	options, _ := json.Marshal(map[string]any{"colors": input.Colors, "sizes": input.Sizes})
	tagList := []string{}
	if input.CategorySlug != "" {
		tagList = append(tagList, input.CategorySlug)
	}
	var updated uuid.UUID
	err = r.pool.QueryRow(ctx, `
		update commerce_products set category_id=$4,title=$5,slug=$6,description=$7,sku=$8,brand=$9,status=$10,price=$11,compare_at_price=$12,cost_price=$13,images=$14,options=$15,is_featured=$16,tags=$17,updated_at=now()
		where organization_id=$1 and store_id=$2 and id=$3
		returning id
	`, organizationID, storeID, productID, categoryID, input.Title, input.Slug, input.Description, input.SKU, input.Brand, input.Status, input.Price, input.CompareAtPrice, input.CostPrice, images, options, input.IsFeatured, tagList).Scan(&updated)
	if err != nil {
		return nil, err
	}
	if err := r.syncChannelListings(ctx, organizationID, storeID, updated, nil, input.ChannelCodes); err != nil {
		return nil, err
	}
	_, _ = r.pool.Exec(ctx, `
		update commerce_product_variants
		set price=case when price=$4 then $5 else price end,
			compare_at_price=case when compare_at_price=$6 then $7 else compare_at_price end,
			cost_price=case when cost_price=$8 then $9 else cost_price end,
			updated_at=now()
		where organization_id=$1 and store_id=$2 and product_id=$3
	`, organizationID, storeID, updated, oldPrice, input.Price, oldCompareAtPrice, input.CompareAtPrice, oldCostPrice, input.CostPrice)
	if err := r.ensureProductVariants(ctx, organizationID, storeID, updated, input); err != nil {
		return nil, err
	}
	if err := r.applyProductVariantOverrides(ctx, organizationID, storeID, updated, input.Variants); err != nil {
		return nil, err
	}
	return r.Product(ctx, organizationID, updated)
}

func (r *CommerceRepository) DeleteProduct(ctx context.Context, organizationID, productID uuid.UUID) error {
	storeID, err := r.DefaultStore(ctx, organizationID)
	if err != nil {
		return err
	}
	ct, err := r.pool.Exec(ctx, `
		update commerce_products set status='deleted',updated_at=now()
		where organization_id=$1 and store_id=$2 and id=$3
	`, organizationID, storeID, productID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	_, err = r.pool.Exec(ctx, `update commerce_product_variants set status='archived',updated_at=now() where organization_id=$1 and store_id=$2 and product_id=$3`, organizationID, storeID, productID)
	return err
}

func (r *CommerceRepository) InventoryItem(ctx context.Context, organizationID, variantID uuid.UUID) (map[string]any, error) {
	storeID, err := r.DefaultStore(ctx, organizationID)
	if err != nil {
		return nil, err
	}
	items, err := queryMaps(ctx, r.pool, `
		select p.id::text product_id,p.title product,p.slug,v.id::text variant_id,v.title variant,v.sku,v.color,v.size,v.price,v.compare_at_price,v.cost_price,v.stock_quantity,v.reserved_quantity,(v.stock_quantity-v.reserved_quantity) available,v.low_stock_threshold,v.status,coalesce((select string_agg(distinct sc.code, ',') from commerce_channel_listings cl join commerce_sales_channels sc on sc.id=cl.channel_id where cl.variant_id=v.id and cl.listing_status='active'),'') listed_channel_codes,v.updated_at
		from commerce_product_variants v join commerce_products p on p.id=v.product_id
		where v.organization_id=$1 and v.store_id=$2 and v.id=$3
	`, organizationID, storeID, variantID)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, pgx.ErrNoRows
	}
	return items[0], nil
}

func (r *CommerceRepository) CreateInventoryItem(ctx context.Context, organizationID uuid.UUID, input CommerceInventoryInput) (map[string]any, error) {
	storeID, err := r.DefaultStore(ctx, organizationID)
	if err != nil {
		return nil, err
	}
	productID, err := uuid.Parse(input.ProductID)
	if err != nil {
		return nil, fmt.Errorf("valid productId is required")
	}
	if input.SKU == "" {
		return nil, fmt.Errorf("sku is required")
	}
	if input.Status == "" {
		input.Status = "active"
	}
	if input.LowStockThreshold <= 0 {
		input.LowStockThreshold = 5
	}
	if input.Title == "" {
		input.Title = strings.TrimSpace(input.Color + " / " + input.Size)
	}
	var variantID uuid.UUID
	err = r.pool.QueryRow(ctx, `
		insert into commerce_product_variants (organization_id,store_id,product_id,title,sku,color,size,price,compare_at_price,cost_price,stock_quantity,reserved_quantity,low_stock_threshold,status)
		select $1,$2,p.id,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14 from commerce_products p
		where p.organization_id=$1 and p.store_id=$2 and p.id=$3
		returning id
	`, organizationID, storeID, productID, input.Title, input.SKU, input.Color, input.Size, input.Price, input.CompareAtPrice, input.CostPrice, input.StockQuantity, input.ReservedQuantity, input.LowStockThreshold, input.Status).Scan(&variantID)
	if err != nil {
		return nil, err
	}
	if err := r.syncChannelListings(ctx, organizationID, storeID, productID, &variantID, input.ChannelCodes); err != nil {
		return nil, err
	}
	return r.InventoryItem(ctx, organizationID, variantID)
}

func (r *CommerceRepository) UpdateInventoryItem(ctx context.Context, organizationID, variantID uuid.UUID, input CommerceInventoryInput) (map[string]any, error) {
	storeID, err := r.DefaultStore(ctx, organizationID)
	if err != nil {
		return nil, err
	}
	if input.Status == "" {
		input.Status = "active"
	}
	if input.LowStockThreshold <= 0 {
		input.LowStockThreshold = 5
	}
	if input.Title == "" {
		input.Title = strings.TrimSpace(input.Color + " / " + input.Size)
	}
	var updated uuid.UUID
	err = r.pool.QueryRow(ctx, `
		update commerce_product_variants set title=$4,sku=$5,color=$6,size=$7,price=$8,compare_at_price=$9,cost_price=$10,stock_quantity=$11,reserved_quantity=$12,low_stock_threshold=$13,status=$14,updated_at=now()
		where organization_id=$1 and store_id=$2 and id=$3
		returning id
	`, organizationID, storeID, variantID, input.Title, input.SKU, input.Color, input.Size, input.Price, input.CompareAtPrice, input.CostPrice, input.StockQuantity, input.ReservedQuantity, input.LowStockThreshold, input.Status).Scan(&updated)
	if err != nil {
		return nil, err
	}
	var productID uuid.UUID
	if err := r.pool.QueryRow(ctx, `select product_id from commerce_product_variants where id=$1 and organization_id=$2 and store_id=$3`, updated, organizationID, storeID).Scan(&productID); err != nil {
		return nil, err
	}
	if err := r.syncChannelListings(ctx, organizationID, storeID, productID, &updated, input.ChannelCodes); err != nil {
		return nil, err
	}
	return r.InventoryItem(ctx, organizationID, updated)
}

func (r *CommerceRepository) DeleteInventoryItem(ctx context.Context, organizationID, variantID uuid.UUID) error {
	storeID, err := r.DefaultStore(ctx, organizationID)
	if err != nil {
		return err
	}
	ct, err := r.pool.Exec(ctx, `update commerce_product_variants set status='archived',updated_at=now() where organization_id=$1 and store_id=$2 and id=$3`, organizationID, storeID, variantID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *CommerceRepository) Checkout(ctx context.Context, storeSlug string, request CheckoutRequest) (map[string]any, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	var storeID, orgID uuid.UUID
	var currency string
	err = tx.QueryRow(ctx, `select id,organization_id,currency_code from commerce_stores where slug=$1`, storeSlug).Scan(&storeID, &orgID, &currency)
	if err != nil {
		return nil, err
	}
	var cartID *uuid.UUID
	if strings.TrimSpace(request.CartToken) != "" {
		var foundCartID uuid.UUID
		err = tx.QueryRow(ctx, `
			select id from commerce_carts where store_id=$1 and cart_token=$2 and status='active'
		`, storeID, strings.TrimSpace(request.CartToken)).Scan(&foundCartID)
		if err == nil {
			cartID = &foundCartID
			_, _ = tx.Exec(ctx, `
				update commerce_carts
				set checkout_started_at=coalesce(checkout_started_at,now()),email=coalesce(nullif($2,''),email),name=coalesce(nullif($3,''),name),phone=coalesce(nullif($4,''),phone),last_activity_at=now(),updated_at=now()
				where id=$1
			`, foundCartID, request.Email, request.Name, request.Phone)
			if len(request.Items) == 0 {
				cartItems, cartErr := tx.Query(ctx, `
					select product_id::text,variant_id::text,quantity
					from commerce_cart_items
					where cart_id=$1 and removed_at is null
					order by added_at
				`, foundCartID)
				if cartErr != nil {
					return nil, cartErr
				}
				for cartItems.Next() {
					var item CheckoutItem
					if scanErr := cartItems.Scan(&item.ProductID, &item.VariantID, &item.Quantity); scanErr != nil {
						cartItems.Close()
						return nil, scanErr
					}
					request.Items = append(request.Items, item)
				}
				if cartErr := cartItems.Err(); cartErr != nil {
					cartItems.Close()
					return nil, cartErr
				}
				cartItems.Close()
			}
		} else if err != pgx.ErrNoRows {
			return nil, err
		}
	}
	if len(request.Items) == 0 {
		return nil, fmt.Errorf("cart is empty")
	}
	if request.Email == "" {
		request.Email = "guest@example.test"
	}
	if request.Name == "" {
		request.Name = "Guest Customer"
	}
	city, _ := request.ShippingAddress["city"].(string)
	region, _ := request.ShippingAddress["state"].(string)
	var customerID uuid.UUID
	err = tx.QueryRow(ctx, `
		insert into commerce_customers (organization_id,store_id,email,name,phone,city,region_code)
		values ($1,$2,$3,$4,$5,$6,$7)
		on conflict (store_id,email) do update set name=excluded.name,phone=excluded.phone,city=excluded.city,region_code=excluded.region_code,updated_at=now()
		returning id
	`, orgID, storeID, request.Email, request.Name, request.Phone, city, region).Scan(&customerID)
	if err != nil {
		return nil, err
	}
	type line struct {
		productID uuid.UUID
		variantID uuid.UUID
		title     string
		sku       string
		qty       int
		price     float64
	}
	lines := []line{}
	subtotal := 0.0
	for _, item := range request.Items {
		if item.Quantity <= 0 {
			item.Quantity = 1
		}
		var l line
		if item.VariantID != "" {
			err = tx.QueryRow(ctx, `select p.id,v.id,p.title,v.sku,v.price from commerce_product_variants v join commerce_products p on p.id=v.product_id where v.id=$1 and v.store_id=$2 and v.stock_quantity-v.reserved_quantity >= $3`, item.VariantID, storeID, item.Quantity).Scan(&l.productID, &l.variantID, &l.title, &l.sku, &l.price)
		} else {
			err = tx.QueryRow(ctx, `select p.id,v.id,p.title,v.sku,v.price from commerce_products p join commerce_product_variants v on v.product_id=p.id where p.id=$1 and p.store_id=$2 and v.stock_quantity-v.reserved_quantity >= $3 order by v.stock_quantity-v.reserved_quantity desc limit 1`, item.ProductID, storeID, item.Quantity).Scan(&l.productID, &l.variantID, &l.title, &l.sku, &l.price)
		}
		if err != nil {
			return nil, fmt.Errorf("item unavailable")
		}
		l.qty = item.Quantity
		subtotal += l.price * float64(l.qty)
		lines = append(lines, l)
	}
	discount := r.couponDiscountTx(ctx, tx, storeID, request.CouponCode, subtotal)
	shipping := r.shippingRateTx(ctx, tx, storeID, region, subtotal-discount)
	tax := math.Round((subtotal-discount)*0.05*100) / 100
	total := subtotal - discount + shipping + tax
	orderNumber := fmt.Sprintf("RV-%d", time.Now().UnixNano()/1e6)
	addressJSON, _ := json.Marshal(request.ShippingAddress)
	var orderID uuid.UUID
	err = tx.QueryRow(ctx, `
		insert into commerce_orders (organization_id,store_id,customer_id,order_number,channel_code,status,payment_status,fulfillment_status,currency_code,subtotal,discount_total,shipping_total,tax_total,total,coupon_code,shipping_address)
		values ($1,$2,$3,$4,'website','paid','paid','unfulfilled',$5,$6,$7,$8,$9,$10,$11,$12)
		returning id
	`, orgID, storeID, customerID, orderNumber, currency, subtotal, discount, shipping, tax, total, strings.ToUpper(request.CouponCode), addressJSON).Scan(&orderID)
	if err != nil {
		return nil, err
	}
	for _, l := range lines {
		_, err = tx.Exec(ctx, `insert into commerce_order_items (organization_id,store_id,order_id,product_id,variant_id,title,sku,quantity,unit_price,total_price) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`, orgID, storeID, orderID, l.productID, l.variantID, l.title, l.sku, l.qty, l.price, l.price*float64(l.qty))
		if err != nil {
			return nil, err
		}
		_, err = tx.Exec(ctx, `update commerce_product_variants set stock_quantity=stock_quantity-$1,updated_at=now() where id=$2`, l.qty, l.variantID)
		if err != nil {
			return nil, err
		}
	}
	_, _ = tx.Exec(ctx, `update commerce_customers set total_spent=total_spent+$1,order_count=order_count+1,updated_at=now() where id=$2`, total, customerID)
	if request.CouponCode != "" {
		_, _ = tx.Exec(ctx, `update commerce_coupons set used_count=used_count+1 where store_id=$1 and code=$2`, storeID, strings.ToUpper(request.CouponCode))
	}
	if cartID != nil {
		_, _ = tx.Exec(ctx, `
			update commerce_carts
			set status='converted',customer_id=$2,converted_order_id=$3,converted_at=now(),last_activity_at=now(),updated_at=now()
			where id=$1
		`, *cartID, customerID, orderID)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return map[string]any{"id": orderID.String(), "orderNumber": orderNumber, "total": total, "currencyCode": currency, "status": "paid"}, nil
}

func (r *CommerceRepository) GenerateDemo(ctx context.Context, organizationID uuid.UUID, months int) (map[string]any, error) {
	if months < 1 {
		months = 6
	}
	if months > 18 {
		months = 18
	}
	storeID, err := r.DefaultStore(ctx, organizationID)
	if err != nil {
		return nil, err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	for _, table := range []string{"commerce_cart_items", "commerce_carts", "commerce_returns", "commerce_order_items", "commerce_orders", "commerce_customers", "commerce_channel_listings", "commerce_product_variants", "commerce_products", "commerce_homepage_sections", "commerce_categories", "commerce_coupons", "commerce_shipping_zones", "commerce_payment_methods", "commerce_cms_pages"} {
		if _, err := tx.Exec(ctx, `delete from `+table+` where organization_id=$1 and store_id=$2`, organizationID, storeID); err != nil {
			return nil, err
		}
	}
	if _, err := tx.Exec(ctx, `
		insert into commerce_sales_channels (organization_id,store_id,code,name,channel_type,status)
		values ($1,$2,'website','Website Storefront','storefront','active'),
			($1,$2,'amazon','Amazon Seller','marketplace','inactive'),
			($1,$2,'google','Google Merchant / Ads','ads_marketplace','inactive'),
			($1,$2,'meta','Meta Shop / Ads','social_commerce','inactive'),
			($1,$2,'myntra','Myntra','marketplace','inactive')
		on conflict (store_id,code) do update set name=excluded.name,channel_type=excluded.channel_type,updated_at=now()
	`, organizationID, storeID); err != nil {
		return nil, err
	}
	type demoProduct struct {
		name        string
		category    string
		baseColor   string
		accentColor string
		price       float64
		image       string
		altImage    string
		featured    bool
	}
	catalog := []demoProduct{
		{"Beige Chanderi Silk Embroidered Anarkali Kurta", "Kurtas", "Beige", "Ivory", 1099, "/images/products/kurta-beige-chanderi-anarkali.png", "/images/products/kurta-rose-pintuck.png", true},
		{"Orange Mangalgiri Cotton Buti A-Line Kurta", "Kurtas", "Orange", "Mustard", 999, "/images/products/kurta-orange-mangalgiri.png", "/images/products/kurta-beige-chanderi-anarkali.png", false},
		{"Black Cotton Metallic Foil A-Line Kurta", "Kurtas", "Black", "Charcoal", 749, "/images/products/kurta-black-foil.png", "/images/products/kurta-orange-mangalgiri.png", true},
		{"Off White Cotton Floral Printed Kurti", "Kurtas", "Off White", "Peach", 699, "/images/products/kurta-rose-pintuck.png", "/images/products/kurta-beige-chanderi-anarkali.png", false},
		{"Maroon Cotton Zig Zag Printed A-Line Kurta", "Kurtas", "Maroon", "Wine", 1899, "/images/products/kurta-black-foil.png", "/images/products/kurta-rose-pintuck.png", true},
		{"Indigo Cotton Silk Embroidered Straight Kurta", "Kurtas", "Indigo", "Navy", 2299, "/images/products/kurta-black-foil.png", "/images/products/kurta-beige-chanderi-anarkali.png", false},
		{"Mustard Sleeveless Cotton Summer Kurta", "Kurtas", "Mustard", "Yellow", 899, "/images/products/kurta-orange-mangalgiri.png", "/images/products/kurta-rose-pintuck.png", false},
		{"Teal Viscose Antique Foil Festive Kurta", "Kurtas", "Teal", "Sea Green", 1999, "/images/products/kurta-beige-chanderi-anarkali.png", "/images/products/kurta-black-foil.png", true},
		{"Rose Cotton Pintuck Everyday Kurta", "Kurtas", "Rose", "Blush", 1299, "/images/products/kurta-rose-pintuck.png", "/images/products/kurta-orange-mangalgiri.png", false},
		{"Sage Handblock Printed Straight Kurta", "Kurtas", "Sage", "Olive", 1499, "/images/products/kurta-orange-mangalgiri.png", "/images/products/kurta-beige-chanderi-anarkali.png", false},
		{"Turquoise Cotton Printed Kurti Pant Set With Dupatta", "Salwar Suits", "Turquoise", "Aqua", 1999, "/images/products/suit-turquoise-printed-dupatta.png", "/images/products/suit-indigo-embroidered.png", true},
		{"Pink Cotton Printed Kurti Pant Set With Dupatta", "Salwar Suits", "Pink", "Rose", 1999, "/images/products/suit-turquoise-printed-dupatta.png", "/images/products/suit-maroon-embroidered.png", false},
		{"Yellow Mangalgiri Cotton Kurta And Palazzo Set", "Salwar Suits", "Yellow", "Mustard", 1999, "/images/products/suit-rust-satin-dupatta.png", "/images/products/suit-turquoise-printed-dupatta.png", true},
		{"Maroon Cotton Silk Embroidered Salwar Suit Set", "Salwar Suits", "Maroon", "Wine", 2999, "/images/products/suit-maroon-embroidered.png", "/images/products/suit-rust-satin-dupatta.png", true},
		{"Sand Beige Maroon Cotton Printed Salwar Suit Set", "Salwar Suits", "Sand Beige", "Maroon", 2499, "/images/products/suit-maroon-embroidered.png", "/images/products/suit-indigo-embroidered.png", false},
		{"Indigo Cotton Silk Embroidered Suit Set With Dupatta", "Salwar Suits", "Indigo", "Navy", 4999, "/images/products/suit-indigo-embroidered.png", "/images/products/suit-maroon-embroidered.png", true},
		{"Yellow Chanderi Cotton Silk Embroidered Suit Set", "Salwar Suits", "Yellow", "Gold", 1999, "/images/products/suit-rust-satin-dupatta.png", "/images/products/suit-turquoise-printed-dupatta.png", false},
		{"Pink Cotton Embroidered Straight Suit Set", "Salwar Suits", "Pink", "Fuchsia", 2299, "/images/products/suit-turquoise-printed-dupatta.png", "/images/products/suit-maroon-embroidered.png", false},
		{"Pista Green Cotton Straight Salwar Suit With Dupatta", "Salwar Suits", "Pista Green", "Mint", 1599, "/images/products/suit-turquoise-printed-dupatta.png", "/images/products/suit-indigo-embroidered.png", false},
		{"Rust Satin Straight Suit With Dupatta", "Salwar Suits", "Rust", "Copper", 2200, "/images/products/suit-rust-satin-dupatta.png", "/images/products/suit-maroon-embroidered.png", true},
	}
	categories := []string{"Kurtas", "Salwar Suits"}
	categoryImages := map[string]string{
		"Kurtas":       "/images/products/kurta-beige-chanderi-anarkali.png",
		"Salwar Suits": "/images/products/suit-turquoise-printed-dupatta.png",
	}
	sizes := []string{"XS", "S", "M", "L", "XL", "XXL"}
	productIDs := []uuid.UUID{}
	variantIDs := []uuid.UUID{}
	for i, category := range categories {
		_, err := tx.Exec(ctx, `insert into commerce_categories (organization_id,store_id,name,slug,image_url,sort_order) values ($1,$2,$3,$4,$5,$6)`, organizationID, storeID, category, slugify(category), categoryImages[category], i+1)
		if err != nil {
			return nil, err
		}
	}
	for i, product := range catalog {
		var categoryID uuid.UUID
		category := product.category
		if err := tx.QueryRow(ctx, `select id from commerce_categories where store_id=$1 and slug=$2`, storeID, slugify(category)).Scan(&categoryID); err != nil {
			return nil, err
		}
		price := product.price
		compare := math.Round((price/0.44)/10) * 10
		cost := price * (0.34 + float64(i%4)*0.03)
		images, _ := json.Marshal([]string{product.image, product.altImage})
		options, _ := json.Marshal(map[string]any{"colors": []string{product.baseColor, product.accentColor}, "sizes": sizes})
		var pid uuid.UUID
		err = tx.QueryRow(ctx, `insert into commerce_products (organization_id,store_id,category_id,title,slug,description,sku,brand,status,price,compare_at_price,cost_price,images,options,is_featured,tags) values ($1,$2,$3,$4,$5,$6,$7,'Rangavali','active',$8,$9,$10,$11,$12,$13,$14) returning id`, organizationID, storeID, categoryID, product.name, slugify(product.name), "Rangavali-style women's ethnicwear with breathable fabrics, easy festive styling, and variant-wise color and size inventory.", fmt.Sprintf("RV-%03d", i+1), price, compare, cost, images, options, product.featured, []string{slugify(category), "women", "ethnicwear", "rangavali-style"}).Scan(&pid)
		if err != nil {
			return nil, err
		}
		productIDs = append(productIDs, pid)
		for _, color := range []string{product.baseColor, product.accentColor} {
			for _, size := range sizes {
				var vid uuid.UUID
				err = tx.QueryRow(ctx, `insert into commerce_product_variants (organization_id,store_id,product_id,title,sku,color,size,price,compare_at_price,cost_price,stock_quantity,low_stock_threshold) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,8) returning id`, organizationID, storeID, pid, color+" / "+size, fmt.Sprintf("RV-%03d-%s-%s", i+1, strings.ToUpper(color[:2]), size), color, size, price, compare, cost, 18+(i*7+len(size))%80).Scan(&vid)
				if err != nil {
					return nil, err
				}
				variantIDs = append(variantIDs, vid)
			}
		}
	}
	if _, err := tx.Exec(ctx, `
		insert into commerce_channel_listings (organization_id,store_id,channel_id,product_id,listing_status,sync_status,last_synced_at)
		select $1,$2,sc.id,p.id,'active','synced',now()
		from commerce_sales_channels sc
		join commerce_products p on p.organization_id=$1 and p.store_id=$2
		where sc.organization_id=$1 and sc.store_id=$2 and sc.code='website'
		on conflict (channel_id,product_id,variant_id) do update set listing_status='active',sync_status='synced',last_synced_at=now(),updated_at=now()
	`, organizationID, storeID); err != nil {
		return nil, err
	}
	for _, c := range []struct {
		code, name, typ string
		val, min        float64
	}{{"WELCOME10", "First order 10% off", "percentage", 10, 999}, {"FREESHIP", "Free shipping above ₹1,999", "free_shipping", 100, 1999}, {"SUIT15", "Salwar suit festive offer", "percentage", 15, 2499}} {
		_, err = tx.Exec(ctx, `insert into commerce_coupons (organization_id,store_id,code,name,discount_type,discount_value,minimum_order_value,status,expires_at) values ($1,$2,$3,$4,$5,$6,$7,'active',now()+interval '120 days')`, organizationID, storeID, c.code, c.name, c.typ, c.val, c.min)
		if err != nil {
			return nil, err
		}
	}
	regions := []struct {
		name, code string
		rate, free float64
	}{{"West India", "MH", 70, 1999}, {"North India", "DL", 90, 2499}, {"South India", "KA", 80, 1999}, {"Pan India", "", 120, 2999}}
	for _, z := range regions {
		_, err = tx.Exec(ctx, `insert into commerce_shipping_zones (organization_id,store_id,name,country_code,region_codes,rate,free_shipping_threshold,estimated_days_min,estimated_days_max,cod_enabled) values ($1,$2,$3,'IN',$4,$5,$6,3,6,true)`, organizationID, storeID, z.name, []string{z.code}, z.rate, z.free)
		if err != nil {
			return nil, err
		}
	}
	customerNames := []string{"Aarohi Sharma", "Nisha Mehta", "Sara Khan", "Priya Iyer", "Meera Patel", "Ananya Rao", "Riya Kapoor", "Ishita Sen"}
	cities := []struct{ city, state string }{{"Mumbai", "MH"}, {"Pune", "MH"}, {"Delhi", "DL"}, {"Bengaluru", "KA"}, {"Hyderabad", "TS"}, {"Ahmedabad", "GJ"}, {"Jaipur", "RJ"}, {"Kolkata", "WB"}}
	rng := rand.New(rand.NewSource(42))
	orderCount := months * 45
	returnCount := 0
	for i := 0; i < orderCount; i++ {
		city := cities[rng.Intn(len(cities))]
		name := customerNames[rng.Intn(len(customerNames))]
		email := strings.ToLower(strings.ReplaceAll(name, " ", ".")) + fmt.Sprintf("%d@example.test", rng.Intn(100))
		var cid uuid.UUID
		if err := tx.QueryRow(ctx, `insert into commerce_customers (organization_id,store_id,email,name,phone,city,region_code) values ($1,$2,$3,$4,$5,$6,$7) on conflict (store_id,email) do update set updated_at=now() returning id`, organizationID, storeID, email, name, fmt.Sprintf("90000%05d", rng.Intn(99999)), city.city, city.state).Scan(&cid); err != nil {
			return nil, err
		}
		lines := 1 + rng.Intn(3)
		subtotal := 0.0
		type line struct {
			pid, vid   uuid.UUID
			title, sku string
			qty        int
			price      float64
		}
		orderLines := []line{}
		for j := 0; j < lines; j++ {
			vid := variantIDs[rng.Intn(len(variantIDs))]
			var l line
			if err := tx.QueryRow(ctx, `select p.id,v.id,p.title,v.sku,v.price from commerce_product_variants v join commerce_products p on p.id=v.product_id where v.id=$1`, vid).Scan(&l.pid, &l.vid, &l.title, &l.sku, &l.price); err != nil {
				return nil, err
			}
			l.qty = 1 + rng.Intn(2)
			subtotal += l.price * float64(l.qty)
			orderLines = append(orderLines, l)
		}
		discount := 0.0
		coupon := ""
		if i%5 == 0 {
			coupon = "WELCOME10"
			discount = math.Round(subtotal*0.10*100) / 100
		}
		shipping := 80.0
		if subtotal-discount > 1999 {
			shipping = 0
		}
		tax := math.Round((subtotal-discount)*0.05*100) / 100
		total := subtotal - discount + shipping + tax
		placed := time.Now().AddDate(0, -months, 0).Add(time.Duration(rng.Intn(months*30*24)) * time.Hour)
		status := []string{"paid", "processing", "shipped", "delivered"}[rng.Intn(4)]
		var oid uuid.UUID
		address, _ := json.Marshal(map[string]any{"city": city.city, "state": city.state, "country": "IN"})
		if err := tx.QueryRow(ctx, `insert into commerce_orders (organization_id,store_id,customer_id,order_number,channel_code,status,payment_status,fulfillment_status,currency_code,subtotal,discount_total,shipping_total,tax_total,total,coupon_code,shipping_address,placed_at) values ($1,$2,$3,$4,'website',$5,'paid',$6,'INR',$7,$8,$9,$10,$11,$12,$13,$14) returning id`, organizationID, storeID, cid, fmt.Sprintf("RV-%06d", 1000+i), status, status, subtotal, discount, shipping, tax, total, coupon, address, placed).Scan(&oid); err != nil {
			return nil, err
		}
		for _, l := range orderLines {
			_, err = tx.Exec(ctx, `insert into commerce_order_items (organization_id,store_id,order_id,product_id,variant_id,title,sku,quantity,unit_price,total_price) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`, organizationID, storeID, oid, l.pid, l.vid, l.title, l.sku, l.qty, l.price, l.price*float64(l.qty))
			if err != nil {
				return nil, err
			}
			if i%17 == 0 {
				returnCount++
				reason := []string{"Size issue", "Color mismatch", "Quality issue", "Late delivery", "Changed mind"}[rng.Intn(5)]
				_, err = tx.Exec(ctx, `insert into commerce_returns (organization_id,store_id,order_id,product_id,reason,status,refund_amount,requested_at) values ($1,$2,$3,$4,$5,'completed',$6,$7)`, organizationID, storeID, oid, l.pid, reason, l.price, placed.AddDate(0, 0, 7+rng.Intn(10)))
				if err != nil {
					return nil, err
				}
			}
		}
		_, _ = tx.Exec(ctx, `update commerce_customers set total_spent=total_spent+$1,order_count=order_count+1 where id=$2`, total, cid)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	if err := r.ensureDefaultCMS(ctx, organizationID, storeID); err != nil {
		return nil, err
	}
	return map[string]any{"storeId": storeID.String(), "products": len(productIDs), "orders": orderCount, "returns": returnCount, "months": months}, nil
}

func (r *CommerceRepository) ensureCategoryTx(ctx context.Context, tx pgx.Tx, organizationID, storeID uuid.UUID, slug string) (*uuid.UUID, error) {
	if slug == "" {
		slug = "uncategorized"
	}
	var id uuid.UUID
	err := tx.QueryRow(ctx, `select id from commerce_categories where store_id=$1 and slug=$2`, storeID, slug).Scan(&id)
	if err == nil {
		return &id, nil
	}
	name := strings.Title(strings.ReplaceAll(slug, "-", " "))
	err = tx.QueryRow(ctx, `insert into commerce_categories (organization_id,store_id,name,slug) values ($1,$2,$3,$4) returning id`, organizationID, storeID, name, slug).Scan(&id)
	return &id, err
}

func (r *CommerceRepository) ensureCategory(ctx context.Context, organizationID, storeID uuid.UUID, slug string) (*uuid.UUID, error) {
	if slug == "" {
		slug = "uncategorized"
	}
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `select id from commerce_categories where store_id=$1 and slug=$2`, storeID, slug).Scan(&id)
	if err == nil {
		return &id, nil
	}
	if err != pgx.ErrNoRows {
		return nil, err
	}
	name := strings.Title(strings.ReplaceAll(slug, "-", " "))
	err = r.pool.QueryRow(ctx, `insert into commerce_categories (organization_id,store_id,name,slug) values ($1,$2,$3,$4) returning id`, organizationID, storeID, name, slug).Scan(&id)
	return &id, err
}

func (r *CommerceRepository) ensureDefaultChannels(ctx context.Context, organizationID, storeID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		insert into commerce_sales_channels (organization_id,store_id,code,name,channel_type,status)
		values ($1,$2,'website','Website Storefront','storefront','active'),
			($1,$2,'amazon','Amazon Seller','marketplace','inactive'),
			($1,$2,'google','Google Merchant / Ads','ads_marketplace','inactive'),
			($1,$2,'meta','Meta Shop / Ads','social_commerce','inactive'),
			($1,$2,'myntra','Myntra','marketplace','inactive')
		on conflict (store_id,code) do update set name=excluded.name,channel_type=excluded.channel_type,updated_at=now()
	`, organizationID, storeID)
	return err
}

func (r *CommerceRepository) ensureDefaultCMS(ctx context.Context, organizationID, storeID uuid.UUID) error {
	var settings map[string]any
	if err := r.pool.QueryRow(ctx, `select settings from commerce_stores where id=$1 and organization_id=$2`, storeID, organizationID).Scan(&settings); err == nil {
		if settings == nil {
			settings = map[string]any{}
		}
		changed := false
		defaults := map[string]any{
			"brandName":       "Rangavali",
			"tagline":         "Modern Indian clothing, thoughtfully made for women who collect moments, not trends.",
			"announcement":    "10% off first order · Free shipping above ₹1,999",
			"supportEmail":    "support@rangavali.test",
			"supportPhone":    "+91 90000 00000",
			"returnPolicy":    "Easy 7-day returns on unworn styles.",
			"primaryColor":    "#6f1d46",
			"accentColor":     "#c8a24a",
			"instagramUrl":    "https://instagram.com/",
			"facebookUrl":     "https://facebook.com/",
			"youtubeUrl":      "https://youtube.com/",
			"newsletterTitle": "₹300 off your first order",
		}
		for key, value := range defaults {
			if _, ok := settings[key]; !ok {
				settings[key] = value
				changed = true
			}
		}
		if changed {
			settingsJSON, _ := json.Marshal(settings)
			if _, err := r.pool.Exec(ctx, `update commerce_stores set settings=$3,updated_at=now() where id=$1 and organization_id=$2`, storeID, organizationID, settingsJSON); err != nil {
				return err
			}
		}
	}
	_, err := r.pool.Exec(ctx, `
		insert into commerce_homepage_sections (organization_id,store_id,section_key,section_type,title,subtitle,layout,image_url,cta_label,cta_href,category_slug,product_source,max_items,content,sort_order,status)
		values
			($1,$2,'hero','hero','Kurtas and suit sets for every day of celebration.','Printed cottons, chanderi textures, and easy salwar suit sets inspired by modern Indian wardrobes.','split','/images/rangavali-hero.png','Shop salwar suits','/category/salwar-suits','salwar-suits','featured',8,$3,1,'active'),
			($1,$2,'category_tiles','category_tiles','Shop by category','','tiles','','View all','/category/all','','categories',8,'{}'::jsonb,2,'active'),
			($1,$2,'trending','product_grid','Trending kurtas and suits','Styles everyone is adding to bag','grid','','View all','/category/all','all','trending',12,'{}'::jsonb,3,'active'),
			($1,$2,'offers','promo_tiles','Offers you cannot miss','','three_tiles','','Shop now','/category/sale','sale','manual',3,$4,4,'active'),
			($1,$2,'bestsellers','product_grid','Bestsellers','Most-loved styles from your storefront','grid','','View all','/category/all','all','bestsellers',12,'{}'::jsonb,5,'active'),
			($1,$2,'feature_banner','banner','The salwar suit edit.','Straight suits, pant sets, and dupatta-ready festive picks.','banner','/images/products/suit-indigo-embroidered.png','Shop suit sets','/category/salwar-suits','salwar-suits','manual',1,$5,6,'active'),
			($1,$2,'benefits','benefits','Why customers choose us','','icons','','','','','manual',4,$6,7,'active'),
			($1,$2,'newsletter','newsletter','₹300 off your first order','Join the store letter for launches, styling notes, and private offers.','centered','','Subscribe','','','manual',1,'{}'::jsonb,8,'active')
		on conflict (store_id,section_key) do nothing
	`, organizationID, storeID,
		map[string]any{"eyebrow": "THE ETHNICWEAR CHAPTER · 2026"},
		map[string]any{"tiles": []map[string]any{
			{"title": "KURTAS UNDER ₹1,999", "subtitle": "Printed, foil, and embroidered everyday styles", "imageUrl": "/images/products/kurta-orange-mangalgiri.png", "href": "/category/kurtas"},
			{"title": "SALWAR SUIT SETS", "subtitle": "Pant sets and dupatta-ready festive looks", "imageUrl": "/images/products/suit-turquoise-printed-dupatta.png", "href": "/category/salwar-suits"},
			{"title": "COTTON & CHANDERI EDIT", "subtitle": "Breathable fabrics with polished detailing", "imageUrl": "/images/products/suit-rust-satin-dupatta.png", "href": "/category/all"},
		}},
		map[string]any{"eyebrow": "SUIT SET SPOTLIGHT"},
		map[string]any{"items": []map[string]any{
			{"icon": "leaf", "label": "Premium fabrics"},
			{"icon": "returns", "label": "Easy 7-day returns"},
			{"icon": "truck", "label": "Tracked delivery"},
			{"icon": "shield", "label": "Secure payments"},
		}},
	)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `
		insert into commerce_payment_methods (organization_id,store_id,code,name,provider,instructions,sort_order,status,settings)
		values
			($1,$2,'cod','Cash on Delivery','manual','Pay when your order is delivered.',1,'active','{}'::jsonb),
			($1,$2,'razorpay','Cards / UPI / Netbanking','razorpay','Use Razorpay checkout in production by adding provider keys later.',2,'inactive',$3),
			($1,$2,'stripe','International cards','stripe','Use Stripe for global card payments when enabled.',3,'inactive',$4)
		on conflict (store_id,code) do nothing
	`, organizationID, storeID, map[string]any{"mode": "test"}, map[string]any{"mode": "test"})
	return err
}

func (r *CommerceRepository) syncChannelListings(ctx context.Context, organizationID, storeID, productID uuid.UUID, variantID *uuid.UUID, channelCodes []string) error {
	if channelCodes == nil {
		return nil
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if variantID == nil {
		_, err = tx.Exec(ctx, `update commerce_channel_listings set listing_status='not_listed',sync_status='pending',updated_at=now() where organization_id=$1 and store_id=$2 and product_id=$3 and variant_id is null`, organizationID, storeID, productID)
	} else {
		_, err = tx.Exec(ctx, `update commerce_channel_listings set listing_status='not_listed',sync_status='pending',updated_at=now() where organization_id=$1 and store_id=$2 and product_id=$3 and variant_id=$4`, organizationID, storeID, productID, *variantID)
	}
	if err != nil {
		return err
	}
	for _, code := range uniqueStrings(channelCodes) {
		code = strings.TrimSpace(strings.ToLower(code))
		if code == "" {
			continue
		}
		if variantID == nil {
			_, err = tx.Exec(ctx, `
		insert into commerce_channel_listings (organization_id,store_id,channel_id,product_id,listing_status,sync_status,last_synced_at)
		select $1,$2,sc.id,$3,'active','synced',now()
		from commerce_sales_channels sc
		where sc.organization_id=$1 and sc.store_id=$2 and sc.code=$4
			and not exists (
				select 1 from commerce_channel_listings existing
				where existing.channel_id=sc.id and existing.product_id=$3 and existing.variant_id is null
			)
	`, organizationID, storeID, productID, code)
		} else {
			_, err = tx.Exec(ctx, `
		insert into commerce_channel_listings (organization_id,store_id,channel_id,product_id,variant_id,listing_status,sync_status,last_synced_at)
		select $1,$2,sc.id,$3,$4,'active','synced',now()
		from commerce_sales_channels sc
		where sc.organization_id=$1 and sc.store_id=$2 and sc.code=$5
			and not exists (
				select 1 from commerce_channel_listings existing
				where existing.channel_id=sc.id and existing.product_id=$3 and existing.variant_id=$4
			)
	`, organizationID, storeID, productID, *variantID, code)
		}
		if err != nil {
			return err
		}
		if variantID == nil {
			_, err = tx.Exec(ctx, `update commerce_channel_listings set listing_status='active',sync_status='synced',error_message='',last_synced_at=now(),updated_at=now() where organization_id=$1 and store_id=$2 and product_id=$3 and variant_id is null and channel_id=(select id from commerce_sales_channels where store_id=$2 and code=$4)`, organizationID, storeID, productID, code)
		} else {
			_, err = tx.Exec(ctx, `update commerce_channel_listings set listing_status='active',sync_status='synced',error_message='',last_synced_at=now(),updated_at=now() where organization_id=$1 and store_id=$2 and product_id=$3 and variant_id=$4 and channel_id=(select id from commerce_sales_channels where store_id=$2 and code=$5)`, organizationID, storeID, productID, *variantID, code)
		}
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *CommerceRepository) couponDiscountTx(ctx context.Context, tx pgx.Tx, storeID uuid.UUID, code string, subtotal float64) float64 {
	if code == "" {
		return 0
	}
	var typ string
	var value, minOrder float64
	err := tx.QueryRow(ctx, `select discount_type,discount_value,minimum_order_value from commerce_coupons where store_id=$1 and code=$2 and status='active' and (expires_at is null or expires_at>now())`, storeID, strings.ToUpper(code)).Scan(&typ, &value, &minOrder)
	if err != nil || subtotal < minOrder {
		return 0
	}
	if typ == "percentage" {
		return math.Round(subtotal*value) / 100
	}
	if typ == "fixed" {
		return math.Min(value, subtotal)
	}
	return 0
}

func (r *CommerceRepository) shippingRateTx(ctx context.Context, tx pgx.Tx, storeID uuid.UUID, region string, subtotal float64) float64 {
	var rate, free float64
	err := tx.QueryRow(ctx, `select rate,free_shipping_threshold from commerce_shipping_zones where store_id=$1 and status='active' and ($2=any(region_codes) or cardinality(region_codes)=0 or region_codes='{""}') order by case when $2=any(region_codes) then 0 else 1 end limit 1`, storeID, region).Scan(&rate, &free)
	if err != nil {
		return 0
	}
	if free > 0 && subtotal >= free {
		return 0
	}
	return rate
}

func (r *CommerceRepository) ensureCart(ctx context.Context, storeSlug string, request CartRequest) (uuid.UUID, uuid.UUID, uuid.UUID, error) {
	var storeID, orgID uuid.UUID
	var currency string
	if err := r.pool.QueryRow(ctx, `select id,organization_id,currency_code from commerce_stores where slug=$1 and status in ('live','draft')`, storeSlug).Scan(&storeID, &orgID, &currency); err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, err
	}
	token := strings.TrimSpace(request.CartToken)
	if token == "" {
		token = "cart_" + uuid.NewString()
	}
	metadataJSON, _ := json.Marshal(request.Metadata)
	var id uuid.UUID
	var status string
	err := r.pool.QueryRow(ctx, `
		select id,status from commerce_carts where store_id=$1 and cart_token=$2
	`, storeID, token).Scan(&id, &status)
	if err == nil && status != "converted" {
		_, err = r.pool.Exec(ctx, `
			update commerce_carts
			set visitor_id=coalesce(nullif($3,''),visitor_id),email=coalesce(nullif($4,''),email),name=coalesce(nullif($5,''),name),phone=coalesce(nullif($6,''),phone),metadata=case when $7::jsonb='null'::jsonb then metadata else metadata || $7::jsonb end,last_activity_at=now(),updated_at=now()
			where id=$1 and store_id=$2
		`, id, storeID, request.VisitorID, request.Email, request.Name, request.Phone, metadataJSON)
		return id, orgID, storeID, err
	}
	if err != nil && err != pgx.ErrNoRows {
		return uuid.Nil, uuid.Nil, uuid.Nil, err
	}
	if status == "converted" {
		token = "cart_" + uuid.NewString()
	}
	err = r.pool.QueryRow(ctx, `
		insert into commerce_carts (organization_id,store_id,cart_token,visitor_id,email,name,phone,currency_code,metadata)
		values ($1,$2,$3,$4,$5,$6,$7,$8,case when $9::jsonb='null'::jsonb then '{}'::jsonb else $9::jsonb end)
		returning id
	`, orgID, storeID, token, request.VisitorID, request.Email, request.Name, request.Phone, currency, metadataJSON).Scan(&id)
	return id, orgID, storeID, err
}

func (r *CommerceRepository) refreshCartTotals(ctx context.Context, cartID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		update commerce_carts c
		set subtotal=coalesce(t.subtotal,0),
			item_count=coalesce(t.item_count,0),
			status=case when c.status='converted' then c.status else 'active' end,
			last_activity_at=now(),
			updated_at=now()
		from (
			select $1::uuid cart_id,
				coalesce(sum(unit_price*quantity),0) subtotal,
				coalesce(sum(quantity),0)::int item_count
			from commerce_cart_items
			where cart_id=$1 and removed_at is null
		) t
		where c.id=t.cart_id
	`, cartID)
	return err
}

func (r *CommerceRepository) cart(ctx context.Context, cartID uuid.UUID) (map[string]any, error) {
	items, err := queryMaps(ctx, r.pool, `
		select c.id::text,c.cart_token,c.visitor_id,c.email,c.name,c.phone,c.status,c.currency_code,c.subtotal,c.item_count,c.first_seen_at,c.last_activity_at,c.checkout_started_at,c.abandoned_at,c.converted_at,c.converted_order_id::text,
			coalesce(jsonb_agg(jsonb_build_object(
				'id',ci.id::text,
				'productId',ci.product_id::text,
				'variantId',ci.variant_id::text,
				'productTitle',ci.product_title,
				'variantTitle',ci.variant_title,
				'sku',ci.sku,
				'imageUrl',ci.image_url,
				'unitPrice',ci.unit_price,
				'quantity',ci.quantity,
				'addedAt',ci.added_at,
				'updatedAt',ci.updated_at
			) order by ci.added_at) filter (where ci.id is not null),'[]'::jsonb) items
		from commerce_carts c
		left join commerce_cart_items ci on ci.cart_id=c.id and ci.removed_at is null
		where c.id=$1
		group by c.id
	`, cartID)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, pgx.ErrNoRows
	}
	return items[0], nil
}

func queryMaps(ctx context.Context, pool *pgxpool.Pool, query string, args ...any) ([]map[string]any, error) {
	rows, err := pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	fields := rows.FieldDescriptions()
	out := []map[string]any{}
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, err
		}
		row := map[string]any{}
		for i, value := range values {
			row[string(fields[i].Name)] = value
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *CommerceRepository) applyProductVariantOverrides(ctx context.Context, organizationID, storeID, productID uuid.UUID, variants []CommerceProductVariantInput) error {
	for _, variant := range variants {
		if variant.ID == "" && strings.TrimSpace(variant.Color) == "" && strings.TrimSpace(variant.Size) == "" {
			continue
		}
		assignments := []string{}
		args := []any{organizationID, storeID, productID}
		if variant.Price != nil {
			args = append(args, *variant.Price)
			assignments = append(assignments, fmt.Sprintf("price=$%d", len(args)))
		}
		if variant.CompareAtPrice != nil {
			args = append(args, *variant.CompareAtPrice)
			assignments = append(assignments, fmt.Sprintf("compare_at_price=$%d", len(args)))
		}
		if variant.CostPrice != nil {
			args = append(args, *variant.CostPrice)
			assignments = append(assignments, fmt.Sprintf("cost_price=$%d", len(args)))
		}
		if variant.StockQuantity != nil {
			args = append(args, *variant.StockQuantity)
			assignments = append(assignments, fmt.Sprintf("stock_quantity=$%d", len(args)))
		}
		if len(assignments) == 0 {
			continue
		}
		assignments = append(assignments, "updated_at=now()")
		query := fmt.Sprintf(`
			update commerce_product_variants
			set %s
			where organization_id=$1 and store_id=$2 and product_id=$3
		`, strings.Join(assignments, ","))
		if variant.ID != "" {
			id, err := uuid.Parse(variant.ID)
			if err != nil {
				return fmt.Errorf("invalid variant id")
			}
			args = append(args, id)
			query += fmt.Sprintf(" and id=$%d", len(args))
		} else {
			args = append(args, strings.TrimSpace(variant.Color), strings.TrimSpace(variant.Size))
			query += fmt.Sprintf(" and color=$%d and size=$%d", len(args)-1, len(args))
		}
		_, err := r.pool.Exec(ctx, query, args...)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *CommerceRepository) ensureProductVariants(ctx context.Context, organizationID, storeID, productID uuid.UUID, input CommerceProductInput) error {
	if len(input.Colors) == 0 || len(input.Sizes) == 0 {
		return nil
	}
	var slug string
	if err := r.pool.QueryRow(ctx, `select slug from commerce_products where organization_id=$1 and store_id=$2 and id=$3`, organizationID, storeID, productID).Scan(&slug); err != nil {
		return err
	}
	if input.StockQuantity <= 0 {
		input.StockQuantity = 24
	}
	variantOverrides := productVariantOverrides(input.Variants)
	for _, color := range input.Colors {
		for _, size := range input.Sizes {
			sku := strings.ToUpper(strings.ReplaceAll(input.SKU, " ", "-"))
			if sku == "" {
				sku = strings.ToUpper(strings.ReplaceAll(slug, "-", ""))
			}
			colorCode := "NA"
			if strings.TrimSpace(color) != "" {
				colorCode = strings.ToUpper(color[:min(2, len(color))])
			}
			sku = fmt.Sprintf("%s-%s-%s", sku, colorCode, size)
			override := variantOverrides[variantKey(color, size)]
			variantPrice := input.Price
			variantCompareAtPrice := input.CompareAtPrice
			variantCostPrice := input.CostPrice
			variantStockQuantity := input.StockQuantity
			if override.Price != nil {
				variantPrice = *override.Price
			}
			if override.CompareAtPrice != nil {
				variantCompareAtPrice = *override.CompareAtPrice
			}
			if override.CostPrice != nil {
				variantCostPrice = *override.CostPrice
			}
			if override.StockQuantity != nil {
				variantStockQuantity = *override.StockQuantity
			}
			_, err := r.pool.Exec(ctx, `
				insert into commerce_product_variants (organization_id,store_id,product_id,title,sku,color,size,price,compare_at_price,cost_price,stock_quantity)
				values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
				on conflict (store_id,sku) do nothing
			`, organizationID, storeID, productID, strings.TrimSpace(color+" / "+size), sku, color, size, variantPrice, variantCompareAtPrice, variantCostPrice, variantStockQuantity)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func productVariantOverrides(variants []CommerceProductVariantInput) map[string]CommerceProductVariantInput {
	out := map[string]CommerceProductVariantInput{}
	for _, variant := range variants {
		key := variantKey(variant.Color, variant.Size)
		if key != "|" {
			out[key] = variant
		}
	}
	return out
}

func variantKey(color, size string) string {
	return strings.ToLower(strings.TrimSpace(color)) + "|" + strings.ToLower(strings.TrimSpace(size))
}

func slugify(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var b strings.Builder
	lastDash := false
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastDash = false
		} else if !lastDash {
			b.WriteRune('-')
			lastDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, value := range values {
		normalized := strings.TrimSpace(strings.ToLower(value))
		if normalized == "" || seen[normalized] {
			continue
		}
		seen[normalized] = true
		out = append(out, normalized)
	}
	return out
}

func stringValue(value any) string {
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}
