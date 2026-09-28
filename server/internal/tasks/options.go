package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"github.com/robfig/cron/v3"
)

type cronSchedule struct {
	expr string
}

func (s cronSchedule) String() string {
	return fmt.Sprintf("cron=%s", s.expr)
}

func (s cronSchedule) Type() asynq.OptionType {
	return asynq.ProcessAtOpt
}

func (s cronSchedule) Value() interface{} {
	return s.expr
}

func (s cronSchedule) Apply(opts *asynq.TaskInfo) {
	schedule, err := cron.ParseStandard(s.expr)
	if err != nil {
		// Fall back to default interval if parsing fails
		opts.NextProcessAt = time.Now().Add(1 * time.Hour)
		return
	}

	// Set next processing time
	now := time.Now()
	next := schedule.Next(now)
	opts.NextProcessAt = next
}

// CronSchedule returns an option to schedule a task with a cron expression
func CronSchedule(expr string) asynq.Option {
	return cronSchedule{expr: expr}
}

// Instead of AfterFunc, we'll use task handlers to manage recurring tasks
// The task handler will need to reschedule the task after processing

// AfterFunc option for handling task completion
type afterOption struct {
	fn func(context.Context, *asynq.Task) error
}

func (o afterOption) String() string {
	return "after"
}

func (o afterOption) Type() asynq.OptionType {
	return asynq.RetentionOpt // Using RetentionOpt as it's processed after task completion
}

func (o afterOption) Value() interface{} {
	return o.fn
}

func (o afterOption) Apply(opts *asynq.TaskInfo) {
	// Store the function in task's metadata
	if opts.Payload == nil {
		opts.Payload = []byte("{}")
	}

	// Unmarshal existing payload
	var payload map[string]interface{}
	if err := json.Unmarshal(opts.Payload, &payload); err != nil {
		payload = make(map[string]interface{})
	}

	// Store the function in the payload
	payload["after_func"] = o.fn

	// Marshal back to bytes
	if newPayload, err := json.Marshal(payload); err == nil {
		opts.Payload = newPayload
	}
}

// AfterFunc returns an option to run a function after task completion
func AfterFunc(fn func(context.Context, *asynq.Task) error) asynq.Option {
	return afterOption{fn: fn}
}
