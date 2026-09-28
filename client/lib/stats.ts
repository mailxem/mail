export interface MetricsData {
  total: number;
  openRate: number;
  clickRate: number;
  bounceRate: number;
  date?: string;
}

export interface TeamMetrics {
  totalEmails: number;
  totalOpens: number;
  totalClicks: number;
  averageOpenRate: number;
  averageClickRate: number;
  bounceRate: number;
  monthlyStats: Array<{
    month: string;
    totalEmails: number;
    openRate: number;
    clickRate: number;
  }>;
}
