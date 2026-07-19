package domain

import "time"

type AnalyticsSummary struct {
	Revenue      float64 `json:"revenue"`
	Profit       float64 `json:"profit"`
	Orders       int     `json:"orders"`
	Units        int     `json:"units"`
	AdSpend      float64 `json:"adSpend"`
	Refunds      float64 `json:"refunds"`
	Products     int     `json:"products"`
	Inventory    int     `json:"inventory"`
	ROAS         float64 `json:"roas"`
	ProfitMargin float64 `json:"profitMargin"`
}

type AnalyticsTrendPoint struct {
	Date    time.Time `json:"date"`
	Revenue float64   `json:"revenue"`
	Profit  float64   `json:"profit"`
	AdSpend float64   `json:"adSpend"`
	Refunds float64   `json:"refunds"`
	Units   int       `json:"units"`
}

type ProductPerformance struct {
	ASIN      string  `json:"asin"`
	SKU       string  `json:"sku"`
	Title     string  `json:"title"`
	Revenue   float64 `json:"revenue"`
	Units     int     `json:"units"`
	Available int     `json:"available"`
}

type CampaignPerformance struct {
	Channel     string  `json:"channel"`
	Name        string  `json:"name"`
	Status      string  `json:"status"`
	Spend       float64 `json:"spend"`
	Sales       float64 `json:"sales"`
	Orders      int     `json:"orders"`
	Impressions int     `json:"impressions"`
	Clicks      int     `json:"clicks"`
	ROAS        float64 `json:"roas"`
}

type AnalyticsDashboard struct {
	Summary   AnalyticsSummary      `json:"summary"`
	Trend     []AnalyticsTrendPoint `json:"trend"`
	Products  []ProductPerformance  `json:"products"`
	Campaigns []CampaignPerformance `json:"campaigns"`
}

type DemoGenerationResult struct {
	Months    int `json:"months"`
	Products  int `json:"products"`
	Orders    int `json:"orders"`
	Campaigns int `json:"campaigns"`
	Days      int `json:"days"`
}

type ProductGrowthInsight struct {
	ProductID       string  `json:"productId"`
	ASIN            string  `json:"asin"`
	SKU             string  `json:"sku"`
	Title           string  `json:"title"`
	Category        string  `json:"category"`
	Revenue         float64 `json:"revenue"`
	Units           int     `json:"units"`
	PreviousUnits   int     `json:"previousUnits"`
	TrendPercent    float64 `json:"trendPercent"`
	AdSpend         float64 `json:"adSpend"`
	AdSales         float64 `json:"adSales"`
	ROAS            float64 `json:"roas"`
	ACOS            float64 `json:"acos"`
	Inventory       int     `json:"inventory"`
	EstimatedProfit float64 `json:"estimatedProfit"`
	Action          string  `json:"action"`
	Reason          string  `json:"reason"`
}

type MarketplaceGrowthInsight struct {
	CountryCode string  `json:"countryCode"`
	Name        string  `json:"name"`
	Orders      int     `json:"orders"`
	Units       int     `json:"units"`
	Revenue     float64 `json:"revenue"`
}

type GrowthIntelligence struct {
	Products     []ProductGrowthInsight     `json:"products"`
	Marketplaces []MarketplaceGrowthInsight `json:"marketplaces"`
}

type SellerHealthMetric struct {
	Key    string  `json:"key"`
	Label  string  `json:"label"`
	Score  int     `json:"score"`
	Status string  `json:"status"`
	Value  string  `json:"value"`
	Detail string  `json:"detail"`
	Weight float64 `json:"weight"`
}

type TodayAction struct {
	ID          string  `json:"id"`
	Priority    string  `json:"priority"`
	Category    string  `json:"category"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Impact      string  `json:"impact"`
	Confidence  float64 `json:"confidence"`
	Product     string  `json:"product,omitempty"`
	Campaign    string  `json:"campaign,omitempty"`
	Channel     string  `json:"channel,omitempty"`
	Region      string  `json:"region,omitempty"`
}

type SellerHealth struct {
	Score       int                  `json:"score"`
	Grade       string               `json:"grade"`
	Summary     string               `json:"summary"`
	DataOrigin  string               `json:"dataOrigin"`
	GeneratedAt time.Time            `json:"generatedAt"`
	Metrics     []SellerHealthMetric `json:"metrics"`
	Actions     []TodayAction        `json:"actions"`
}
