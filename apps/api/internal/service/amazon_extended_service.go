package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/google/uuid"

	"retailpulse/apps/api/internal/domain"
	platformaws "retailpulse/apps/api/internal/platform/aws"
	"retailpulse/apps/api/internal/repository"
)

func (s *AmazonService) ImportDataset(ctx context.Context, principal domain.Principal, storeID uuid.UUID, dataset, marketplaceID string) (domain.AmazonDatasetImportResult, error) {
	started := time.Now().UTC()
	store, encrypted, err := s.repo.GetStore(ctx, principal.OrganizationID, storeID)
	if err != nil {
		return domain.AmazonDatasetImportResult{}, err
	}
	refresh, err := s.encryptor.Decrypt(encrypted)
	if err != nil {
		return domain.AmazonDatasetImportResult{}, err
	}
	clientID, secret, endpoint := s.config.ClientID, s.config.ClientSecret, s.spAPIEndpoint(store.Region)
	if store.Environment == "sandbox" {
		clientID, secret, endpoint = s.config.SandboxClientID, s.config.SandboxSecret, s.config.SandboxEndpoint
	}
	token, err := s.refreshAccessToken(ctx, string(refresh), clientID, secret)
	if err != nil {
		return domain.AmazonDatasetImportResult{}, err
	}
	sandbox := store.Environment == "sandbox"
	if marketplaceID == "" {
		marketplaceID = "ATVPDKIKX0DER"
	}
	count := 0
	switch dataset {
	case "inventory", "catalog":
		items, e := s.fetchInventory(ctx, token.AccessToken, endpoint, marketplaceID, sandbox)
		if e != nil {
			err = e
		} else {
			count, err = s.repo.UpsertInventory(ctx, principal.OrganizationID, storeID, items)
		}
	case "finances":
		events, e := s.fetchFinances(ctx, token.AccessToken, endpoint, time.Now().UTC().AddDate(0, -6, 0), sandbox)
		if e != nil {
			err = e
		} else {
			count, err = s.repo.UpsertFinancialEvents(ctx, principal.OrganizationID, storeID, events)
		}
	case "reports":
		reports, e := s.fetchReports(ctx, token.AccessToken, endpoint, sandbox)
		if e != nil {
			err = e
		} else {
			count, err = s.repo.UpsertReportRecords(ctx, principal.OrganizationID, storeID, reports)
		}
	case "campaigns":
		campaigns, e := s.fetchCampaigns(ctx)
		if e != nil {
			err = e
		} else {
			count, err = s.repo.UpsertCampaigns(ctx, principal.OrganizationID, storeID, campaigns)
		}
	default:
		return domain.AmazonDatasetImportResult{}, fmt.Errorf("%w: unsupported dataset %s", domain.ErrValidation, dataset)
	}
	finished := time.Now().UTC()
	status := "completed"
	message := ""
	if err != nil {
		status = "failed"
		message = err.Error()
	}
	_ = s.repo.RecordSync(ctx, principal.OrganizationID, storeID, dataset, status, count, message, started, finished)
	if err != nil {
		return domain.AmazonDatasetImportResult{}, err
	}
	return domain.AmazonDatasetImportResult{StoreID: storeID, Dataset: dataset, RecordsImported: count, StartedAt: started, FinishedAt: finished}, nil
}

func (s *AmazonService) fetchCampaigns(ctx context.Context) ([]repository.ImportedCampaign, error) {
	if s.config.AdsClientID == "" || s.config.AdsClientSecret == "" || s.config.AdsRefreshToken == "" || s.config.AdsProfileID == "" {
		return nil, fmt.Errorf("%w: Amazon Ads credentials and profile are not configured", domain.ErrConfiguration)
	}
	token, err := s.refreshAccessToken(ctx, s.config.AdsRefreshToken, s.config.AdsClientID, s.config.AdsClientSecret)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, s.config.AdsEndpoint+"/v2/sp/campaigns", nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+token.AccessToken)
	request.Header.Set("Amazon-Advertising-API-ClientId", s.config.AdsClientID)
	request.Header.Set("Amazon-Advertising-API-Scope", s.config.AdsProfileID)
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
		return nil, fmt.Errorf("Amazon Ads request failed: status %d: %s", response.StatusCode, string(body))
	}
	var parsed []struct {
		CampaignID    any     `json:"campaignId"`
		Name          string  `json:"name"`
		State         string  `json:"state"`
		TargetingType string  `json:"targetingType"`
		DailyBudget   float64 `json:"dailyBudget"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	campaigns := make([]repository.ImportedCampaign, 0, len(parsed))
	for _, x := range parsed {
		campaigns = append(campaigns, repository.ImportedCampaign{CampaignID: fmt.Sprint(x.CampaignID), Channel: "amazon_ads", Name: x.Name, CampaignType: x.TargetingType, Status: x.State, Budget: x.DailyBudget})
	}
	return campaigns, nil
}

func (s *AmazonService) spGet(ctx context.Context, accessToken, rawURL string, sandbox bool) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
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
		return nil, fmt.Errorf("Amazon API request failed: status %d: %s", response.StatusCode, string(body))
	}
	return body, nil
}

func (s *AmazonService) fetchInventory(ctx context.Context, token, base, marketplace string, sandbox bool) ([]repository.ImportedInventory, error) {
	endpoint, _ := url.Parse(base + "/fba/inventory/v1/summaries")
	q := endpoint.Query()
	q.Set("granularityType", "Marketplace")
	q.Set("granularityId", marketplace)
	q.Set("marketplaceIds", marketplace)
	q.Set("details", "true")
	endpoint.RawQuery = q.Encode()
	body, err := s.spGet(ctx, token, endpoint.String(), sandbox)
	if err != nil {
		return nil, err
	}
	var parsed struct {
		Payload struct {
			InventorySummaries []struct {
				ASIN        string `json:"asin"`
				SellerSKU   string `json:"sellerSku"`
				ProductName string `json:"productName"`
				Total       int    `json:"totalQuantity"`
				Details     struct {
					Fulfillable    int `json:"fulfillableQuantity"`
					InboundWorking int `json:"inboundWorkingQuantity"`
					InboundShipped int `json:"inboundShippedQuantity"`
					Reserved       int `json:"reservedQuantity"`
				} `json:"inventoryDetails"`
			} `json:"inventorySummaries"`
		} `json:"payload"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	items := make([]repository.ImportedInventory, 0, len(parsed.Payload.InventorySummaries))
	for _, x := range parsed.Payload.InventorySummaries {
		fulfillable := x.Details.Fulfillable
		if fulfillable == 0 {
			fulfillable = x.Total
		}
		items = append(items, repository.ImportedInventory{ASIN: x.ASIN, SKU: x.SellerSKU, Title: x.ProductName, Fulfillable: fulfillable, Inbound: x.Details.InboundWorking + x.Details.InboundShipped, Reserved: x.Details.Reserved})
	}
	return items, nil
}

func (s *AmazonService) fetchFinances(ctx context.Context, token, base string, after time.Time, sandbox bool) ([]repository.ImportedFinancialEvent, error) {
	endpoint, _ := url.Parse(base + "/finances/v0/financialEvents")
	q := endpoint.Query()
	q.Set("PostedAfter", after.Format(time.RFC3339))
	endpoint.RawQuery = q.Encode()
	body, err := s.spGet(ctx, token, endpoint.String(), sandbox)
	if err != nil {
		return nil, err
	}
	var parsed struct {
		Payload struct {
			FinancialEvents map[string][]json.RawMessage `json:"FinancialEvents"`
		} `json:"payload"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	var events []repository.ImportedFinancialEvent
	for kind, list := range parsed.Payload.FinancialEvents {
		for _, raw := range list {
			var value map[string]any
			_ = json.Unmarshal(raw, &value)
			amount, currency := findMoney(value)
			posted := findTime(value)
			events = append(events, repository.ImportedFinancialEvent{EventType: kind, Amount: amount, Currency: currency, PostedAt: posted, RawPayload: raw})
		}
	}
	return events, nil
}

func findMoney(value map[string]any) (float64, string) {
	for _, v := range value {
		if m, ok := v.(map[string]any); ok {
			if raw, yes := m["CurrencyAmount"]; yes {
				amount, _ := strconv.ParseFloat(fmt.Sprint(raw), 64)
				currency := fmt.Sprint(m["CurrencyCode"])
				if currency == "<nil>" {
					currency = "USD"
				}
				return amount, currency
			}
			if amount, currency := findMoney(m); amount != 0 {
				return amount, currency
			}
		}
	}
	return 0, "USD"
}
func findTime(value map[string]any) time.Time {
	for key, v := range value {
		if key == "PostedDate" || key == "PostedDateTime" {
			parsed, _ := time.Parse(time.RFC3339, fmt.Sprint(v))
			return parsed
		}
		if m, ok := v.(map[string]any); ok {
			if parsed := findTime(m); !parsed.IsZero() {
				return parsed
			}
		}
	}
	return time.Now().UTC()
}

func (s *AmazonService) fetchReports(ctx context.Context, token, base string, sandbox bool) ([]repository.ImportedReport, error) {
	body, err := s.spGet(ctx, token, base+"/reports/2021-06-30/reports?pageSize=100", sandbox)
	if err != nil {
		return nil, err
	}
	var parsed struct {
		Reports []struct {
			ReportID         string `json:"reportId"`
			ReportType       string `json:"reportType"`
			ProcessingStatus string `json:"processingStatus"`
			CreatedTime      string `json:"createdTime"`
		} `json:"reports"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	reports := make([]repository.ImportedReport, 0, len(parsed.Reports))
	for _, x := range parsed.Reports {
		created, _ := time.Parse(time.RFC3339, x.CreatedTime)
		reports = append(reports, repository.ImportedReport{ReportID: x.ReportID, ReportType: x.ReportType, Status: x.ProcessingStatus, CreatedAt: created})
	}
	return reports, nil
}
