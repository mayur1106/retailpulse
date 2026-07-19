"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { BadgePercent, BarChart3, Boxes, ClipboardCheck, Eye, FileText, Globe2, LayoutDashboard, LayoutTemplate, Lightbulb, LogOut, Megaphone, Menu, Network, PackageSearch, ReceiptText, Search, ShoppingCart, Store, TrendingUp, Truck, Users, WalletCards, X } from "lucide-react";
import { useState } from "react";
import { logout } from "@/lib/api";

const navigation = [
  { href: "/dashboard", label: "Overview", icon: LayoutDashboard },
  { href: "/dashboard/products", label: "Products", icon: PackageSearch },
  { href: "/dashboard/inventory", label: "SKU Inventory", icon: Boxes },
  { href: "/dashboard/orders", label: "Orders", icon: ShoppingCart },
  { href: "/dashboard/customers", label: "Customers", icon: Users },
  { href: "/dashboard/channels", label: "Channels", icon: Network },
  { href: "/dashboard/storefront", label: "Storefront CMS", icon: LayoutTemplate },
  { href: "/dashboard/store", label: "Store settings", icon: Store },
  { href: "/dashboard/coupons", label: "Coupons", icon: BadgePercent },
  { href: "/dashboard/shipping", label: "Shipping", icon: Truck },
  { href: "/dashboard/actions", label: "Today’s actions", icon: ClipboardCheck },
  { href: "/dashboard/growth", label: "Growth insights", icon: TrendingUp },
  { href: "/dashboard/decisions", label: "Product decisions", icon: PackageSearch },
  { href: "/dashboard/recommendations", label: "Recommendations", icon: Lightbulb },
  { href: "/dashboard/traffic", label: "Amazon traffic", icon: Eye },
  { href: "/dashboard/search", label: "Amazon search", icon: Search },
  { href: "/dashboard/regions", label: "Regions", icon: Globe2 },
  { href: "/dashboard/returns", label: "Returns", icon: ReceiptText },
  { href: "/dashboard/profit", label: "Profit", icon: WalletCards },
  { href: "/dashboard/ads-optimization", label: "Ads optimization", icon: Megaphone },
  { href: "/dashboard/ads", label: "Ads intelligence", icon: Megaphone },
  { href: "/dashboard/campaigns", label: "Ad campaigns", icon: Megaphone },
  { href: "/dashboard/reports", label: "Reports", icon: FileText },
];

export function DashboardShell({ title, description, action, children }: { title: string; description: string; action?: React.ReactNode; children: React.ReactNode }) {
  const pathname = usePathname();
  const [open, setOpen] = useState(false);
  const signOut = async () => { await logout(); window.location.replace("/login"); };
  return <div className="min-h-screen bg-[#f6f8fb] text-[#17202a]">
    {open ? <button aria-label="Close navigation" className="fixed inset-0 z-30 bg-black/40 lg:hidden" onClick={() => setOpen(false)} /> : null}
    <aside className={`fixed inset-y-0 left-0 z-40 flex w-64 flex-col border-r border-[#0f172a] bg-[#0f172a] text-white shadow-2xl shadow-slate-950/20 transition-transform lg:translate-x-0 ${open ? "translate-x-0" : "-translate-x-full"}`}>
      <div className="flex h-18 items-center justify-between border-b border-white/10 px-5"><Link href="/dashboard" className="flex items-center gap-3"><span className="grid h-9 w-9 place-items-center rounded-xl bg-[#14b8a6] shadow-lg shadow-teal-500/20"><BarChart3 className="h-5 w-5" /></span><span><span className="block text-[15px] font-semibold tracking-tight">RetailPulse</span><span className="block text-[10px] uppercase tracking-[0.18em] text-slate-400">Commerce OS</span></span></Link><button className="lg:hidden" onClick={() => setOpen(false)}><X className="h-5 w-5" /></button></div>
      <nav className="flex-1 space-y-1 overflow-y-auto p-3"><p className="px-3 pb-2 pt-3 text-[10px] font-semibold uppercase tracking-[0.18em] text-slate-500">Workspace</p>{navigation.map((item) => { const active=item.href==="/dashboard"?pathname===item.href:pathname.startsWith(item.href);const Icon=item.icon;return <Link key={item.href} href={item.href} onClick={()=>setOpen(false)} className={`flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-medium transition ${active?"bg-[#14b8a6] text-white shadow-sm shadow-teal-950/20":"text-slate-300 hover:bg-white/10 hover:text-white"}`}><Icon className="h-4 w-4" />{item.label}</Link>})}</nav>
      <div className="border-t border-white/10 p-3"><div className="mb-3 flex items-center gap-3 rounded-xl bg-white/8 p-3"><span className="grid h-8 w-8 place-items-center rounded-full bg-[#334155] text-xs font-semibold">RP</span><div className="min-w-0"><p className="truncate text-sm font-medium">Workspace owner</p><p className="truncate text-xs text-slate-400">Multi-channel seller</p></div></div><button onClick={signOut} className="flex w-full items-center gap-3 rounded-xl px-3 py-2 text-sm text-slate-300 hover:bg-white/10 hover:text-white"><LogOut className="h-4 w-4" />Sign out</button></div>
    </aside>
    <div className="lg:pl-64"><header className="sticky top-0 z-20 flex h-18 items-center justify-between border-b border-[#e4e7ec] bg-white/95 px-4 shadow-sm shadow-slate-200/50 backdrop-blur sm:px-7"><div className="flex items-center gap-3"><button className="rounded-xl border border-[#d0d5dd] bg-white p-2 text-[#344054] lg:hidden" onClick={()=>setOpen(true)}><Menu className="h-5 w-5" /></button><div><h1 className="text-lg font-semibold tracking-tight text-[#101828] sm:text-xl">{title}</h1><p className="hidden text-xs font-medium text-[#667085] sm:block">{description}</p></div></div><div className="flex items-center gap-3">{action}<span className="hidden items-center gap-2 rounded-full bg-[#ecfdf5] px-3 py-1.5 text-xs font-semibold text-[#047857] ring-1 ring-[#bbf7d0] sm:flex"><span className="h-2 w-2 rounded-full bg-[#10b981]" />System online</span></div></header><main className="p-4 sm:p-7">{children}</main></div>
  </div>;
}

export function PageCard({ children, className="" }: { children: React.ReactNode; className?: string }) { return <section className={`rounded-2xl border border-[#e4e7ec] bg-white shadow-sm shadow-slate-200/70 ${className}`}>{children}</section> }
