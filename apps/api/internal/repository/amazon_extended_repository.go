package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

func (r *PostgresAmazonRepository) UpsertInventory(ctx context.Context, organizationID, storeID uuid.UUID, items []ImportedInventory) (int, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)
	for _, item := range items {
		title := item.Title
		if title == "" {
			title = item.SKU
		}
		asin := item.ASIN
		if asin == "" {
			asin = "SKU-" + item.SKU
		}
		var productID uuid.UUID
		err = tx.QueryRow(ctx, `insert into products(organization_id,store_id,asin,sku,title,status,data_origin) values($1,$2,$3,nullif($4,''),$5,'active','amazon') on conflict(organization_id,store_id,asin,coalesce(sku,'')) do update set title=excluded.title,data_origin='amazon',updated_at=now() returning id`, organizationID, storeID, asin, item.SKU, title).Scan(&productID)
		if err != nil {
			return 0, err
		}
		_, err = tx.Exec(ctx, `insert into inventory(organization_id,product_id,fulfillable_quantity,inbound_quantity,reserved_quantity,data_origin) values($1,$2,$3,$4,$5,'amazon') on conflict(organization_id,product_id) do update set fulfillable_quantity=excluded.fulfillable_quantity,inbound_quantity=excluded.inbound_quantity,reserved_quantity=excluded.reserved_quantity,data_origin='amazon',updated_at=now()`, organizationID, productID, item.Fulfillable, item.Inbound, item.Reserved)
		if err != nil {
			return 0, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return len(items), nil
}

func (r *PostgresAmazonRepository) UpsertFinancialEvents(ctx context.Context, organizationID, storeID uuid.UUID, events []ImportedFinancialEvent) (int, error) {
	for _, event := range events {
		if event.PostedAt.IsZero() {
			event.PostedAt = time.Now().UTC()
		}
		_, err := r.pool.Exec(ctx, `insert into financial_transactions(organization_id,store_id,transaction_type,amount,currency_code,posted_at,raw_payload,data_origin) select $1,$2,$3,$4,$5,$6,$7,'amazon' where not exists(select 1 from financial_transactions where organization_id=$1 and store_id=$2 and transaction_type=$3 and amount=$4 and posted_at=$6 and raw_payload=$7)`, organizationID, storeID, event.EventType, event.Amount, event.Currency, event.PostedAt, event.RawPayload)
		if err != nil {
			return 0, err
		}
	}
	return len(events), nil
}

func (r *PostgresAmazonRepository) UpsertReportRecords(ctx context.Context, organizationID uuid.UUID, storeID uuid.UUID, records []ImportedReport) (int, error) {
	for _, report := range records {
		_, err := r.pool.Exec(ctx, `insert into reports(organization_id,store_id,report_type,format,status,storage_key,created_at,data_origin) select $1,$2,$3,'amazon',$4,$5,$6,'amazon' where not exists(select 1 from reports where organization_id=$1 and store_id=$2 and storage_key=$5)`, organizationID, storeID, report.ReportType, report.Status, "amazon-report/"+report.ReportID, report.CreatedAt)
		if err != nil {
			return 0, err
		}
	}
	return len(records), nil
}

func (r *PostgresAmazonRepository) UpsertCampaigns(ctx context.Context, organizationID, storeID uuid.UUID, campaigns []ImportedCampaign) (int, error) {
	for _, campaign := range campaigns {
		channel := campaign.Channel
		if channel == "" {
			channel = "amazon_ads"
		}
		_, err := r.pool.Exec(ctx, `insert into campaigns(organization_id,store_id,amazon_campaign_id,channel,name,campaign_type,status,budget,data_origin) values($1,$2,$3,$4,$5,$6,$7,$8,'amazon') on conflict(organization_id,store_id,amazon_campaign_id) do update set channel=excluded.channel,name=excluded.name,campaign_type=excluded.campaign_type,status=excluded.status,budget=excluded.budget,data_origin='amazon',updated_at=now()`, organizationID, storeID, campaign.CampaignID, channel, campaign.Name, campaign.CampaignType, campaign.Status, campaign.Budget)
		if err != nil {
			return 0, err
		}
	}
	return len(campaigns), nil
}
