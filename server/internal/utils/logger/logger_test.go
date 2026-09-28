package logger

import (
	"bytes"
	"errors"
	"net/textproto"
	"strings"
	"testing"

	"github.com/fatih/color"
)

func TestErrorWrapsAndLogsSMTPFailure(t *testing.T) {
	var output bytes.Buffer
	previous := color.Output
	color.Output = &output
	t.Cleanup(func() { color.Output = previous })

	cause := &textproto.Error{Code: 501, Msg: "Error: Bad sender address syntax (100% rejected)"}
	inner := New("EMAIL_HANDLER").Error("failed to send email: %w", cause)
	outer := New("task_handler").Error("task %s failed: %w", inner, "send")
	want := "task send failed: failed to send email: 501 Error: Bad sender address syntax (100% rejected)"
	if outer.Error() != want {
		t.Fatalf("error = %q, want %q", outer, want)
	}
	if !errors.Is(outer, cause) {
		t.Fatal("SMTP error was not preserved in the error chain")
	}
	var smtpError *textproto.Error
	if !errors.As(outer, &smtpError) || smtpError.Code != 501 {
		t.Fatal("SMTP status code was not preserved")
	}
	if !strings.Contains(output.String(), want) || strings.Contains(output.String(), "%!") {
		t.Fatalf("unexpected log output: %s", output.String())
	}
}
