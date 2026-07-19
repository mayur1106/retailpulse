"use client";
import { CommerceResourcePage, commerceDate } from "@/components/commerce-resource-page";
export default function StorePage(){return <CommerceResourcePage resource="store" title="Store settings" description="Owned storefront identity, currency, domain, and status" columns={[{key:"name",label:"Store"},{key:"slug",label:"Slug"},{key:"domain",label:"Domain"},{key:"currency_code",label:"Currency"},{key:"country_code",label:"Country"},{key:"timezone",label:"Timezone"},{key:"status",label:"Status"},{key:"updated_at",label:"Updated",format:commerceDate}]}/>;}
