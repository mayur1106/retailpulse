"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { BadgeCheck, Boxes, CreditCard, Eye, Globe2, GripVertical, ImageIcon, LayoutTemplate, Megaphone, Palette, Plus, Save, Settings2, Sparkles, ToggleLeft, Trash2, Truck, WalletCards } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { authedRequest } from "@/lib/api";
import { DashboardShell, PageCard } from "@/components/dashboard-shell";

type CMSStore = Record<string, unknown> & { name?: string; slug?: string; domain?: string; logo_url?: string; currency_code?: string; country_code?: string; timezone?: string; status?: string; settings?: Record<string, unknown> };
type CMSCategory = Record<string, unknown> & { id?: string; name?: string; slug?: string; description?: string; image_url?: string; sort_order?: number; status?: string };
type CMSSection = Record<string, unknown> & { id?: string; section_key?: string; section_type?: string; title?: string; subtitle?: string; layout?: string; image_url?: string; cta_label?: string; cta_href?: string; category_slug?: string; product_source?: string; max_items?: number; content?: Record<string, unknown>; sort_order?: number; status?: string };
type CMSPayment = Record<string, unknown> & { id?: string; code?: string; name?: string; provider?: string; instructions?: string; sort_order?: number; status?: string; settings?: Record<string, unknown> };
type CMSShipping = Record<string, unknown> & { id?: string; name?: string; country_code?: string; region_codes?: string[]; rate_type?: string; rate?: number; free_shipping_threshold?: number; estimated_days_min?: number; estimated_days_max?: number; cod_enabled?: boolean; status?: string };
type CMSConfig = { store: CMSStore; categories: CMSCategory[]; sections: CMSSection[]; paymentMethods: CMSPayment[]; shippingOptions: CMSShipping[] };
type TabKey = "brand" | "homepage" | "categories" | "payments" | "shipping";
type StoreForm = {
  name: string;
  slug: string;
  domain: string;
  logoUrl: string;
  currencyCode: string;
  countryCode: string;
  timezone: string;
  status: string;
  brandName: string;
  tagline: string;
  announcement: string;
  supportEmail: string;
  supportPhone: string;
  returnPolicy: string;
  primaryColor: string;
  accentColor: string;
  newsletterTitle: string;
};

const sectionTypes = ["hero", "category_tiles", "product_grid", "promo_tiles", "banner", "benefits", "newsletter"];
const productSources = ["all", "featured", "trending", "bestsellers", "new", "manual", "categories"];
const tabs: Array<{ key: TabKey; label: string; icon: React.ComponentType<{ className?: string }> }> = [
  { key: "brand", label: "Brand", icon: Palette },
  { key: "homepage", label: "Homepage", icon: LayoutTemplate },
  { key: "categories", label: "Categories", icon: Boxes },
  { key: "payments", label: "Payments", icon: CreditCard },
  { key: "shipping", label: "Shipping", icon: Truck },
];

export function StorefrontCMSPage() {
  const queryClient = useQueryClient();
  const query = useQuery({ queryKey: ["commerce", "cms"], queryFn: () => authedRequest<CMSConfig>("/v1/commerce/cms"), retry: false });
  const [activeTab, setActiveTab] = useState<TabKey>("brand");
  const [store, setStore] = useState<StoreForm>({
    name: "Rangavali", slug: "rangavali", domain: "", logoUrl: "", currencyCode: "INR", countryCode: "IN", timezone: "Asia/Kolkata", status: "live",
    brandName: "Rangavali", tagline: "", announcement: "", supportEmail: "", supportPhone: "", returnPolicy: "", primaryColor: "#6f1d46", accentColor: "#c8a24a", newsletterTitle: "₹300 off your first order",
  });
  const [categories, setCategories] = useState<CMSCategory[]>([]);
  const [sections, setSections] = useState<CMSSection[]>([]);
  const [payments, setPayments] = useState<CMSPayment[]>([]);
  const [shipping, setShipping] = useState<CMSShipping[]>([]);

  useEffect(() => {
    if (!query.data) return;
    const s = query.data.store ?? {};
    const settings = (s.settings ?? {}) as Record<string, unknown>;
    setStore({
      name: text(s.name, "Rangavali"),
      slug: text(s.slug, "rangavali"),
      domain: text(s.domain),
      logoUrl: text(s.logo_url),
      currencyCode: text(s.currency_code, "INR"),
      countryCode: text(s.country_code, "IN"),
      timezone: text(s.timezone, "Asia/Kolkata"),
      status: text(s.status, "live"),
      brandName: text(settings.brandName, text(s.name, "Rangavali")),
      tagline: text(settings.tagline),
      announcement: text(settings.announcement),
      supportEmail: text(settings.supportEmail),
      supportPhone: text(settings.supportPhone),
      returnPolicy: text(settings.returnPolicy),
      primaryColor: text(settings.primaryColor, "#6f1d46"),
      accentColor: text(settings.accentColor, "#c8a24a"),
      newsletterTitle: text(settings.newsletterTitle, "₹300 off your first order"),
    });
    setCategories(query.data.categories ?? []);
    setSections(query.data.sections ?? []);
    setPayments(query.data.paymentMethods ?? []);
    setShipping(query.data.shippingOptions ?? []);
  }, [query.data]);

  const activeSummary = useMemo(() => ({
    sections: sections.filter(item => item.status === "active").length,
    categories: categories.filter(item => item.status === "active").length,
    payments: payments.filter(item => item.status === "active").length,
    shipping: shipping.filter(item => item.status === "active").length,
  }), [sections, categories, payments, shipping]);

  const save = useMutation({
    mutationFn: () => authedRequest<CMSConfig>("/v1/commerce/cms", {
      method: "PUT",
      body: JSON.stringify({
        store: {
          name: store.name,
          slug: store.slug,
          domain: store.domain,
          logoUrl: store.logoUrl,
          currencyCode: store.currencyCode,
          countryCode: store.countryCode,
          timezone: store.timezone,
          status: store.status,
          settings: {
            brandName: store.brandName,
            tagline: store.tagline,
            announcement: store.announcement,
            supportEmail: store.supportEmail,
            supportPhone: store.supportPhone,
            returnPolicy: store.returnPolicy,
            primaryColor: store.primaryColor,
            accentColor: store.accentColor,
            newsletterTitle: store.newsletterTitle,
          },
        },
        categories: categories.map((item, index) => ({
          id: item.id,
          name: text(item.name),
          slug: text(item.slug),
          description: text(item.description),
          imageUrl: text(item.image_url),
          sortOrder: number(item.sort_order, index + 1),
          status: text(item.status, "active"),
        })),
        sections: sections.map((item, index) => ({
          id: item.id,
          sectionKey: text(item.section_key),
          sectionType: text(item.section_type, "product_grid"),
          title: text(item.title),
          subtitle: text(item.subtitle),
          layout: text(item.layout, "grid"),
          imageUrl: text(item.image_url),
          ctaLabel: text(item.cta_label),
          ctaHref: text(item.cta_href),
          categorySlug: text(item.category_slug),
          productSource: text(item.product_source, "all"),
          maxItems: number(item.max_items, 12),
          content: item.content ?? {},
          sortOrder: number(item.sort_order, index + 1),
          status: text(item.status, "active"),
        })),
        paymentMethods: payments.map((item, index) => ({
          id: item.id,
          code: text(item.code),
          name: text(item.name),
          provider: text(item.provider, text(item.code, "manual")),
          instructions: text(item.instructions),
          sortOrder: number(item.sort_order, index + 1),
          status: text(item.status, "inactive"),
          settings: item.settings ?? {},
        })),
        shippingOptions: shipping.map(item => ({
          id: item.id,
          name: text(item.name),
          countryCode: text(item.country_code, "IN"),
          regionCodes: Array.isArray(item.region_codes) ? item.region_codes : split(text(item.region_codes)),
          rateType: text(item.rate_type, "flat"),
          rate: number(item.rate, 0),
          freeShippingThreshold: number(item.free_shipping_threshold, 0),
          estimatedDaysMin: number(item.estimated_days_min, 3),
          estimatedDaysMax: number(item.estimated_days_max, 7),
          codEnabled: Boolean(item.cod_enabled),
          status: text(item.status, "active"),
        })),
      }),
    }),
    onSuccess: data => {
      queryClient.setQueryData(["commerce", "cms"], data);
      queryClient.invalidateQueries({ queryKey: ["commerce"] });
    },
  });

  const updateStore = (key: keyof typeof store) => (event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => setStore(value => ({ ...value, [key]: event.target.value }));
  const storefrontURL = "http://localhost:3006";

  return <DashboardShell title="Storefront CMS" description="Design storefront content, commerce toggles, and checkout options." action={<div className="flex gap-2"><a href={storefrontURL} target="_blank" className="hidden rounded-xl border border-[#d0d5dd] bg-white px-3 py-2 text-xs font-semibold text-[#344054] shadow-sm hover:bg-[#f8fafc] sm:inline-flex"><Eye className="mr-2 h-4 w-4" />Preview</a><button onClick={() => save.mutate()} disabled={save.isPending || query.isLoading} className="rounded-xl bg-[#0f172a] px-3 py-2 text-xs font-semibold text-white shadow-sm hover:bg-[#1e293b] disabled:opacity-60"><Save className="mr-2 inline h-4 w-4" />{save.isPending ? "Saving..." : "Publish changes"}</button></div>}>
    <div className="mb-5 overflow-hidden rounded-[28px] border border-[#e4e7ec] bg-white shadow-sm shadow-slate-200/70">
      <div className="grid gap-0 xl:grid-cols-[1.1fr_.9fr]">
        <div className="relative overflow-hidden bg-[#101828] p-6 text-white sm:p-8">
          <div className="absolute -right-16 -top-16 h-52 w-52 rounded-full bg-[#14b8a6]/20 blur-2xl" />
          <div className="absolute bottom-0 right-10 h-28 w-28 rounded-full bg-[#c8a24a]/20 blur-xl" />
          <div className="relative">
            <span className="inline-flex items-center rounded-full bg-white/10 px-3 py-1 text-xs font-semibold text-white ring-1 ring-white/15"><Sparkles className="mr-2 h-3.5 w-3.5 text-[#5eead4]" />Live storefront workspace</span>
            <h2 className="mt-5 text-3xl font-semibold tracking-tight sm:text-4xl">{store.brandName || store.name}</h2>
            <p className="mt-3 max-w-xl text-sm leading-6 text-white/65">{store.tagline || "Create a polished storefront experience from one central CMS."}</p>
            <div className="mt-6 flex flex-wrap gap-2 text-xs font-medium text-white/75">
              <span className="rounded-full bg-white/10 px-3 py-1.5 ring-1 ring-white/10">/{store.slug}</span>
              <span className="rounded-full bg-white/10 px-3 py-1.5 ring-1 ring-white/10">{store.currencyCode}</span>
              <span className="rounded-full bg-[#14b8a6]/20 px-3 py-1.5 text-[#99f6e4] ring-1 ring-[#5eead4]/20">{store.status}</span>
            </div>
          </div>
        </div>
        <div className="bg-gradient-to-br from-[#f8fafc] to-white p-5 sm:p-6">
          <div className="rounded-3xl border border-[#e4e7ec] bg-white p-4 shadow-sm">
            <div className="flex items-center justify-between border-b border-[#eef2f6] pb-3">
              <div className="flex items-center gap-2"><span className="h-2.5 w-2.5 rounded-full bg-[#ef4444]" /><span className="h-2.5 w-2.5 rounded-full bg-[#f59e0b]" /><span className="h-2.5 w-2.5 rounded-full bg-[#10b981]" /></div>
              <span className="text-[10px] font-semibold uppercase tracking-[0.18em] text-[#98a2b3]">Preview shell</span>
            </div>
            <div className="mt-4 rounded-2xl bg-[#24191e] p-4 text-white">
              <p className="text-[10px] font-semibold uppercase tracking-[0.2em] text-white/60">{store.announcement || "Announcement bar"}</p>
              <div className="mt-5 grid min-h-36 place-items-center rounded-2xl bg-white/10 px-5 text-center">
                {store.logoUrl ? <img src={store.logoUrl} alt={store.brandName} className="max-h-14 object-contain" /> : <div><p className="text-2xl font-semibold tracking-[0.16em]">{(store.brandName || store.name).toUpperCase()}</p><p className="mt-2 text-xs text-white/55">{store.newsletterTitle}</p></div>}
              </div>
            </div>
            <div className="mt-4 grid grid-cols-4 gap-2">
              <MiniStat label="Sections" value={activeSummary.sections} />
              <MiniStat label="Cats" value={activeSummary.categories} />
              <MiniStat label="Pay" value={activeSummary.payments} />
              <MiniStat label="Ship" value={activeSummary.shipping} />
            </div>
          </div>
        </div>
      </div>
    </div>

    {query.error ? <Notice tone="error">{query.error.message}</Notice> : null}
    {save.error ? <Notice tone="error">{save.error.message}</Notice> : null}
    {save.isSuccess ? <Notice tone="success">Storefront CMS saved. Refresh the storefront preview to see the latest published config.</Notice> : null}

    <div className="mb-5 grid gap-3 rounded-2xl border border-[#e4e7ec] bg-white p-2 shadow-sm shadow-slate-200/70 md:grid-cols-5">
      {tabs.map(tab => {
        const Icon = tab.icon;
        const active = activeTab === tab.key;
        return <button key={tab.key} onClick={() => setActiveTab(tab.key)} className={`flex items-center justify-center gap-2 rounded-xl px-3 py-3 text-sm font-semibold transition ${active ? "bg-[#0f172a] text-white shadow-sm" : "text-[#475467] hover:bg-[#f8fafc] hover:text-[#101828]"}`}><Icon className="h-4 w-4" />{tab.label}</button>;
      })}
    </div>

    {activeTab === "brand" ? <BrandPanel store={store} updateStore={updateStore} /> : null}
    {activeTab === "homepage" ? <HomepagePanel sections={sections} setSections={setSections} /> : null}
    {activeTab === "categories" ? <CategoriesPanel categories={categories} setCategories={setCategories} /> : null}
    {activeTab === "payments" ? <PaymentsPanel payments={payments} setPayments={setPayments} /> : null}
    {activeTab === "shipping" ? <ShippingPanel shipping={shipping} setShipping={setShipping} /> : null}
  </DashboardShell>;
}

function BrandPanel({ store, updateStore }: { store: StoreForm; updateStore: (key: keyof StoreForm) => (event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => void }) {
  return <div className="grid gap-5 xl:grid-cols-[1fr_380px]">
    <PageCard className="p-5 sm:p-6">
      <PanelHeader icon={Settings2} title="Brand identity" note="Control the visible brand details used by the header, footer, homepage, and checkout trust blocks." />
      <div className="mt-6 grid gap-4 md:grid-cols-4">
        <Field label="Store name" span="md:col-span-2"><input value={store.name} onChange={updateStore("name")} className={inputClass} /></Field>
        <Field label="Brand display name" span="md:col-span-2"><input value={store.brandName} onChange={updateStore("brandName")} className={inputClass} /></Field>
        <Field label="Store slug"><input value={store.slug} onChange={updateStore("slug")} className={inputClass} /></Field>
        <Field label="Domain"><input value={store.domain} onChange={updateStore("domain")} placeholder="store.example.com" className={inputClass} /></Field>
        <Field label="Currency"><input value={store.currencyCode} onChange={updateStore("currencyCode")} className={inputClass} /></Field>
        <Field label="Status"><select value={store.status} onChange={updateStore("status")} className={inputClass}><option value="live">Live</option><option value="draft">Draft</option><option value="paused">Paused</option></select></Field>
        <Field label="Logo URL/path" span="md:col-span-2"><input value={store.logoUrl} onChange={updateStore("logoUrl")} placeholder="/images/logo.png" className={inputClass} /></Field>
        <Field label="Announcement bar" span="md:col-span-2"><input value={store.announcement} onChange={updateStore("announcement")} className={inputClass} /></Field>
        <Field label="Brand tagline" span="md:col-span-2"><textarea value={store.tagline} onChange={updateStore("tagline")} className={`${inputClass} min-h-24 resize-none`} /></Field>
        <Field label="Return policy" span="md:col-span-2"><textarea value={store.returnPolicy} onChange={updateStore("returnPolicy")} className={`${inputClass} min-h-24 resize-none`} /></Field>
        <Field label="Support email"><input value={store.supportEmail} onChange={updateStore("supportEmail")} className={inputClass} /></Field>
        <Field label="Support phone"><input value={store.supportPhone} onChange={updateStore("supportPhone")} className={inputClass} /></Field>
        <Field label="Newsletter headline" span="md:col-span-2"><input value={store.newsletterTitle} onChange={updateStore("newsletterTitle")} className={inputClass} /></Field>
      </div>
    </PageCard>
    <PageCard className="overflow-hidden">
      <div className="border-b border-[#e4e7ec] p-5">
        <PanelHeader icon={Palette} title="Theme preview" note="Visual language that shoppers will feel first." compact />
      </div>
      <div className="space-y-4 p-5">
        <div className="rounded-3xl p-5 text-white shadow-sm" style={{ background: `linear-gradient(135deg, ${store.primaryColor || "#6f1d46"}, #24191e)` }}>
          <p className="text-[10px] font-semibold uppercase tracking-[0.2em] text-white/60">{store.announcement || "Announcement"}</p>
          <h3 className="mt-12 text-3xl font-semibold tracking-tight">{store.brandName || store.name}</h3>
          <p className="mt-2 text-sm text-white/65">{store.tagline || "Your storefront tagline appears here."}</p>
          <button className="mt-5 rounded-xl bg-white px-4 py-2 text-xs font-bold" style={{ color: store.primaryColor || "#6f1d46" }}>Shop now</button>
        </div>
        <div className="grid grid-cols-2 gap-3">
          <ColorInput label="Primary" value={store.primaryColor} onChange={updateStore("primaryColor")} />
          <ColorInput label="Accent" value={store.accentColor} onChange={updateStore("accentColor")} />
        </div>
      </div>
    </PageCard>
  </div>;
}

function HomepagePanel({ sections, setSections }: { sections: CMSSection[]; setSections: React.Dispatch<React.SetStateAction<CMSSection[]>> }) {
  return <PageCard className="p-5 sm:p-6">
    <PanelToolbar icon={LayoutTemplate} title="Homepage builder" note="Each enabled block appears on the storefront in sort order." actionLabel="Add section" onAdd={() => setSections(value => [...value, { section_key: `section-${value.length + 1}`, section_type: "product_grid", title: "New section", subtitle: "", layout: "grid", product_source: "all", max_items: 8, sort_order: value.length + 1, status: "active", content: {} }])} />
    <div className="mt-5 grid gap-4">
      {sections.map((section, index) => <div key={`${section.id ?? section.section_key}-${index}`} className="rounded-2xl border border-[#e4e7ec] bg-[#fcfcfd] p-4 transition hover:border-[#cbd5e1] hover:bg-white">
        <div className="flex flex-col gap-4 lg:flex-row lg:items-start">
          <div className="flex min-w-0 flex-1 gap-3">
            <div className="mt-1 hidden h-10 w-10 shrink-0 place-items-center rounded-xl bg-white text-[#98a2b3] ring-1 ring-[#e4e7ec] sm:grid"><GripVertical className="h-4 w-4" /></div>
            <div className="min-w-0 flex-1">
              <div className="flex flex-wrap items-center gap-2">
                <StatusPill active={section.status === "active"} />
                <span className="rounded-full bg-white px-2.5 py-1 text-xs font-semibold text-[#475467] ring-1 ring-[#e4e7ec]">{pretty(text(section.section_type, "section"))}</span>
                <span className="text-xs font-medium text-[#98a2b3]">Order {number(section.sort_order, index + 1)}</span>
              </div>
              <h3 className="mt-2 truncate text-base font-semibold text-[#101828]">{text(section.title, "Untitled section")}</h3>
              <p className="mt-1 line-clamp-1 text-xs text-[#667085]">{text(section.subtitle, "No subtitle added yet.")}</p>
            </div>
          </div>
          <div className="flex shrink-0 items-center gap-2">
            <Toggle value={section.status === "active"} onChange={checked => updateRow(setSections, index, { status: checked ? "active" : "inactive" })} compact />
            <button onClick={() => setSections(value => value.filter((_, i) => i !== index))} className="rounded-xl border border-[#fee4e2] bg-white p-2 text-[#b42318] hover:bg-[#fff5f5]"><Trash2 className="h-4 w-4" /></button>
          </div>
        </div>
        <div className="mt-4 grid gap-3 rounded-2xl bg-white p-3 ring-1 ring-[#eef2f6] md:grid-cols-12">
          <Field label="Key" span="md:col-span-2"><input value={text(section.section_key)} onChange={event => updateRow(setSections, index, { section_key: event.target.value })} className={inputClass} /></Field>
          <Field label="Type" span="md:col-span-2"><select value={text(section.section_type, "product_grid")} onChange={event => updateRow(setSections, index, { section_type: event.target.value })} className={inputClass}>{sectionTypes.map(type => <option key={type} value={type}>{pretty(type)}</option>)}</select></Field>
          <Field label="Title" span="md:col-span-4"><input value={text(section.title)} onChange={event => updateRow(setSections, index, { title: event.target.value })} className={inputClass} /></Field>
          <Field label="Order" span="md:col-span-1"><input value={text(section.sort_order, String(index + 1))} onChange={event => updateRow(setSections, index, { sort_order: Number(event.target.value) })} className={inputClass} /></Field>
          <Field label="Max items" span="md:col-span-1"><input value={text(section.max_items, "8")} onChange={event => updateRow(setSections, index, { max_items: Number(event.target.value) })} className={inputClass} /></Field>
          <Field label="Source" span="md:col-span-2"><select value={text(section.product_source, "all")} onChange={event => updateRow(setSections, index, { product_source: event.target.value })} className={inputClass}>{productSources.map(source => <option key={source} value={source}>{pretty(source)}</option>)}</select></Field>
          <Field label="Subtitle" span="md:col-span-5"><input value={text(section.subtitle)} onChange={event => updateRow(setSections, index, { subtitle: event.target.value })} className={inputClass} /></Field>
          <Field label="Image URL/path" span="md:col-span-3"><input value={text(section.image_url)} onChange={event => updateRow(setSections, index, { image_url: event.target.value })} className={inputClass} /></Field>
          <Field label="CTA label" span="md:col-span-2"><input value={text(section.cta_label)} onChange={event => updateRow(setSections, index, { cta_label: event.target.value })} className={inputClass} /></Field>
          <Field label="CTA href" span="md:col-span-2"><input value={text(section.cta_href)} onChange={event => updateRow(setSections, index, { cta_href: event.target.value })} className={inputClass} /></Field>
          <Field label="Category slug" span="md:col-span-3"><input value={text(section.category_slug)} onChange={event => updateRow(setSections, index, { category_slug: event.target.value })} className={inputClass} /></Field>
          <Field label="Layout" span="md:col-span-3"><input value={text(section.layout, "grid")} onChange={event => updateRow(setSections, index, { layout: event.target.value })} className={inputClass} /></Field>
          <Field label="Advanced content JSON" span="md:col-span-6"><input value={JSON.stringify(section.content ?? {})} onChange={event => updateJSONRow(setSections, index, event.target.value)} className={inputClass} /></Field>
        </div>
      </div>)}
    </div>
  </PageCard>;
}

function CategoriesPanel({ categories, setCategories }: { categories: CMSCategory[]; setCategories: React.Dispatch<React.SetStateAction<CMSCategory[]>> }) {
  return <PageCard className="p-5 sm:p-6">
    <PanelToolbar icon={ImageIcon} title="Navigation categories" note="Active categories show in header navigation and homepage category tiles." actionLabel="Add category" onAdd={() => setCategories(value => [...value, { name: "New category", slug: "new-category", description: "", image_url: "/images/catalog/look-1.jpg", sort_order: value.length + 1, status: "active" }])} />
    <div className="mt-5 grid gap-4 md:grid-cols-2 xl:grid-cols-3">
      {categories.map((row, index) => <EditableCard key={`${row.id ?? row.slug}-${index}`} title={text(row.name, "Untitled category")} subtitle={`/${text(row.slug, "category")}`} active={row.status === "active"} onToggle={checked => updateRow(setCategories, index, { status: checked ? "active" : "inactive" })} onDelete={() => setCategories(value => value.filter((_, i) => i !== index))}>
        <div className="grid gap-3">
          <Field label="Name"><input value={text(row.name)} onChange={event => updateRow(setCategories, index, { name: event.target.value })} className={inputClass} /></Field>
          <Field label="Slug"><input value={text(row.slug)} onChange={event => updateRow(setCategories, index, { slug: event.target.value })} className={inputClass} /></Field>
          <Field label="Image URL"><input value={text(row.image_url)} onChange={event => updateRow(setCategories, index, { image_url: event.target.value })} className={inputClass} /></Field>
          <div className="grid grid-cols-2 gap-3">
            <Field label="Sort order"><input value={text(row.sort_order, String(index + 1))} onChange={event => updateRow(setCategories, index, { sort_order: Number(event.target.value) })} className={inputClass} /></Field>
            <Field label="Status"><select value={text(row.status, "active")} onChange={event => updateRow(setCategories, index, { status: event.target.value })} className={inputClass}><option value="active">Active</option><option value="inactive">Hidden</option></select></Field>
          </div>
          <Field label="Description"><textarea value={text(row.description)} onChange={event => updateRow(setCategories, index, { description: event.target.value })} className={`${inputClass} min-h-20 resize-none`} /></Field>
        </div>
      </EditableCard>)}
    </div>
  </PageCard>;
}

function PaymentsPanel({ payments, setPayments }: { payments: CMSPayment[]; setPayments: React.Dispatch<React.SetStateAction<CMSPayment[]>> }) {
  return <PageCard className="p-5 sm:p-6">
    <PanelToolbar icon={WalletCards} title="Payment methods" note="Enable only payment methods you want buyers to see at checkout." actionLabel="Add method" onAdd={() => setPayments(value => [...value, { code: "manual", name: "Manual payment", provider: "manual", instructions: "", sort_order: value.length + 1, status: "inactive", settings: {} }])} />
    <div className="mt-5 grid gap-4 lg:grid-cols-2">
      {payments.map((row, index) => <EditableCard key={`${row.id ?? row.code}-${index}`} title={text(row.name, "Payment method")} subtitle={`${pretty(text(row.provider, "manual"))} · ${text(row.code, "code")}`} active={row.status === "active"} onToggle={checked => updateRow(setPayments, index, { status: checked ? "active" : "inactive" })} onDelete={() => setPayments(value => value.filter((_, i) => i !== index))}>
        <div className="grid gap-3 md:grid-cols-2">
          <Field label="Code"><input value={text(row.code)} onChange={event => updateRow(setPayments, index, { code: event.target.value })} className={inputClass} /></Field>
          <Field label="Provider"><input value={text(row.provider)} onChange={event => updateRow(setPayments, index, { provider: event.target.value })} className={inputClass} /></Field>
          <Field label="Name" span="md:col-span-2"><input value={text(row.name)} onChange={event => updateRow(setPayments, index, { name: event.target.value })} className={inputClass} /></Field>
          <Field label="Sort order"><input value={text(row.sort_order, String(index + 1))} onChange={event => updateRow(setPayments, index, { sort_order: Number(event.target.value) })} className={inputClass} /></Field>
          <Field label="Status"><select value={text(row.status, "inactive")} onChange={event => updateRow(setPayments, index, { status: event.target.value })} className={inputClass}><option value="active">Active</option><option value="inactive">Hidden</option></select></Field>
          <Field label="Checkout instructions" span="md:col-span-2"><textarea value={text(row.instructions)} onChange={event => updateRow(setPayments, index, { instructions: event.target.value })} className={`${inputClass} min-h-20 resize-none`} /></Field>
        </div>
      </EditableCard>)}
    </div>
  </PageCard>;
}

function ShippingPanel({ shipping, setShipping }: { shipping: CMSShipping[]; setShipping: React.Dispatch<React.SetStateAction<CMSShipping[]>> }) {
  return <PageCard className="p-5 sm:p-6">
    <PanelToolbar icon={Truck} title="Shipping zones" note="Enabled zones appear at checkout and feed storefront delivery promises." actionLabel="Add zone" onAdd={() => setShipping(value => [...value, { name: "New zone", country_code: "IN", region_codes: [], rate_type: "flat", rate: 99, free_shipping_threshold: 1999, estimated_days_min: 3, estimated_days_max: 7, cod_enabled: true, status: "active" }])} />
    <div className="mt-5 grid gap-4 lg:grid-cols-2">
      {shipping.map((row, index) => <EditableCard key={`${row.id ?? row.name}-${index}`} title={text(row.name, "Shipping zone")} subtitle={`${money(row.rate)} · ${number(row.estimated_days_min, 3)}-${number(row.estimated_days_max, 7)} days`} active={row.status === "active"} onToggle={checked => updateRow(setShipping, index, { status: checked ? "active" : "inactive" })} onDelete={() => setShipping(value => value.filter((_, i) => i !== index))}>
        <div className="grid gap-3 md:grid-cols-2">
          <Field label="Zone name"><input value={text(row.name)} onChange={event => updateRow(setShipping, index, { name: event.target.value })} className={inputClass} /></Field>
          <Field label="Country"><input value={text(row.country_code, "IN")} onChange={event => updateRow(setShipping, index, { country_code: event.target.value })} className={inputClass} /></Field>
          <Field label="Region codes"><input value={Array.isArray(row.region_codes) ? row.region_codes.join(",") : text(row.region_codes)} onChange={event => updateRow(setShipping, index, { region_codes: split(event.target.value) })} className={inputClass} /></Field>
          <Field label="Rate type"><input value={text(row.rate_type, "flat")} onChange={event => updateRow(setShipping, index, { rate_type: event.target.value })} className={inputClass} /></Field>
          <Field label="Rate"><input value={text(row.rate, "0")} onChange={event => updateRow(setShipping, index, { rate: Number(event.target.value) })} className={inputClass} /></Field>
          <Field label="Free above"><input value={text(row.free_shipping_threshold, "0")} onChange={event => updateRow(setShipping, index, { free_shipping_threshold: Number(event.target.value) })} className={inputClass} /></Field>
          <Field label="Min days"><input value={text(row.estimated_days_min, "3")} onChange={event => updateRow(setShipping, index, { estimated_days_min: Number(event.target.value) })} className={inputClass} /></Field>
          <Field label="Max days"><input value={text(row.estimated_days_max, "7")} onChange={event => updateRow(setShipping, index, { estimated_days_max: Number(event.target.value) })} className={inputClass} /></Field>
          <label className="flex items-center justify-between rounded-xl border border-[#d0d5dd] bg-white px-3 py-2.5 text-sm font-semibold text-[#344054] md:col-span-2"><span>Cash on delivery</span><input type="checkbox" checked={Boolean(row.cod_enabled)} onChange={event => updateRow(setShipping, index, { cod_enabled: event.target.checked })} /></label>
        </div>
      </EditableCard>)}
    </div>
  </PageCard>;
}

function PanelHeader({ icon: Icon, title, note, compact = false }: { icon: React.ComponentType<{ className?: string }>; title: string; note: string; compact?: boolean }) {
  return <div className="flex items-start gap-3"><span className={`rounded-2xl bg-[#ecfdf5] text-[#047857] ${compact ? "p-2" : "p-3"}`}><Icon className={compact ? "h-4 w-4" : "h-5 w-5"} /></span><div><h2 className={`${compact ? "text-sm" : "text-base"} font-semibold text-[#101828]`}>{title}</h2><p className="mt-1 text-xs leading-5 text-[#667085]">{note}</p></div></div>;
}

function PanelToolbar({ icon, title, note, actionLabel, onAdd }: { icon: React.ComponentType<{ className?: string }>; title: string; note: string; actionLabel: string; onAdd: () => void }) {
  return <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between"><PanelHeader icon={icon} title={title} note={note} /><button onClick={onAdd} className="inline-flex items-center justify-center rounded-xl border border-[#d0d5dd] bg-white px-4 py-2.5 text-sm font-semibold text-[#344054] shadow-sm hover:bg-[#f8fafc]"><Plus className="mr-2 h-4 w-4" />{actionLabel}</button></div>;
}

function EditableCard({ title, subtitle, active, onToggle, onDelete, children }: { title: string; subtitle: string; active: boolean; onToggle: (checked: boolean) => void; onDelete: () => void; children: React.ReactNode }) {
  return <div className="rounded-2xl border border-[#e4e7ec] bg-[#fcfcfd] p-4 transition hover:border-[#cbd5e1] hover:bg-white">
    <div className="mb-4 flex items-start justify-between gap-3">
      <div className="min-w-0"><div className="flex flex-wrap items-center gap-2"><StatusPill active={active} /><p className="truncate text-sm font-semibold text-[#101828]">{title}</p></div><p className="mt-1 truncate text-xs text-[#667085]">{subtitle}</p></div>
      <div className="flex shrink-0 gap-2"><Toggle value={active} onChange={onToggle} compact /><button onClick={onDelete} className="rounded-xl border border-[#fee4e2] bg-white p-2 text-[#b42318] hover:bg-[#fff5f5]"><Trash2 className="h-4 w-4" /></button></div>
    </div>
    <div className="rounded-2xl bg-white p-3 ring-1 ring-[#eef2f6]">{children}</div>
  </div>;
}

function MiniStat({ label, value }: { label: string; value: number }) {
  return <div className="rounded-2xl bg-[#f8fafc] p-3 text-center ring-1 ring-[#eef2f6]"><p className="text-lg font-semibold text-[#101828]">{value}</p><p className="mt-0.5 text-[10px] font-semibold uppercase tracking-wide text-[#98a2b3]">{label}</p></div>;
}

function Notice({ tone, children }: { tone: "success" | "error"; children: React.ReactNode }) {
  return <div className={`mb-5 rounded-2xl border px-4 py-3 text-sm font-medium ${tone === "success" ? "border-[#bbf7d0] bg-[#ecfdf5] text-[#047857]" : "border-[#fecaca] bg-[#fef3f2] text-[#b42318]"}`}>{children}</div>;
}

function StatusPill({ active }: { active: boolean }) {
  return <span className={`inline-flex items-center rounded-full px-2.5 py-1 text-[11px] font-semibold ${active ? "bg-[#ecfdf5] text-[#047857] ring-1 ring-[#bbf7d0]" : "bg-[#f2f4f7] text-[#667085] ring-1 ring-[#e4e7ec]"}`}>{active ? <BadgeCheck className="mr-1 h-3 w-3" /> : <ToggleLeft className="mr-1 h-3 w-3" />}{active ? "Active" : "Hidden"}</span>;
}

function ColorInput({ label, value, onChange }: { label: string; value: string; onChange: (event: React.ChangeEvent<HTMLInputElement>) => void }) {
  return <label className="rounded-2xl border border-[#e4e7ec] bg-white p-3"><span className="text-xs font-semibold text-[#667085]">{label}</span><div className="mt-2 flex items-center gap-2"><input type="color" value={value || "#000000"} onChange={onChange} className="h-9 w-10 rounded-lg border border-[#d0d5dd] bg-white p-1" /><input value={value} onChange={onChange} className="min-w-0 flex-1 bg-transparent text-sm font-semibold text-[#344054] outline-none" /></div></label>;
}

function Field({ label, span = "", children }: { label: string; span?: string; children: React.ReactNode }) {
  return <label className={`block ${span}`}><span className="mb-1.5 block text-xs font-semibold text-[#475467]">{label}</span>{children}</label>;
}

function Toggle({ value, onChange, compact = false }: { value: boolean; onChange: (value: boolean) => void; compact?: boolean }) {
  return <button type="button" onClick={() => onChange(!value)} className={`inline-flex items-center justify-center rounded-xl text-xs font-semibold transition ${compact ? "h-9 px-3" : "h-10 w-full"} ${value ? "bg-[#ecfdf5] text-[#047857] ring-1 ring-[#bbf7d0]" : "bg-[#f2f4f7] text-[#667085] ring-1 ring-[#e4e7ec]"}`}>{value ? "On" : "Off"}</button>;
}

function updateRow<T>(setter: React.Dispatch<React.SetStateAction<T[]>>, index: number, patch: Partial<T>) {
  setter(value => value.map((row, i) => i === index ? { ...row, ...patch } : row));
}

function updateJSONRow(setter: React.Dispatch<React.SetStateAction<CMSSection[]>>, index: number, raw: string) {
  try { updateRow(setter, index, { content: JSON.parse(raw || "{}") }); } catch { updateRow(setter, index, { content: { raw } }); }
}

function text(value: unknown, fallback = "") {
  return value === undefined || value === null ? fallback : String(value);
}

function number(value: unknown, fallback = 0) {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : fallback;
}

function split(value: string) {
  return value.split(",").map(item => item.trim()).filter(Boolean);
}

function pretty(value: string) {
  return value.replaceAll("_", " ").replace(/\b\w/g, c => c.toUpperCase());
}

function money(value: unknown) {
  return new Intl.NumberFormat("en-IN", { style: "currency", currency: "INR", maximumFractionDigits: 0 }).format(number(value, 0));
}

const inputClass = "w-full rounded-xl border border-[#d0d5dd] bg-white px-3 py-2.5 text-sm text-[#344054] outline-none transition placeholder:text-[#98a2b3] focus:border-[#14b8a6] focus:ring-4 focus:ring-[#ccfbf1]";
