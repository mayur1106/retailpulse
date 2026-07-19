"use client";
import { ResourcePage, formatMoney } from "@/components/resource-page";
const pct=(value:unknown)=>`${Number(value??0).toFixed(2)}%`;
export default function SearchTermsPage(){return <ResourcePage resource="search-terms" title="Search terms" description="Search, keyword, and product-level demand signals" columns={[{key:"search_term",label:"Search term"},{key:"product",label:"Product"},{key:"campaign",label:"Campaign"},{key:"channel",label:"Channel"},{key:"match_type",label:"Match"},{key:"impressions",label:"Impressions"},{key:"clicks",label:"Clicks"},{key:"orders",label:"Orders"},{key:"conversion_rate",label:"CVR",format:pct},{key:"spend",label:"Spend",format:formatMoney},{key:"sales",label:"Sales",format:formatMoney},{key:"roas",label:"ROAS"}]}/>}
