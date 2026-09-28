# Billing Cycle Toggle Feature

## Overview
Added a beautiful toggle switch to allow users to switch between monthly and yearly billing plans.

## Component: BillingCycleToggle

### Location
`components/billing/billing-cycle-toggle.tsx`

### Features
✅ Toggle between Monthly and Yearly plans
✅ Automatic discount badge for yearly plans
✅ Smooth transitions
✅ Accessible with proper labels
✅ Dark theme compatible

### Props
```typescript
interface BillingCycleToggleProps {
  cycle: "monthly" | "yearly";           // Current selected cycle
  onCycleChange: (cycle) => void;        // Callback when cycle changes
  yearlyDiscount?: number;               // Discount percentage (default: 20%)
}
```

### Usage
```tsx
<BillingCycleToggle
  cycle={billingCycle}
  onCycleChange={setBillingCycle}
  yearlyDiscount={calculateYearlyDiscount()}
/>
```

## Implementation Details

### Filtering Logic
- Free plans (price = "0") are shown in both monthly and yearly views
- Paid plans are filtered by `billing_cycle` field
- Plans automatically update when toggle is switched

### Discount Calculation
The component automatically calculates yearly savings:
```typescript
const calculateYearlyDiscount = () => {
  const monthlyAnnual = monthlyPrice * 12;
  const yearly = yearlyPrice;
  const discount = ((monthlyAnnual - yearly) / monthlyAnnual) * 100;
  return Math.round(discount);
};
```

### Visual Design
```
┌─────────────────────────────────────────┐
│  Monthly  ⚪──────⚪  Yearly  [Save 20%] │
└─────────────────────────────────────────┘
```

When yearly is selected:
```
┌─────────────────────────────────────────┐
│  Monthly  ──────⚪⚪  Yearly  [Save 20%] │
└─────────────────────────────────────────┘
```

### State Management
```typescript
const [billingCycle, setBillingCycle] = useState<"monthly" | "yearly">("monthly");

// Filter plans based on cycle
const filteredPlans = plans.filter(
  (plan) => plan.billing_cycle === billingCycle || plan.price === "0"
);
```

## User Experience Flow

1. **Page Load**
   - Defaults to "Monthly" view
   - Shows all monthly plans + free plans

2. **Toggle to Yearly**
   - Switch animates to right
   - "Save 20%" badge appears
   - Plans smoothly transition to yearly options
   - Free plan remains visible

3. **Toggle back to Monthly**
   - Switch animates to left
   - Discount badge fades out
   - Plans transition back to monthly options

## Visual Features

### Colors
- **Active state**: `text-foreground` (full opacity)
- **Inactive state**: `text-muted-foreground` (reduced opacity)
- **Discount badge**: Green (`bg-green-500/10`, `text-green-600`)
- **Background**: `bg-muted/30` with border

### Animations
- Smooth switch transition
- Text color fade
- Badge appear/disappear
- Plan card transitions

### Accessibility
- Proper `<Label>` associations
- Keyboard navigable
- Screen reader friendly
- Focus states

## Integration with Existing Features

### Works With
✅ Plan comparison modal (shows filtered plans)
✅ Current subscription indicator
✅ Plan selection/checkout flow
✅ All existing billing features

### Maintains
✅ Current subscription status
✅ Usage statistics
✅ Payment history
✅ All existing functionality

## Example Plans Structure

```typescript
// Backend should return plans like:
[
  {
    id: "free-001",
    name: "Free",
    price: "0",
    billing_cycle: "monthly", // Can be either, doesn't matter
    // ...
  },
  {
    id: "pro-monthly-001",
    name: "Pro",
    price: "16.99",
    billing_cycle: "monthly",
    // ...
  },
  {
    id: "pro-yearly-001",
    name: "Pro",
    price: "163.00",
    billing_cycle: "yearly",
    // ...
  }
]
```

## Responsive Design

### Desktop
- Full toggle with both labels visible
- Badge displayed inline
- Centered layout

### Mobile
- Toggle scales appropriately
- Labels remain readable
- Badge wraps if needed

## Testing Scenarios

- [ ] Toggle switches between monthly/yearly
- [ ] Free plans show in both views
- [ ] Correct plans filter based on cycle
- [ ] Discount badge shows/hides appropriately
- [ ] Discount calculation is accurate
- [ ] Current plan indicator works in both views
- [ ] Plan selection works in both views
- [ ] Toggle state persists during page navigation
- [ ] Keyboard navigation works
- [ ] Mobile layout looks good

## Future Enhancements (Optional)

1. **Remember User Preference**
   ```typescript
   const [billingCycle, setBillingCycle] = useLocalStorage(
     "billing-cycle-preference",
     "monthly"
   );
   ```

2. **Animated Savings Counter**
   - Show "Save $XX.XX" instead of percentage
   - Animate counter when toggling

3. **Popular Badge Logic**
   - Show different "popular" plans for monthly vs yearly

4. **Trial Period Indicator**
   - Highlight trial length difference between cycles

5. **URL State**
   - Reflect toggle state in URL
   - Deep link to specific cycle view
