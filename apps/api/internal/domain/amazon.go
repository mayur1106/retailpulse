package domain

import (
	"time"

	"github.com/google/uuid"
)

type AmazonStore struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID uuid.UUID  `json:"organizationId"`
	Name           string     `json:"name"`
	SellerID       string     `json:"sellerId"`
	Region         string     `json:"region"`
	Environment    string     `json:"environment"`
	Status         string     `json:"status"`
	LastImportedAt *time.Time `json:"lastImportedAt"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

type AmazonOrderImportResult struct {
	StoreID        uuid.UUID `json:"storeId"`
	OrdersImported int       `json:"ordersImported"`
	StartedAt      time.Time `json:"startedAt"`
	FinishedAt     time.Time `json:"finishedAt"`
}

type AmazonDatasetImportResult struct {
	StoreID         uuid.UUID `json:"storeId"`
	Dataset         string    `json:"dataset"`
	RecordsImported int       `json:"recordsImported"`
	StartedAt       time.Time `json:"startedAt"`
	FinishedAt      time.Time `json:"finishedAt"`
}
