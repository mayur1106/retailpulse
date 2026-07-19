"use client";
import { ResourcePage,formatDate } from "@/components/resource-page";
export default function ReportsPage(){return <ResourcePage resource="reports" title="Reports" description="Generated and imported Amazon reports" columns={[{key:"report_type",label:"Report type"},{key:"format",label:"Format"},{key:"status",label:"Status"},{key:"storage_key",label:"Storage key"},{key:"data_origin",label:"Origin"},{key:"created_at",label:"Created",format:formatDate}]}/>}
