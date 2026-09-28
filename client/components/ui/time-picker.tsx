// components/ui/time-picker.tsx
import * as React from "react";
import { cn } from "@/lib/utils";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "./select";

interface TimePickerProps {
  value?: string;
  onChange?: (time: string) => void;
  className?: string;
  disabled?: boolean;
}

const TimePicker = React.forwardRef<HTMLDivElement, TimePickerProps>(
  ({ value, onChange, className, disabled }, ref) => {
    // Parse initial value in 24h format to 12h format
    const parseInitialValue = () => {
      if (!value) return { hours: "12", minutes: "00", period: "AM" };

      const [hoursStr, minutes] = value.split(":");
      const hours = parseInt(hoursStr);
      
      if (hours === 0) return { hours: "12", minutes, period: "AM" };
      if (hours === 12) return { hours: "12", minutes, period: "PM" };
      if (hours > 12) return { hours: (hours - 12).toString().padStart(2, "0"), minutes, period: "PM" };
      return { hours: hours.toString().padStart(2, "0"), minutes, period: "AM" };
    };

    const { hours: initialHours, minutes: initialMinutes, period: initialPeriod } = parseInitialValue();

    const [hours, setHours] = React.useState<string>(initialHours);
    const [minutes, setMinutes] = React.useState<string>(initialMinutes);
    const [period, setPeriod] = React.useState<"AM" | "PM">(
      initialPeriod as "AM" | "PM"
    );

    const hoursArray = Array.from({ length: 12 }, (_, i) =>
      (i + 1).toString().padStart(2, "0")
    );
    const minutesArray = Array.from({ length: 60 }, (_, i) =>
      i.toString().padStart(2, "0")
    );

    const handleTimeChange = (
      type: "hours" | "minutes" | "period",
      newValue: string
    ) => {
      switch (type) {
        case "hours":
          setHours(newValue);
          break;
        case "minutes":
          setMinutes(newValue);
          break;
        case "period":
          setPeriod(newValue as "AM" | "PM");
          break;
      }

      if (onChange) {
        let formattedHours: number;
        const hoursNum = parseInt(type === "hours" ? newValue : hours);
        const newPeriod = type === "period" ? newValue : period;

        if (newPeriod === "AM") {
          formattedHours = hoursNum === 12 ? 0 : hoursNum;
        } else {
          formattedHours = hoursNum === 12 ? 12 : hoursNum + 12;
        }

        onChange(`${formattedHours.toString().padStart(2, "0")}:${type === "minutes" ? newValue : minutes}`);
      }
    };

    return (
      <div
        ref={ref}
        className={cn("flex justify-center items-center gap-1", className)}
      >
        <Select
          value={hours}
          onValueChange={(value) => handleTimeChange("hours", value)}
          disabled={disabled}
        >
          <SelectTrigger className="w-[70px]">
            <SelectValue placeholder="HH" />
          </SelectTrigger>
          <SelectContent>
            {hoursArray.map((hour) => (
              <SelectItem key={hour} value={hour}>
                {hour}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        <span className="text-slate-500">:</span>

        <Select
          value={minutes}
          onValueChange={(value) => handleTimeChange("minutes", value)}
          disabled={disabled}
        >
          <SelectTrigger className="w-[70px]">
            <SelectValue placeholder="MM" />
          </SelectTrigger>
          <SelectContent>
            {minutesArray.map((minute) => (
              <SelectItem key={minute} value={minute}>
                {minute}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        <Select
          value={period}
          onValueChange={(value) => handleTimeChange("period", value)}
          disabled={disabled}
        >
          <SelectTrigger className="w-[70px]">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="AM">AM</SelectItem>
            <SelectItem value="PM">PM</SelectItem>
          </SelectContent>
        </Select>
      </div>
    );
  }
);

TimePicker.displayName = "TimePicker";

export { TimePicker };