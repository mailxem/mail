# Billing Components

Beautiful, comprehensive billing management components for the application.

## Components

### PricingCard
Displays individual pricing plans with features, pricing, and CTA buttons.
- Shows plan icon, name, description
- Displays pricing with billing cycle
- Lists all features with checkmarks
- Highlights popular plans
- Indicates current plan

### SubscriptionStatus
Shows current subscription status with visual indicators.
- Displays subscription plan name and status
- Shows renewal date and seat count
- Provides "Manage Billing" button
- Visual status indicators (active, pending, cancelled, etc.)

### UsageDisplay
Displays usage statistics for the current billing period.
- Shows usage for each feature (queries, data sources, chats, members)
- Progress bars for quota limits
- Handles unlimited quotas
- Visual icons for each feature type

### PlanComparison
Modal dialog for comparing all available plans side-by-side.
- Tabular comparison of all features
- Shows pricing for each plan
- Highlights popular plans
- Indicates which features are included in each plan

### BillingCycleToggle
Toggle switch for switching between monthly and yearly billing cycles.
- Smooth toggle animation
- Automatic discount badge for yearly plans
- Filters plans by billing cycle
- Accessible and keyboard navigable

## Usage

```tsx
import {
  PricingCard,
  SubscriptionStatus,
  UsageDisplay,
  PlanComparison,
  BillingCycleToggle
} from "@/components/billing";

// In your component
<SubscriptionStatus
  subscription={subscription}
  onManageBilling={handleManageBilling}
/>

<PricingCard
  plan={plan}
  currentPlanId={currentPlanId}
  onSelectPlan={handleSelectPlan}
/>

<UsageDisplay
  usageStats={usageStats}
  planFeatures={planFeatures}
/>

<PlanComparison
  plans={plans}
  open={showComparison}
  onOpenChange={setShowComparison}
/>

<BillingCycleToggle
  cycle={billingCycle}
  onCycleChange={setBillingCycle}
  yearlyDiscount={20}
/>
```

## Features

- ✅ View all available plans
- ✅ Toggle between monthly and yearly billing cycles
- ✅ Automatic yearly discount badge
- ✅ Change plans with confirmation dialog
- ✅ Cancel subscription
- ✅ View usage statistics
- ✅ View payment history
- ✅ Compare plans side-by-side
- ✅ Manage billing through customer portal
- ✅ Handle checkout for plan upgrades
- ✅ Beautiful, responsive UI matching design system
- ✅ Real-time data with React Query
- ✅ Toast notifications for actions
- ✅ Loading states

## Routes

- `/settings` (Billing tab) - Main billing interface
- `/settings/billing` - Standalone billing page

## Dependencies

- `@tanstack/react-query` - Data fetching and caching
- `date-fns` - Date formatting
- `sonner` - Toast notifications
- `lucide-react` - Icons
- UI components from `@/components/ui`
