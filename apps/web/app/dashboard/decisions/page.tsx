"use client";
import { ResourcePage, formatMoney } from "@/components/resource-page";

const pct = (value: unknown) => `${Number(value ?? 0).toFixed(2)}%`;
const days = (value: unknown) => Number(value ?? 0) >= 999 ? "No recent velocity" : `${Number(value ?? 0).toFixed(1)} days`;

export default function ProductDecisionsPage() {
  return <ResourcePage
    resource="product-decisions"
    title="Product decisions"
    description="Scale, hold, restock, reduce ads, and fix-listing decisions for every product"
    columns={[
      { key: "decision", label: "Decision" },
      { key: "product", label: "Product" },
      { key: "category", label: "Category" },
      { key: "units", label: "Units" },
      { key: "revenue", label: "Revenue", format: formatMoney },
      { key: "margin", label: "Margin", format: pct },
      { key: "roas", label: "ROAS" },
      { key: "return_rate", label: "Return rate", format: pct },
      { key: "conversion_rate", label: "CVR", format: pct },
      { key: "inventory", label: "Inventory" },
      { key: "days_cover", label: "Days cover", format: days },
      { key: "reason", label: "Reason" },
    ]}
  />;
}
