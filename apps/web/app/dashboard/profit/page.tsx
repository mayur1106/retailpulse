"use client";
import { ResourcePage, formatMoney } from "@/components/resource-page";
const pct=(value:unknown)=>`${Number(value??0).toFixed(2)}%`;
export default function ProfitPage(){return <ResourcePage resource="profit" title="Profit" description="Estimated product profit after COGS, ads, refunds, and fees" columns={[{key:"title",label:"Product"},{key:"category",label:"Category"},{key:"units",label:"Units"},{key:"revenue",label:"Revenue",format:formatMoney},{key:"cogs",label:"COGS",format:formatMoney},{key:"ad_spend",label:"Ads",format:formatMoney},{key:"refunds",label:"Refunds",format:formatMoney},{key:"estimated_fees",label:"Fees",format:formatMoney},{key:"estimated_profit",label:"Profit",format:formatMoney},{key:"margin",label:"Margin",format:pct}]}/>}
