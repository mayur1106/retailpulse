package service

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func TestFetchOrdersFollowsNextToken(t *testing.T) {
	t.Parallel()

	requests := 0
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests++
		body := `{"payload":{"NextToken":"page-2","Orders":[{"AmazonOrderId":"order-1","MarketplaceId":"ATVPDKIKX0DER","OrderStatus":"Unshipped","PurchaseDate":"2026-07-09T10:00:00Z","OrderTotal":{"CurrencyCode":"USD","Amount":"10.00"}}]}}`
		if request.URL.Query().Get("NextToken") == "page-2" {
			body = `{"payload":{"Orders":[{"AmazonOrderId":"order-2","MarketplaceId":"ATVPDKIKX0DER","OrderStatus":"Shipped","PurchaseDate":"2026-07-10T10:00:00Z","OrderTotal":{"CurrencyCode":"USD","Amount":"20.00"}}]}}`
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(body)),
		}, nil
	})}

	service := &AmazonService{
		client: client,
		config: AmazonConfig{
			SPAPIEndpoint: "https://sellingpartnerapi.example.com",
			AWSAccessKey:  "access",
			AWSSecretKey:  "secret",
			AWSRegion:     "us-east-1",
		},
	}

	orders, err := service.fetchOrders(context.Background(), "lwa-token", "ATVPDKIKX0DER", time.Now().Add(-24*time.Hour))
	if err != nil {
		t.Fatalf("fetchOrders returned error: %v", err)
	}
	if requests != 2 {
		t.Fatalf("expected 2 requests, got %d", requests)
	}
	if len(orders) != 2 || orders[0].AmazonOrderID != "order-1" || orders[1].AmazonOrderID != "order-2" {
		t.Fatalf("unexpected orders: %#v", orders)
	}
}
