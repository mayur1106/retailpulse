"use client";
import { ResourcePage, formatDate } from "@/components/resource-page";
const score=(value:unknown)=>Number(value??0).toFixed(2);
export default function RecommendationsPage(){return <ResourcePage resource="recommendations" title="Recommendations" description="Prioritized seller actions with evidence-backed reasons" columns={[{key:"recommendation_type",label:"Type"},{key:"title",label:"Recommendation"},{key:"product",label:"Product"},{key:"campaign",label:"Campaign"},{key:"reason",label:"Reason"},{key:"impact_score",label:"Impact",format:score},{key:"confidence",label:"Confidence",format:score},{key:"status",label:"Status"},{key:"data_origin",label:"Origin"},{key:"created_at",label:"Created",format:formatDate}]}/>}
