package aws

import (
	"net/http"
	"testing"
	"time"
)

func TestSigV4SignerAddsAuthorizationHeaders(t *testing.T) {
	request, err := http.NewRequest(http.MethodGet, "https://sellingpartnerapi-na.amazon.com/orders/v0/orders?CreatedAfter=2026-07-01T00%3A00%3A00Z&MarketplaceIds=ATVPDKIKX0DER", nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	request.Header.Set("x-amz-access-token", "lwa-token")
	request.Header.Set("accept", "application/json")

	signer := NewSigV4Signer(Credentials{
		AccessKey: "AKIDEXAMPLE",
		SecretKey: "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY",
		Region:    "us-east-1",
	})
	now := time.Date(2026, 7, 5, 12, 30, 0, 0, time.UTC)
	if err := signer.Sign(request, now); err != nil {
		t.Fatalf("sign: %v", err)
	}
	if request.Header.Get("X-Amz-Date") != "20260705T123000Z" {
		t.Fatalf("unexpected x-amz-date: %s", request.Header.Get("X-Amz-Date"))
	}
	if request.Header.Get("Authorization") == "" {
		t.Fatal("expected authorization header")
	}
}
