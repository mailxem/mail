import { Users, Mail, MessageSquare, TrendingUp } from "lucide-react";
import { StatCard } from "./stat-card";

const stats = [
  {
    name: "Total Contacts",
    value: "8",
    icon: Users,
    change: "+2.5%",
  },
  {
    name: "Email Sent",
    value: "300",
    icon: Mail,
    change: "+12%",
  },
  {
    name: "Conversations",
    value: "24",
    icon: MessageSquare,
    change: "+5%",
  },
  {
    name: "Conversion Rate",
    value: "2.4%",
    icon: TrendingUp,
    change: "+1.2%",
  },
];

export function Stats() {
  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
      {stats.map((stat) => (
        <StatCard key={stat.name} {...stat} />
      ))}
    </div>
  );
}