"use client";

import { useMutation, useQuery } from "@tanstack/react-query";
import { Activity, ArrowUpRight, Database, RefreshCw, Store } from "lucide-react";
import { authedRequest, getAccessToken } from "@/lib/api";
import { useEffect, useState } from "react";
import { DashboardShell } from "@/components/dashboard-shell";

type AmazonStore = {
  id: string;
  name: string;
  sellerId: string;
  region: string;
  environment: "production" | "sandbox";
  status: string;
  lastImportedAt: string | null;
};

type StoresResponse = {
  stores: AmazonStore[] | null;
};

type OAuthStartResponse = {
  authorizationUrl: string;
  state: string;
};

type AmazonConnectionStatus = {
  ready: boolean;
  mode: string;
  sellerMessage: string;
  adminMessage?: string;
};

type ImportResult = {
  storeId: string;
  ordersImported: number;
  startedAt: string;
  finishedAt: string;
};

type Analytics = {
  summary: { revenue: number; profit: number; orders: number; units: number; adSpend: number; refunds: number; products: number; inventory: number; roas: number; profitMargin: number };
  trend: { date: string; revenue: number; profit: number; adSpend: number; refunds: number; units: number }[];
  products: { asin: string; sku: string; title: string; revenue: number; units: number; available: number }[];
  campaigns: { channel: string; name: string; status: string; spend: number; sales: number; orders: number; impressions: number; clicks: number; roas: number }[];
};

const money = new Intl.NumberFormat("en-US", { style: "currency", currency: "USD", maximumFractionDigits: 0 });

export default function DashboardPage() {
  const [amazonConnected, setAmazonConnected] = useState(false);
  const [selectedStoreId, setSelectedStoreId] = useState("");
  const [connectRegion, setConnectRegion] = useState("NA");

  useEffect(() => {
    if (!getAccessToken()) window.location.replace("/login");
    setAmazonConnected(new URLSearchParams(window.location.search).has("amazonConnected"));
  }, []);

  const storesQuery = useQuery({
    queryKey: ["amazon-stores"],
    queryFn: () => authedRequest<StoresResponse>("/v1/amazon/stores"),
    retry: false,
  });

  const analyticsQuery = useQuery({
    queryKey: ["analytics-dashboard", selectedStoreId],
    queryFn: () => authedRequest<Analytics>(`/v1/analytics/dashboard?days=180&storeId=${selectedStoreId}`),
    enabled: Boolean(selectedStoreId),
    retry: false,
  });

  const amazonStatusQuery = useQuery({
    queryKey: ["amazon-oauth-status"],
    queryFn: () => authedRequest<AmazonConnectionStatus>("/v1/amazon/oauth/status"),
    retry: false,
  });

  const connectMutation = useMutation({
    mutationFn: () =>
      authedRequest<OAuthStartResponse>("/v1/amazon/oauth/start", {
        method: "POST",
        body: JSON.stringify({ region: connectRegion, marketplaceId: marketplaceForRegion(connectRegion) }),
      }),
    onSuccess: (data) => {
      window.location.href = data.authorizationUrl;
    },
  });

  const sandboxMutation = useMutation({
	mutationFn: () => authedRequest<AmazonStore>("/v1/amazon/sandbox/connect", {
		method: "POST",
		body: JSON.stringify({ region: "NA" }),
	}),
	onSuccess: () => storesQuery.refetch(),
  });

  const importMutation = useMutation({
    mutationFn: (storeId: string) =>
      authedRequest<ImportResult>(`/v1/amazon/stores/${storeId}/import/orders`, {
        method: "POST",
        body: JSON.stringify({
          marketplaceId: "ATVPDKIKX0DER",
          createdAfter: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString(),
        }),
      }),
    onSuccess: () => Promise.all([analyticsQuery.refetch(), storesQuery.refetch()]),
  });

  const datasetMutation = useMutation({
    mutationFn: ({ storeId, dataset }: { storeId: string; dataset: string }) => authedRequest<{ dataset: string; recordsImported: number }>(`/v1/amazon/stores/${storeId}/import/${dataset}`, { method: "POST", body: JSON.stringify({ marketplaceId: "ATVPDKIKX0DER" }) }),
    onSuccess: () => Promise.all([analyticsQuery.refetch(), storesQuery.refetch()]),
  });

  const demoMutation = useMutation({
    mutationFn: (storeId: string) => authedRequest<{ orders: number; products: number; months: number }>("/v1/analytics/demo/generate", { method: "POST", body: JSON.stringify({ storeId, months: 6 }) }),
    onSuccess: () => Promise.all([analyticsQuery.refetch(), storesQuery.refetch()]),
  });

  const stores = storesQuery.data?.stores ?? [];
  const selectedStore = stores.find((store) => store.id === selectedStoreId);

  useEffect(() => {
    if (stores.length === 0) return;
    if (selectedStoreId && stores.some((store) => store.id === selectedStoreId)) return;
    setSelectedStoreId(stores.find((store) => store.environment === "production")?.id ?? stores[0].id);
  }, [stores, selectedStoreId]);

  const analytics = analyticsQuery.data;
  const amazonStatus = amazonStatusQuery.data;
  const maxRevenue = Math.max(...(analytics?.trend.map((point) => point.revenue) ?? [1]), 1);

  return (
    <DashboardShell title="Overview" description="Performance and operations across your Amazon business" action={
          <div className="hidden items-center gap-2 md:flex">
            <select value={connectRegion} onChange={(event) => setConnectRegion(event.target.value)} className="rounded-lg border border-[#d0d5dd] bg-white px-3 py-2 text-xs font-semibold text-[#344054] outline-none focus:border-[#14b8a6]">
              <option value="NA">North America</option>
              <option value="EU">Europe / India</option>
              <option value="FE">Far East</option>
            </select>
            <button
              className="inline-flex items-center gap-2 rounded-lg bg-[#0f766e] px-3 py-2 text-xs font-semibold text-white disabled:opacity-60"
              onClick={() => connectMutation.mutate()}
              disabled={connectMutation.isPending || amazonStatusQuery.isLoading || amazonStatus?.ready === false}
            >
              <ArrowUpRight className="h-4 w-4" aria-hidden />
              Connect Amazon seller account
            </button>
			<button className="inline-flex items-center gap-2 rounded-lg border border-[#d0d5dd] bg-white px-3 py-2 text-xs font-semibold text-[#344054] disabled:opacity-60" onClick={() => sandboxMutation.mutate()} disabled={sandboxMutation.isPending}>
			  <Database className="h-4 w-4" aria-hidden />
			  {sandboxMutation.isPending ? "Connecting..." : "Connect sandbox"}
			</button>
          </div>
    }><div className="mx-auto max-w-7xl">

        {amazonConnected ? <div className="mt-6 rounded-md border border-[#86d3c9] bg-[#f0fdfa] p-4 text-sm text-[#0f766e]">Amazon store connected successfully.</div> : null}
        {amazonStatus ? <div className={`mt-6 rounded-md border p-4 text-sm ${amazonStatus.ready ? "border-[#86d3c9] bg-[#f0fdfa] text-[#0f766e]" : "border-[#fedf89] bg-[#fffaeb] text-[#b54708]"}`}>
          <div className="font-semibold">Amazon seller connection</div>
          <p className="mt-1">{amazonStatus.sellerMessage}</p>
          {!amazonStatus.ready ? <p className="mt-2 text-xs text-[#667085]">This is an internal SaaS setup issue. Sellers should only sign in on Amazon and approve access; they should never enter Solution Provider credentials.</p> : null}
        </div> : null}

        <section className="mt-8">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <div><h2 className="text-xl font-semibold tracking-tight text-[#101828]">Business overview</h2><p className="mt-1 text-sm text-[#667085]">Last 180 days across your workspace</p></div>
            <div className="flex flex-wrap items-center gap-2">
              <select value={selectedStoreId} onChange={(event) => setSelectedStoreId(event.target.value)} className="rounded-xl border border-[#d0d5dd] bg-white px-3 py-2 text-sm text-[#344054] outline-none focus:border-[#14b8a6] focus:ring-4 focus:ring-[#ccfbf1]">
                {stores.map((store) => <option key={store.id} value={store.id}>{store.name} ({store.environment})</option>)}
              </select>
              <button className="rounded-md bg-[#2563eb] px-4 py-2 text-sm font-semibold text-white disabled:opacity-60" disabled={!selectedStore || selectedStore.environment !== "sandbox" || demoMutation.isPending} onClick={() => selectedStore && demoMutation.mutate(selectedStore.id)}>{demoMutation.isPending ? "Generating..." : "Generate sandbox demo data"}</button>
            </div>
          </div>
          {selectedStore ? <p className="mt-3 text-sm text-[#667085]">Viewing {selectedStore.name} in {selectedStore.environment} mode. Demo generation is allowed only for sandbox stores.</p> : null}
          {demoMutation.isSuccess ? <p className="mt-3 text-sm text-[#0f766e]">Demo dataset generated successfully.</p> : null}
          {demoMutation.error ? <p className="mt-3 text-sm text-[#b42318]">{demoMutation.error.message}</p> : null}
          <div className="mt-4 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            {[
              ["Revenue", money.format(analytics?.summary.revenue ?? 0)], ["Profit", money.format(analytics?.summary.profit ?? 0)],
              ["Orders", String(analytics?.summary.orders ?? 0)], ["Units sold", String(analytics?.summary.units ?? 0)],
              ["Ad spend", money.format(analytics?.summary.adSpend ?? 0)], ["ROAS", `${(analytics?.summary.roas ?? 0).toFixed(2)}×`],
              ["Profit margin", `${(analytics?.summary.profitMargin ?? 0).toFixed(1)}%`], ["Available inventory", String(analytics?.summary.inventory ?? 0)],
            ].map(([label, value]) => <div key={label} className="rounded-2xl border border-[#e4e7ec] bg-white p-4 shadow-sm shadow-slate-200/70"><p className="text-xs font-semibold uppercase tracking-wide text-[#667085]">{label}</p><p className="mt-2 text-2xl font-semibold tracking-tight text-[#101828]">{value}</p></div>)}
          </div>
        </section>

        <section className="mt-8 grid gap-6 lg:grid-cols-2">
          <div className="rounded-2xl border border-[#e4e7ec] bg-white p-5 shadow-sm shadow-slate-200/70">
            <h2 className="font-semibold">Revenue trend</h2>
            <div className="mt-5 flex h-48 items-end gap-[2px]" aria-label="Daily revenue chart">
              {(analytics?.trend ?? []).map((point) => <div key={point.date} title={`${new Date(point.date).toLocaleDateString()}: ${money.format(point.revenue)}`} className="min-w-[2px] flex-1 rounded-t bg-[#0f766e]" style={{ height: `${Math.max(3, point.revenue / maxRevenue * 100)}%` }} />)}
            </div>
          </div>
          <div className="overflow-hidden rounded-2xl border border-[#e4e7ec] bg-white shadow-sm shadow-slate-200/70">
            <h2 className="p-5 font-semibold">Campaign performance</h2>
            <div className="overflow-x-auto"><table className="w-full text-sm"><thead className="bg-[#f8fafc] text-left text-xs uppercase tracking-wide text-[#667085]"><tr><th className="px-4 py-3">Campaign</th><th className="px-4 py-3">Channel</th><th className="px-4 py-3">Spend</th><th className="px-4 py-3">Sales</th><th className="px-4 py-3">ROAS</th></tr></thead><tbody>{(analytics?.campaigns ?? []).map((campaign) => <tr key={campaign.name} className="border-t border-[#eaecf0] text-[#344054]"><td className="px-4 py-3 font-medium text-[#101828]">{campaign.name}</td><td className="px-4 py-3">{campaign.channel}</td><td className="px-4 py-3">{money.format(campaign.spend)}</td><td className="px-4 py-3">{money.format(campaign.sales)}</td><td className="px-4 py-3">{campaign.roas.toFixed(2)}×</td></tr>)}</tbody></table></div>
          </div>
        </section>

        <section className="mt-8 overflow-hidden rounded-2xl border border-[#e4e7ec] bg-white shadow-sm shadow-slate-200/70">
          <h2 className="p-5 font-semibold">Product performance</h2>
          <div className="overflow-x-auto"><table className="w-full text-sm"><thead className="bg-[#f8fafc] text-left text-xs uppercase tracking-wide text-[#667085]"><tr><th className="px-4 py-3">Product</th><th className="px-4 py-3">SKU</th><th className="px-4 py-3">Revenue</th><th className="px-4 py-3">Units</th><th className="px-4 py-3">Available</th></tr></thead><tbody>{(analytics?.products ?? []).map((product) => <tr key={product.asin} className="border-t border-[#eaecf0] text-[#344054]"><td className="px-4 py-3"><div className="font-medium text-[#101828]">{product.title}</div><div className="text-xs text-[#667085]">{product.asin}</div></td><td className="px-4 py-3">{product.sku}</td><td className="px-4 py-3">{money.format(product.revenue)}</td><td className="px-4 py-3">{product.units}</td><td className="px-4 py-3">{product.available}</td></tr>)}</tbody></table></div>
        </section>

        {connectMutation.error ? (
          <div className="mt-6 rounded-md border border-[#f2b8b5] bg-[#fff7f7] p-4 text-sm text-[#b42318]">
            {sellerSafeAmazonError(connectMutation.error.message)}
          </div>
        ) : null}
		{sandboxMutation.error ? <div className="mt-6 rounded-md border border-[#f2b8b5] bg-[#fff7f7] p-4 text-sm text-[#b42318]">{sandboxMutation.error.message}</div> : null}
		{sandboxMutation.isSuccess ? <div className="mt-6 rounded-md border border-[#86d3c9] bg-[#f0fdfa] p-4 text-sm text-[#0f766e]">Amazon sandbox connected.</div> : null}

        <section className="mt-8 grid gap-4 md:grid-cols-3">
          <div className="rounded-2xl border border-[#e4e7ec] bg-white p-5 shadow-sm shadow-slate-200/70">
            <Store className="h-6 w-6 text-[#0f766e]" aria-hidden />
            <h2 className="mt-3 text-xl font-semibold">{stores.length}</h2>
            <p className="mt-2 text-sm text-[#667085]">Connected Amazon stores</p>
          </div>
          <div className="rounded-2xl border border-[#e4e7ec] bg-white p-5 shadow-sm shadow-slate-200/70">
            <Database className="h-6 w-6 text-[#b45309]" aria-hidden />
            <h2 className="mt-3 text-xl font-semibold">Orders API</h2>
            <p className="mt-2 text-sm text-[#667085]">Imports recent seller orders into PostgreSQL</p>
          </div>
          <div className="rounded-2xl border border-[#e4e7ec] bg-white p-5 shadow-sm shadow-slate-200/70">
            <Activity className="h-6 w-6 text-[#2563eb]" aria-hidden />
            <h2 className="mt-3 text-xl font-semibold">OAuth only</h2>
            <p className="mt-2 text-sm text-[#667085]">No Amazon passwords are collected</p>
          </div>
        </section>

        <section className="mt-8">
          <h2 className="text-xl font-semibold">Amazon stores</h2>
          {storesQuery.error ? (
            <p className="mt-3 text-sm text-[#b42318]">{storesQuery.error.message}</p>
          ) : null}
          <div className="mt-4 overflow-hidden rounded-2xl border border-[#e4e7ec] bg-white shadow-sm shadow-slate-200/70">
            {stores.length === 0 ? (
              <div className="p-6 text-sm text-[#667085]">No stores connected yet.</div>
            ) : (
              <div className="divide-y divide-[#eaecf0]">
                {stores.map((store) => (
                  <div key={store.id} className="grid gap-4 p-5 lg:grid-cols-[1fr_auto] lg:items-center">
                    <div>
                      <div className="font-semibold">{store.name}</div>
                      <div className="mt-1 text-sm text-[#667085]">
						Seller {store.sellerId} · {store.region} · {store.environment} · {store.status}
                      </div>
                      <div className="mt-1 text-sm text-[#667085]">
                        Last import: {store.lastImportedAt ? new Date(store.lastImportedAt).toLocaleString() : "Never"}
                      </div>
                    </div>
					<div className="flex max-w-xl flex-wrap justify-end gap-2">
					  <button className="inline-flex items-center gap-2 rounded-md border border-[#c8ced8] px-3 py-2 text-xs font-semibold disabled:opacity-60" onClick={() => importMutation.mutate(store.id)} disabled={importMutation.isPending}><RefreshCw className="h-4 w-4" aria-hidden />Orders + items</button>
					  {["catalog", "inventory", "reports", "finances", "campaigns"].map((dataset) => <button key={dataset} className="rounded-md border border-[#c8ced8] px-3 py-2 text-xs font-semibold capitalize disabled:opacity-60" onClick={() => datasetMutation.mutate({ storeId: store.id, dataset })} disabled={datasetMutation.isPending}>{dataset}</button>)}
					</div>
                  </div>
                ))}
              </div>
            )}
          </div>
          {importMutation.data ? (
            <p className="mt-4 text-sm text-[#0f766e]">Imported {importMutation.data.ordersImported} orders.</p>
          ) : null}
          {importMutation.error ? (
            <p className="mt-4 text-sm text-[#b42318]">{importMutation.error.message}</p>
          ) : null}
		  {datasetMutation.data ? <p className="mt-4 text-sm text-[#0f766e]">Imported {datasetMutation.data.recordsImported} {datasetMutation.data.dataset} records.</p> : null}
		  {datasetMutation.error ? <p className="mt-4 text-sm text-[#b42318]">{datasetMutation.error.message}</p> : null}
        </section>
      </div>
    </DashboardShell>
  );
}

function marketplaceForRegion(region: string) {
  if (region === "EU") return "A1F83G8C2ARO7P";
  if (region === "FE") return "A1VC38T7YXB528";
  return "ATVPDKIKX0DER";
}

function sellerSafeAmazonError(message: string) {
  if (message.toLowerCase().includes("configuration error")) {
    return "Amazon seller connection is not available yet. This is an internal SaaS setup issue; sellers do not need Solution Provider credentials.";
  }
  return message;
}
