"use client";
import { CommerceResourcePage, commerceDate, commerceMoney } from "@/components/commerce-resource-page";
export default function CommerceReturnsPage(){return <CommerceResourcePage resource="returns" title="Commerce returns" description="Return/refund requests, reasons, and product quality signals" columns={[{key:"order_number",label:"Order"},{key:"product",label:"Product"},{key:"reason",label:"Reason"},{key:"refund_amount",label:"Refund",format:commerceMoney},{key:"requested_at",label:"Requested",format:commerceDate},{key:"status",label:"Status"}]}/>;}
