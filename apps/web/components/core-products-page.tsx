"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Edit3, Plus, RefreshCw, Search, Trash2, X } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { authedRequest } from "@/lib/api";
import { DashboardShell, PageCard } from "@/components/dashboard-shell";

type Channel = { code: string; name: string; status: string; channel_type: string };
type ProductVariantRow = { id?: string; title?: string; sku?: string; color?: string; size?: string; price?: number; compareAtPrice?: number; costPrice?: number; stockQuantity?: number };
type VariantFormRow = { id?: string; color: string; size: string; price: string; compareAtPrice: string; costPrice: string; stockQuantity: string };
type ProductRow = Record<string, unknown> & { id: string; title: string; slug?: string; description?: string; category?: string; category_slug?: string; sku?: string; brand?: string; status?: string; price?: number; compare_at_price?: number; cost_price?: number; available_stock?: number; website_status?: string; amazon_status?: string; google_status?: string; meta_status?: string; listed_channel_codes?: string; is_featured?: boolean; images?: string[]; variants?: ProductVariantRow[] };

const inr = new Intl.NumberFormat("en-IN", { style: "currency", currency: "INR", maximumFractionDigits: 0 });
const money = (value: unknown) => inr.format(Number(value ?? 0));
const emptyProduct = { title: "", slug: "", description: "", categorySlug: "dresses", sku: "", brand: "Rangavali", status: "active", price: "2490", compareAtPrice: "3990", costPrice: "920", images: "/images/catalog/look-1.jpg", colors: "Wine,Ivory", sizes: "S,M,L,XL", stockQuantity: "36", isFeatured: false, channelCodes: ["website"], variants: [] as VariantFormRow[] };

export function CoreProductsPage() {
  const [search, setSearch] = useState("");
  const [editing, setEditing] = useState<ProductRow | null>(null);
  const [creating, setCreating] = useState(false);
  const [deleting, setDeleting] = useState<ProductRow | null>(null);
  const queryClient = useQueryClient();
  const query = useQuery({ queryKey: ["commerce", "products"], queryFn: () => authedRequest<{ items: ProductRow[] }>("/v1/commerce/products"), retry: false });
  const channels = useQuery({ queryKey: ["commerce", "channels"], queryFn: () => authedRequest<{ items: Channel[] }>("/v1/commerce/channels"), retry: false });
  const demo = useMutation({ mutationFn: () => authedRequest("/v1/commerce/demo/generate", { method: "POST", body: JSON.stringify({ months: 8 }) }), onSuccess: () => queryClient.invalidateQueries({ queryKey: ["commerce"] }) });
  const remove = useMutation({ mutationFn: (id: string) => authedRequest(`/v1/commerce/products/${id}`, { method: "DELETE" }), onSuccess: () => { queryClient.invalidateQueries({ queryKey: ["commerce"] }); setDeleting(null); } });
  const rows = useMemo(() => (query.data?.items ?? []).filter(row => Object.values(row).some(value => String(value ?? "").toLowerCase().includes(search.toLowerCase()))), [query.data, search]);
  return <DashboardShell title="Products" description="Master catalog. Create the product once, generate its variants/SKUs, then list it to sales channels.">
    <div className="mb-5 grid gap-4 xl:grid-cols-4">
      <PageCard className="p-5"><Metric label="Master products" value={query.data?.items.length ?? 0} note="Catalog records." /></PageCard>
      <PageCard className="p-5"><Metric label="Website listed" value={(query.data?.items ?? []).filter(p => p.website_status === "active").length} note="Visible to storefront." /></PageCard>
      <PageCard className="p-5"><Metric label="Ready to sell" value={(query.data?.items ?? []).filter(p => Number(p.available_stock ?? 0) > 0 && p.status === "active").length} note="Active with SKU stock." /></PageCard>
      <PageCard className="p-5"><Metric label="Low / no SKU stock" value={(query.data?.items ?? []).filter(p => Number(p.available_stock ?? 0) <= 5).length} note="Review inventory." /></PageCard>
    </div>
    <PageCard className="mb-5 p-5">
      <div className="flex flex-col gap-3 lg:flex-row">
        <label className="relative block flex-1"><Search className="absolute left-3 top-3 h-4 w-4 text-[#98a2b3]" /><input value={search} onChange={event => setSearch(event.target.value)} placeholder="Search catalog products, base SKU, category, channel status..." className="w-full rounded-xl border border-[#d0d5dd] py-2.5 pl-10 pr-3 text-sm outline-none focus:border-[#14b8a6] focus:ring-4 focus:ring-[#ccfbf1]" /></label>
        <button onClick={() => setCreating(true)} className="inline-flex items-center justify-center rounded-xl bg-[#14b8a6] px-4 py-2.5 text-sm font-semibold text-white"><Plus className="mr-2 h-4 w-4" />Add product</button>
        <button onClick={() => demo.mutate()} disabled={demo.isPending} className="inline-flex items-center justify-center rounded-xl border border-[#d0d5dd] px-4 py-2.5 text-sm font-semibold text-[#344054]"><RefreshCw className="mr-2 h-4 w-4" />Generate demo</button>
      </div>
      <p className="mt-2 text-xs text-[#667085]">Use this page for catalog details and initial variant/SKU setup. Use SKU Inventory later for stock adjustments and operational maintenance.</p>
    </PageCard>
    {(creating || editing) ? <ProductModal channels={channels.data?.items ?? []} product={editing} onClose={() => { setCreating(false); setEditing(null); }} /> : null}
    {deleting ? <ConfirmModal title="Archive product?" text={`This will archive “${deleting.title}” and hide its variants from active inventory.`} busy={remove.isPending} onCancel={() => setDeleting(null)} onConfirm={() => remove.mutate(deleting.id)} /> : null}
    <PageCard className="overflow-hidden"><div className="overflow-x-auto"><table className="w-full min-w-[1180px] text-sm"><thead className="bg-[#f8fafc] text-left text-xs uppercase tracking-wide text-[#667085]"><tr>{["Product", "Base SKU", "Category", "From price", "Total SKU stock", "Website", "Amazon", "Google", "Meta", "Listed channels", "Actions"].map(h => <th key={h} className="px-5 py-3.5 font-semibold">{h}</th>)}</tr></thead><tbody>{rows.map(row => <tr key={row.id} className="border-t border-[#eaecf0] text-[#344054] hover:bg-[#f8fafc]"><td className="px-5 py-4"><div className="font-semibold text-[#101828]">{row.title}</div><div className="text-xs text-[#667085]">{row.slug}</div></td><td className="px-5 py-4">{row.sku || "—"}</td><td className="px-5 py-4">{row.category || "—"}</td><td className="px-5 py-4">{money(row.price)}</td><td className="px-5 py-4">{String(row.available_stock ?? 0)}</td><td className="px-5 py-4"><Status value={row.website_status} /></td><td className="px-5 py-4"><Status value={row.amazon_status} /></td><td className="px-5 py-4"><Status value={row.google_status} /></td><td className="px-5 py-4"><Status value={row.meta_status} /></td><td className="px-5 py-4 text-xs">{formatChannels(row.listed_channel_codes)}</td><td className="px-5 py-4"><div className="flex gap-2"><button onClick={() => setEditing(row)} className="rounded-lg border border-[#d0d5dd] p-2 text-[#344054] hover:bg-white"><Edit3 className="h-4 w-4" /></button><button onClick={() => setDeleting(row)} className="rounded-lg border border-[#fecaca] p-2 text-[#b42318] hover:bg-[#fef3f2]"><Trash2 className="h-4 w-4" /></button></div></td></tr>)}</tbody></table>{query.isLoading ? <div className="p-8 text-center text-sm text-[#667085]">Loading catalog products...</div> : null}{!query.isLoading && rows.length === 0 ? <div className="p-10 text-center text-sm text-[#667085]">No catalog products found.</div> : null}{query.error ? <div className="p-8 text-center text-sm text-[#b42318]">{query.error.message}</div> : null}</div></PageCard>
  </DashboardShell>;
}

function ProductModal({ product, channels, onClose }: { product: ProductRow | null; channels: Channel[]; onClose: () => void }) {
  const queryClient = useQueryClient();
  const selectedCodes = product?.listed_channel_codes ? String(product.listed_channel_codes).split(",").filter(Boolean) : emptyProduct.channelCodes;
  const detail = useQuery({ queryKey: ["commerce", "product", product?.id], queryFn: () => authedRequest<ProductRow>(`/v1/commerce/products/${product?.id}`), enabled: Boolean(product?.id), retry: false });
  const [form, setForm] = useState(() => product ? { ...emptyProduct, title: String(product.title ?? ""), slug: String(product.slug ?? ""), categorySlug: String(product.category_slug ?? product.category ?? "dresses").toLowerCase().replaceAll(" ", "-"), sku: String(product.sku ?? ""), brand: String(product.brand ?? "Rangavali"), status: String(product.status ?? "active"), price: String(product.price ?? 0), compareAtPrice: String(product.compare_at_price ?? 0), costPrice: String(product.cost_price ?? 0), images: Array.isArray(product.images) ? product.images.join(",") : "/images/catalog/look-1.jpg", isFeatured: Boolean(product.is_featured), channelCodes: selectedCodes } : emptyProduct);
  useEffect(() => {
    if (!detail.data) return;
    const data = detail.data;
    const variants = Array.isArray(data.variants) ? data.variants : [];
    setForm(value => ({
      ...value,
      description: String(data.description ?? ""),
      images: Array.isArray(data.images) ? data.images.join(",") : value.images,
      colors: variants.length ? uniqueList(variants.map(v => String(v.color ?? "")).filter(Boolean)).join(",") : value.colors,
      sizes: variants.length ? uniqueList(variants.map(v => String(v.size ?? "")).filter(Boolean)).join(",") : value.sizes,
      variants: variants.map(v => ({
        id: v.id,
        color: String(v.color ?? ""),
        size: String(v.size ?? ""),
        price: Number(v.price ?? value.price) === Number(value.price) ? "" : String(v.price ?? ""),
        compareAtPrice: Number(v.compareAtPrice ?? value.compareAtPrice) === Number(value.compareAtPrice) ? "" : String(v.compareAtPrice ?? ""),
        costPrice: Number(v.costPrice ?? value.costPrice) === Number(value.costPrice) ? "" : String(v.costPrice ?? ""),
        stockQuantity: String(v.stockQuantity ?? value.stockQuantity),
      })),
    }));
  }, [detail.data]);
  const variantRows = useMemo(() => buildVariantRows(splitList(form.colors), splitList(form.sizes), form.variants, form.stockQuantity), [form.colors, form.sizes, form.variants, form.stockQuantity]);
  const mutation = useMutation({ mutationFn: () => authedRequest(product ? `/v1/commerce/products/${product.id}` : "/v1/commerce/products", { method: product ? "PUT" : "POST", body: JSON.stringify({ ...form, price: Number(form.price), compareAtPrice: Number(form.compareAtPrice), costPrice: Number(form.costPrice), stockQuantity: Number(form.stockQuantity), images: form.images.split(",").map(x => x.trim()).filter(Boolean), colors: splitList(form.colors), sizes: splitList(form.sizes), variants: variantRows.map(row => ({ id: row.id, color: row.color, size: row.size, price: Number(row.price || form.price), compareAtPrice: Number(row.compareAtPrice || form.compareAtPrice), costPrice: Number(row.costPrice || form.costPrice), stockQuantity: Number(row.stockQuantity || form.stockQuantity) })) }) }), onSuccess: () => { queryClient.invalidateQueries({ queryKey: ["commerce"] }); onClose(); } });
  const update = (key: keyof typeof emptyProduct) => (event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => setForm(value => ({ ...value, [key]: event.target.type === "checkbox" ? (event.target as HTMLInputElement).checked : event.target.value }));
  const toggleChannel = (code: string) => setForm(value => ({ ...value, channelCodes: value.channelCodes.includes(code) ? value.channelCodes.filter(x => x !== code) : [...value.channelCodes, code] }));
  const updateVariant = (index: number, key: keyof VariantFormRow, value: string) => setForm(current => {
    const rows = buildVariantRows(splitList(current.colors), splitList(current.sizes), current.variants, current.stockQuantity);
    rows[index] = { ...rows[index], [key]: value };
    return { ...current, variants: rows };
  });
  return <Modal title={product ? "Edit product + variants" : "Add product + variants"} subtitle="Define the master product once. Colors and sizes generate SKU inventory rows automatically." onClose={onClose}>
    <div className="grid gap-3 md:grid-cols-4">
      <Field label="Product title" span="md:col-span-2"><input value={form.title} onChange={update("title")} className={inputClass} /></Field>
      <Field label="Slug"><input value={form.slug} onChange={update("slug")} className={inputClass} /></Field>
      <Field label="Status"><select value={form.status} onChange={update("status")} className={inputClass}><option value="active">Active</option><option value="draft">Draft</option><option value="archived">Archived</option></select></Field>
      <Field label="Category"><input value={form.categorySlug} onChange={update("categorySlug")} className={inputClass} /></Field>
      <Field label="Base SKU"><input value={form.sku} onChange={update("sku")} className={inputClass} /></Field>
      <Field label="Brand"><input value={form.brand} onChange={update("brand")} className={inputClass} /></Field>
      <label className="mt-5 flex items-center gap-2 rounded-xl border border-[#d0d5dd] px-3 py-2 text-sm"><input type="checkbox" checked={form.isFeatured} onChange={update("isFeatured")} /> Featured</label>
      <Field label="Base price"><input value={form.price} onChange={update("price")} className={inputClass} /></Field>
      <Field label="Base MRP"><input value={form.compareAtPrice} onChange={update("compareAtPrice")} className={inputClass} /></Field>
      <Field label="Base cost"><input value={form.costPrice} onChange={update("costPrice")} className={inputClass} /></Field>
      <Field label="Opening stock per SKU"><input value={form.stockQuantity} onChange={update("stockQuantity")} className={inputClass} /></Field>
      <Field label="Colors" span="md:col-span-2"><input value={form.colors} onChange={update("colors")} className={inputClass} /></Field>
      <Field label="Sizes" span="md:col-span-2"><input value={form.sizes} onChange={update("sizes")} className={inputClass} /></Field>
      <Field label="Images" span="md:col-span-4"><input value={form.images} onChange={update("images")} className={inputClass} /></Field>
      <Field label="Description" span="md:col-span-4"><textarea value={form.description} onChange={update("description")} className={`${inputClass} min-h-20`} /></Field>
      <div className="md:col-span-4 rounded-2xl border border-[#eaecf0] bg-[#f8fafc] p-3">
        <div className="flex flex-col gap-1 sm:flex-row sm:items-end sm:justify-between">
          <div><p className="text-xs font-semibold uppercase tracking-wide text-[#344054]">Manage variants / SKU inventory</p><p className="text-xs text-[#667085]">Each color × size becomes one sellable SKU. Leave price/MRP/cost blank to use base values; stock is per SKU.</p></div>
          <span className="rounded-full bg-white px-2.5 py-1 text-xs font-semibold text-[#344054] ring-1 ring-[#eaecf0]">{variantRows.length} SKUs</span>
        </div>
        <div className="mt-3 max-h-72 overflow-auto rounded-xl border border-[#eaecf0] bg-white">
          <table className="w-full min-w-[760px] text-xs">
            <thead className="bg-white text-left text-[#667085]"><tr>{["Color", "Size", "SKU price", "SKU MRP", "SKU cost", "Opening stock"].map(head => <th key={head} className="px-3 py-2 font-semibold">{head}</th>)}</tr></thead>
            <tbody>{variantRows.map((row, index) => <tr key={`${row.color}-${row.size}`} className="border-t border-[#eaecf0]">
              <td className="px-3 py-2 font-semibold text-[#344054]">{row.color || "Default"}</td>
              <td className="px-3 py-2 font-semibold text-[#344054]">{row.size || "Default"}</td>
              <td className="px-3 py-2"><input value={row.price} onChange={event => updateVariant(index, "price", event.target.value)} placeholder={form.price} className={miniInputClass} /></td>
              <td className="px-3 py-2"><input value={row.compareAtPrice} onChange={event => updateVariant(index, "compareAtPrice", event.target.value)} placeholder={form.compareAtPrice} className={miniInputClass} /></td>
              <td className="px-3 py-2"><input value={row.costPrice} onChange={event => updateVariant(index, "costPrice", event.target.value)} placeholder={form.costPrice} className={miniInputClass} /></td>
              <td className="px-3 py-2"><input value={row.stockQuantity} onChange={event => updateVariant(index, "stockQuantity", event.target.value)} placeholder={form.stockQuantity} className={miniInputClass} /></td>
            </tr>)}</tbody>
          </table>
        </div>
      </div>
      <div className="md:col-span-4"><p className="mb-2 text-xs font-semibold uppercase tracking-wide text-[#667085]">List to sales channels</p><div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-4">{channels.map(channel => <label key={channel.code} className="flex items-center justify-between rounded-xl border border-[#d0d5dd] px-3 py-2 text-sm"><span><span className="block font-semibold text-[#344054]">{channel.name}</span><span className="text-xs text-[#667085]">{channel.status}</span></span><input type="checkbox" checked={form.channelCodes.includes(channel.code)} onChange={() => toggleChannel(channel.code)} /></label>)}</div></div>
    </div>
    <div className="mt-5 flex justify-end gap-3"><button onClick={onClose} className="rounded-xl border border-[#d0d5dd] px-4 py-2.5 text-sm font-semibold text-[#344054]">Cancel</button><button onClick={() => mutation.mutate()} disabled={mutation.isPending || !form.title} className="rounded-xl bg-[#14b8a6] px-4 py-2.5 text-sm font-semibold text-white disabled:opacity-60">{mutation.isPending ? "Saving..." : "Save product"}</button></div>
    {mutation.error ? <p className="mt-2 text-xs text-[#b42318]">{mutation.error.message}</p> : null}
  </Modal>;
}

function Metric({ label, value, note }: { label: string; value: number; note: string }) { return <><p className="text-xs font-semibold uppercase tracking-wide text-[#667085]">{label}</p><p className="mt-2 text-3xl font-semibold text-[#101828]">{value}</p><p className="mt-2 text-xs text-[#667085]">{note}</p></>; }
function Status({ value }: { value: unknown }) { const text = String(value ?? "not_listed"); const good = ["active", "synced", "paid", "delivered"].includes(text); return <span className={`rounded-full px-2.5 py-1 text-xs font-semibold ${good ? "bg-[#ecfdf5] text-[#047857]" : "bg-[#f2f4f7] text-[#475467]"}`}>{text.replaceAll("_", " ")}</span>; }
function formatChannels(value: unknown) { const list = String(value ?? "").split(",").filter(Boolean); return list.length ? list.join(", ") : "—"; }
const inputClass = "w-full rounded-xl border border-[#d0d5dd] px-3 py-2 text-sm text-[#344054] outline-none focus:border-[#14b8a6] focus:ring-4 focus:ring-[#ccfbf1]";
const miniInputClass = "w-full rounded-lg border border-[#d0d5dd] px-2 py-1.5 text-xs text-[#344054] outline-none focus:border-[#14b8a6] focus:ring-2 focus:ring-[#ccfbf1]";
function Field({ label, span = "", children }: { label: string; span?: string; children: React.ReactNode }) { return <label className={`block ${span}`}><span className="mb-1 block text-xs font-semibold text-[#344054]">{label}</span>{children}</label>; }
function Modal({ title, subtitle, onClose, children }: { title: string; subtitle: string; onClose: () => void; children: React.ReactNode }) { return <div className="fixed inset-0 z-50 flex items-center justify-center bg-[#0f172a]/50 p-4"><div className="max-h-[92vh] w-full max-w-5xl overflow-y-auto rounded-2xl bg-white shadow-2xl"><div className="sticky top-0 z-10 flex items-start justify-between gap-4 border-b border-[#eaecf0] bg-white px-5 py-4"><div><h2 className="text-lg font-semibold text-[#101828]">{title}</h2><p className="mt-1 text-xs text-[#667085]">{subtitle}</p></div><button onClick={onClose} className="rounded-lg p-2 text-[#667085] hover:bg-[#f2f4f7]"><X className="h-5 w-5" /></button></div><div className="p-5">{children}</div></div></div>; }
function ConfirmModal({ title, text, busy, onCancel, onConfirm }: { title: string; text: string; busy: boolean; onCancel: () => void; onConfirm: () => void }) { return <div className="fixed inset-0 z-50 flex items-center justify-center bg-[#0f172a]/50 p-4"><div className="w-full max-w-md rounded-2xl bg-white p-5 shadow-2xl"><h2 className="text-lg font-semibold text-[#101828]">{title}</h2><p className="mt-2 text-sm text-[#667085]">{text}</p><div className="mt-5 flex justify-end gap-3"><button onClick={onCancel} className="rounded-xl border border-[#d0d5dd] px-4 py-2.5 text-sm font-semibold text-[#344054]">Cancel</button><button onClick={onConfirm} disabled={busy} className="rounded-xl bg-[#b42318] px-4 py-2.5 text-sm font-semibold text-white disabled:opacity-60">{busy ? "Archiving..." : "Archive"}</button></div></div></div>; }
function splitList(value: string) { return value.split(",").map(item => item.trim()).filter(Boolean); }
function uniqueList(values: string[]) { return Array.from(new Set(values.map(value => value.trim()).filter(Boolean))); }
function variantKey(color: string, size: string) { return `${color.trim().toLowerCase()}|${size.trim().toLowerCase()}`; }
function buildVariantRows(colors: string[], sizes: string[], existing: VariantFormRow[], defaultStock: string): VariantFormRow[] {
  const lookup = new Map(existing.map(row => [variantKey(row.color, row.size), row]));
  const safeColors = colors.length ? colors : ["Default"];
  const safeSizes = sizes.length ? sizes : ["Default"];
  return safeColors.flatMap(color => safeSizes.map(size => ({ id: lookup.get(variantKey(color, size))?.id, color, size, price: lookup.get(variantKey(color, size))?.price ?? "", compareAtPrice: lookup.get(variantKey(color, size))?.compareAtPrice ?? "", costPrice: lookup.get(variantKey(color, size))?.costPrice ?? "", stockQuantity: lookup.get(variantKey(color, size))?.stockQuantity ?? defaultStock })));
}
