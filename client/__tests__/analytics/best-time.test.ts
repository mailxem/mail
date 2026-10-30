import { nextOccurrence, peakHour, formatHour } from "@/lib/analytics/best-time";

const cell = (day: number, hour: number, count: number) => ({ day, hour, count });

test("needs enough clicks before suggesting a time", () => {
  expect(peakHour([cell(6, 9, 10)])).toBeNull();
});

test("prefers a busy block over a single noisy hour", () => {
  const hours = [
    cell(0, 3, 30), // one spike on Monday 3 AM
    cell(6, 8, 18),
    cell(6, 9, 25),
    cell(6, 10, 18),
  ];
  expect(peakHour(hours)).toMatchObject({ day: 6, hour: 9, dayName: "Sunday" });
});

test("next occurrence is later today, next week, or later this week", () => {
  const wedMorning = new Date(2026, 9, 28, 8, 30); // Wed Oct 28 2026
  expect(nextOccurrence(2, 9, wedMorning)).toEqual(new Date(2026, 9, 28, 9));
  expect(nextOccurrence(2, 8, wedMorning)).toEqual(new Date(2026, 10, 4, 8));
  expect(nextOccurrence(6, 9, wedMorning)).toEqual(new Date(2026, 10, 1, 9));
});

test("formats hours in 12-hour time", () => {
  expect([0, 9, 12, 21].map(formatHour)).toEqual(["12 AM", "9 AM", "12 PM", "9 PM"]);
});
