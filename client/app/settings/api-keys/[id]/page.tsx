"use client";

import { useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, Download, Activity, CheckCircle2, CircleAlert } from "lucide-react";
import { PageHeader } from "@/components/page-header";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { QueryState, Empty, Metric } from "@/components/marketing/shared";
import { CollectionPagination } from "@/components/ui/collection-pagination";
import { useResourcePage } from "@/hooks/use-resource-page";
import { workspaceClassName } from "@/lib/workspace-styles";

interface Usage { id:string; endpoint:string; method:string; timestamp:string; success:boolean; error?:string; ipAddress?:string; }
export default function APIKeyDetailsPage() {
 const { id } = useParams<{id:string}>();
 const [page,setPage] = useState(1);
 const [period,setPeriod] = useState("24h");
 const query = useResourcePage<Usage>("api-key-usage",page,50,{api_key_id:id,period});
 const rows = query.data?.data || [];
 const successes = rows.filter(row=>row.success).length;
 const exportLogs = () => {
  const url = URL.createObjectURL(new Blob([JSON.stringify(rows,null,2)],{type:"application/json"}));
  const link = document.createElement("a"); link.href=url; link.download=`xem-api-usage-page-${page}.json`; link.click(); URL.revokeObjectURL(url);
 };
 return <div className={workspaceClassName("workspace-page")}>
  <PageHeader heading="API key usage" description="Requests made with this key, with their recorded delivery status.">
   <Button variant="outline" asChild><Link href="/settings/api-keys"><ArrowLeft size={16}/>API keys</Link></Button>
   <Button variant="outline" disabled={!rows.length || query.isLoading} onClick={exportLogs}><Download size={16}/>Export this page</Button>
  </PageHeader>
  <div className="grid gap-4 sm:grid-cols-3">
   <Metric icon={<Activity size={18}/>} label="Requests in period" value={query.isLoading ? "—" : query.data?.total ?? 0}/>
   <Metric icon={<CheckCircle2 size={18}/>} label="Successful on this page" value={query.isLoading ? "—" : successes}/>
   <Metric icon={<CircleAlert size={18}/>} label="Failed on this page" value={query.isLoading ? "—" : rows.length-successes}/>
  </div>
  <section className={workspaceClassName("product-panel")}>
   <div className={workspaceClassName("panel-toolbar")}><h2>Activity log</h2><Select value={period} onValueChange={value=>{setPeriod(value);setPage(1);}}><SelectTrigger className="w-44"><SelectValue/></SelectTrigger><SelectContent><SelectItem value="1h">Last hour</SelectItem><SelectItem value="24h">Last 24 hours</SelectItem><SelectItem value="7d">Last 7 days</SelectItem><SelectItem value="30d">Last 30 days</SelectItem></SelectContent></Select></div>
   {query.isLoading || query.error ? <QueryState loading={query.isLoading} error={query.error} retry={()=>void query.refetch()}/> : !rows.length ? <Empty title="No requests in this period" description="API calls made with this key will appear here."/> : <Table><TableHeader><TableRow><TableHead>Time</TableHead><TableHead>Method</TableHead><TableHead>Endpoint</TableHead><TableHead>Status</TableHead><TableHead>IP address</TableHead></TableRow></TableHeader><TableBody>{rows.map(row=><TableRow key={row.id}><TableCell className="whitespace-nowrap">{new Date(row.timestamp).toLocaleString()}</TableCell><TableCell><Badge variant="outline">{row.method}</Badge></TableCell><TableCell className="max-w-xs break-all">{row.endpoint}</TableCell><TableCell><Badge variant={row.success ? "success":"destructive"}>{row.success ? "Success":"Failed"}</Badge>{row.error && <p className="mt-1 text-xs text-muted-foreground">{row.error}</p>}</TableCell><TableCell>{row.ipAddress || "—"}</TableCell></TableRow>)}</TableBody></Table>}
   <CollectionPagination page={page} limit={50} total={query.data?.total || 0} onPageChange={setPage}/>
  </section>
 </div>;
}
