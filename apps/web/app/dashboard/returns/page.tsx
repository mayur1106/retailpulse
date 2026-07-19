"use client";
import { ResourcePage, formatDate, formatMoney } from "@/components/resource-page";
export default function ReturnsPage(){return <ResourcePage resource="returns" title="Returns" description="Returned products, refund impact, and reason analysis" columns={[{key:"return_date",label:"Returned",format:formatDate},{key:"product",label:"Product"},{key:"category",label:"Category"},{key:"reason",label:"Reason"},{key:"status",label:"Status"},{key:"quantity",label:"Qty"},{key:"refund_amount",label:"Refund",format:formatMoney},{key:"data_origin",label:"Origin"}]}/>}
