import { APIService } from "./api";

export interface EmailAnalytics {
  openCount: number;
  clickCount: number;
  bounceCount: number;
  complaintCount: number;
  openRate: number;
  clickRate: number;
  averageReadTime: number;
  uniqueOpens: number;
  uniqueClicks: number;
  repeatOpens: number;
  repeatClicks: number;
  geoBreakdown: Record<string, number>;
  deviceBreakdown: Record<string, number>;
  browserBreakdown: Record<string, number>;
  osBreakdown: Record<string, number>;
  cityBreakdown: Record<string, number>;
  regionBreakdown: Record<string, number>;
  dayOfWeekBreakdown: Record<string, number>;
  hourlyBreakdown: Record<string, number>;
  industryAvgOpenRate: number;
  industryAvgClickRate: number;
  timelineData: TimelineDataPoint[];
  clickedLinks: LinkAnalytics[];
}

export interface TimelineDataPoint {
  timestamp: string;
  eventType: string;
  count: number;
  uniqueCount: number;
  browser?: string;
  deviceType?: string;
  country?: string;
}

export interface LinkAnalytics {
  url: string;
  clickCount: number;
  uniqueClicks: number;
  clickRate: number;
  deviceBreakdown: Record<string, number>;
  averageTimeToClick: number;
  firstClickTime: string;
  lastClickTime: string;
}

export interface AudienceInsights {
  demographics: Record<string, number>;
  behaviors: Record<string, BehaviorMetrics>;
  preferences: Record<string, PreferenceMetrics>;
  segments: AudienceSegment[];
}

export interface BehaviorMetrics {
  count: number;
  engagementRate: number;
  trend: string;
}

export interface PreferenceMetrics {
  count: number;
  score: number;
  confidence: number;
}

export interface AudienceSegment {
  name: string;
  size: number;
  openRate: number;
  clickRate: number;
  growth: number;
}

export interface TeamOverview {
  totalEmails: number;
  totalOpens: number;
  totalClicks: number;
  averageOpenRate: number;
  averageClickRate: number;
  deviceStats: Record<string, number>;
  geoStats: Record<string, number>;
  monthlyStats: MonthlyEngagement[];
  topCampaigns: CampaignSummary[];
  topPerformers: PerformerMetrics[];
}

export interface MonthlyEngagement {
  month: string;
  totalEmails: number;
  openRate: number;
  clickRate: number;
}

export interface CampaignSummary {
  campaignId: string;
  name: string;
  openRate: number;
  clickRate: number;
  engagementScore: number;
}

export interface PerformerMetrics {
  type: string;
  value: string;
  engagementRate: number;
  sampleSize: number;
}

export interface HeatmapData {
  url: string;
  clicks: ClickPoint[];
  segments: ClickSegment[];
}

export interface ClickPoint {
  x: number;
  y: number;
  count: number;
  element: string;
}

export interface ClickSegment {
  selector: string;
  count: number;
  clickRate: number;
}

export class AnalyticsService extends APIService {
  constructor(session?: any) {
    super("api/v1/analytics", session);
  }

  async getAudienceInsights(teamId: string): Promise<AudienceInsights> {
    return this.get<AudienceInsights>("audience", { teamId });
  }

  async getCampaignAnalytics(campaignId: string): Promise<EmailAnalytics> {
    return this.get<EmailAnalytics>("campaign", { campaignId });
  }

  async compareCampaigns(campaignIds: string[]): Promise<Record<string, EmailAnalytics>> {
    return this.get<Record<string, EmailAnalytics>>("campaign/compare", {
      campaignIds: campaignIds.join(","),
    });
  }

  async getEmailAnalytics(emailId: string): Promise<EmailAnalytics> {
    return this.get<EmailAnalytics>("email", { emailId });
  }

  async exportCampaignAnalytics(campaignId: string): Promise<Uint8Array> {
    return this.get<Uint8Array>("export/campaign", { campaignId });
  }

  async exportEmailAnalytics(emailId: string): Promise<Uint8Array> {
    return this.get<Uint8Array>("export/email", { emailId });
  }

  async getClickHeatmap(emailId: string): Promise<HeatmapData> {
    return this.get<HeatmapData>("heatmap", { emailId });
  }

  async getTeamOverview(teamId: string): Promise<TeamOverview> {
    return this.get<TeamOverview>("team/overview", { teamId });
  }

  async getTrendAnalysis(teamId: string): Promise<TimelineDataPoint[]> {
    return this.get<TimelineDataPoint[]>("trends", { teamId });
  }

  async getCampaignsTimeline(teamId: string, startDate: string): Promise<{
    timelineData: Array<{
      timestamp: string;
      metrics: {
        openRate: number;
        clickRate: number;
        bounceRate: number;
        total: number;
      };
    }>;
  }> {
    return this.get("campaigns/timeline", { teamId, startDate });
  }
}
