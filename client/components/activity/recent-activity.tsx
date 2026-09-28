import { Card } from "../ui/card";
import { Mail, MessageSquare, UserPlus } from "lucide-react";
import { ActivityItem } from "./activity-item";

const activities = [
  {
    icon: Mail,
    description: "Campaign 'Welcome Series' sent to 150 contacts",
    time: "2 hours ago",
  },
  {
    icon: UserPlus,
    description: "New contact John Doe added",
    time: "4 hours ago",
  },
  {
    icon: MessageSquare,
    description: "New conversation started with Sarah Smith",
    time: "5 hours ago",
  },
];

export function RecentActivity() {
  return (
    <Card className="p-6">
      <h2 className="text-lg font-medium mb-4">Recent Activity</h2>
      <div className="space-y-4">
        {activities.map((activity, index) => (
          <ActivityItem key={index} {...activity} />
        ))}
      </div>
    </Card>
  );
}