import { formatRate, rateDifference } from "@/lib/analytics/types";
test("rates format fractions once and distinguish missing data from zero", () => {
  expect(formatRate({ value: 1, numerator: 1, denominator: 1 })).toBe("100.0%");
  expect(formatRate({ value: 0, numerator: 0, denominator: 10 })).toBe("0.0%");
  expect(formatRate({ value: null, numerator: 0, denominator: 0 })).toBe("—");
  expect(formatRate(undefined)).toBe("—");
});
test("rate comparisons are percentage points and do not invent missing baselines", () => {
  expect(
    rateDifference(
      { value: 0.08, numerator: 8, denominator: 100 },
      { value: 0.05, numerator: 5, denominator: 100 },
    ),
  ).toBe("+3.0 pp");
  expect(
    rateDifference(
      { value: 0.08, numerator: 8, denominator: 100 },
      { value: null, numerator: 0, denominator: 0 },
    ),
  ).toBe("No comparable rate");
});
