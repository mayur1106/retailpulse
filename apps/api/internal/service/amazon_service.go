package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"retailpulse/apps/api/internal/domain"
	platformaws "retailpulse/apps/api/internal/platform/aws"
	"retailpulse/apps/api/internal/platform/security"
	"retailpulse/apps/api/internal/repository"
)

type AmazonConfig struct {
	ClientID         string
	ClientSecret     string
	ApplicationID    string
	AuthVersion      string
	RedirectURL      string
	SellerCentralURL string
	SPAPIEndpoint    string
	AWSAccessKey     string
	AWSSecretKey     string
	AWSSessionToken  string
	AWSRegion        string
	SandboxClientID  string
	SandboxSecret    string
	SandboxToken     string
	SandboxEndpoint  string
	AdsClientID      string
	AdsClientSecret  string
	AdsRefreshToken  string
	AdsProfileID     string
	AdsEndpoint      string
}

type AmazonService struct {
	repo      repository.AmazonRepository
	auditRepo repository.AuditRepository
	encryptor security.Encryptor
	client    *http.Client
	config    AmazonConfig
}

type AmazonOAuthStart struct {
	AuthorizationURL string `json:"authorizationUrl"`
	State            string `json:"state"`
}

type AmazonConnectionStatus struct {
	Ready         bool   `json:"ready"`
	Mode          string `json:"mode"`
	SellerMessage string `json:"sellerMessage"`
	AdminMessage  string `json:"adminMessage,omitempty"`
}

type AmazonOAuthCallbackInput struct {
	State            string
	OAuthCode        string
	SellingPartnerID string
}

func NewAmazonService(repo repository.AmazonRepository, auditRepo repository.AuditRepository, encryptor security.Encryptor, config AmazonConfig) *AmazonService {
	return &AmazonService{
		repo:      repo,
		auditRepo: auditRepo,
		encryptor: encryptor,
		client:    &http.Client{Timeout: 30 * time.Second},
		config:    config,
	}
}

func (s *AmazonService) sellerCentralURL(region string) string {
	configured := strings.TrimRight(s.config.SellerCentralURL, "/")
	defaultNA := "https://sellercentral.amazon.com"
	if configured != "" && configured != defaultNA {
		return configured
	}
	switch strings.ToUpper(region) {
	case "EU":
		return "https://sellercentral-europe.amazon.com"
	case "FE":
		return "https://sellercentral.amazon.co.jp"
	default:
		return defaultNA
	}
}

func (s *AmazonService) spAPIEndpoint(region string) string {
	configured := strings.TrimRight(s.config.SPAPIEndpoint, "/")
	defaultNA := "https://sellingpartnerapi-na.amazon.com"
	if configured != "" && configured != defaultNA {
		return configured
	}
	switch strings.ToUpper(region) {
	case "EU":
		return "https://sellingpartnerapi-eu.amazon.com"
	case "FE":
		return "https://sellingpartnerapi-fe.amazon.com"
	default:
		return defaultNA
	}
}

func (s *AmazonService) ConnectionStatus() AmazonConnectionStatus {
	if s.config.ClientID == "" || s.config.ClientSecret == "" || s.config.ApplicationID == "" {
		return AmazonConnectionStatus{
			Ready:         false,
			Mode:          "not_configured",
			SellerMessage: "Amazon seller connection is not available yet. Please contact support.",
			AdminMessage:  "Configure AMAZON_LWA_CLIENT_ID, AMAZON_LWA_CLIENT_SECRET, and AMAZON_SPAPI_APP_ID on the SaaS backend.",
		}
	}
	if strings.HasPrefix(s.config.ApplicationID, "amzn1.sp.solution.") {
		return AmazonConnectionStatus{
			Ready:         false,
			Mode:          "invalid_application_id",
			SellerMessage: "Amazon seller connection is not available yet. Please contact support.",
			AdminMessage:  "AMAZON_SPAPI_APP_ID is currently a Solution ID. Configure the SP-API application ID sellers authorize, usually amzn1.sellerapps.app...",
		}
	}
	if strings.EqualFold(s.config.AuthVersion, "beta") {
		return AmazonConnectionStatus{
			Ready:         true,
			Mode:          "draft_beta",
			SellerMessage: "Amazon connection is available for draft/beta authorization. Sellers will approve access in Seller Central.",
			AdminMessage:  "Draft app mode is enabled with AMAZON_SPAPI_AUTH_VERSION=beta. Publish the app and clear this value for public onboarding.",
		}
	}
	return AmazonConnectionStatus{
		Ready:         true,
		Mode:          "production",
		SellerMessage: "Amazon connection is ready. Sellers approve access in Seller Central; no Solution Provider credentials are requested from sellers.",
	}
}

func (s *AmazonService) StartOAuth(ctx context.Context, principal domain.Principal, region string, marketplaceID string) (AmazonOAuthStart, error) {
	status := s.ConnectionStatus()
	if !status.Ready {
		return AmazonOAuthStart{}, errors.Join(domain.ErrConfiguration, errors.New(status.SellerMessage))
	}
	if region == "" {
		region = "NA"
	}
	if marketplaceID == "" {
		marketplaceID = "ATVPDKIKX0DER"
	}
	state, err := randomState()
	if err != nil {
		return AmazonOAuthStart{}, err
	}
	if err := s.repo.CreateOAuthState(ctx, state, principal, region, marketplaceID, time.Now().UTC().Add(15*time.Minute)); err != nil {
		return AmazonOAuthStart{}, err
	}
	authURL, err := url.Parse(s.sellerCentralURL(region) + "/apps/authorize/consent")
	if err != nil {
		return AmazonOAuthStart{}, err
	}
	query := authURL.Query()
	query.Set("application_id", s.config.ApplicationID)
	query.Set("state", state)
	query.Set("redirect_uri", s.config.RedirectURL)
	if strings.EqualFold(s.config.AuthVersion, "beta") {
		query.Set("version", "beta")
	}
	authURL.RawQuery = query.Encode()
	return AmazonOAuthStart{AuthorizationURL: authURL.String(), State: state}, nil
}

func (s *AmazonService) CompleteOAuth(ctx context.Context, input AmazonOAuthCallbackInput) (domain.AmazonStore, error) {
	if input.State == "" || input.OAuthCode == "" {
		return domain.AmazonStore{}, domain.ErrValidation
	}
	principal, region, _, err := s.repo.ConsumeOAuthState(ctx, input.State)
	if err != nil {
		return domain.AmazonStore{}, err
	}
	token, err := s.exchangeAuthorizationCode(ctx, input.OAuthCode)
	if err != nil {
		return domain.AmazonStore{}, err
	}
	if token.RefreshToken == "" {
		return domain.AmazonStore{}, errors.New("Amazon authorization did not return a refresh token")
	}
	encrypted, err := s.encryptor.Encrypt([]byte(token.RefreshToken))
	if err != nil {
		return domain.AmazonStore{}, err
	}
	sellerID := input.SellingPartnerID
	if sellerID == "" {
		sellerID = "amazon-authorized-seller"
	}
	store, err := s.repo.UpsertStore(ctx, principal.OrganizationID, "Amazon Store", sellerID, region, "production", encrypted)
	if err != nil {
		return domain.AmazonStore{}, err
	}
	_ = s.auditRepo.Record(ctx, principal.OrganizationID, principal.UserID, "amazon.connect", "store", store.ID.String(), "", "")
	return store, nil
}

func (s *AmazonService) ConnectSandbox(ctx context.Context, principal domain.Principal, region string) (domain.AmazonStore, error) {
	if s.config.SandboxClientID == "" || s.config.SandboxSecret == "" || s.config.SandboxToken == "" {
		return domain.AmazonStore{}, errors.Join(domain.ErrConfiguration, errors.New("Amazon sandbox credentials are not configured"))
	}
	if region == "" {
		region = "NA"
	}
	encrypted, err := s.encryptor.Encrypt([]byte(s.config.SandboxToken))
	if err != nil {
		return domain.AmazonStore{}, err
	}
	store, err := s.repo.UpsertStore(ctx, principal.OrganizationID, "Amazon Sandbox", "amazon-sandbox-seller", region, "sandbox", encrypted)
	if err != nil {
		return domain.AmazonStore{}, err
	}
	_ = s.auditRepo.Record(ctx, principal.OrganizationID, principal.UserID, "amazon.connect.sandbox", "store", store.ID.String(), "", "")
	return store, nil
}

func (s *AmazonService) ListStores(ctx context.Context, principal domain.Principal) ([]domain.AmazonStore, error) {
	return s.repo.ListStores(ctx, principal.OrganizationID)
}

func (s *AmazonService) ImportOrders(ctx context.Context, principal domain.Principal, storeID uuid.UUID, marketplaceID string, createdAfter time.Time) (domain.AmazonOrderImportResult, error) {
	startedAt := time.Now().UTC()
	store, encryptedRefreshToken, err := s.repo.GetStore(ctx, principal.OrganizationID, storeID)
	if err != nil {
		return domain.AmazonOrderImportResult{}, err
	}
	refreshTokenBytes, err := s.encryptor.Decrypt(encryptedRefreshToken)
	if err != nil {
		return domain.AmazonOrderImportResult{}, err
	}
	clientID, clientSecret, endpoint := s.config.ClientID, s.config.ClientSecret, s.spAPIEndpoint(store.Region)
	if store.Environment == "sandbox" {
		clientID, clientSecret, endpoint = s.config.SandboxClientID, s.config.SandboxSecret, s.config.SandboxEndpoint
	}
	lwaToken, err := s.refreshAccessToken(ctx, string(refreshTokenBytes), clientID, clientSecret)
	if err != nil {
		_ = s.repo.RecordSync(ctx, principal.OrganizationID, store.ID, "orders", "failed", 0, err.Error(), startedAt, time.Now().UTC())
		return domain.AmazonOrderImportResult{}, err
	}
	orders, err := s.fetchOrdersFromEndpoint(ctx, lwaToken.AccessToken, marketplaceID, createdAfter, endpoint, store.Environment == "sandbox")
	if err != nil {
		_ = s.repo.RecordSync(ctx, principal.OrganizationID, store.ID, "orders", "failed", 0, err.Error(), startedAt, time.Now().UTC())
		return domain.AmazonOrderImportResult{}, err
	}
	for index := range orders {
		items, itemErr := s.fetchOrderItems(ctx, lwaToken.AccessToken, endpoint, orders[index].AmazonOrderID, store.Environment == "sandbox")
		if itemErr == nil {
			orders[index].Items = items
		}
	}
	count, err := s.repo.UpsertOrders(ctx, principal.OrganizationID, store.ID, orders)
	if err != nil {
		_ = s.repo.RecordSync(ctx, principal.OrganizationID, store.ID, "orders", "failed", 0, err.Error(), startedAt, time.Now().UTC())
		return domain.AmazonOrderImportResult{}, err
	}
	finishedAt := time.Now().UTC()
	_ = s.repo.RecordSync(ctx, principal.OrganizationID, store.ID, "orders", "completed", count, "", startedAt, finishedAt)
	_ = s.auditRepo.Record(ctx, principal.OrganizationID, principal.UserID, "amazon.import.orders", "store", store.ID.String(), "", "")
	return domain.AmazonOrderImportResult{StoreID: store.ID, OrdersImported: count, StartedAt: startedAt, FinishedAt: finishedAt}, nil
}

func (s *AmazonService) fetchOrderItems(ctx context.Context, accessToken, baseEndpoint, orderID string, sandbox bool) ([]repository.ImportedOrderItem, error) {
	endpoint, err := url.Parse(baseEndpoint + "/orders/v0/orders/" + url.PathEscape(orderID) + "/orderItems")
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("x-amz-access-token", accessToken)
	request.Header.Set("Accept", "application/json")
	if !sandbox {
		signer := platformaws.NewSigV4Signer(platformaws.Credentials{AccessKey: s.config.AWSAccessKey, SecretKey: s.config.AWSSecretKey, SessionToken: s.config.AWSSessionToken, Region: s.config.AWSRegion, Service: "execute-api"})
		if err := signer.Sign(request, time.Now().UTC()); err != nil {
			return nil, err
		}
	}
	response, err := s.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("SP-API order items request failed: status %d: %s", response.StatusCode, string(body))
	}
	var parsed struct {
		Payload struct {
			OrderItems []struct {
				OrderItemID string `json:"OrderItemId"`
				ASIN        string `json:"ASIN"`
				SellerSKU   string `json:"SellerSKU"`
				Title       string `json:"Title"`
				Quantity    int    `json:"QuantityOrdered"`
				ItemPrice   *struct {
					Currency string `json:"CurrencyCode"`
					Amount   string `json:"Amount"`
				} `json:"ItemPrice"`
			} `json:"OrderItems"`
		} `json:"payload"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	items := make([]repository.ImportedOrderItem, 0, len(parsed.Payload.OrderItems))
	for _, item := range parsed.Payload.OrderItems {
		price := 0.0
		currency := "USD"
		if item.ItemPrice != nil {
			price, _ = strconv.ParseFloat(item.ItemPrice.Amount, 64)
			currency = item.ItemPrice.Currency
		}
		items = append(items, repository.ImportedOrderItem{AmazonOrderItemID: item.OrderItemID, ASIN: item.ASIN, SKU: item.SellerSKU, Title: item.Title, Quantity: item.Quantity, ItemPrice: price, CurrencyCode: currency})
	}
	return items, nil
}

type lwaTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

func (s *AmazonService) exchangeAuthorizationCode(ctx context.Context, code string) (lwaTokenResponse, error) {
	values := url.Values{}
	values.Set("grant_type", "authorization_code")
	values.Set("code", code)
	values.Set("client_id", s.config.ClientID)
	values.Set("client_secret", s.config.ClientSecret)
	values.Set("redirect_uri", s.config.RedirectURL)
	return s.callLWAToken(ctx, values)
}

func (s *AmazonService) refreshAccessToken(ctx context.Context, refreshToken string, clientID string, clientSecret string) (lwaTokenResponse, error) {
	values := url.Values{}
	values.Set("grant_type", "refresh_token")
	values.Set("refresh_token", refreshToken)
	values.Set("client_id", clientID)
	values.Set("client_secret", clientSecret)
	return s.callLWAToken(ctx, values)
}

func (s *AmazonService) callLWAToken(ctx context.Context, values url.Values) (lwaTokenResponse, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.amazon.com/auth/o2/token", bytes.NewBufferString(values.Encode()))
	if err != nil {
		return lwaTokenResponse{}, err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := s.client.Do(request)
	if err != nil {
		return lwaTokenResponse{}, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return lwaTokenResponse{}, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return lwaTokenResponse{}, fmt.Errorf("Amazon LWA token request failed: status %d: %s", response.StatusCode, string(body))
	}
	var token lwaTokenResponse
	if err := json.Unmarshal(body, &token); err != nil {
		return lwaTokenResponse{}, err
	}
	return token, nil
}

type spOrdersResponse struct {
	Payload struct {
		NextToken string `json:"NextToken"`
		Orders    []struct {
			AmazonOrderID string `json:"AmazonOrderId"`
			MarketplaceID string `json:"MarketplaceId"`
			OrderStatus   string `json:"OrderStatus"`
			PurchaseDate  string `json:"PurchaseDate"`
			OrderTotal    *struct {
				CurrencyCode string `json:"CurrencyCode"`
				Amount       string `json:"Amount"`
			} `json:"OrderTotal"`
		} `json:"Orders"`
	} `json:"payload"`
}

func (s *AmazonService) fetchOrders(ctx context.Context, accessToken string, marketplaceID string, createdAfter time.Time) ([]repository.ImportedOrder, error) {
	return s.fetchOrdersFromEndpoint(ctx, accessToken, marketplaceID, createdAfter, s.config.SPAPIEndpoint, false)
}

func (s *AmazonService) fetchOrdersFromEndpoint(ctx context.Context, accessToken string, marketplaceID string, createdAfter time.Time, baseEndpoint string, sandbox bool) ([]repository.ImportedOrder, error) {
	if marketplaceID == "" {
		marketplaceID = "ATVPDKIKX0DER"
	}
	endpoint, err := url.Parse(baseEndpoint + "/orders/v0/orders")
	if err != nil {
		return nil, err
	}
	query := endpoint.Query()
	query.Set("MarketplaceIds", marketplaceID)
	query.Set("CreatedAfter", createdAfter.UTC().Format(time.RFC3339))
	endpoint.RawQuery = query.Encode()

	return s.fetchOrderPages(ctx, accessToken, endpoint, nil, baseEndpoint, sandbox)
}

func (s *AmazonService) fetchOrderPages(ctx context.Context, accessToken string, endpoint *url.URL, orders []repository.ImportedOrder, baseEndpoint string, sandbox bool) ([]repository.ImportedOrder, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("x-amz-access-token", accessToken)
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "RetailPulseAI/0.1")
	signer := platformaws.NewSigV4Signer(platformaws.Credentials{
		AccessKey:    s.config.AWSAccessKey,
		SecretKey:    s.config.AWSSecretKey,
		SessionToken: s.config.AWSSessionToken,
		Region:       s.config.AWSRegion,
		Service:      "execute-api",
	})
	if !sandbox {
		if err := signer.Sign(request, time.Now().UTC()); err != nil {
			return nil, errors.Join(domain.ErrConfiguration, err)
		}
	}
	response, err := s.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("SP-API orders request failed: status %d: %s", response.StatusCode, string(body))
	}
	var parsed spOrdersResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	for _, order := range parsed.Payload.Orders {
		purchaseDate, err := time.Parse(time.RFC3339, order.PurchaseDate)
		if err != nil {
			continue
		}
		total := 0.0
		currency := "USD"
		if order.OrderTotal != nil {
			currency = order.OrderTotal.CurrencyCode
			total, _ = strconv.ParseFloat(order.OrderTotal.Amount, 64)
		}
		orders = append(orders, repository.ImportedOrder{
			AmazonOrderID:       order.AmazonOrderID,
			AmazonMarketplaceID: order.MarketplaceID,
			OrderStatus:         order.OrderStatus,
			PurchaseDate:        purchaseDate,
			OrderTotal:          total,
			CurrencyCode:        currency,
		})
	}
	if parsed.Payload.NextToken != "" {
		nextEndpoint, err := url.Parse(baseEndpoint + "/orders/v0/orders")
		if err != nil {
			return nil, err
		}
		query := nextEndpoint.Query()
		query.Set("NextToken", parsed.Payload.NextToken)
		nextEndpoint.RawQuery = query.Encode()
		return s.fetchOrderPages(ctx, accessToken, nextEndpoint, orders, baseEndpoint, sandbox)
	}
	return orders, nil
}

func randomState() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
