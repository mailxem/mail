export type ClickHour = { day: number; hour: number; count: number };

const DAYS = [
  "Monday",
  "Tuesday",
  "Wednesday",
  "Thursday",
  "Friday",
  "Saturday",
  "Sunday",
];

/** Below this many clicks a "best" hour is mostly noise, so we don't suggest one. */
export const MIN_CLICKS = 50;

/**
 * The weekday (0 = Monday, as the report returns it) and hour with the most
 * clicks, smoothed over the neighbouring hours so one lucky click doesn't win.
 */
export function peakHour(hours: ClickHour[]) {
  const grid = Array.from({ length: 7 }, () => Array(24).fill(0) as number[]);
  let total = 0;
  for (const h of hours) {
    grid[h.day][h.hour] += h.count;
    total += h.count;
  }
  if (total < MIN_CLICKS) return null;
  let best = { day: 0, hour: 0, score: -1 };
  for (let day = 0; day < 7; day++)
    for (let hour = 0; hour < 24; hour++) {
      // Neighbours wrap across midnight into the adjacent day.
      const at = (offset: number) => {
        const i = day * 24 + hour + offset;
        const wrapped = (i + 168) % 168;
        return grid[Math.floor(wrapped / 24)][wrapped % 24];
      };
      const score = at(-1) + 2 * at(0) + at(1);
      if (score > best.score) best = { day, hour, score };
    }
  return { day: best.day, hour: best.hour, dayName: DAYS[best.day], total };
}

/** Next local time strictly after `now` that falls on `day` (0 = Monday) at `hour`:00. */
export function nextOccurrence(day: number, hour: number, now = new Date()) {
  const next = new Date(now);
  next.setHours(hour, 0, 0, 0);
  const today = (now.getDay() + 6) % 7;
  let add = (day - today + 7) % 7;
  if (add === 0 && next <= now) add = 7;
  next.setDate(next.getDate() + add);
  return next;
}

export function formatHour(hour: number) {
  const h = hour % 12 || 12;
  return `${h} ${hour < 12 ? "AM" : "PM"}`;
}
