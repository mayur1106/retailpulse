"use client";

import { useQuery } from "@tanstack/react-query";
import { AlertTriangle, CheckCircle2, Gauge, Target, Zap } from "lucide-react";
import { useEffect, useState } from "react";
import { DashboardShell, PageCard } from "@/components/dashboard-shell";
import { authedRequest } from "@/lib/api";

type AmazonStore = { id: string; name: string; environment: "production" | "sandbox" };
type HealthMetric = { key: string; label: string; score: number; status: string; value: string; detail: string; weight: number };
type TodayAction = { id: string; priority: string; category: string; title: string; description: string; impact: string; confidence: number; product?: string; campaign?: string; channel?: string; region?: string };
type SellerHealth = { score: number; grade: string; summary: string; dataOrigin: string; generatedAt: string; metrics: HealthMetric[]; actions: TodayAction[] };

const priorityStyle: Record<string, string> = {
  High: "bg-[#fff1f3] text-[#c01048] ring-[#fecdd3]",
  Medium: "bg-[#fffaeb] text-[#b54708] ring-[#fedf89]",
  Low: "bg-[#eff8ff] text-[#175cd3] ring-[#b2ddff]",
};

const statusStyle: Record<string, string> = {
  Healthy: "text-[#027a48]",
  Watch: "text-[#b54708]",
  "Needs attention": "text-[#c01048]",
};

export default function TodayActionsPage() {
  const [selectedStoreId, setSelectedStoreId] = useState("");
  const storesQuery = useQuery({ queryKey: ["amazon-stores"], queryFn: () => authedRequest<{ stores: AmazonStore[] | null }>("/v1/amazon/stores"), retry: false });
  const stores = storesQuery.data?.stores ?? [];

  useEffect(() => {
    if (stores.length === 0) return;
    if (selectedStoreId && stores.some((store) => store.id === selectedStoreId)) return;
    setSelectedStoreId(stores.find((store) => store.environment === "production")?.id ?? stores[0].id);
  }, [stores, selectedStoreId]);

  const selectedStore = stores.find((store) => store.id === selectedStoreId);
  const healthQuery = useQuery({
    queryKey: ["seller-health", selectedStoreId],
    queryFn: () => authedRequest<SellerHealth>(`/v1/analytics/health?days=90&storeId=${selectedStoreId}`),
    enabled: Boolean(selectedStoreId),
    retry: false,
  });

  const health = healthQuery.data;
  const score = health?.score ?? 0;

  return <DashboardShell title="Today’s actions" description="Seller health score and the highest-leverage actions to take next">
    <div className="mx-auto max-w-7xl space-y-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 className="text-lg font-semibold tracking-tight text-[#101828]">Seller health command center</h2>
          {selectedStore ? <p className="mt-1 text-sm text-[#667085]">{selectedStore.name} · {selectedStore.environment} · {health?.dataOrigin ?? "data"} data</p> : null}
        </div>
        <select value={selectedStoreId} onChange={(event) => setSelectedStoreId(event.target.value)} className="rounded-xl border border-[#d0d5dd] bg-white px-3 py-2.5 text-sm text-[#344054] outline-none transition focus:border-[#14b8a6] focus:ring-4 focus:ring-[#ccfbf1]">
          {stores.map((store) => <option key={store.id} value={store.id}>{store.name} ({store.environment})</option>)}
        </select>
      </div>

      <div className="grid gap-5 lg:grid-cols-[360px_1fr]">
        <PageCard className="p-6">
          <div className="flex items-start justify-between gap-4">
            <div>
              <p className="text-xs font-semibold uppercase tracking-wide text-[#667085]">Seller health</p>
              <div className="mt-3 flex items-end gap-3">
                <span className="text-6xl font-semibold tracking-tight text-[#101828]">{score}</span>
                <span className="mb-2 rounded-full bg-[#ecfdf3] px-3 py-1 text-sm font-semibold text-[#027a48] ring-1 ring-[#bbf7d0]">Grade {health?.grade ?? "—"}</span>
              </div>
            </div>
            <span className="grid h-12 w-12 place-items-center rounded-2xl bg-[#ecfdf5] text-[#0f766e]"><Gauge className="h-6 w-6" /></span>
          </div>
          <div className="mt-5 h-3 overflow-hidden rounded-full bg-[#eef2f6]">
            <div className="h-full rounded-full bg-[#14b8a6]" style={{ width: `${Math.min(100, Math.max(0, score))}%` }} />
          </div>
          <p className="mt-5 text-sm leading-6 text-[#667085]">{health?.summary ?? "Load a connected store to calculate seller health."}</p>
          {health?.generatedAt ? <p className="mt-3 text-xs text-[#98a2b3]">Generated {new Date(health.generatedAt).toLocaleString()}</p> : null}
        </PageCard>

        <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
          {(health?.metrics ?? []).map((metric) => <PageCard key={metric.key} className="p-5">
            <div className="flex items-start justify-between gap-3">
              <div>
                <p className="text-sm font-semibold text-[#101828]">{metric.label}</p>
                <p className="mt-1 text-xs text-[#667085]">{metric.value}</p>
              </div>
              <span className={`text-sm font-semibold ${statusStyle[metric.status] ?? "text-[#667085]"}`}>{metric.score}</span>
            </div>
            <div className="mt-4 h-2 overflow-hidden rounded-full bg-[#eef2f6]"><div className="h-full rounded-full bg-[#14b8a6]" style={{ width: `${metric.score}%` }} /></div>
            <p className="mt-3 text-xs leading-5 text-[#667085]">{metric.detail}</p>
          </PageCard>)}
        </div>
      </div>

      <PageCard className="overflow-hidden">
        <div className="flex items-center justify-between border-b border-[#eaecf0] p-5">
          <div>
            <h2 className="font-semibold text-[#101828]">Action queue</h2>
            <p className="mt-1 text-sm text-[#667085]">Prioritized from inventory, ads, returns, regional demand, and profit signals.</p>
          </div>
          <span className="hidden rounded-full bg-[#f8fafc] px-3 py-1 text-xs font-semibold text-[#475467] ring-1 ring-[#e4e7ec] sm:inline-flex">{health?.actions.length ?? 0} actions</span>
        </div>
        <div className="divide-y divide-[#eaecf0]">
          {(health?.actions ?? []).map((action) => {
            const Icon = action.priority === "High" ? AlertTriangle : action.category === "Ads" ? Zap : action.category === "Region" ? Target : CheckCircle2;
            return <div key={action.id} className="grid gap-4 p-5 lg:grid-cols-[1fr_auto] lg:items-start">
              <div className="flex gap-4">
                <span className="mt-1 grid h-10 w-10 shrink-0 place-items-center rounded-2xl bg-[#f8fafc] text-[#344054] ring-1 ring-[#e4e7ec]"><Icon className="h-5 w-5" /></span>
                <div>
                  <div className="flex flex-wrap items-center gap-2">
                    <span className={`rounded-full px-2.5 py-1 text-xs font-semibold ring-1 ${priorityStyle[action.priority] ?? priorityStyle.Low}`}>{action.priority}</span>
                    <span className="rounded-full bg-[#f8fafc] px-2.5 py-1 text-xs font-semibold text-[#475467] ring-1 ring-[#e4e7ec]">{action.category}</span>
                    {action.channel ? <span className="rounded-full bg-[#eef4ff] px-2.5 py-1 text-xs font-semibold text-[#3538cd] ring-1 ring-[#c7d7fe]">{action.channel}</span> : null}
                  </div>
                  <h3 className="mt-3 font-semibold text-[#101828]">{action.title}</h3>
                  <p className="mt-1 max-w-3xl text-sm leading-6 text-[#667085]">{action.description}</p>
                  <p className="mt-2 text-sm font-medium text-[#0f766e]">{action.impact}</p>
                  <div className="mt-2 flex flex-wrap gap-2 text-xs text-[#667085]">
                    {action.product ? <span>Product: {action.product}</span> : null}
                    {action.campaign ? <span>Campaign: {action.campaign}</span> : null}
                    {action.region ? <span>Region: {action.region}</span> : null}
                  </div>
                </div>
              </div>
              <div className="rounded-xl bg-[#f8fafc] px-3 py-2 text-right ring-1 ring-[#e4e7ec]">
                <p className="text-xs text-[#667085]">Confidence</p>
                <p className="font-semibold text-[#101828]">{Math.round(action.confidence * 100)}%</p>
              </div>
            </div>;
          })}
        </div>
        {healthQuery.isLoading || storesQuery.isLoading ? <div className="p-8 text-center text-sm text-[#667085]">Calculating seller health...</div> : null}
        {!healthQuery.isLoading && !storesQuery.isLoading && !health ? <div className="p-10 text-center text-sm text-[#667085]">No health data found yet.</div> : null}
        {healthQuery.error ? <div className="p-8 text-center text-sm text-[#b42318]">{healthQuery.error.message}</div> : null}
      </PageCard>
    </div>
  </DashboardShell>;
}
