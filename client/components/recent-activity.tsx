"use client";

import { useEffect, useState } from "react";
import { format } from "date-fns";
import { Mail, MousePointerClick, Eye, AlertTriangle } from "lucide-react";
import { cn } from "@/lib/utils";
import { useTeam } from "@/app/providers/team-provider";

interface EmailEvent {
  eventId: string;
  eventType: "OPENED" | "CLICKED" | "BOUNCED" | "FAILED";
  timestamp: string;
  campaignName: string;
  emailSubject: string;
  recipient: string;
  deviceInfo?: {
    type: string;
    browser: string;
  };
  location?: {
    country: string;
    city: string;
  };
}

export function RecentActivity() {
  const [events, setEvents] = useState<EmailEvent[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const { team } = useTeam();
  useEffect(() => {
    const fetchEvents = async () => {
      try {
        const response = await fetch(
          `/api/analytics/team/overview?teamId=${team?.id}`
        );
        if (!response.ok) {
          throw new Error("Failed to fetch team overview");
        }
        const { data } = await response.json();

        console.log(data, "data recent activity pain");

        // Convert tracking events from team overview to email events
        const recentEvents = (data.recentActivity || [])
          .slice(0, 10)
          .map((activity: any) => ({
            eventId: activity.id,
            eventType: activity.type.toUpperCase(),
            timestamp: activity.timestamp,
            campaignName: activity.campaignName || "",
            emailSubject: activity.emailSubject || "Email Campaign",
            recipient: activity.recipient,
            deviceInfo: activity.deviceInfo,
            location: activity.location,
          }));

        setEvents(recentEvents);
      } catch (error) {
        console.error("Failed to fetch events:", error);
      } finally {
        setIsLoading(false);
      }
    };

    if (team?.id) {
      fetchEvents();
    }
  }, [team?.id]);

  const getEventIcon = (type: EmailEvent["eventType"]) => {
    switch (type) {
      case "OPENED":
        return <Eye className="h-4 w-4" />;
      case "CLICKED":
        return <MousePointerClick className="h-4 w-4" />;
      case "BOUNCED":
      case "FAILED":
        return <AlertTriangle className="h-4 w-4" />;
      default:
        return <Mail className="h-4 w-4" />;
    }
  };

  const getEventColor = (type: EmailEvent["eventType"]) => {
    switch (type) {
      case "OPENED":
        return "text-blue-500";
      case "CLICKED":
        return "text-green-500";
      case "BOUNCED":
      case "FAILED":
        return "text-red-500";
      default:
        return "text-gray-500";
    }
  };

  if (isLoading) {
    return <div>Loading...</div>;
  }

  return (
    <div className="space-y-8">
      {events.map((event) => (
        <div key={event.eventId} className="flex items-start space-x-4">
          <div className={cn("mt-1", getEventColor(event.eventType))}>
            {getEventIcon(event.eventType)}
          </div>
          <div className="space-y-1">
            <p className="text-sm font-medium leading-none">
              {event.emailSubject}
            </p>
            <div className="flex items-center gap-2">
              <p className="text-sm text-muted-foreground">{event.recipient}</p>
              {event.deviceInfo && (
                <span className="text-xs text-muted-foreground">
                  via {event.deviceInfo.type} ({event.deviceInfo.browser})
                </span>
              )}
            </div>
            <div className="flex items-center gap-2">
              <p className="text-xs text-muted-foreground">
                {format(new Date(event.timestamp), "MMM d, yyyy 'at' h:mm a")}
              </p>
              {event.location && (
                <span className="text-xs text-muted-foreground">
                  from {event.location.city}, {event.location.country}
                </span>
              )}
            </div>
          </div>
        </div>
      ))}
      {events.length === 0 && (
        <div className="text-center text-sm text-muted-foreground">
          No recent activity
        </div>
      )}
    </div>
  );
}
