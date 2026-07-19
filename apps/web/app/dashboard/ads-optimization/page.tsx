"use client";
import { ResourcePage, formatMoney } from "@/components/resource-page";

const pct = (value: unknown) => `${Number(value ?? 0).toFixed(2)}%`;

export default function AdsOptimizationPage() {
  return <ResourcePage
    resource="ads-optimization"
    title="Ads optimization"
    description="Campaign-level budget guidance across Amazon Ads, Google Ads, and Meta Ads"
    columns={[
      { key: "decision", label: "Decision" },
      { key: "budget_action", label: "Budget action" },
      { key: "campaign", label: "Campaign" },
      { key: "channel", label: "Channel" },
      { key: "campaign_type", label: "Type" },
      { key: "spend", label: "Spend", format: formatMoney },
      { key: "sales", label: "Sales", format: formatMoney },
      { key: "roas", label: "ROAS" },
      { key: "acos", label: "ACOS", format: pct },
      { key: "conversion_rate", label: "CVR", format: pct },
      { key: "clicks", label: "Clicks" },
      { key: "orders", label: "Orders" },
      { key: "reason", label: "Reason" },
    ]}
  />;
}
