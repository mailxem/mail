# Workflow builder

Drafts can be saved before configuration is complete. Validation and publishing check the entire graph and workspace-owned resources. Pause a published workflow before editing; changing a graph with in-flight executions remains blocked.

## Steps

| Type | Configuration and behavior |
| --- | --- |
| `EMAIL` | Workspace template and sender, optional subject, postal address for marketing templates. Existing consent and suppression enforcement applies. |
| `WAIT` | `duration` from 1 second through 365 days, or `until` as an RFC3339 timestamp including its timezone. Past dates continue immediately. Waiting runs retain their variables and execution identity. |
| `CONDITION` | 1–20 rules joined with `AND` or `OR`; `true` and `false` graph edges. Supports equality, numeric/ordered-text comparisons, contains/not-contains, prefix/suffix, existence and empty-value checks. Text substring comparisons ignore case. |
| `TAG` | `action` is `add` or `remove`, with 1–20 tag names. Adding can create workspace tags. Repeating the same association is safe. |
| `ADD_TO_LIST` | Moves the contact to a workspace-owned list; contacts have one list. Both lists' active subscriber counts are recalculated in the same transaction. Subscription status is preserved. |
| `UPDATE_SUBSCRIBER` | Literal text updates to name, company, location, address, phone and social-profile fields. Empty text clears a field. Email, subscription status and team ownership cannot be changed. |
| `PERCENTAGE_SPLIT` | `percentage` is the share on path `A` (1–99); the remainder uses path `B`. Assignment hashes workflow, node and contact IDs and remains stable on retries. Small samples may differ from the target ratio. |
| `SET_VARIABLE` | Sets a text value under a `workflow_` name. Conditions later in the run can read it; reserved contact/event variables cannot be overwritten by this step. |

The editor includes step search, optional step names, workflow descriptions, separate branch handles, multiple-rule editing, wait presets and date selection, and per-step execution history. Selecting a finish step inserts before it, including inside an existing branch. All paths must terminate, and graphs cannot contain cycles.

Supported entry events remain manual, contact created, email opened, email clicked, and opted-in form completed. These changes do not expose legacy webhook or AI processors as publishable steps.

## Validation

Focused Go tests cover action execution, workspace ownership, forbidden subscription changes, tag retries, list counts, deterministic splits, condition routing, and a real engine sequence that sets a variable, waits, resumes, checks conditions, updates a profile and finishes. Frontend tests cover inserting within a branch, preserving multiple rules and published edit locks.
