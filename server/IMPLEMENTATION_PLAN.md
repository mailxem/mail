# Implementation Plan - Critical Features

## Phase 1: Segmentation Engine (Weeks 1-4)

### Technical Architecture

#### Database Schema
```sql
-- Segments table
segments:
  - id (uuid, primary key)
  - team_id (uuid, foreign key)
  - name (string)
  - description (text)
  - rules (jsonb) - Rule builder configuration
  - match_type (enum: ALL, ANY) - AND/OR logic
  - contact_count (integer) - Cached count
  - is_dynamic (boolean) - Auto-update membership
  - created_at, updated_at, is_deleted

-- Segment membership (for static/cached segments)
segment_contacts:
  - segment_id (uuid, foreign key)
  - contact_id (uuid, foreign key)
  - added_at (timestamp)
  - primary key (segment_id, contact_id)

-- Rule structure (stored in segments.rules as JSONB)
{
  "rules": [
    {
      "field": "email",
      "operator": "contains",
      "value": "@gmail.com"
    },
    {
      "field": "tags",
      "operator": "includes",
      "value": "vip"
    }
  ],
  "matchType": "ALL" // AND logic
}
```

#### Rule Operators by Field Type
```go
String fields: ==, !=, contains, not_contains, starts_with, ends_with
Number fields: ==, !=, >, <, >=, <=
Date fields: ==, !=, >, <, is_set, is_not_set, within_days
Array fields: includes, not_includes, is_empty, is_not_empty
Boolean fields: is_true, is_false
```

#### API Endpoints
```
POST   /api/v1/segments                  Create segment
GET    /api/v1/segments                  List segments
GET    /api/v1/segments/:id              Get segment details
PUT    /api/v1/segments/:id              Update segment
DELETE /api/v1/segments/:id              Delete segment
GET    /api/v1/segments/:id/contacts     Get segment contacts (paginated)
POST   /api/v1/segments/:id/refresh      Recalculate membership
GET    /api/v1/segments/:id/preview      Preview contacts (before save)
POST   /api/v1/segments/:id/export       Export contacts to CSV
```

#### Automation Integration
```
New Node: SEGMENT_FILTER
- Routes contacts based on segment membership
- If contact in segment -> true path
- If contact not in segment -> false path

New Node: ADD_TO_SEGMENT
- Adds contact to static segment
- Updates segment_contacts table

New Node: REMOVE_FROM_SEGMENT
- Removes contact from segment
- Updates segment_contacts table

New Trigger: segment.entered
- Fires when contact joins dynamic segment
- Can trigger automations

New Trigger: segment.exited
- Fires when contact leaves dynamic segment
- Can trigger automations
```

---

## Phase 2: Lead Scoring (Weeks 5-7)

### Technical Architecture

#### Database Schema
```sql
-- Lead scores (current score per contact)
lead_scores:
  - contact_id (uuid, primary key, foreign key)
  - score (integer, default 0)
  - grade (enum: A, B, C, D, F)
  - last_activity_at (timestamp)
  - created_at, updated_at

-- Score activities (audit log)
score_activities:
  - id (uuid, primary key)
  - contact_id (uuid, foreign key)
  - activity_type (string) - "email_opened", "link_clicked", etc.
  - points (integer) - Can be negative for decay
  - automation_id (uuid, nullable)
  - campaign_id (uuid, nullable)
  - created_at

-- Score rules (configurable per team)
score_rules:
  - id (uuid, primary key)
  - team_id (uuid, foreign key)
  - activity_type (string)
  - points (integer)
  - decay_days (integer, nullable) - Auto-subtract after X days
  - is_active (boolean)
  - created_at, updated_at

-- Default score rules
email_opened: +5
email_clicked: +10
form_submitted: +20
page_visited: +2
demo_requested: +50
purchase: +100
unsubscribed: -50
```

#### Score Grades
```go
A: 80-100 (Hot lead, sales ready)
B: 60-79 (Warm lead, nurture more)
C: 40-59 (Cold lead, engage)
D: 20-39 (Very cold, re-engagement campaign)
F: 0-19 (Inactive, consider removing)
```

#### API Endpoints
```
GET    /api/v1/contacts/:id/score        Get contact score
POST   /api/v1/contacts/:id/score        Update score manually
GET    /api/v1/contacts/:id/score/history Score activity log
GET    /api/v1/score-rules                List scoring rules
POST   /api/v1/score-rules                Create scoring rule
PUT    /api/v1/score-rules/:id            Update rule
DELETE /api/v1/score-rules/:id            Delete rule
GET    /api/v1/analytics/scores           Score distribution
```

#### Automation Integration
```
New Node: SCORE_CHANGE
- Add or subtract points
- Example: "Add 20 points for form submission"

New Node: SCORE_THRESHOLD
- Routes based on score
- If score >= 80 -> "hot lead" path
- If score < 80 -> "nurture" path

New Trigger: score.threshold_reached
- Fires when contact hits score threshold
- Example: "When score reaches 80, notify sales"

New Trigger: score.grade_changed
- Fires when grade changes (C->B->A)
- Can trigger different nurture tracks

Background Job: score_decay
- Runs daily
- Subtracts points for old activities
- Example: "Email open from 30 days ago loses 5 points"
```

---

## Phase 3: A/B Testing (Weeks 8-11)

### Technical Architecture

#### Database Schema
```sql
-- A/B tests
ab_tests:
  - id (uuid, primary key)
  - automation_id (uuid, foreign key, nullable)
  - campaign_id (uuid, foreign key, nullable)
  - name (string)
  - test_type (enum: SUBJECT, CONTENT, SEND_TIME, FULL)
  - winner_metric (enum: OPEN_RATE, CLICK_RATE, CONVERSION_RATE)
  - status (enum: DRAFT, RUNNING, COMPLETED)
  - winner_variant_id (uuid, nullable)
  - confidence_threshold (float, default 0.95) - 95% confidence
  - started_at, ended_at
  - created_at, updated_at

-- Test variants
test_variants:
  - id (uuid, primary key)
  - test_id (uuid, foreign key)
  - variant_name (string) - "A", "B", "C"
  - split_percentage (integer) - 33, 33, 34 for 3-way split
  - subject_line (string, nullable)
  - email_content (text, nullable)
  - send_hour (integer, nullable) - For send-time tests
  - created_at

-- Test results
test_results:
  - id (uuid, primary key)
  - test_id (uuid, foreign key)
  - variant_id (uuid, foreign key)
  - sends (integer)
  - opens (integer)
  - clicks (integer)
  - conversions (integer)
  - revenue (decimal, nullable)
  - updated_at

-- Variant assignments (track which contact got which variant)
variant_assignments:
  - test_id (uuid, foreign key)
  - contact_id (uuid, foreign key)
  - variant_id (uuid, foreign key)
  - assigned_at (timestamp)
  - primary key (test_id, contact_id)
```

#### Statistical Analysis
```go
// Chi-squared test for statistical significance
func CalculateSignificance(variantA, variantB TestResult) float64 {
    // Compare open rates or click rates
    // Return p-value
    // If p < 0.05, difference is statistically significant
}

// Declare winner when:
// 1. Minimum sample size reached (typically 100+ per variant)
// 2. Statistical significance achieved (p < 0.05)
// 3. Confidence threshold met (95%+)
```

#### API Endpoints
```
POST   /api/v1/ab-tests                   Create test
GET    /api/v1/ab-tests                   List tests
GET    /api/v1/ab-tests/:id               Get test details
PUT    /api/v1/ab-tests/:id               Update test
DELETE /api/v1/ab-tests/:id               Delete test
POST   /api/v1/ab-tests/:id/start         Start test
POST   /api/v1/ab-tests/:id/stop          Stop test
GET    /api/v1/ab-tests/:id/results       Get test results
POST   /api/v1/ab-tests/:id/declare-winner Manual winner selection
GET    /api/v1/ab-tests/:id/preview       Preview variants
```

#### Automation Integration
```
New Node: AB_SPLIT
- Randomly assigns contacts to variants
- Tracks assignment in variant_assignments
- Routes to different email nodes based on variant

Example:
START -> AB_SPLIT (33% A, 33% B, 34% C)
  -> Variant A: EMAIL (subject: "Sale")
  -> Variant B: EMAIL (subject: "Discount")
  -> Variant C: EMAIL (subject: "Limited Time")

Background Job: ab_test_analyzer
- Runs every hour for active tests
- Calculates statistical significance
- Auto-declares winner when threshold met
- Stops test and routes all to winner
```

---

## Phase 4: Goal Tracking (Week 12)

### Technical Architecture

#### Database Schema
```sql
-- Automation goals
automation_goals:
  - id (uuid, primary key)
  - automation_id (uuid, foreign key)
  - name (string) - "Book Demo", "Purchase", "Download"
  - goal_type (enum: EVENT, REVENUE, ENGAGEMENT)
  - target_event (string) - Event name to track
  - target_value (decimal, nullable) - Revenue target
  - created_at, updated_at

-- Goal conversions
goal_conversions:
  - id (uuid, primary key)
  - automation_id (uuid, foreign key)
  - goal_id (uuid, foreign key)
  - contact_id (uuid, foreign key)
  - execution_id (uuid, foreign key) - Which execution achieved goal
  - value (decimal, nullable) - Revenue value
  - converted_at (timestamp)
  - days_to_convert (integer) - Time from enrollment to conversion

-- Attribution touches (multi-touch attribution)
attribution_touches:
  - id (uuid, primary key)
  - contact_id (uuid, foreign key)
  - automation_id (uuid, foreign key)
  - touchpoint_type (enum: EMAIL_SENT, EMAIL_OPENED, EMAIL_CLICKED)
  - touchpoint_data (jsonb) - Additional context
  - created_at

-- Attribution models
first_touch: 100% credit to first automation
last_touch: 100% credit to last automation
linear: Equal credit across all touches
time_decay: More credit to recent touches
```

#### Goal Types
```go
EVENT_GOAL:
  - Track when specific event happens
  - Example: "form_submitted", "demo_booked"

REVENUE_GOAL:
  - Track revenue from automation
  - Example: "Generate $10,000 in sales"

ENGAGEMENT_GOAL:
  - Track engagement metrics
  - Example: "Achieve 40% open rate"
```

#### API Endpoints
```
POST   /api/v1/automations/:id/goals          Create goal
GET    /api/v1/automations/:id/goals          List goals
PUT    /api/v1/automations/:id/goals/:goalId  Update goal
DELETE /api/v1/automations/:id/goals/:goalId  Delete goal
GET    /api/v1/automations/:id/conversions    Get conversions
POST   /api/v1/conversions                    Track conversion
GET    /api/v1/analytics/attribution          Attribution report
GET    /api/v1/analytics/roi                  ROI by automation
```

#### Automation Integration
```
New Node: GOAL_ACHIEVED
- Marks goal as achieved for contact
- Records conversion with timestamp
- Triggers goal.achieved event

New Node: TRACK_REVENUE
- Records revenue for attribution
- Links to specific automation execution

New Trigger: goal.achieved
- Fires when goal is completed
- Can trigger celebration/follow-up automation

Metrics Calculated:
- Conversion rate per automation
- Average time to conversion
- Revenue per automation
- ROI (revenue - cost)
- Goal completion funnel
```

---

## Implementation Order

### Week 1-4: Segmentation Engine
**Day 1-2:** Database models and migrations
**Day 3-5:** Segment service with rule engine
**Day 6-8:** Segment APIs (CRUD)
**Day 9-11:** Contact query engine (dynamic segment evaluation)
**Day 12-15:** Automation nodes (SEGMENT_FILTER, ADD_TO_SEGMENT, REMOVE_FROM_SEGMENT)
**Day 16-18:** Testing and optimization
**Day 19-20:** Documentation and examples

### Week 5-7: Lead Scoring
**Day 1-2:** Database models and migrations
**Day 3-4:** Score rules configuration
**Day 5-7:** Scoring service (add points, calculate grades)
**Day 8-10:** Score APIs
**Day 11-13:** Automation nodes (SCORE_CHANGE, SCORE_THRESHOLD)
**Day 14-15:** Background job for score decay
**Day 16-18:** Testing and analytics
**Day 19-21:** Documentation

### Week 8-11: A/B Testing
**Day 1-3:** Database models and migrations
**Day 4-7:** Test management service
**Day 8-11:** Variant assignment logic
**Day 12-15:** Statistical analysis engine
**Day 16-20:** Automation node (AB_SPLIT)
**Day 21-25:** Results dashboard APIs
**Day 26-28:** Background jobs (analyzer, winner selection)
**Day 29-30:** Testing and documentation

### Week 12: Goal Tracking
**Day 1-2:** Database models and migrations
**Day 3-4:** Goal tracking service
**Day 5-6:** Conversion APIs
**Day 7-8:** Attribution logic
**Day 9:** Automation node (GOAL_ACHIEVED)
**Day 10:** Testing and documentation

---

## Success Metrics

### Segmentation Engine
- [ ] Can create segments with 5+ rule combinations
- [ ] Dynamic segments update in < 5 seconds
- [ ] 50% of customers use segments within first month
- [ ] Segment preview shows accurate counts

### Lead Scoring
- [ ] Scoring rules apply in real-time (< 1 second)
- [ ] Score decay runs daily without errors
- [ ] 80% adoption rate
- [ ] Score-based routing works in automations

### A/B Testing
- [ ] Tests accurately split traffic by percentage
- [ ] Statistical significance calculated correctly
- [ ] Winner auto-selected within 24 hours of confidence threshold
- [ ] 30% of automations use A/B testing

### Goal Tracking
- [ ] Goals track conversions accurately
- [ ] Attribution logic handles multi-touch scenarios
- [ ] ROI calculations are correct
- [ ] 70% of automations have goals defined

---

## Technical Standards

### Code Quality
- [ ] All code follows existing patterns
- [ ] 80%+ test coverage
- [ ] No SQL injection vulnerabilities
- [ ] Input validation on all APIs
- [ ] Rate limiting on expensive operations

### Performance
- [ ] Segment evaluation < 5 seconds for 100K contacts
- [ ] Score updates < 1 second
- [ ] A/B variant assignment < 100ms
- [ ] Goal tracking < 500ms

### Documentation
- [ ] API documentation with examples
- [ ] UI mockups for frontends
- [ ] Architecture diagrams
- [ ] Migration guides

---

**Ready to start implementation!** 🚀
