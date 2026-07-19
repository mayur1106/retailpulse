"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Plus, RefreshCw, Search, Sparkles } from "lucide-react";
import { useMemo, useState } from "react";
import { authedRequest } from "@/lib/api";
import { DashboardShell, PageCard } from "@/components/dashboard-shell";

type Column = { key: string; label: string; format?: (value: unknown, row: Record<string, unknown>) => string };
const inr = new Intl.NumberFormat("en-IN", { style: "currency", currency: "INR", maximumFractionDigits: 0 });
export const commerceMoney = (value: unknown) => inr.format(Number(value ?? 0));
export const commerceDate = (value: unknown) => value ? new Date(String(value)).toLocaleString("en-IN") : "—";

export function CommerceResourcePage({ resource, title, description, columns, allowProductCreate = false }: { resource: string; title: string; description: string; columns: Column[]; allowProductCreate?: boolean }) {
  const [search, setSearch] = useState("");
  const [showCreate, setShowCreate] = useState(false);
  const queryClient = useQueryClient();
  const query = useQuery({ queryKey: ["commerce", resource], queryFn: () => authedRequest<{ items: Record<string, unknown>[] }>(`/v1/commerce/${resource}`), retry: false });
  const demo = useMutation({
    mutationFn: () => authedRequest<{ products: number; orders: number }>("/v1/commerce/demo/generate", { method: "POST", body: JSON.stringify({ months: 8 }) }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["commerce"] }),
  });
  const items = useMemo(() => (query.data?.items ?? []).filter(item => Object.values(item).some(value => String(value ?? "").toLowerCase().includes(search.toLowerCase()))), [query.data, search]);
  return <DashboardShell title={title} description={description} action={<button onClick={() => demo.mutate()} disabled={demo.isPending} className="hidden rounded-xl bg-[#0f172a] px-3 py-2 text-xs font-semibold text-white shadow-sm hover:bg-[#1e293b] disabled:opacity-60 sm:inline-flex"><Sparkles className="mr-2 h-4 w-4" />Generate demo</button>}>
    <div className="mb-5 grid gap-4 lg:grid-cols-3">
      <PageCard className="p-5">
        <p className="text-xs font-semibold uppercase tracking-wide text-[#667085]">Commerce records</p>
        <p className="mt-2 text-3xl font-semibold tracking-tight text-[#101828]">{query.data?.items?.length ?? 0}</p>
        <p className="mt-2 text-xs text-[#667085]">Central commerce data used by storefront and future channel syncs.</p>
      </PageCard>
      <PageCard className="p-5 lg:col-span-2">
        <div className="flex flex-col gap-3 md:flex-row">
          <label className="relative block flex-1"><Search className="absolute left-3 top-3 h-4 w-4 text-[#98a2b3]" /><input value={search} onChange={event => setSearch(event.target.value)} placeholder={`Search ${title.toLowerCase()}...`} className="w-full rounded-xl border border-[#d0d5dd] bg-white py-2.5 pl-10 pr-3 text-sm text-[#344054] outline-none transition placeholder:text-[#98a2b3] focus:border-[#14b8a6] focus:ring-4 focus:ring-[#ccfbf1]" /></label>
          {allowProductCreate ? <button onClick={() => setShowCreate(value => !value)} className="inline-flex items-center justify-center rounded-xl border border-[#d0d5dd] bg-white px-4 py-2.5 text-sm font-semibold text-[#344054] hover:bg-[#f8fafc]"><Plus className="mr-2 h-4 w-4" />Add product</button> : null}
          <button onClick={() => query.refetch()} className="inline-flex items-center justify-center rounded-xl border border-[#d0d5dd] bg-white px-4 py-2.5 text-sm font-semibold text-[#344054] hover:bg-[#f8fafc]"><RefreshCw className="mr-2 h-4 w-4" />Refresh</button>
        </div>
        <p className="mt-2 text-xs text-[#667085]">Showing {items.length} matching records. Demo seed is realistic women’s apparel data.</p>
        {demo.data ? <p className="mt-2 text-xs font-medium text-[#047857]">Generated {demo.data.products} products and {demo.data.orders} orders.</p> : null}
        {demo.error ? <p className="mt-2 text-xs font-medium text-[#b42318]">{demo.error.message}</p> : null}
      </PageCard>
    </div>
    {showCreate ? <ProductCreateCard onDone={() => { setShowCreate(false); queryClient.invalidateQueries({ queryKey: ["commerce"] }); }} /> : null}
    <PageCard className="overflow-hidden">
      <div className="overflow-x-auto">
        <table className="w-full min-w-[860px] text-sm">
          <thead className="bg-[#f8fafc] text-left text-xs uppercase tracking-wide text-[#667085]"><tr>{columns.map(column => <th key={column.key} className="px-5 py-3.5 font-semibold">{column.label}</th>)}</tr></thead>
          <tbody>{items.map((row, index) => <tr key={String(row.id ?? row.order_number ?? row.code ?? index)} className="border-t border-[#eaecf0] text-[#344054] hover:bg-[#f8fafc]">{columns.map(column => <td key={column.key} className="max-w-xs px-5 py-4"><span className={["status", "payment_status", "fulfillment_status"].includes(column.key) ? "rounded-full bg-[#ecfdf5] px-2.5 py-1 text-xs font-semibold text-[#047857]" : ""}>{column.format ? column.format(row[column.key], row) : String(row[column.key] ?? "—")}</span></td>)}</tr>)}</tbody>
        </table>
        {query.isLoading ? <div className="p-8 text-center text-sm text-[#667085]">Loading commerce data...</div> : null}
        {!query.isLoading && items.length === 0 ? <div className="p-10 text-center text-sm text-[#667085]">No records found. Use “Generate demo” to seed the central commerce core.</div> : null}
        {query.error ? <div className="p-8 text-center text-sm text-[#b42318]">{query.error.message}</div> : null}
      </div>
    </PageCard>
  </DashboardShell>;
}

function ProductCreateCard({ onDone }: { onDone: () => void }) {
  const [form, setForm] = useState({ title: "", categorySlug: "dresses", price: "2490", compareAtPrice: "3990", costPrice: "920", stockQuantity: "36", colors: "Wine,Ivory", sizes: "S,M,L,XL" });
  const mutation = useMutation({
    mutationFn: () => authedRequest("/v1/commerce/products", {
      method: "POST",
      body: JSON.stringify({
        ...form,
        price: Number(form.price),
        compareAtPrice: Number(form.compareAtPrice),
        costPrice: Number(form.costPrice),
        stockQuantity: Number(form.stockQuantity),
        colors: form.colors.split(",").map(x => x.trim()).filter(Boolean),
        sizes: form.sizes.split(",").map(x => x.trim()).filter(Boolean),
      }),
    }),
    onSuccess: onDone,
  });
  const update = (key: keyof typeof form) => (event: React.ChangeEvent<HTMLInputElement>) => setForm(value => ({ ...value, [key]: event.target.value }));
  return <PageCard className="mb-5 p-5">
    <h2 className="text-base font-semibold text-[#101828]">Add ecommerce product</h2>
    <p className="mt-1 text-xs text-[#667085]">Creates variants automatically from colors and sizes, with inventory per variant.</p>
    <div className="mt-4 grid gap-3 md:grid-cols-4">
      <input value={form.title} onChange={update("title")} placeholder="Product title" className="rounded-xl border border-[#d0d5dd] px-3 py-2.5 text-sm md:col-span-2" />
      <input value={form.categorySlug} onChange={update("categorySlug")} placeholder="Category slug" className="rounded-xl border border-[#d0d5dd] px-3 py-2.5 text-sm" />
      <input value={form.stockQuantity} onChange={update("stockQuantity")} placeholder="Stock / variant" className="rounded-xl border border-[#d0d5dd] px-3 py-2.5 text-sm" />
      <input value={form.price} onChange={update("price")} placeholder="Price" className="rounded-xl border border-[#d0d5dd] px-3 py-2.5 text-sm" />
      <input value={form.compareAtPrice} onChange={update("compareAtPrice")} placeholder="MRP" className="rounded-xl border border-[#d0d5dd] px-3 py-2.5 text-sm" />
      <input value={form.costPrice} onChange={update("costPrice")} placeholder="Cost" className="rounded-xl border border-[#d0d5dd] px-3 py-2.5 text-sm" />
      <input value={form.colors} onChange={update("colors")} placeholder="Colors" className="rounded-xl border border-[#d0d5dd] px-3 py-2.5 text-sm" />
      <input value={form.sizes} onChange={update("sizes")} placeholder="Sizes" className="rounded-xl border border-[#d0d5dd] px-3 py-2.5 text-sm" />
    </div>
    <button onClick={() => mutation.mutate()} disabled={mutation.isPending || !form.title} className="mt-4 rounded-xl bg-[#14b8a6] px-4 py-2.5 text-sm font-semibold text-white disabled:opacity-60">Save product</button>
    {mutation.error ? <p className="mt-2 text-xs text-[#b42318]">{mutation.error.message}</p> : null}
  </PageCard>;
}
