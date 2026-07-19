package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"retailpulse/apps/api/internal/domain"
)

type AnalyticsRepository struct{ pool *pgxpool.Pool }

func NewAnalyticsRepository(pool *pgxpool.Pool) *AnalyticsRepository {
	return &AnalyticsRepository{pool: pool}
}

func (r *AnalyticsRepository) StoreExists(ctx context.Context, organizationID, storeID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `select exists(select 1 from stores where id=$1 and organization_id=$2)`, storeID, organizationID).Scan(&exists)
	return exists, err
}

func (r *AnalyticsRepository) StoreEnvironment(ctx context.Context, organizationID, storeID uuid.UUID) (string, error) {
	var environment string
	err := r.pool.QueryRow(ctx, `select environment from stores where id=$1 and organization_id=$2`, storeID, organizationID).Scan(&environment)
	return environment, err
}

func (r *AnalyticsRepository) ResourceList(ctx context.Context, organizationID, storeID uuid.UUID, resource string) ([]map[string]any, error) {
	queries := map[string]string{
		"products":          `select p.id::text id,p.asin,coalesce(p.sku,'') sku,p.title,p.category,p.selling_price,p.status,p.data_origin,coalesce(i.fulfillable_quantity,0) available,coalesce(i.inbound_quantity,0) inbound from products p left join inventory i on i.product_id=p.id where p.organization_id=$1 and p.store_id=$2 order by p.updated_at desc limit 200`,
		"orders":            `select o.id::text id,o.amazon_order_id,o.order_status,o.purchase_date,o.order_total,o.currency_code,o.data_origin,coalesce(m.country_code,'') country_code,coalesce(sum(oi.quantity_ordered),0)::int items from orders o left join marketplaces m on m.id=o.marketplace_id left join order_items oi on oi.order_id=o.id where o.organization_id=$1 and o.store_id=$2 group by o.id,m.country_code order by o.purchase_date desc limit 200`,
		"inventory":         `select p.asin,coalesce(p.sku,'') sku,p.title,i.fulfillable_quantity,i.inbound_quantity,i.reserved_quantity,i.data_origin,i.updated_at from inventory i join products p on p.id=i.product_id where i.organization_id=$1 and p.store_id=$2 order by i.fulfillable_quantity asc limit 200`,
		"reports":           `select id::text id,report_type,format,status,storage_key,data_origin,created_at from reports where organization_id=$1 and store_id=$2 order by created_at desc limit 200`,
		"campaigns":         `select c.id::text id,c.name,c.channel,c.campaign_type,c.status,c.budget,c.data_origin,coalesce(sum(m.spend),0) spend,coalesce(sum(m.sales),0) sales,coalesce(sum(m.orders),0)::int orders,coalesce(sum(m.impressions),0)::int impressions,coalesce(sum(m.clicks),0)::int clicks from campaigns c left join campaign_metrics m on m.campaign_id=c.id where c.organization_id=$1 and c.store_id=$2 group by c.id order by sales desc limit 200`,
		"traffic":           `select p.id::text id,p.title,p.category,p.asin,coalesce(p.sku,'') sku,sum(t.sessions)::int sessions,sum(t.page_views)::int page_views,sum(t.units_ordered)::int units,coalesce(sum(t.ordered_revenue),0) revenue,case when sum(t.sessions)>0 then round((sum(t.units_ordered)::numeric/sum(t.sessions)::numeric*100),2) else 0 end conversion_rate,round(avg(t.buy_box_percentage),2) buy_box_percentage,max(t.data_origin) data_origin from product_traffic_metrics t join products p on p.id=t.product_id where t.organization_id=$1 and t.store_id=$2 group by p.id order by revenue desc limit 200`,
		"search-terms":      `select min(st.id::text) id,st.search_term,st.keyword_text,st.match_type,p.title product,c.name campaign,c.channel,sum(st.impressions)::int impressions,sum(st.clicks)::int clicks,coalesce(sum(st.spend),0) spend,coalesce(sum(st.sales),0) sales,sum(st.orders)::int orders,case when sum(st.clicks)>0 then round((sum(st.orders)::numeric/sum(st.clicks)::numeric*100),2) else 0 end conversion_rate,case when sum(st.spend)>0 then round((sum(st.sales)/sum(st.spend))::numeric,2) else 0 end roas,max(st.data_origin) data_origin from search_term_metrics st left join products p on p.id=st.product_id left join campaigns c on c.id=st.campaign_id where st.organization_id=$1 and st.store_id=$2 group by st.search_term,st.keyword_text,st.match_type,p.title,c.name,c.channel order by sales desc limit 200`,
		"returns":           `select re.id::text id,re.return_date,p.title product,p.category,p.asin,coalesce(p.sku,'') sku,re.reason,re.status,re.quantity,re.refund_amount,re.currency_code,re.data_origin from return_events re left join products p on p.id=re.product_id where re.organization_id=$1 and re.store_id=$2 order by re.return_date desc limit 200`,
		"regions":           `select min(id::text) id,country_code,region_code,city,sum(orders)::int orders,sum(units)::int units,coalesce(sum(revenue),0) revenue,coalesce(sum(refunds),0) refunds,coalesce(sum(ad_spend),0) ad_spend,case when sum(ad_spend)>0 then round((sum(revenue)/sum(ad_spend))::numeric,2) else 0 end roas,max(data_origin) data_origin from regional_sales_metrics where organization_id=$1 and store_id=$2 group by country_code,region_code,city order by revenue desc limit 200`,
		"profit":            `with sales as (select oi.product_id,coalesce(sum(oi.quantity_ordered),0)::int units,coalesce(sum(oi.item_price*oi.quantity_ordered),0) revenue from order_items oi join orders o on o.id=oi.order_id where o.organization_id=$1 and o.store_id=$2 group by oi.product_id), ads as (select product_id,sum(spend) spend from advertised_product_metrics where organization_id=$1 and store_id=$2 group by product_id), refunds as (select product_id,sum(refund_amount) refunds from return_events where organization_id=$1 and store_id=$2 group by product_id) select p.id::text id,p.title,p.category,p.asin,coalesce(p.sku,'') sku,coalesce(s.units,0)::int units,coalesce(s.revenue,0) revenue,coalesce(s.units*p.cost_price,0) cogs,coalesce(a.spend,0) ad_spend,coalesce(r.refunds,0) refunds,round((coalesce(s.revenue,0)*0.15)::numeric,2) estimated_fees,round((coalesce(s.revenue,0)-coalesce(s.units*p.cost_price,0)-coalesce(a.spend,0)-coalesce(r.refunds,0)-(coalesce(s.revenue,0)*0.15))::numeric,2) estimated_profit,case when coalesce(s.revenue,0)>0 then round(((coalesce(s.revenue,0)-coalesce(s.units*p.cost_price,0)-coalesce(a.spend,0)-coalesce(r.refunds,0)-(coalesce(s.revenue,0)*0.15))/coalesce(s.revenue,0)*100)::numeric,2) else 0 end margin from products p left join sales s on s.product_id=p.id left join ads a on a.product_id=p.id left join refunds r on r.product_id=p.id where p.organization_id=$1 and p.store_id=$2 group by p.id,s.units,s.revenue,a.spend,r.refunds order by estimated_profit desc limit 200`,
		"ads-intelligence":  `select min(apm.id::text) id,p.title product,p.category,c.name campaign,c.channel,sum(apm.impressions)::int impressions,sum(apm.clicks)::int clicks,coalesce(sum(apm.spend),0) spend,coalesce(sum(apm.sales),0) sales,sum(apm.orders)::int orders,case when sum(apm.clicks)>0 then round((sum(apm.orders)::numeric/sum(apm.clicks)::numeric*100),2) else 0 end conversion_rate,case when sum(apm.spend)>0 then round((sum(apm.sales)/sum(apm.spend))::numeric,2) else 0 end roas,case when sum(apm.sales)>0 then round((sum(apm.spend)/sum(apm.sales)*100)::numeric,2) else 0 end acos,case when sum(apm.spend)>0 and sum(apm.sales)/sum(apm.spend) < 1.5 then 'Stop or reduce' when sum(apm.spend)>0 and sum(apm.sales)/sum(apm.spend) >= 3 then 'Scale' else 'Monitor' end recommendation,max(apm.data_origin) data_origin from advertised_product_metrics apm join products p on p.id=apm.product_id left join campaigns c on c.id=apm.campaign_id where apm.organization_id=$1 and apm.store_id=$2 group by p.title,p.category,c.name,c.channel order by sales desc limit 200`,
		"ads-optimization":  `with campaign as (select c.id::text id,c.name campaign,c.channel,c.campaign_type,c.status,c.budget,coalesce(sum(m.impressions),0)::int impressions,coalesce(sum(m.clicks),0)::int clicks,coalesce(sum(m.orders),0)::int orders,coalesce(sum(m.spend),0) spend,coalesce(sum(m.sales),0) sales from campaigns c left join campaign_metrics m on m.campaign_id=c.id and m.metric_date >= current_date-89 where c.organization_id=$1 and c.store_id=$2 group by c.id), metrics as (select *,case when clicks>0 then round((orders::numeric/clicks::numeric*100),2) else 0 end conversion_rate,case when spend>0 then round((sales/spend)::numeric,2) else 0 end roas,case when sales>0 then round((spend/sales*100)::numeric,2) else 0 end acos from campaign) select id,campaign,channel,campaign_type,status,budget,impressions,clicks,orders,spend,sales,conversion_rate,roas,acos,case when spend >= 50 and roas >= 4 and conversion_rate >= 8 then 'Scale' when spend >= 50 and roas < 1.25 then 'Pause' when spend >= 50 and roas < 2 then 'Reduce' when clicks > 200 and conversion_rate < 2 then 'Fix targeting' else 'Monitor' end decision,case when spend >= 50 and roas >= 4 and conversion_rate >= 8 then '+20% budget test' when spend >= 50 and roas < 1.25 then 'Pause or cap spend' when spend >= 50 and roas < 2 then '-25% budget' when clicks > 200 and conversion_rate < 2 then 'Refine audience/keywords' else 'No immediate budget change' end budget_action,case when spend >= 50 and roas >= 4 and conversion_rate >= 8 then 'High ROAS and conversion indicate room to scale.' when spend >= 50 and roas < 1.25 then 'Spend is not converting into profitable sales.' when spend >= 50 and roas < 2 then 'Campaign needs lower bids or tighter targeting.' when clicks > 200 and conversion_rate < 2 then 'Traffic is not converting; audience or keyword intent may be weak.' else 'Performance is stable; keep monitoring.' end reason from metrics order by case when spend >= 50 and roas >= 4 and conversion_rate >= 8 then 1 when spend >= 50 and roas < 1.25 then 2 when spend >= 50 and roas < 2 then 3 when clicks > 200 and conversion_rate < 2 then 4 else 5 end,spend desc limit 200`,
		"product-decisions": `with sales as (select oi.product_id,coalesce(sum(oi.quantity_ordered),0)::int units,coalesce(sum(oi.item_price*oi.quantity_ordered),0) revenue from order_items oi join orders o on o.id=oi.order_id where o.organization_id=$1 and o.store_id=$2 and o.purchase_date >= current_date-89 group by oi.product_id), ads as (select product_id,coalesce(sum(spend),0) ad_spend,coalesce(sum(sales),0) ad_sales from advertised_product_metrics where organization_id=$1 and store_id=$2 and metric_date >= current_date-89 group by product_id), returns as (select product_id,count(*)::int returns,coalesce(sum(refund_amount),0) refunds from return_events where organization_id=$1 and store_id=$2 and return_date >= current_date-89 group by product_id), traffic as (select product_id,coalesce(sum(sessions),0)::int sessions,coalesce(sum(units_ordered),0)::int traffic_units from product_traffic_metrics where organization_id=$1 and store_id=$2 and metric_date >= current_date-89 group by product_id), metrics as (select p.id::text id,p.title product,p.category,p.asin,coalesce(p.sku,'') sku,coalesce(s.units,0)::int units,coalesce(s.revenue,0) revenue,coalesce(i.fulfillable_quantity,0)::int inventory,case when coalesce(s.units,0)>0 then round((coalesce(i.fulfillable_quantity,0)::numeric/(coalesce(s.units,0)::numeric/90))::numeric,1) else 999 end days_cover,coalesce(a.ad_spend,0) ad_spend,coalesce(a.ad_sales,0) ad_sales,case when coalesce(a.ad_spend,0)>0 then round((coalesce(a.ad_sales,0)/coalesce(a.ad_spend,0))::numeric,2) else 0 end roas,coalesce(r.returns,0)::int returns,case when coalesce(s.units,0)>0 then round((coalesce(r.returns,0)::numeric/coalesce(s.units,0)::numeric*100),2) else 0 end return_rate,case when coalesce(t.sessions,0)>0 then round((coalesce(t.traffic_units,0)::numeric/coalesce(t.sessions,0)::numeric*100),2) else 0 end conversion_rate,case when coalesce(s.revenue,0)>0 then round(((coalesce(s.revenue,0)-coalesce(s.units*p.cost_price,0)-coalesce(a.ad_spend,0)-coalesce(r.refunds,0)-(coalesce(s.revenue,0)*0.15))/coalesce(s.revenue,0)*100)::numeric,2) else 0 end margin from products p left join sales s on s.product_id=p.id left join ads a on a.product_id=p.id left join returns r on r.product_id=p.id left join traffic t on t.product_id=p.id left join inventory i on i.product_id=p.id where p.organization_id=$1 and p.store_id=$2) select id,product,category,asin,sku,units,revenue,inventory,days_cover,ad_spend,roas,returns,return_rate,conversion_rate,margin,case when inventory < 20 or days_cover < 21 then 'Restock' when return_rate >= 8 then 'Fix listing' when revenue > 0 and margin < 12 and ad_spend > 0 then 'Reduce ads' when revenue > 0 and roas >= 3 and margin >= 20 and return_rate < 6 then 'Scale' when revenue > 0 and margin >= 15 then 'Hold' else 'Watch' end decision,case when inventory < 20 or days_cover < 21 then 'Inventory risk can block sales momentum.' when return_rate >= 8 then 'Return rate is high; inspect fit, quality, images, and size guidance.' when revenue > 0 and margin < 12 and ad_spend > 0 then 'Margin is weak after ad spend, refunds, costs, and estimated fees.' when revenue > 0 and roas >= 3 and margin >= 20 and return_rate < 6 then 'Strong ads, margin, and return profile make this a scale candidate.' when revenue > 0 and margin >= 15 then 'Product is profitable enough to maintain while monitoring trend and stock.' else 'Needs more demand or cleaner economics before major investment.' end reason from metrics order by case when inventory < 20 or days_cover < 21 then 2 when return_rate >= 8 then 3 when revenue > 0 and margin < 12 and ad_spend > 0 then 4 when revenue > 0 and roas >= 3 and margin >= 20 and return_rate < 6 then 1 when revenue > 0 and margin >= 15 then 5 else 6 end,revenue desc limit 200`,
		"recommendations":   `select gr.id::text id,gr.recommendation_type,gr.title,gr.reason,gr.region_code,coalesce(p.title,'') product,coalesce(c.name,'') campaign,gr.impact_score,gr.confidence,gr.status,gr.data_origin,gr.created_at from growth_recommendations gr left join products p on p.id=gr.product_id left join campaigns c on c.id=gr.campaign_id where gr.organization_id=$1 and gr.store_id=$2 order by gr.impact_score desc,gr.created_at desc limit 200`,
	}
	query, ok := queries[resource]
	if !ok {
		return nil, fmt.Errorf("unsupported resource")
	}
	rows, err := r.pool.Query(ctx, query, organizationID, storeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	fields := rows.FieldDescriptions()
	result := make([]map[string]any, 0)
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, err
		}
		item := make(map[string]any, len(values))
		for i, value := range values {
			item[string(fields[i].Name)] = value
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *AnalyticsRepository) Dashboard(ctx context.Context, organizationID, storeID uuid.UUID, days int) (domain.AnalyticsDashboard, error) {
	from := time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -days+1)
	var out domain.AnalyticsDashboard
	err := r.pool.QueryRow(ctx, `
		select coalesce(sum(revenue),0), coalesce(sum(profit),0), coalesce(sum(units_sold),0),
			coalesce(sum(ad_spend),0), coalesce(sum(refunds),0),
			(select count(*) from orders where organization_id=$1 and store_id=$2 and purchase_date >= $3),
			(select count(*) from products where organization_id=$1 and store_id=$2),
			(select coalesce(sum(i.fulfillable_quantity),0) from inventory i join products p on p.id=i.product_id where i.organization_id=$1 and p.store_id=$2)
		from daily_metrics where organization_id=$1 and store_id=$2 and metric_date >= $3
	`, organizationID, storeID, from).Scan(&out.Summary.Revenue, &out.Summary.Profit, &out.Summary.Units, &out.Summary.AdSpend, &out.Summary.Refunds, &out.Summary.Orders, &out.Summary.Products, &out.Summary.Inventory)
	if err != nil {
		return out, err
	}
	if out.Summary.AdSpend > 0 {
		out.Summary.ROAS = out.Summary.Revenue / out.Summary.AdSpend
	}
	if out.Summary.Revenue > 0 {
		out.Summary.ProfitMargin = out.Summary.Profit / out.Summary.Revenue * 100
	}

	rows, err := r.pool.Query(ctx, `select metric_date, sum(revenue), sum(profit), sum(ad_spend), sum(refunds), sum(units_sold) from daily_metrics where organization_id=$1 and store_id=$2 and metric_date >= $3 group by metric_date order by metric_date`, organizationID, storeID, from)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var p domain.AnalyticsTrendPoint
		if err := rows.Scan(&p.Date, &p.Revenue, &p.Profit, &p.AdSpend, &p.Refunds, &p.Units); err != nil {
			return out, err
		}
		out.Trend = append(out.Trend, p)
	}

	productRows, err := r.pool.Query(ctx, `
		select p.asin, coalesce(p.sku,''), p.title,
			coalesce(sum(case when o.id is null then 0 else oi.item_price*oi.quantity_ordered end),0),
			coalesce(sum(case when o.id is null then 0 else oi.quantity_ordered end),0),
			coalesce(i.fulfillable_quantity,0)
		from products p
		left join order_items oi on oi.product_id=p.id
		left join orders o on o.id=oi.order_id and o.organization_id=$1 and o.store_id=$2 and o.purchase_date >= $3
		left join inventory i on i.product_id=p.id
		where p.organization_id=$1 and p.store_id=$2
		group by p.id,i.fulfillable_quantity order by 4 desc limit 8`, organizationID, storeID, from)
	if err != nil {
		return out, err
	}
	defer productRows.Close()
	for productRows.Next() {
		var p domain.ProductPerformance
		if err := productRows.Scan(&p.ASIN, &p.SKU, &p.Title, &p.Revenue, &p.Units, &p.Available); err != nil {
			return out, err
		}
		out.Products = append(out.Products, p)
	}

	campaignRows, err := r.pool.Query(ctx, `select c.channel,c.name,c.status,coalesce(sum(m.spend),0),coalesce(sum(m.sales),0),coalesce(sum(m.orders),0),coalesce(sum(m.impressions),0),coalesce(sum(m.clicks),0) from campaigns c left join campaign_metrics m on m.campaign_id=c.id and m.metric_date >= $3 where c.organization_id=$1 and c.store_id=$2 group by c.id order by 5 desc`, organizationID, storeID, from)
	if err != nil {
		return out, err
	}
	defer campaignRows.Close()
	for campaignRows.Next() {
		var c domain.CampaignPerformance
		if err := campaignRows.Scan(&c.Channel, &c.Name, &c.Status, &c.Spend, &c.Sales, &c.Orders, &c.Impressions, &c.Clicks); err != nil {
			return out, err
		}
		if c.Spend > 0 {
			c.ROAS = c.Sales / c.Spend
		}
		out.Campaigns = append(out.Campaigns, c)
	}
	return out, rows.Err()
}

func (r *AnalyticsRepository) GrowthIntelligence(ctx context.Context, organizationID, storeID uuid.UUID, days int) (domain.GrowthIntelligence, error) {
	if days < 30 {
		days = 30
	}
	if days > 365 {
		days = 365
	}
	from := time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -days+1)
	previousFrom := from.AddDate(0, 0, -days)
	var out domain.GrowthIntelligence

	rows, err := r.pool.Query(ctx, `
		with current_sales as (
			select oi.product_id,
				coalesce(sum(oi.item_price * oi.quantity_ordered),0) revenue,
				coalesce(sum(oi.quantity_ordered),0)::int units
			from order_items oi
			join orders o on o.id=oi.order_id
			where o.organization_id=$1 and o.store_id=$2 and o.purchase_date >= $3
			group by oi.product_id
		),
		previous_sales as (
			select oi.product_id,
				coalesce(sum(oi.quantity_ordered),0)::int units
			from order_items oi
			join orders o on o.id=oi.order_id
			where o.organization_id=$1 and o.store_id=$2 and o.purchase_date >= $4 and o.purchase_date < $3
			group by oi.product_id
		),
		ads as (
			select product_id,
				coalesce(sum(spend),0) spend,
				coalesce(sum(sales),0) sales
			from advertised_product_metrics
			where organization_id=$1 and store_id=$2 and metric_date >= $3::date
			group by product_id
		)
		select p.id::text, p.asin, coalesce(p.sku,''), p.title, coalesce(p.category,''),
			coalesce(cs.revenue,0), coalesce(cs.units,0), coalesce(ps.units,0),
			coalesce(a.spend,0), coalesce(a.sales,0), coalesce(i.fulfillable_quantity,0), coalesce(p.cost_price,0)
		from products p
		left join current_sales cs on cs.product_id=p.id
		left join previous_sales ps on ps.product_id=p.id
		left join ads a on a.product_id=p.id
		left join inventory i on i.product_id=p.id
		where p.organization_id=$1 and p.store_id=$2
		order by coalesce(cs.revenue,0) desc, p.updated_at desc
		limit 100
	`, organizationID, storeID, from, previousFrom)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var insight domain.ProductGrowthInsight
		var costPrice float64
		if err := rows.Scan(&insight.ProductID, &insight.ASIN, &insight.SKU, &insight.Title, &insight.Category, &insight.Revenue, &insight.Units, &insight.PreviousUnits, &insight.AdSpend, &insight.AdSales, &insight.Inventory, &costPrice); err != nil {
			return out, err
		}
		if insight.PreviousUnits > 0 {
			insight.TrendPercent = (float64(insight.Units-insight.PreviousUnits) / float64(insight.PreviousUnits)) * 100
		} else if insight.Units > 0 {
			insight.TrendPercent = 100
		}
		if insight.AdSpend > 0 {
			insight.ROAS = insight.AdSales / insight.AdSpend
		}
		if insight.AdSales > 0 {
			insight.ACOS = insight.AdSpend / insight.AdSales * 100
		}
		if costPrice > 0 {
			insight.EstimatedProfit = insight.Revenue - (float64(insight.Units) * costPrice) - insight.AdSpend
		} else {
			insight.EstimatedProfit = (insight.Revenue * 0.35) - insight.AdSpend
		}
		insight.Action, insight.Reason = growthAction(insight, days)
		out.Products = append(out.Products, insight)
	}
	if err := rows.Err(); err != nil {
		return out, err
	}

	marketRows, err := r.pool.Query(ctx, `
		select coalesce(m.country_code,'Unknown'), coalesce(m.name,'Unknown marketplace'),
			count(distinct o.id)::int,
			coalesce(sum(oi.quantity_ordered),0)::int,
			coalesce(sum(oi.item_price * oi.quantity_ordered),0)
		from orders o
		left join marketplaces m on m.id=o.marketplace_id
		left join order_items oi on oi.order_id=o.id
		where o.organization_id=$1 and o.store_id=$2 and o.purchase_date >= $3
		group by m.country_code,m.name
		order by 5 desc
	`, organizationID, storeID, from)
	if err != nil {
		return out, err
	}
	defer marketRows.Close()
	for marketRows.Next() {
		var marketplace domain.MarketplaceGrowthInsight
		if err := marketRows.Scan(&marketplace.CountryCode, &marketplace.Name, &marketplace.Orders, &marketplace.Units, &marketplace.Revenue); err != nil {
			return out, err
		}
		out.Marketplaces = append(out.Marketplaces, marketplace)
	}
	return out, marketRows.Err()
}

func (r *AnalyticsRepository) SellerHealth(ctx context.Context, organizationID, storeID uuid.UUID, days int) (domain.SellerHealth, error) {
	if days < 30 {
		days = 30
	}
	if days > 365 {
		days = 365
	}
	from := time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -days+1)
	previousFrom := from.AddDate(0, 0, -days)

	dashboard, err := r.Dashboard(ctx, organizationID, storeID, days)
	if err != nil {
		return domain.SellerHealth{}, err
	}

	var previousRevenue float64
	if err := r.pool.QueryRow(ctx, `
		select coalesce(sum(revenue),0)
		from daily_metrics
		where organization_id=$1 and store_id=$2 and metric_date >= $3 and metric_date < $4
	`, organizationID, storeID, previousFrom, from).Scan(&previousRevenue); err != nil {
		return domain.SellerHealth{}, err
	}

	var productCount, lowStockProducts int
	var averageDaysCover float64
	if err := r.pool.QueryRow(ctx, `
		with velocity as (
			select oi.product_id, coalesce(sum(oi.quantity_ordered),0)::numeric / 30 daily_units
			from order_items oi
			join orders o on o.id=oi.order_id
			where o.organization_id=$1 and o.store_id=$2 and o.purchase_date >= current_date-29
			group by oi.product_id
		),
		stock as (
			select p.id,
				coalesce(i.fulfillable_quantity,0) available,
				coalesce(v.daily_units,0) daily_units,
				case when coalesce(v.daily_units,0) > 0 then coalesce(i.fulfillable_quantity,0)::numeric / v.daily_units else 999 end days_cover
			from products p
			left join inventory i on i.product_id=p.id
			left join velocity v on v.product_id=p.id
			where p.organization_id=$1 and p.store_id=$2
		)
		select count(*)::int,
			coalesce(sum(case when days_cover < 21 or available < 20 then 1 else 0 end),0)::int,
			coalesce(avg(case when days_cover < 999 then days_cover else null end),0)
		from stock
	`, organizationID, storeID).Scan(&productCount, &lowStockProducts, &averageDaysCover); err != nil {
		return domain.SellerHealth{}, err
	}

	var returnEvents int
	var unitsSold int
	var refundAmount float64
	if err := r.pool.QueryRow(ctx, `
		select
			(select count(*)::int from return_events where organization_id=$1 and store_id=$2 and return_date >= current_date-89),
			(select coalesce(sum(quantity_ordered),0)::int from order_items oi join orders o on o.id=oi.order_id where o.organization_id=$1 and o.store_id=$2 and o.purchase_date >= current_date-89),
			(select coalesce(sum(refund_amount),0) from return_events where organization_id=$1 and store_id=$2 and return_date >= current_date-89)
	`, organizationID, storeID).Scan(&returnEvents, &unitsSold, &refundAmount); err != nil {
		return domain.SellerHealth{}, err
	}

	dataOrigin := "production"
	if err := r.pool.QueryRow(ctx, `
		select coalesce(max(data_origin),'production')
		from daily_metrics
		where organization_id=$1 and store_id=$2
	`, organizationID, storeID).Scan(&dataOrigin); err != nil {
		return domain.SellerHealth{}, err
	}

	revenueTrend := 0.0
	if previousRevenue > 0 {
		revenueTrend = ((dashboard.Summary.Revenue - previousRevenue) / previousRevenue) * 100
	} else if dashboard.Summary.Revenue > 0 {
		revenueTrend = 100
	}
	returnRate := 0.0
	if unitsSold > 0 {
		returnRate = float64(returnEvents) / float64(unitsSold) * 100
	}

	revenueScore := scoreRevenueTrend(revenueTrend, dashboard.Summary.Revenue)
	profitScore := scoreProfitMargin(dashboard.Summary.ProfitMargin)
	adsScore := scoreROAS(dashboard.Summary.ROAS, dashboard.Summary.AdSpend)
	inventoryScore := scoreInventory(productCount, lowStockProducts, averageDaysCover)
	returnScore := scoreReturns(returnRate, refundAmount)

	metrics := []domain.SellerHealthMetric{
		{Key: "revenue", Label: "Revenue momentum", Score: revenueScore, Status: scoreStatus(revenueScore), Value: fmt.Sprintf("%.1f%%", revenueTrend), Detail: "Compares current period revenue with the previous matching period.", Weight: 0.25},
		{Key: "profit", Label: "Profit quality", Score: profitScore, Status: scoreStatus(profitScore), Value: fmt.Sprintf("%.1f%% margin", dashboard.Summary.ProfitMargin), Detail: "Measures estimated profit margin after ads, refunds, and known costs.", Weight: 0.25},
		{Key: "ads", Label: "Ads efficiency", Score: adsScore, Status: scoreStatus(adsScore), Value: fmt.Sprintf("%.2fx ROAS", dashboard.Summary.ROAS), Detail: "Blends ad spend and attributed sales across connected ad channels.", Weight: 0.20},
		{Key: "inventory", Label: "Inventory health", Score: inventoryScore, Status: scoreStatus(inventoryScore), Value: fmt.Sprintf("%d low-stock SKUs", lowStockProducts), Detail: fmt.Sprintf("Average stock cover is %.0f days for products with recent demand.", averageDaysCover), Weight: 0.15},
		{Key: "returns", Label: "Return risk", Score: returnScore, Status: scoreStatus(returnScore), Value: fmt.Sprintf("%.1f%% return rate", returnRate), Detail: "Watches refund volume and return reasons from recent orders.", Weight: 0.15},
	}

	weighted := 0.0
	for _, metric := range metrics {
		weighted += float64(metric.Score) * metric.Weight
	}
	score := int(weighted + 0.5)
	actions, err := r.todayActions(ctx, organizationID, storeID)
	if err != nil {
		return domain.SellerHealth{}, err
	}

	return domain.SellerHealth{
		Score:       score,
		Grade:       healthGrade(score),
		Summary:     healthSummary(score, len(actions)),
		DataOrigin:  dataOrigin,
		GeneratedAt: time.Now().UTC(),
		Metrics:     metrics,
		Actions:     actions,
	}, nil
}

func (r *AnalyticsRepository) todayActions(ctx context.Context, organizationID, storeID uuid.UUID) ([]domain.TodayAction, error) {
	actions := make([]domain.TodayAction, 0, 8)

	stockRows, err := r.pool.Query(ctx, `
		with velocity as (
			select oi.product_id, coalesce(sum(oi.quantity_ordered),0)::numeric / 30 daily_units
			from order_items oi
			join orders o on o.id=oi.order_id
			where o.organization_id=$1 and o.store_id=$2 and o.purchase_date >= current_date-29
			group by oi.product_id
		)
		select p.title, coalesce(i.fulfillable_quantity,0)::int available, coalesce(v.daily_units,0) daily_units,
			case when coalesce(v.daily_units,0)>0 then coalesce(i.fulfillable_quantity,0)::numeric/v.daily_units else 999 end days_cover
		from products p
		left join inventory i on i.product_id=p.id
		left join velocity v on v.product_id=p.id
		where p.organization_id=$1 and p.store_id=$2 and (coalesce(i.fulfillable_quantity,0) < 20 or (coalesce(v.daily_units,0)>0 and coalesce(i.fulfillable_quantity,0)::numeric/v.daily_units < 21))
		order by days_cover asc, available asc
		limit 2
	`, organizationID, storeID)
	if err != nil {
		return nil, err
	}
	defer stockRows.Close()
	for stockRows.Next() {
		var title string
		var available int
		var dailyUnits, daysCover float64
		if err := stockRows.Scan(&title, &available, &dailyUnits, &daysCover); err != nil {
			return nil, err
		}
		priority := "Medium"
		if daysCover < 14 || available < 10 {
			priority = "High"
		}
		actions = append(actions, domain.TodayAction{
			ID:          fmt.Sprintf("stock-%d", len(actions)+1),
			Priority:    priority,
			Category:    "Inventory",
			Title:       "Restock before sales velocity drops",
			Description: fmt.Sprintf("%s has %d units available and about %.0f days of cover at current velocity.", title, available, daysCover),
			Impact:      "Avoid stockouts and lost ranking momentum.",
			Confidence:  0.84,
			Product:     title,
		})
	}
	if err := stockRows.Err(); err != nil {
		return nil, err
	}

	adRows, err := r.pool.Query(ctx, `
		select c.name,c.channel,coalesce(sum(m.spend),0) spend,coalesce(sum(m.sales),0) sales,
			case when coalesce(sum(m.spend),0)>0 then coalesce(sum(m.sales),0)/coalesce(sum(m.spend),0) else 0 end roas
		from campaigns c
		join campaign_metrics m on m.campaign_id=c.id
		where c.organization_id=$1 and c.store_id=$2 and m.metric_date >= current_date-29
		group by c.id
		having coalesce(sum(m.spend),0) > 50
		order by roas desc
	`, organizationID, storeID)
	if err != nil {
		return nil, err
	}
	defer adRows.Close()
	addedScale := false
	addedReduce := false
	for adRows.Next() {
		var name, channel string
		var spend, sales, roas float64
		if err := adRows.Scan(&name, &channel, &spend, &sales, &roas); err != nil {
			return nil, err
		}
		if roas >= 3.0 && !addedScale {
			actions = append(actions, domain.TodayAction{
				ID:          fmt.Sprintf("ads-scale-%d", len(actions)+1),
				Priority:    "High",
				Category:    "Ads",
				Title:       "Scale a profitable campaign",
				Description: fmt.Sprintf("%s is producing %.2fx ROAS from $%.0f spend.", name, roas, spend),
				Impact:      "Increase budget gradually while monitoring ACOS and stock cover.",
				Confidence:  0.78,
				Campaign:    name,
				Channel:     channel,
			})
			addedScale = true
		}
		if roas < 1.5 && !addedReduce {
			actions = append(actions, domain.TodayAction{
				ID:          fmt.Sprintf("ads-reduce-%d", len(actions)+1),
				Priority:    "High",
				Category:    "Ads",
				Title:       "Reduce spend on weak campaign",
				Description: fmt.Sprintf("%s is below target at %.2fx ROAS after $%.0f spend.", name, roas, spend),
				Impact:      "Cut wasted spend or move budget into stronger campaigns.",
				Confidence:  0.76,
				Campaign:    name,
				Channel:     channel,
			})
			addedReduce = true
		}
	}
	if err := adRows.Err(); err != nil {
		return nil, err
	}

	var returnProduct, returnReason string
	var returnCount int
	var refundAmount float64
	err = r.pool.QueryRow(ctx, `
		select coalesce(p.title,'Unknown product'), re.reason, count(*)::int, coalesce(sum(re.refund_amount),0)
		from return_events re
		left join products p on p.id=re.product_id
		where re.organization_id=$1 and re.store_id=$2 and re.return_date >= current_date-89
		group by p.title,re.reason
		order by count(*) desc, sum(re.refund_amount) desc
		limit 1
	`, organizationID, storeID).Scan(&returnProduct, &returnReason, &returnCount, &refundAmount)
	if err == nil && returnCount > 0 {
		actions = append(actions, domain.TodayAction{
			ID:          fmt.Sprintf("returns-%d", len(actions)+1),
			Priority:    "Medium",
			Category:    "Returns",
			Title:       "Fix top return driver",
			Description: fmt.Sprintf("%s has %d recent returns, mostly: %s.", returnProduct, returnCount, returnReason),
			Impact:      fmt.Sprintf("Reduce refunds currently worth about $%.0f in this cluster.", refundAmount),
			Confidence:  0.72,
			Product:     returnProduct,
		})
	}

	var profitProduct string
	var productRevenue, productMargin float64
	err = r.pool.QueryRow(ctx, `
		with sales as (
			select oi.product_id, coalesce(sum(oi.quantity_ordered),0)::int units, coalesce(sum(oi.item_price*oi.quantity_ordered),0) revenue
			from order_items oi
			join orders o on o.id=oi.order_id
			where o.organization_id=$1 and o.store_id=$2 and o.purchase_date >= current_date-89
			group by oi.product_id
		),
		ads as (
			select product_id, coalesce(sum(spend),0) spend
			from advertised_product_metrics
			where organization_id=$1 and store_id=$2 and metric_date >= current_date-89
			group by product_id
		),
		refunds as (
			select product_id, coalesce(sum(refund_amount),0) refunds
			from return_events
			where organization_id=$1 and store_id=$2 and return_date >= current_date-89
			group by product_id
		),
		profit as (
			select p.title, coalesce(s.revenue,0) revenue,
				case when coalesce(s.revenue,0)>0 then
					((coalesce(s.revenue,0)-coalesce(s.units*p.cost_price,0)-coalesce(a.spend,0)-coalesce(r.refunds,0)-(coalesce(s.revenue,0)*0.15))/coalesce(s.revenue,0)*100)
				else 0 end margin
			from products p
			join sales s on s.product_id=p.id
			left join ads a on a.product_id=p.id
			left join refunds r on r.product_id=p.id
			where p.organization_id=$1 and p.store_id=$2
		)
		select title,revenue,margin
		from profit
		where revenue > 0
		order by margin asc
		limit 1
	`, organizationID, storeID).Scan(&profitProduct, &productRevenue, &productMargin)
	if err == nil && productRevenue > 0 && productMargin < 18 {
		actions = append(actions, domain.TodayAction{
			ID:          fmt.Sprintf("profit-%d", len(actions)+1),
			Priority:    "High",
			Category:    "Profit",
			Title:       "Protect margin before scaling",
			Description: fmt.Sprintf("%s has only %.1f%% estimated margin on $%.0f recent revenue.", profitProduct, productMargin, productRevenue),
			Impact:      "Review price, bids, fees, and return leakage before increasing spend.",
			Confidence:  0.74,
			Product:     profitProduct,
		})
	}

	var city, region string
	var regionalRevenue, regionalROAS float64
	var regionalOrders int
	err = r.pool.QueryRow(ctx, `
		select city,region_code,coalesce(sum(revenue),0),coalesce(sum(orders),0)::int,
			case when coalesce(sum(ad_spend),0)>0 then coalesce(sum(revenue),0)/coalesce(sum(ad_spend),0) else 0 end roas
		from regional_sales_metrics
		where organization_id=$1 and store_id=$2 and metric_date >= current_date-179
		group by city,region_code
		order by 3 desc
		limit 1
	`, organizationID, storeID).Scan(&city, &region, &regionalRevenue, &regionalOrders, &regionalROAS)
	if err == nil && regionalRevenue > 0 {
		actions = append(actions, domain.TodayAction{
			ID:          fmt.Sprintf("region-%d", len(actions)+1),
			Priority:    "Medium",
			Category:    "Region",
			Title:       "Double down on strongest region",
			Description: fmt.Sprintf("%s, %s generated $%.0f from %d recent orders with %.2fx regional ROAS.", city, region, regionalRevenue, regionalOrders, regionalROAS),
			Impact:      "Create region-specific creatives, offers, or budget allocation.",
			Confidence:  0.70,
			Region:      region,
		})
	}

	if len(actions) == 0 {
		actions = append(actions, domain.TodayAction{
			ID:          "monitor-1",
			Priority:    "Low",
			Category:    "Monitoring",
			Title:       "Keep monitoring the account",
			Description: "No urgent issues were detected from the available data.",
			Impact:      "Use imports and connected ad channels to unlock deeper recommendations.",
			Confidence:  0.60,
		})
	}
	if len(actions) > 8 {
		actions = actions[:8]
	}
	return actions, nil
}

func scoreRevenueTrend(trendPercent, revenue float64) int {
	if revenue <= 0 {
		return 35
	}
	switch {
	case trendPercent >= 20:
		return 95
	case trendPercent >= 8:
		return 85
	case trendPercent >= 0:
		return 75
	case trendPercent >= -10:
		return 55
	default:
		return 35
	}
}

func scoreProfitMargin(margin float64) int {
	switch {
	case margin >= 35:
		return 95
	case margin >= 25:
		return 82
	case margin >= 15:
		return 65
	case margin >= 5:
		return 45
	default:
		return 30
	}
}

func scoreROAS(roas, adSpend float64) int {
	if adSpend <= 0 {
		return 60
	}
	switch {
	case roas >= 4:
		return 92
	case roas >= 3:
		return 82
	case roas >= 2:
		return 65
	case roas >= 1.25:
		return 45
	default:
		return 28
	}
}

func scoreInventory(productCount, lowStockProducts int, averageDaysCover float64) int {
	if productCount == 0 {
		return 50
	}
	lowShare := float64(lowStockProducts) / float64(productCount)
	switch {
	case lowShare == 0 && averageDaysCover >= 30:
		return 92
	case lowShare <= 0.15:
		return 80
	case lowShare <= 0.35:
		return 62
	case lowShare <= 0.60:
		return 44
	default:
		return 25
	}
}

func scoreReturns(returnRate, refunds float64) int {
	switch {
	case returnRate <= 2 && refunds < 100:
		return 92
	case returnRate <= 5:
		return 78
	case returnRate <= 9:
		return 60
	case returnRate <= 14:
		return 42
	default:
		return 25
	}
}

func scoreStatus(score int) string {
	switch {
	case score >= 80:
		return "Healthy"
	case score >= 60:
		return "Watch"
	default:
		return "Needs attention"
	}
}

func healthGrade(score int) string {
	switch {
	case score >= 90:
		return "A"
	case score >= 80:
		return "B"
	case score >= 70:
		return "C"
	case score >= 60:
		return "D"
	default:
		return "E"
	}
}

func healthSummary(score int, actionCount int) string {
	if score >= 80 {
		return fmt.Sprintf("The account is healthy. Focus on the %d highest-leverage actions to keep growth efficient.", actionCount)
	}
	if score >= 60 {
		return fmt.Sprintf("The account has growth potential but needs attention. Start with the %d recommended actions.", actionCount)
	}
	return fmt.Sprintf("The account needs corrective action. Prioritize the %d actions below before scaling spend.", actionCount)
}

func growthAction(insight domain.ProductGrowthInsight, days int) (string, string) {
	dailyVelocity := float64(insight.Units) / float64(days)
	reorderPoint := int(dailyVelocity * 21)
	if reorderPoint < 10 {
		reorderPoint = 10
	}
	if insight.Units == 0 && insight.Inventory > 0 {
		return "Hold", "No sales in the selected period; avoid new ad spend until demand appears."
	}
	if insight.Inventory <= reorderPoint && insight.Units > 0 {
		return "Restock", "Inventory is close to the next 21 days of sales velocity."
	}
	if insight.AdSpend > 0 && insight.ROAS > 0 && insight.ROAS < 1.5 {
		return "Reduce ads", "Ad ROAS is below 1.5x, so spend is likely dragging profit."
	}
	if insight.TrendPercent >= 10 && (insight.AdSpend == 0 || insight.ROAS >= 2.5) {
		return "Scale", "Sales trend is rising and ads are efficient or not yet being used."
	}
	if insight.TrendPercent >= 10 && insight.AdSpend == 0 && insight.Units > 0 {
		return "Test ads", "Organic demand is growing without ad support."
	}
	if insight.TrendPercent <= -25 && insight.Inventory > reorderPoint {
		return "Hold", "Demand is falling while inventory is still available."
	}
	return "Maintain", "Performance is stable; keep monitoring trend, stock, and ad efficiency."
}

func (r *AnalyticsRepository) GenerateDemo(ctx context.Context, organizationID, storeID uuid.UUID, months int) (domain.DemoGenerationResult, error) {
	if months < 1 {
		months = 6
	}
	if months > 24 {
		months = 24
	}
	environment, err := r.StoreEnvironment(ctx, organizationID, storeID)
	if err != nil {
		return domain.DemoGenerationResult{}, err
	}
	if environment != "sandbox" {
		return domain.DemoGenerationResult{}, fmt.Errorf("%w: demo data can only be generated for sandbox stores", domain.ErrValidation)
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.DemoGenerationResult{}, err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, `delete from orders where organization_id=$1 and store_id=$2 and amazon_order_id like 'DEMO-%'`, organizationID, storeID)
	if err != nil {
		return domain.DemoGenerationResult{}, err
	}
	_, err = tx.Exec(ctx, `delete from products where organization_id=$1 and store_id=$2 and data_origin='demo' and asin like 'B0DEMO%'`, organizationID, storeID)
	if err != nil {
		return domain.DemoGenerationResult{}, err
	}
	for _, statement := range []string{
		`delete from product_traffic_metrics where organization_id=$1 and store_id=$2 and data_origin='demo'`,
		`delete from search_term_metrics where organization_id=$1 and store_id=$2 and data_origin='demo'`,
		`delete from return_events where organization_id=$1 and store_id=$2 and data_origin='demo'`,
		`delete from regional_sales_metrics where organization_id=$1 and store_id=$2 and data_origin='demo'`,
		`delete from growth_recommendations where organization_id=$1 and store_id=$2 and data_origin='demo'`,
		`delete from advertised_product_metrics where organization_id=$1 and store_id=$2 and data_origin='demo'`,
		`delete from campaign_metrics using campaigns c where campaign_metrics.campaign_id=c.id and c.organization_id=$1 and c.store_id=$2 and c.data_origin='demo'`,
		`delete from campaigns where organization_id=$1 and store_id=$2 and data_origin='demo' and amazon_campaign_id like 'DEMO-%'`,
	} {
		if _, err = tx.Exec(ctx, statement, organizationID, storeID); err != nil {
			return domain.DemoGenerationResult{}, err
		}
	}
	products := []struct {
		asin         string
		sku          string
		title        string
		category     string
		costPrice    float64
		sellingPrice float64
	}{
		{"B0WMAP0001", "WM-DRESS-MIDI-BLK", "Avery Ribbed Knit Midi Dress - Black", "Dresses", 18.40, 54.00},
		{"B0WMAP0002", "WM-DRESS-WRAP-FLR", "Mila Floral Wrap Dress - Sage Floral", "Dresses", 16.75, 49.00},
		{"B0WMAP0003", "WM-BLAZER-LIN-OAT", "Lena Linen-Blend Relaxed Blazer - Oatmeal", "Outerwear", 27.50, 79.00},
		{"B0WMAP0004", "WM-TOP-SATIN-IVY", "Serena Satin Cowl Neck Camisole - Ivory", "Tops", 9.80, 34.00},
		{"B0WMAP0005", "WM-JEANS-HR-STR", "Nova High Rise Straight Leg Jeans - Vintage Blue", "Denim", 24.60, 68.00},
		{"B0WMAP0006", "WM-SET-KNIT-TAU", "Elise Knit Lounge Set - Taupe", "Matching Sets", 22.20, 64.00},
		{"B0WMAP0007", "WM-LEGGING-SCULPT", "Studio Sculpt High Waist Leggings - Charcoal", "Activewear", 12.90, 42.00},
		{"B0WMAP0008", "WM-BRA-SEAMLESS-ROSE", "CloudSoft Seamless Bralette - Dusty Rose", "Intimates", 6.40, 24.00},
		{"B0WMAP0009", "WM-SKIRT-SLIP-CHAMP", "Gia Satin Bias Slip Skirt - Champagne", "Skirts", 13.30, 46.00},
		{"B0WMAP0010", "WM-SWEATER-CARD-CRE", "Harper Oversized Cable Cardigan - Cream", "Sweaters", 19.10, 58.00},
		{"B0WMAP0011", "WM-BLOUSE-PLEAT-NAV", "Clara Pleated Sleeve Blouse - Navy", "Tops", 11.90, 39.00},
		{"B0WMAP0012", "WM-JUMPSUIT-WIDE-OLV", "Sloane Wide Leg Utility Jumpsuit - Olive", "Jumpsuits", 23.80, 72.00},
	}
	for _, product := range products {
		_, err = tx.Exec(ctx, `insert into products(organization_id,store_id,asin,sku,title,status,data_origin,category,cost_price,selling_price) values($1,$2,$3,$4,$5,'active','demo',$6,$7,$8) on conflict(organization_id,store_id,asin,coalesce(sku,'')) do update set title=excluded.title,data_origin='demo',category=excluded.category,cost_price=excluded.cost_price,selling_price=excluded.selling_price,updated_at=now()`, organizationID, storeID, product.asin, product.sku, product.title, product.category, product.costPrice, product.sellingPrice)
		if err != nil {
			return domain.DemoGenerationResult{}, err
		}
	}
	_, err = tx.Exec(ctx, `insert into inventory(organization_id,product_id,fulfillable_quantity,inbound_quantity,reserved_quantity,data_origin) select $1,id,35+(row_number() over(order by asin)*11)::int%180,5+(row_number() over(order by asin)*3)::int%30,(row_number() over(order by asin)*2)::int%15,'demo' from products where organization_id=$1 and store_id=$2 and asin like 'B0WMAP%' on conflict(organization_id,product_id) do update set fulfillable_quantity=excluded.fulfillable_quantity,inbound_quantity=excluded.inbound_quantity,reserved_quantity=excluded.reserved_quantity,data_origin='demo',updated_at=now()`, organizationID, storeID)
	if err != nil {
		return domain.DemoGenerationResult{}, err
	}
	campaigns := []struct {
		id           string
		channel      string
		name         string
		campaignType string
		budget       int
	}{
		{"DEMO-AMZ-CAMP-1", "amazon_ads", "Amazon SP - Spring Dresses - Exact + Phrase", "sponsored_products", 85},
		{"DEMO-AMZ-CAMP-2", "amazon_ads", "Amazon SP - Workwear Blazers & Blouses", "sponsored_products", 65},
		{"DEMO-GOOGLE-CAMP-1", "google_ads", "Google Search - Midi Dresses & Blazers", "search", 95},
		{"DEMO-GOOGLE-CAMP-2", "google_ads", "Google Performance Max - Women's Apparel", "performance_max", 120},
		{"DEMO-META-CAMP-1", "meta_ads", "Meta Prospecting - Spring Outfit Video", "paid_social", 75},
		{"DEMO-META-CAMP-2", "meta_ads", "Meta Retargeting - Cart & Product Viewers", "retargeting", 60},
	}
	for _, campaign := range campaigns {
		_, err = tx.Exec(ctx, `insert into campaigns(organization_id,store_id,amazon_campaign_id,channel,name,campaign_type,status,budget,data_origin) values($1,$2,$3,$4,$5,$6,'enabled',$7,'demo') on conflict(organization_id,store_id,amazon_campaign_id) do update set channel=excluded.channel,name=excluded.name,campaign_type=excluded.campaign_type,budget=excluded.budget,data_origin='demo',updated_at=now()`, organizationID, storeID, campaign.id, campaign.channel, campaign.name, campaign.campaignType, campaign.budget)
		if err != nil {
			return domain.DemoGenerationResult{}, err
		}
	}
	days := months * 30
	_, err = tx.Exec(ctx, `insert into daily_metrics(organization_id,store_id,metric_date,revenue,profit,units_sold,ad_spend,refunds,data_origin) select $1,$2,d::date,round((520+extract(doy from d)::int%17*31+extract(dow from d)::int*44)::numeric,2),round((180+extract(doy from d)::int%13*17)::numeric,2),18+extract(doy from d)::int%24,round((65+extract(doy from d)::int%9*7)::numeric,2),case when extract(doy from d)::int%11=0 then 42 else 0 end,'demo' from generate_series(current_date-($3-1),current_date,'1 day') d on conflict(organization_id,store_id,metric_date) do update set revenue=excluded.revenue,profit=excluded.profit,units_sold=excluded.units_sold,ad_spend=excluded.ad_spend,refunds=excluded.refunds,data_origin='demo'`, organizationID, storeID, days)
	if err != nil {
		return domain.DemoGenerationResult{}, err
	}
	_, err = tx.Exec(ctx, `insert into campaign_metrics(organization_id,campaign_id,metric_date,impressions,clicks,spend,sales,orders,data_origin) select $1,c.id,d::date,900+(extract(doy from d)::int%19)*70,25+(extract(doy from d)::int%13),round((14+(extract(doy from d)::int%8)*2.5)::numeric,2),round((55+(extract(doy from d)::int%11)*12)::numeric,2),2+(extract(doy from d)::int%6),'demo' from campaigns c cross join generate_series(current_date-($3-1),current_date,'1 day') d where c.organization_id=$1 and c.store_id=$2 and c.amazon_campaign_id like 'DEMO-%' on conflict(campaign_id,metric_date) do update set impressions=excluded.impressions,clicks=excluded.clicks,spend=excluded.spend,sales=excluded.sales,orders=excluded.orders,data_origin='demo'`, organizationID, storeID, days)
	if err != nil {
		return domain.DemoGenerationResult{}, err
	}
	_, err = tx.Exec(ctx, `delete from orders where organization_id=$1 and store_id=$2 and amazon_order_id like 'DEMO-%'`, organizationID, storeID)
	if err != nil {
		return domain.DemoGenerationResult{}, err
	}
	_, err = tx.Exec(ctx, `
		insert into orders(organization_id,store_id,amazon_order_id,marketplace_id,order_status,purchase_date,order_total,currency_code,data_origin)
		select $1,$2,'DEMO-'||to_char(d,'YYYYMMDD')||'-'||n,
			(select id from marketplaces where amazon_marketplace_id='ATVPDKIKX0DER'),
			case when extract(doy from d)::int%37=0 then 'Refunded' else 'Shipped' end,
			d+(n||' hours')::interval,
			round((p.selling_price * (1+(extract(doy from d)::int%2)) * (case when extract(dow from d)::int in (0,6) then 1.08 else 1 end))::numeric,2),
			'USD','demo'
		from generate_series(current_date-($3-1),current_date,'1 day') d
		cross join generate_series(1,3) n
		join lateral(
			select * from products p
			where p.organization_id=$1 and p.store_id=$2 and p.asin like 'B0WMAP%'
			order by p.asin
			offset ((extract(doy from d)::int+n)::int%12)
			limit 1
		)p on true`, organizationID, storeID, days)
	if err != nil {
		return domain.DemoGenerationResult{}, err
	}
	_, err = tx.Exec(ctx, `insert into order_items(order_id,product_id,amazon_order_item_id,asin,sku,title,quantity_ordered,item_price,currency_code) select o.id,p.id,o.amazon_order_id||'-ITEM',p.asin,p.sku,p.title,1+(extract(doy from o.purchase_date)::int%2),round((o.order_total/(1+(extract(doy from o.purchase_date)::int%2)))::numeric,2),'USD' from orders o join lateral(select * from products p where p.organization_id=$1 and p.store_id=$2 and p.asin like 'B0WMAP%' order by p.asin offset ((extract(doy from o.purchase_date)::int+substring(o.amazon_order_id from '-([0-9]+)$')::int)::int%12) limit 1)p on true where o.organization_id=$1 and o.store_id=$2 and o.amazon_order_id like 'DEMO-%' on conflict(order_id,amazon_order_item_id) do update set product_id=excluded.product_id,asin=excluded.asin,sku=excluded.sku,title=excluded.title,quantity_ordered=excluded.quantity_ordered,item_price=excluded.item_price`, organizationID, storeID)
	if err != nil {
		return domain.DemoGenerationResult{}, err
	}
	_, err = tx.Exec(ctx, `
		insert into return_events(organization_id,store_id,order_id,product_id,return_date,quantity,reason,status,refund_amount,currency_code,data_origin)
		select $1,$2,o.id,oi.product_id,(o.purchase_date::date + interval '6 days')::date,1,
			case
				when p.category='Dresses' then 'Fit was smaller than expected'
				when p.category='Denim' then 'Size did not match chart'
				when p.category='Intimates' then 'Changed mind after delivery'
				when p.category='Tops' then 'Color looked different from images'
				when p.category='Activewear' then 'Fabric compression not preferred'
				else 'Item did not match expectation'
			end,
			'completed',
			round((oi.item_price*least(oi.quantity_ordered,1))::numeric,2),
			oi.currency_code,
			'demo'
		from orders o
		join order_items oi on oi.order_id=o.id
		join products p on p.id=oi.product_id
		where o.organization_id=$1 and o.store_id=$2 and o.amazon_order_id like 'DEMO-%'
			and (o.order_status='Refunded' or extract(doy from o.purchase_date)::int%29=0)
		on conflict(store_id,order_id,product_id,return_date,reason) do update set
			quantity=excluded.quantity,refund_amount=excluded.refund_amount,status=excluded.status,
			data_origin='demo',updated_at=now()
	`, organizationID, storeID)
	if err != nil {
		return domain.DemoGenerationResult{}, err
	}
	_, err = tx.Exec(ctx, `
		with sales as (
			select o.purchase_date::date metric_date,
				coalesce(sum(oi.item_price*oi.quantity_ordered),0) revenue,
				coalesce(sum(oi.quantity_ordered),0)::int units,
				coalesce(sum(case when o.order_status='Refunded' then oi.item_price*oi.quantity_ordered else 0 end),0) refunds
			from orders o
			join order_items oi on oi.order_id=o.id
			where o.organization_id=$1 and o.store_id=$2 and o.amazon_order_id like 'DEMO-%'
			group by o.purchase_date::date
		)
		update daily_metrics dm set
			revenue=round(s.revenue::numeric,2),
			units_sold=s.units,
			refunds=round(s.refunds::numeric,2),
			profit=round((s.revenue*0.46-dm.ad_spend-s.refunds)::numeric,2),
			data_origin='demo'
		from sales s
		where dm.organization_id=$1 and dm.store_id=$2 and dm.metric_date=s.metric_date
	`, organizationID, storeID)
	if err != nil {
		return domain.DemoGenerationResult{}, err
	}
	_, err = tx.Exec(ctx, `
		with product_day as (
			select p.id product_id, d::date metric_date,
				coalesce(s.units,0)::int units,
				coalesce(s.revenue,0) revenue
			from products p
			cross join generate_series(current_date-($3-1),current_date,'1 day') d
			left join (
				select oi.product_id,o.purchase_date::date metric_date,
					coalesce(sum(oi.quantity_ordered),0)::int units,
					coalesce(sum(oi.item_price*oi.quantity_ordered),0) revenue
				from order_items oi
				join orders o on o.id=oi.order_id
				where o.organization_id=$1 and o.store_id=$2 and o.amazon_order_id like 'DEMO-%'
				group by oi.product_id,o.purchase_date::date
			) s on s.product_id=p.id and s.metric_date=d::date
			where p.organization_id=$1 and p.store_id=$2 and p.asin like 'B0WMAP%'
		)
		insert into product_traffic_metrics(organization_id,store_id,product_id,metric_date,sessions,page_views,buy_box_percentage,units_ordered,ordered_revenue,data_origin)
		select $1,$2,pd.product_id,pd.metric_date,
			(pd.units*18 + 35 + extract(doy from pd.metric_date)::int%27 + row_number() over(partition by pd.metric_date order by pd.product_id)::int%12)::int,
			(pd.units*24 + 48 + extract(doy from pd.metric_date)::int%31 + row_number() over(partition by pd.metric_date order by pd.product_id)::int%18)::int,
			round((88 + extract(doy from pd.metric_date)::int%10)::numeric,2),
			pd.units,
			round(pd.revenue::numeric,2),
			'demo'
		from product_day pd
		on conflict(store_id,product_id,metric_date) do update set
			sessions=excluded.sessions,page_views=excluded.page_views,buy_box_percentage=excluded.buy_box_percentage,
			units_ordered=excluded.units_ordered,ordered_revenue=excluded.ordered_revenue,data_origin='demo',updated_at=now()
	`, organizationID, storeID, days)
	if err != nil {
		return domain.DemoGenerationResult{}, err
	}
	_, err = tx.Exec(ctx, `
		with order_regions as (
			select o.id,o.purchase_date::date metric_date,
				case ((extract(doy from o.purchase_date)::int + substring(o.amazon_order_id from '-([0-9]+)$')::int) % 8)
					when 0 then 'CA' when 1 then 'NY' when 2 then 'TX' when 3 then 'FL'
					when 4 then 'IL' when 5 then 'WA' when 6 then 'GA' else 'AZ'
				end region_code,
				case ((extract(doy from o.purchase_date)::int + substring(o.amazon_order_id from '-([0-9]+)$')::int) % 8)
					when 0 then 'Los Angeles' when 1 then 'New York' when 2 then 'Austin' when 3 then 'Miami'
					when 4 then 'Chicago' when 5 then 'Seattle' when 6 then 'Atlanta' else 'Phoenix'
				end city,
				o.marketplace_id
			from orders o
			where o.organization_id=$1 and o.store_id=$2 and o.amazon_order_id like 'DEMO-%'
		),
		agg as (
			select r.metric_date,r.marketplace_id,'US' country_code,r.region_code,r.city,
				count(distinct r.id)::int orders,
				coalesce(sum(oi.quantity_ordered),0)::int units,
				coalesce(sum(oi.item_price*oi.quantity_ordered),0) revenue,
				coalesce(sum(case when o.order_status='Refunded' then oi.item_price*oi.quantity_ordered else 0 end),0) refunds
			from order_regions r
			join orders o on o.id=r.id
			join order_items oi on oi.order_id=o.id
			group by r.metric_date,r.marketplace_id,r.region_code,r.city
		)
		insert into regional_sales_metrics(organization_id,store_id,marketplace_id,country_code,region_code,city,metric_date,orders,units,revenue,refunds,ad_spend,data_origin)
		select $1,$2,marketplace_id,country_code,region_code,city,metric_date,orders,units,round(revenue::numeric,2),round(refunds::numeric,2),round((revenue*0.075)::numeric,2),'demo'
		from agg
		on conflict(store_id,country_code,region_code,city,metric_date) do update set
			orders=excluded.orders,units=excluded.units,revenue=excluded.revenue,refunds=excluded.refunds,
			ad_spend=excluded.ad_spend,data_origin='demo',updated_at=now()
	`, organizationID, storeID)
	if err != nil {
		return domain.DemoGenerationResult{}, err
	}
	_, err = tx.Exec(ctx, `delete from financial_transactions where organization_id=$1 and store_id=$2 and raw_payload->>'source'='demo'`, organizationID, storeID)
	if err != nil {
		return domain.DemoGenerationResult{}, err
	}
	_, err = tx.Exec(ctx, `
		insert into financial_transactions(organization_id,store_id,transaction_type,amount,currency_code,posted_at,raw_payload,data_origin)
		select $1,$2,x.kind,x.amount,'USD',dm.metric_date::timestamptz,jsonb_build_object('source','demo','business','women_apparel'),'demo'
		from daily_metrics dm
		cross join lateral(values
			('sale',dm.revenue),
			('amazon_fee',round((-dm.revenue*0.15)::numeric,2)),
			('refund',-dm.refunds)
		)x(kind,amount)
		where dm.organization_id=$1 and dm.store_id=$2 and dm.metric_date >= current_date-($3-1) and x.amount <> 0
	`, organizationID, storeID, days)
	if err != nil {
		return domain.DemoGenerationResult{}, err
	}
	_, err = tx.Exec(ctx, `insert into reports(organization_id,store_id,created_by,report_type,format,status,storage_key,data_origin) select $1,$2,null,x,'json','completed','demo/'||lower(replace(x,' ','-'))||'.json','demo' from unnest(array['Listings Report','FBA Inventory Report','Orders Report','Finances Report'])x where not exists(select 1 from reports r where r.organization_id=$1 and r.store_id=$2 and r.report_type=x and r.storage_key like 'demo/%')`, organizationID, storeID)
	if err != nil {
		return domain.DemoGenerationResult{}, err
	}
	_, err = tx.Exec(ctx, `
		insert into advertised_product_metrics(organization_id,store_id,product_id,campaign_id,metric_date,impressions,clicks,spend,sales,orders,data_origin)
		select $1,$2,p.id,c.id,d::date,
			350+(extract(doy from d)::int%17)*40,
			8+(extract(doy from d)::int%9),
			round((0.75+(extract(doy from d)::int%6)*0.22+(row_number() over(order by p.asin)%4)*0.15)::numeric,2),
			round((12+(extract(doy from d)::int%11)*2.7+(row_number() over(order by p.asin)%5)*2.5)::numeric,2),
			1+(extract(doy from d)::int%4),
			'demo'
		from products p
		join lateral (
			select id from campaigns c
			where c.organization_id=$1 and c.store_id=$2 and c.amazon_campaign_id like 'DEMO-%'
			order by c.amazon_campaign_id
			offset ((right(p.asin, 2)::int - 1) % 6)
			limit 1
		) c on true
		cross join generate_series(current_date-($3-1),current_date,'1 day') d
		where p.organization_id=$1 and p.store_id=$2 and p.asin like 'B0WMAP%'
		on conflict(store_id,product_id,campaign_id,metric_date) do update set
			impressions=excluded.impressions,clicks=excluded.clicks,spend=excluded.spend,
			sales=excluded.sales,orders=excluded.orders,data_origin='demo',updated_at=now()
	`, organizationID, storeID, days)
	if err != nil {
		return domain.DemoGenerationResult{}, err
	}
	_, err = tx.Exec(ctx, `
		insert into search_term_metrics(organization_id,store_id,product_id,campaign_id,search_term,keyword_text,match_type,metric_date,impressions,clicks,spend,sales,orders,data_origin)
		select $1,$2,p.id,c.id,
			case
				when p.category='Dresses' then 'women midi dress'
				when p.category='Outerwear' then 'linen blazer women'
				when p.category='Denim' then 'high rise straight jeans women'
				when p.category='Matching Sets' then 'women lounge set'
				when p.category='Activewear' then 'high waist workout leggings'
				when p.category='Intimates' then 'seamless bralette women'
				when p.category='Skirts' then 'satin slip skirt'
				when p.category='Sweaters' then 'oversized cable cardigan women'
				when p.category='Jumpsuits' then 'wide leg utility jumpsuit women'
				else 'women dressy blouse'
			end,
			case
				when p.category='Dresses' then 'midi dress'
				when p.category='Outerwear' then 'women blazer'
				when p.category='Denim' then 'straight leg jeans'
				when p.category='Matching Sets' then 'lounge set women'
				when p.category='Activewear' then 'leggings women'
				when p.category='Intimates' then 'bralette'
				when p.category='Skirts' then 'slip skirt'
				when p.category='Sweaters' then 'cardigan women'
				when p.category='Jumpsuits' then 'jumpsuit women'
				else 'women blouse'
			end,
			case when extract(doy from d)::int%3=0 then 'exact' when extract(doy from d)::int%3=1 then 'phrase' else 'broad' end,
			d::date,
			620+(extract(doy from d)::int%23)*38,
			18+(extract(doy from d)::int%11),
			round((6.50+(extract(doy from d)::int%7)*1.15+(row_number() over(order by p.asin)%5)*0.35)::numeric,2),
			round((p.selling_price*(1+(extract(doy from d)::int%3))*0.82)::numeric,2),
			1+(extract(doy from d)::int%4),
			'demo'
		from products p
		join lateral (
			select id from campaigns c
			where c.organization_id=$1 and c.store_id=$2 and c.amazon_campaign_id like 'DEMO-%'
			order by c.amazon_campaign_id
			offset ((right(p.asin, 2)::int - 1) % 6)
			limit 1
		)c on true
		cross join generate_series(current_date-($3-1),current_date,'1 day') d
		where p.organization_id=$1 and p.store_id=$2 and p.asin like 'B0WMAP%'
		on conflict(store_id,product_id,campaign_id,search_term,metric_date) do update set
			keyword_text=excluded.keyword_text,match_type=excluded.match_type,impressions=excluded.impressions,
			clicks=excluded.clicks,spend=excluded.spend,sales=excluded.sales,orders=excluded.orders,
			data_origin='demo',updated_at=now()
	`, organizationID, storeID, days)
	if err != nil {
		return domain.DemoGenerationResult{}, err
	}
	_, err = tx.Exec(ctx, `
		insert into growth_recommendations(organization_id,store_id,product_id,campaign_id,region_code,recommendation_type,title,reason,evidence,impact_score,confidence,data_origin)
		select $1,$2,p.id,null,'',
			case
				when i.fulfillable_quantity < 60 then 'restock'
				when coalesce(a.sales,0) > coalesce(a.spend,0)*5 then 'scale_ads'
				else 'monitor'
			end,
			case
				when i.fulfillable_quantity < 60 then 'Restock '||p.title
				when coalesce(a.sales,0) > coalesce(a.spend,0)*5 then 'Scale ads for '||p.title
				else 'Monitor '||p.title
			end,
			case
				when i.fulfillable_quantity < 60 then 'Inventory is below the apparel safety threshold while recent demand is active.'
				when coalesce(a.sales,0) > coalesce(a.spend,0)*5 then 'Search and product ads show strong ROAS for this SKU.'
				else 'Performance is stable; keep watching sales velocity and returns.'
			end,
			jsonb_build_object('inventory',i.fulfillable_quantity,'adSales',coalesce(a.sales,0),'adSpend',coalesce(a.spend,0)),
			case when i.fulfillable_quantity < 60 then 88 when coalesce(a.sales,0) > coalesce(a.spend,0)*5 then 82 else 55 end,
			case when i.fulfillable_quantity < 60 then 0.84 when coalesce(a.sales,0) > coalesce(a.spend,0)*5 then 0.78 else 0.62 end,
			'demo'
		from products p
		left join inventory i on i.product_id=p.id
		left join (
			select product_id,sum(spend) spend,sum(sales) sales
			from advertised_product_metrics
			where organization_id=$1 and store_id=$2 and metric_date >= current_date-29
			group by product_id
		)a on a.product_id=p.id
		where p.organization_id=$1 and p.store_id=$2 and p.asin like 'B0WMAP%'
		order by 11 desc
		limit 12
	`, organizationID, storeID)
	if err != nil {
		return domain.DemoGenerationResult{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return domain.DemoGenerationResult{}, err
	}
	return domain.DemoGenerationResult{Months: months, Products: 12, Orders: days * 3, Campaigns: len(campaigns), Days: days}, nil
}
