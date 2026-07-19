"use client";
import { CommerceResourcePage } from "@/components/commerce-resource-page";
export default function CommerceAnalyticsPage(){return <CommerceResourcePage resource="analytics" title="Commerce analytics" description="Revenue, product performance, and owned-store signals" columns={[{key:"kind",label:"Type"},{key:"label",label:"Metric"},{key:"value",label:"Value"},{key:"secondary",label:"Secondary"}]}/>;}
