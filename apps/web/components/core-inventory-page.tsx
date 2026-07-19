"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Edit3, Plus, RefreshCw, Search, Trash2, X } from "lucide-react";
import { useMemo, useState } from "react";
import { authedRequest } from "@/lib/api";
import { DashboardShell, PageCard } from "@/components/dashboard-shell";

type Channel = { code: string; name: string; status: string; channel_type: string };
type ProductRow = { id: string; title: string; price?: number; compare_at_price?: number; cost_price?: number };
type InventoryRow = { product_id: string; title: string; variant_id: string; variant: string; sku: string; color: string; size: string; price?: number; compare_at_price?: number; cost_price?: number; stock_quantity: number; reserved_quantity: number; available: number; low_stock_threshold: number; status: string; listed_channel_codes?: string };

const inr = new Intl.NumberFormat("en-IN", { style: "currency", currency: "INR", maximumFractionDigits: 0 });
const money = (value: unknown) => inr.format(Number(value ?? 0));
const emptyInventory = { productId: "", title: "", sku: "", color: "", size: "", price: "0", compareAtPrice: "0", costPrice: "0", stockQuantity: "20", reservedQuantity: "0", lowStockThreshold: "5", status: "active", channelCodes: ["website"] };

export function CoreInventoryPage() {
  const [search, setSearch] = useState("");
  const [creating, setCreating] = useState(false);
  const [editing, setEditing] = useState<InventoryRow | null>(null);
  const [deleting, setDeleting] = useState<InventoryRow | null>(null);
  const queryClient = useQueryClient();
  const inventory = useQuery({ queryKey: ["commerce", "inventory"], queryFn: () => authedRequest<{ items: InventoryRow[] }>("/v1/commerce/inventory"), retry: false });
  const products = useQuery({ queryKey: ["commerce", "products"], queryFn: () => authedRequest<{ items: ProductRow[] }>("/v1/commerce/products"), retry: false });
  const channels = useQuery({ queryKey: ["commerce", "channels"], queryFn: () => authedRequest<{ items: Channel[] }>("/v1/commerce/channels"), retry: false });
  const remove = useMutation({ mutationFn: (id: string) => authedRequest(`/v1/commerce/inventory/${id}`, { method: "DELETE" }), onSuccess: () => { queryClient.invalidateQueries({ queryKey: ["commerce"] }); setDeleting(null); } });
  const rows = useMemo(() => (inventory.data?.items ?? []).filter(row => Object.values(row).some(value => String(value ?? "").toLowerCase().includes(search.toLowerCase()))), [inventory.data, search]);
  return <DashboardShell title="SKU Inventory" description="Operational stock control for product variants. Manage SKU quantities, reserved stock, low-stock alerts, and channel availability.">
    <div className="mb-5 grid gap-4 xl:grid-cols-4">
      <PageCard className="p-5"><Metric label="SKU records" value={inventory.data?.items.length ?? 0} note="One row per variant." /></PageCard>
      <PageCard className="p-5"><Metric label="Available units" value={(inventory.data?.items ?? []).reduce((sum, row) => sum + Number(row.available ?? 0), 0)} note="On hand minus reserved." /></PageCard>
      <PageCard className="p-5"><Metric label="Reserved units" value={(inventory.data?.items ?? []).reduce((sum, row) => sum + Number(row.reserved_quantity ?? 0), 0)} note="Held by orders/syncs." /></PageCard>
      <PageCard className="p-5"><Metric label="Low-stock SKUs" value={(inventory.data?.items ?? []).filter(row => Number(row.available ?? 0) <= Number(row.low_stock_threshold ?? 0)).length} note="Needs replenishment." /></PageCard>
    </div>
    <PageCard className="mb-5 p-5">
      <div className="flex flex-col gap-3 lg:flex-row">
        <label className="relative block flex-1"><Search className="absolute left-3 top-3 h-4 w-4 text-[#98a2b3]" /><input value={search} onChange={event => setSearch(event.target.value)} placeholder="Search SKU, product, color, size, stock status, channel..." className="w-full rounded-xl border border-[#d0d5dd] py-2.5 pl-10 pr-3 text-sm outline-none focus:border-[#14b8a6] focus:ring-4 focus:ring-[#ccfbf1]" /></label>
        <button onClick={() => setCreating(true)} className="inline-flex items-center justify-center rounded-xl bg-[#14b8a6] px-4 py-2.5 text-sm font-semibold text-white"><Plus className="mr-2 h-4 w-4" />Add SKU</button>
        <button onClick={() => inventory.refetch()} className="inline-flex items-center justify-center rounded-xl border border-[#d0d5dd] px-4 py-2.5 text-sm font-semibold text-[#344054]"><RefreshCw className="mr-2 h-4 w-4" />Refresh</button>
      </div>
      <p className="mt-2 text-xs text-[#667085]">Use Products to create the master item and generate variants. Use SKU Inventory to maintain stock, per-SKU pricing, and channel availability.</p>
    </PageCard>
    {(creating || editing) ? <InventoryModal products={products.data?.items ?? []} channels={channels.data?.items ?? []} item={editing} onClose={() => { setCreating(false); setEditing(null); }} /> : null}
    {deleting ? <ConfirmModal title="Archive SKU?" text={`This will archive SKU “${deleting.sku}” and remove it from active selling.`} busy={remove.isPending} onCancel={() => setDeleting(null)} onConfirm={() => remove.mutate(deleting.variant_id)} /> : null}
    <PageCard className="overflow-hidden"><div className="overflow-x-auto"><table className="w-full min-w-[1120px] text-sm"><thead className="bg-[#f8fafc] text-left text-xs uppercase tracking-wide text-[#667085]"><tr>{["Product", "Variant", "SKU", "SKU price", "On hand", "Reserved", "Available", "Low alert", "Channel availability", "Status", "Actions"].map(h => <th key={h} className="px-5 py-3.5 font-semibold">{h}</th>)}</tr></thead><tbody>{rows.map(row => <tr key={row.variant_id} className="border-t border-[#eaecf0] text-[#344054] hover:bg-[#f8fafc]"><td className="px-5 py-4"><div className="font-semibold text-[#101828]">{row.title}</div><div className="text-xs text-[#667085]">Master product</div></td><td className="px-5 py-4">{row.color || "—"} / {row.size || "—"}</td><td className="px-5 py-4">{row.sku}</td><td className="px-5 py-4">{money(row.price)}</td><td className="px-5 py-4">{row.stock_quantity}</td><td className="px-5 py-4">{row.reserved_quantity}</td><td className="px-5 py-4"><span className={Number(row.available) <= Number(row.low_stock_threshold) ? "font-semibold text-[#b42318]" : "font-semibold text-[#047857]"}>{row.available}</span></td><td className="px-5 py-4">{row.low_stock_threshold}</td><td className="px-5 py-4 text-xs">{formatChannels(row.listed_channel_codes)}</td><td className="px-5 py-4"><Status value={row.status} /></td><td className="px-5 py-4"><div className="flex gap-2"><button onClick={() => setEditing(row)} className="rounded-lg border border-[#d0d5dd] p-2 text-[#344054] hover:bg-white"><Edit3 className="h-4 w-4" /></button><button onClick={() => setDeleting(row)} className="rounded-lg border border-[#fecaca] p-2 text-[#b42318] hover:bg-[#fef3f2]"><Trash2 className="h-4 w-4" /></button></div></td></tr>)}</tbody></table>{inventory.isLoading ? <div className="p-8 text-center text-sm text-[#667085]">Loading SKU inventory...</div> : null}{!inventory.isLoading && rows.length === 0 ? <div className="p-10 text-center text-sm text-[#667085]">No SKU inventory records found. Create a product first, then generate variants/SKUs.</div> : null}{inventory.error ? <div className="p-8 text-center text-sm text-[#b42318]">{inventory.error.message}</div> : null}</div></PageCard>
  </DashboardShell>;
}

function InventoryModal({ products, channels, item, onClose }: { products: ProductRow[]; channels: Channel[]; item: InventoryRow | null; onClose: () => void }) {
  const queryClient = useQueryClient();
  const selectedCodes = item?.listed_channel_codes ? String(item.listed_channel_codes).split(",").filter(Boolean) : emptyInventory.channelCodes;
  const [form, setForm] = useState(() => item ? { productId: item.product_id, title: item.variant, sku: item.sku, color: item.color, size: item.size, price: String(item.price ?? 0), compareAtPrice: String(item.compare_at_price ?? 0), costPrice: String(item.cost_price ?? 0), stockQuantity: String(item.stock_quantity ?? 0), reservedQuantity: String(item.reserved_quantity ?? 0), lowStockThreshold: String(item.low_stock_threshold ?? 5), status: item.status ?? "active", channelCodes: selectedCodes } : { ...emptyInventory, productId: products[0]?.id ?? "" });
  const mutation = useMutation({ mutationFn: () => authedRequest(item ? `/v1/commerce/inventory/${item.variant_id}` : "/v1/commerce/inventory", { method: item ? "PUT" : "POST", body: JSON.stringify({ ...form, price: Number(form.price), compareAtPrice: Number(form.compareAtPrice), costPrice: Number(form.costPrice), stockQuantity: Number(form.stockQuantity), reservedQuantity: Number(form.reservedQuantity), lowStockThreshold: Number(form.lowStockThreshold) }) }), onSuccess: () => { queryClient.invalidateQueries({ queryKey: ["commerce"] }); onClose(); } });
  const update = (key: keyof typeof emptyInventory) => (event: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => setForm(value => ({ ...value, [key]: event.target.value }));
  const toggleChannel = (code: string) => setForm(value => ({ ...value, channelCodes: value.channelCodes.includes(code) ? value.channelCodes.filter(x => x !== code) : [...value.channelCodes, code] }));
  const selectedProduct = products.find(product => product.id === form.productId);
  return <Modal title={item ? "Edit SKU inventory" : "Add SKU to product"} subtitle="Use this for stock, reserved quantity, low-stock threshold, SKU price override, and channel availability." onClose={onClose}>
    <div className="grid gap-3 md:grid-cols-4">
      <Field label="Product" span="md:col-span-2"><select value={form.productId} onChange={update("productId")} disabled={Boolean(item)} className={inputClass}><option value="">Select product</option>{products.map(product => <option key={product.id} value={product.id}>{product.title}</option>)}</select></Field>
      <Field label="SKU"><input value={form.sku} onChange={update("sku")} className={inputClass} /></Field>
      <Field label="Status"><select value={form.status} onChange={update("status")} className={inputClass}><option value="active">Active</option><option value="archived">Archived</option><option value="draft">Draft</option></select></Field>
      <Field label="Color"><input value={form.color} onChange={update("color")} className={inputClass} /></Field>
      <Field label="Size"><input value={form.size} onChange={update("size")} className={inputClass} /></Field>
      <Field label="SKU title" span="md:col-span-2"><input value={form.title} onChange={update("title")} className={inputClass} /></Field>
      <Field label="SKU price"><input value={form.price} onChange={update("price")} className={inputClass} /></Field>
      <Field label="SKU MRP"><input value={form.compareAtPrice} onChange={update("compareAtPrice")} className={inputClass} /></Field>
      <Field label="SKU cost"><input value={form.costPrice} onChange={update("costPrice")} className={inputClass} /></Field>
      <Field label="Low alert"><input value={form.lowStockThreshold} onChange={update("lowStockThreshold")} className={inputClass} /></Field>
      <Field label="On-hand stock"><input value={form.stockQuantity} onChange={update("stockQuantity")} className={inputClass} /></Field>
      <Field label="Reserved quantity"><input value={form.reservedQuantity} onChange={update("reservedQuantity")} className={inputClass} /></Field>
      <div className="mt-5 rounded-xl bg-[#f8fafc] px-3 py-2 text-sm text-[#667085] md:col-span-2">Product price reference: {selectedProduct ? money(selectedProduct.price) : "Select product"}</div>
      <div className="md:col-span-4"><p className="mb-2 text-xs font-semibold uppercase tracking-wide text-[#667085]">SKU channel availability</p><div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-4">{channels.map(channel => <label key={channel.code} className="flex items-center justify-between rounded-xl border border-[#d0d5dd] px-3 py-2 text-sm"><span><span className="block font-semibold text-[#344054]">{channel.name}</span><span className="text-xs text-[#667085]">{channel.status}</span></span><input type="checkbox" checked={form.channelCodes.includes(channel.code)} onChange={() => toggleChannel(channel.code)} /></label>)}</div></div>
    </div>
    <div className="mt-5 flex justify-end gap-3"><button onClick={onClose} className="rounded-xl border border-[#d0d5dd] px-4 py-2.5 text-sm font-semibold text-[#344054]">Cancel</button><button onClick={() => mutation.mutate()} disabled={mutation.isPending || !form.productId || !form.sku} className="rounded-xl bg-[#14b8a6] px-4 py-2.5 text-sm font-semibold text-white disabled:opacity-60">{mutation.isPending ? "Saving..." : "Save SKU"}</button></div>
    {mutation.error ? <p className="mt-2 text-xs text-[#b42318]">{mutation.error.message}</p> : null}
  </Modal>;
}

function Metric({ label, value, note }: { label: string; value: number; note: string }) { return <><p className="text-xs font-semibold uppercase tracking-wide text-[#667085]">{label}</p><p className="mt-2 text-3xl font-semibold text-[#101828]">{value}</p><p className="mt-2 text-xs text-[#667085]">{note}</p></>; }
function Status({ value }: { value: unknown }) { const text = String(value ?? "not_listed"); const good = ["active", "synced"].includes(text); return <span className={`rounded-full px-2.5 py-1 text-xs font-semibold ${good ? "bg-[#ecfdf5] text-[#047857]" : "bg-[#f2f4f7] text-[#475467]"}`}>{text.replaceAll("_", " ")}</span>; }
function formatChannels(value: unknown) { const list = String(value ?? "").split(",").filter(Boolean); return list.length ? list.join(", ") : "—"; }
const inputClass = "w-full rounded-xl border border-[#d0d5dd] px-3 py-2 text-sm text-[#344054] outline-none focus:border-[#14b8a6] focus:ring-4 focus:ring-[#ccfbf1]";
function Field({ label, span = "", children }: { label: string; span?: string; children: React.ReactNode }) { return <label className={`block ${span}`}><span className="mb-1 block text-xs font-semibold text-[#344054]">{label}</span>{children}</label>; }
function Modal({ title, subtitle, onClose, children }: { title: string; subtitle: string; onClose: () => void; children: React.ReactNode }) { return <div className="fixed inset-0 z-50 flex items-center justify-center bg-[#0f172a]/50 p-4"><div className="max-h-[92vh] w-full max-w-5xl overflow-y-auto rounded-2xl bg-white shadow-2xl"><div className="sticky top-0 z-10 flex items-start justify-between gap-4 border-b border-[#eaecf0] bg-white px-5 py-4"><div><h2 className="text-lg font-semibold text-[#101828]">{title}</h2><p className="mt-1 text-xs text-[#667085]">{subtitle}</p></div><button onClick={onClose} className="rounded-lg p-2 text-[#667085] hover:bg-[#f2f4f7]"><X className="h-5 w-5" /></button></div><div className="p-5">{children}</div></div></div>; }
function ConfirmModal({ title, text, busy, onCancel, onConfirm }: { title: string; text: string; busy: boolean; onCancel: () => void; onConfirm: () => void }) { return <div className="fixed inset-0 z-50 flex items-center justify-center bg-[#0f172a]/50 p-4"><div className="w-full max-w-md rounded-2xl bg-white p-5 shadow-2xl"><h2 className="text-lg font-semibold text-[#101828]">{title}</h2><p className="mt-2 text-sm text-[#667085]">{text}</p><div className="mt-5 flex justify-end gap-3"><button onClick={onCancel} className="rounded-xl border border-[#d0d5dd] px-4 py-2.5 text-sm font-semibold text-[#344054]">Cancel</button><button onClick={onConfirm} disabled={busy} className="rounded-xl bg-[#b42318] px-4 py-2.5 text-sm font-semibold text-white disabled:opacity-60">{busy ? "Archiving..." : "Archive"}</button></div></div></div>; }
