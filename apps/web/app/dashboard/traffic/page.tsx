"use client";
import { ResourcePage, formatMoney } from "@/components/resource-page";
const pct=(value:unknown)=>`${Number(value??0).toFixed(2)}%`;
export default function TrafficPage(){return <ResourcePage resource="traffic" title="Traffic" description="Sessions, page views, conversion, and ordered revenue by product" columns={[{key:"title",label:"Product"},{key:"category",label:"Category"},{key:"sessions",label:"Sessions"},{key:"page_views",label:"Page views"},{key:"units",label:"Units"},{key:"conversion_rate",label:"CVR",format:pct},{key:"buy_box_percentage",label:"Buy box",format:pct},{key:"revenue",label:"Revenue",format:formatMoney},{key:"data_origin",label:"Origin"}]}/>}
