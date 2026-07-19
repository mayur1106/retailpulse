"use client";

import { useQuery } from "@tanstack/react-query";
import { ArrowDownCircle, ArrowUpCircle, LineChart, PackageCheck, PauseCircle } from "lucide-react";
import type { ElementType } from "react";
import { useEffect, useMemo, useState } from "react";
import { DashboardShell, PageCard } from "@/components/dashboard-shell";
import { authedRequest } from "@/lib/api";

type AmazonStore = { id: string; name: string; environment: "production" | "sandbox" };
type ProductInsight = {
  productId: string;
  asin: string;
  sku: string;
  title: string;
  category: string;
  revenue: number;
  units: number;
  previousUnits: number;
  trendPercent: number;
  adSpend: number;
  adSales: number;
  roas: number;
  acos: number;
  inventory: number;
  estimatedProfit: number;
  action: string;
  reason: string;
};
type MarketplaceInsight = { countryCode: string; name: string; orders: number; units: number; revenue: number };
type GrowthResponse = { products: ProductInsight[]; marketplaces: MarketplaceInsight[] };

const money = new Intl.NumberFormat("en-US", { style: "currency", currency: "USD", maximumFractionDigits: 0 });
const number = new Intl.NumberFormat("en-US");

const actionStyles: Record<string, string> = {
  Scale: "bg-[#ecfdf3] text-[#027a48]",
  Restock: "bg-[#eff8ff] text-[#175cd3]",
  "Test ads": "bg-[#f4f3ff] text-[#5925dc]",
  "Reduce ads": "bg-[#fff1f3] text-[#c01048]",
  Hold: "bg-[#fffaeb] text-[#b54708]",
  Maintain: "bg-[#f2f4f7] text-[#344054]",
};

const actionIcons: Record<string, ElementType> = {
  Scale: ArrowUpCircle,
  Restock: PackageCheck,
  "Test ads": LineChart,
  "Reduce ads": ArrowDownCircle,
  Hold: PauseCircle,
  Maintain: LineChart,
};

export default function GrowthPage() {
  const [selectedStoreId, setSelectedStoreId] = useState("");
  const storesQuery = useQuery({ queryKey: ["amazon-stores"], queryFn: () => authedRequest<{ stores: AmazonStore[] | null }>("/v1/amazon/stores"), retry: false });
  const stores = storesQuery.data?.stores ?? [];

  useEffect(() => {
    if (stores.length === 0) return;
    if (selectedStoreId && stores.some((store) => store.id === selectedStoreId)) return;
    setSelectedStoreId(stores.find((store) => store.environment === "production")?.id ?? stores[0].id);
  }, [stores, selectedStoreId]);

  const selectedStore = stores.find((store) => store.id === selectedStoreId);
  const growthQuery = useQuery({
    queryKey: ["growth-intelligence", selectedStoreId],
    queryFn: () => authedRequest<GrowthResponse>(`/v1/analytics/growth?days=90&storeId=${selectedStoreId}`),
    enabled: Boolean(selectedStoreId),
    retry: false,
  });

  const products = growthQuery.data?.products ?? [];
  const marketplaces = growthQuery.data?.marketplaces ?? [];
  const totals = useMemo(() => ({
    revenue: products.reduce((sum, item) => sum + item.revenue, 0),
    profit: products.reduce((sum, item) => sum + item.estimatedProfit, 0),
    adSpend: products.reduce((sum, item) => sum + item.adSpend, 0),
    scale: products.filter((item) => item.action === "Scale" || item.action === "Test ads").length,
  }), [products]);

  return <DashboardShell title="Growth insights" description="Product, marketplace, and ad actions for the selected Amazon account">
    <div className="mx-auto max-w-7xl space-y-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 className="text-lg font-semibold">90-day growth model</h2>
          {selectedStore ? <p className="mt-1 text-sm text-[#667085]">{selectedStore.name} · {selectedStore.environment}</p> : null}
        </div>
        <select value={selectedStoreId} onChange={(event) => setSelectedStoreId(event.target.value)} className="rounded-xl border border-[#d0d5dd] bg-white px-3 py-2.5 text-sm text-[#344054] outline-none transition focus:border-[#14b8a6] focus:ring-4 focus:ring-[#ccfbf1]">
          {stores.map((store) => <option key={store.id} value={store.id}>{store.name} ({store.environment})</option>)}
        </select>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <PageCard className="p-5"><p className="text-xs font-semibold uppercase tracking-wide text-[#667085]">Revenue analyzed</p><p className="mt-2 text-2xl font-semibold">{money.format(totals.revenue)}</p></PageCard>
        <PageCard className="p-5"><p className="text-xs font-semibold uppercase tracking-wide text-[#667085]">Estimated profit</p><p className="mt-2 text-2xl font-semibold">{money.format(totals.profit)}</p></PageCard>
        <PageCard className="p-5"><p className="text-xs font-semibold uppercase tracking-wide text-[#667085]">Ad spend mapped</p><p className="mt-2 text-2xl font-semibold">{money.format(totals.adSpend)}</p></PageCard>
        <PageCard className="p-5"><p className="text-xs font-semibold uppercase tracking-wide text-[#667085]">Growth candidates</p><p className="mt-2 text-2xl font-semibold">{totals.scale}</p></PageCard>
      </div>

      <PageCard className="overflow-hidden">
        <div className="border-b border-[#eaecf0] p-5">
          <h2 className="font-semibold">Product actions</h2>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full min-w-[980px] text-sm">
            <thead className="bg-[#f8fafc] text-left text-xs uppercase tracking-wide text-[#667085]">
              <tr><th className="px-5 py-3.5">Product</th><th className="px-5 py-3.5">Action</th><th className="px-5 py-3.5">Revenue</th><th className="px-5 py-3.5">Units</th><th className="px-5 py-3.5">Trend</th><th className="px-5 py-3.5">Inventory</th><th className="px-5 py-3.5">ROAS</th><th className="px-5 py-3.5">Reason</th></tr>
            </thead>
            <tbody>
              {products.map((product) => {
                const Icon = actionIcons[product.action] ?? LineChart;
                return <tr key={product.productId} className="border-t border-[#eaecf0] align-top text-[#344054] hover:bg-[#f8fafc]">
                  <td className="px-5 py-4"><div className="font-medium">{product.title}</div><div className="mt-1 text-xs text-[#667085]">{product.asin} · {product.sku || "No SKU"}</div></td>
                  <td className="px-5 py-4"><span className={`inline-flex items-center gap-1 rounded-full px-2.5 py-1 text-xs font-semibold ${actionStyles[product.action] ?? actionStyles.Maintain}`}><Icon className="h-3.5 w-3.5" />{product.action}</span></td>
                  <td className="px-5 py-4">{money.format(product.revenue)}</td>
                  <td className="px-5 py-4">{number.format(product.units)}</td>
                  <td className="px-5 py-4">{product.trendPercent.toFixed(0)}%</td>
                  <td className="px-5 py-4">{number.format(product.inventory)}</td>
                  <td className="px-5 py-4">{product.roas ? `${product.roas.toFixed(2)}x` : "-"}</td>
                  <td className="max-w-sm px-5 py-4 text-[#667085]">{product.reason}</td>
                </tr>;
              })}
            </tbody>
          </table>
          {growthQuery.isLoading || storesQuery.isLoading ? <div className="p-8 text-center text-sm text-[#667085]">Loading insights...</div> : null}
          {!growthQuery.isLoading && products.length === 0 ? <div className="p-10 text-center text-sm text-[#667085]">No product data found for this account yet.</div> : null}
          {growthQuery.error ? <div className="p-8 text-center text-sm text-[#b42318]">{growthQuery.error.message}</div> : null}
        </div>
      </PageCard>

      <PageCard className="overflow-hidden">
        <div className="border-b border-[#eaecf0] p-5"><h2 className="font-semibold">Marketplace performance</h2></div>
        <div className="overflow-x-auto">
          <table className="w-full min-w-[640px] text-sm">
            <thead className="bg-[#f8fafc] text-left text-xs uppercase tracking-wide text-[#667085]"><tr><th className="px-5 py-3.5">Marketplace</th><th className="px-5 py-3.5">Orders</th><th className="px-5 py-3.5">Units</th><th className="px-5 py-3.5">Revenue</th></tr></thead>
            <tbody>{marketplaces.map((marketplace) => <tr key={marketplace.countryCode} className="border-t border-[#eaecf0] text-[#344054]"><td className="px-5 py-4"><div className="font-medium text-[#101828]">{marketplace.name}</div><div className="text-xs text-[#667085]">{marketplace.countryCode}</div></td><td className="px-5 py-4">{number.format(marketplace.orders)}</td><td className="px-5 py-4">{number.format(marketplace.units)}</td><td className="px-5 py-4">{money.format(marketplace.revenue)}</td></tr>)}</tbody>
          </table>
        </div>
      </PageCard>
    </div>
  </DashboardShell>;
}
