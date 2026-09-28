import { MetricsData } from "@/lib/stats";
import { format } from "date-fns";
import Papa from "papaparse";

export interface ReportOptions {
  startDate?: Date;
  endDate?: Date;
  campaignId?: string;
  teamId: string;
  format: "csv" | "json";
  groupBy?: "day" | "week" | "month" | "campaign";
}

export interface ReportMetrics extends MetricsData {
  campaignName?: string;
  period?: string;
}

export async function generateReport(
  options: ReportOptions
): Promise<string | Blob> {
  const queryParams = new URLSearchParams();

  if (options.startDate) {
    queryParams.append("startDate", options.startDate.toISOString());
  }
  if (options.endDate) {
    queryParams.append("endDate", options.endDate.toISOString());
  }
  if (options.campaignId) {
    queryParams.append("campaignId", options.campaignId);
  }
  if (options.groupBy) {
    queryParams.append("groupBy", options.groupBy);
  }
  queryParams.append("teamId", options.teamId);

  const response = await fetch(
    `/api/email/track/metrics?${queryParams.toString()}`
  );
  const data: ReportMetrics[] = await response.json();

  // Process the metrics data
  const processedData = data.map((metric) => ({
    Date: format(new Date(metric.date), "yyyy-MM-dd"),
    Period: metric.period || "N/A",
    Campaign: metric.campaignName || "All Campaigns",
    "Total Sent": metric.total,
    Opens: metric.openRate,
    "Open Rate": `${(metric.openRate * 100).toFixed(1)}%`,
    Clicks: metric.clickRate,
    "Click Rate": `${(metric.clickRate * 100).toFixed(1)}%`,
    Bounces: metric.bounceRate,
    "Bounce Rate": `${(metric.bounceRate * 100).toFixed(1)}%`,
  }));

  if (options.format === "csv") {
    const csv = Papa.unparse(processedData);
    const blob = new Blob([csv], { type: "text/csv;charset=utf-8;" });
    return blob;
  }

  return JSON.stringify(processedData, null, 2);
}

export async function downloadReport(options: ReportOptions): Promise<void> {
  const report = await generateReport(options);

  if (report instanceof Blob) {
    // For CSV files
    const url = window.URL.createObjectURL(report);
    const link = document.createElement("a");
    const timestamp = format(new Date(), "yyyy-MM-dd");

    link.href = url;
    link.setAttribute(
      "download",
      `email-report-${timestamp}${options.campaignId ? `-campaign-${options.campaignId}` : ""}.csv`
    );
    document.body.appendChild(link);
    link.click();
    link.remove();
  } else {
    // For JSON data
    const blob = new Blob([report], { type: "application/json" });
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement("a");
    const timestamp = format(new Date(), "yyyy-MM-dd");

    link.href = url;
    link.setAttribute(
      "download",
      `email-report-${timestamp}${options.campaignId ? `-campaign-${options.campaignId}` : ""}.json`
    );
    document.body.appendChild(link);
    link.click();
    link.remove();
  }
}
