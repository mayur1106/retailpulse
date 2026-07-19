"use client";

import { useQuery } from "@tanstack/react-query";
import { Search } from "lucide-react";
import { useEffect, useState } from "react";
import { authedRequest } from "@/lib/api";
import { DashboardShell, PageCard } from "@/components/dashboard-shell";

type Column = { key: string; label: string; format?: (value: unknown, row: Record<string, unknown>) => string };
type AmazonStore = { id: string; name: string; environment: "production" | "sandbox" };
const money = new Intl.NumberFormat("en-US", { style: "currency", currency: "USD" });
export const formatMoney=(value:unknown)=>money.format(Number(value??0));
export const formatDate=(value:unknown)=>value?new Date(String(value)).toLocaleString():"—";

export function ResourcePage({ resource,title,description,columns }: { resource:string;title:string;description:string;columns:Column[] }) {
  const [search,setSearch]=useState("");
  const [selectedStoreId,setSelectedStoreId]=useState("");
  const storesQuery=useQuery({queryKey:["amazon-stores"],queryFn:()=>authedRequest<{stores:AmazonStore[]|null}>("/v1/amazon/stores"),retry:false});
  const stores=storesQuery.data?.stores??[];
  useEffect(()=>{if(stores.length===0)return;if(selectedStoreId&&stores.some(store=>store.id===selectedStoreId))return;setSelectedStoreId(stores.find(store=>store.environment==="production")?.id??stores[0].id)},[stores,selectedStoreId]);
  const selectedStore=stores.find(store=>store.id===selectedStoreId);
  const query=useQuery({queryKey:["resource",resource,selectedStoreId],queryFn:()=>authedRequest<{items:Record<string,unknown>[]}>(`/v1/analytics/data/${resource}?storeId=${selectedStoreId}`),enabled:Boolean(selectedStoreId),retry:false});
  const items=(query.data?.items??[]).filter(item=>Object.values(item).some(value=>String(value??"").toLowerCase().includes(search.toLowerCase())));
  return <DashboardShell title={title} description={description}><div className="mb-5 grid gap-4 sm:grid-cols-3"><PageCard className="p-5"><p className="text-xs font-semibold uppercase tracking-wide text-[#667085]">Total records</p><p className="mt-2 text-3xl font-semibold tracking-tight text-[#101828]">{query.data?.items.length??0}</p>{selectedStore?<p className="mt-2 text-xs text-[#667085]">{selectedStore.name} · {selectedStore.environment}</p>:null}</PageCard><PageCard className="p-5 sm:col-span-2"><div className="grid gap-3 md:grid-cols-[220px_1fr]"><select value={selectedStoreId} onChange={event=>setSelectedStoreId(event.target.value)} className="rounded-xl border border-[#d0d5dd] bg-white px-3 py-2.5 text-sm text-[#344054] outline-none transition focus:border-[#14b8a6] focus:ring-4 focus:ring-[#ccfbf1]">{stores.map(store=><option key={store.id} value={store.id}>{store.name} ({store.environment})</option>)}</select><label className="relative block"><Search className="absolute left-3 top-3 h-4 w-4 text-[#98a2b3]"/><input value={search} onChange={event=>setSearch(event.target.value)} placeholder={`Search ${title.toLowerCase()}...`} className="w-full rounded-xl border border-[#d0d5dd] bg-white py-2.5 pl-10 pr-3 text-sm text-[#344054] outline-none transition placeholder:text-[#98a2b3] focus:border-[#14b8a6] focus:ring-4 focus:ring-[#ccfbf1]"/></label></div><p className="mt-2 text-xs text-[#667085]">Showing {items.length} matching records</p></PageCard></div><PageCard className="overflow-hidden"><div className="overflow-x-auto"><table className="w-full min-w-[760px] text-sm"><thead className="bg-[#f8fafc] text-left text-xs uppercase tracking-wide text-[#667085]"><tr>{columns.map(column=><th key={column.key} className="px-5 py-3.5 font-semibold">{column.label}</th>)}</tr></thead><tbody>{items.map((row,index)=><tr key={String(row.id??row.asin??index)} className="border-t border-[#eaecf0] text-[#344054] hover:bg-[#f8fafc]">{columns.map(column=><td key={column.key} className="max-w-xs px-5 py-4"><span className={column.key==="status"||column.key==="order_status"?"rounded-full bg-[#ecfdf5] px-2.5 py-1 text-xs font-semibold text-[#047857]":""}>{column.format?column.format(row[column.key],row):String(row[column.key]??"—")}</span></td>)}</tr>)}</tbody></table>{query.isLoading||storesQuery.isLoading?<div className="p-8 text-center text-sm text-[#667085]">Loading data...</div>:null}{!query.isLoading&&!storesQuery.isLoading&&items.length===0?<div className="p-10 text-center text-sm text-[#667085]">No records found.</div>:null}{query.error?<div className="p-8 text-center text-sm text-[#b42318]">{query.error.message}</div>:null}</div></PageCard></DashboardShell>
}
