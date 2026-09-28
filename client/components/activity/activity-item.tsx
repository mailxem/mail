import { LucideIcon } from "lucide-react";

interface ActivityItemProps {
  icon: LucideIcon;
  description: string;
  time: string;
}

export function ActivityItem({
  icon: Icon,
  description,
  time,
}: ActivityItemProps) {
  return (
    <div className="flex items-center gap-4">
      <div className=" p-2 bg-primary/10">
        <Icon className="h-4 w-4 text-primary" />
      </div>
      <div>
        <p className="text-sm">{description}</p>
        <p className="text-xs text-muted-foreground">{time}</p>
      </div>
    </div>
  );
}
