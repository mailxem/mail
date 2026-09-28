package sending

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"

	"github.com/aws/smithy-go"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// Preserve the cause for operators while exposing only curated messages to clients.
type domainCheckError struct {
	stage string
	cause error
}

func (e *domainCheckError) Error() string { return "domain check " + e.stage + ": " + e.cause.Error() }
func (e *domainCheckError) Unwrap() error { return e.cause }

func dnsRecordMissing(err error) bool {
	var dnsErr *net.DNSError
	return errors.As(err, &dnsErr) && dnsErr.IsNotFound
}

func domainCheckHTTPError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return httpError(err)
	}
	if errors.Is(err, ErrDenied) {
		return echo.NewHTTPError(http.StatusForbidden, "Domain verification is unavailable or this check is no longer current. Refresh the page and check your workspace status.")
	}
	var timeout interface{ Timeout() bool }
	if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &timeout) && timeout.Timeout()) {
		return echo.NewHTTPError(http.StatusGatewayTimeout, "The domain check timed out. Try checking your records again shortly.")
	}
	var check *domainCheckError
	if errors.As(err, &check) {
		switch check.stage {
		case "provision":
			return echo.NewHTTPError(http.StatusBadGateway, "Unable to prepare your sending domain with the email provider. Contact your Xem operator to check the managed sending configuration.")
		case "identity":
			return echo.NewHTTPError(http.StatusBadGateway, "Unable to retrieve your domain verification status from the email provider. Try again shortly; if this continues, contact your Xem operator.")
		case "ownership_dns", "dmarc_dns":
			return echo.NewHTTPError(http.StatusServiceUnavailable, "DNS lookup is temporarily unavailable. Try checking your records again shortly.")
		}
	}
	return echo.NewHTTPError(http.StatusInternalServerError, "Unable to complete domain verification. Try again shortly; if this continues, contact your Xem operator.")
}

// Do not log raw provider/database errors: they may contain credentials, SQL,
// ownership tokens, or customer data. Codes and operation names locate failures
// for both HTTP checks and periodic verification without exposing those values.
func logDomainCheckError(team, domain, region string, err error) {
	if errors.Is(err, ErrDenied) || errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, context.Canceled) {
		return
	}
	var check *domainCheckError
	if !errors.As(err, &check) {
		return
	}
	var code, operation, requestID, sqlState string
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		code = apiErr.ErrorCode()
	}
	var opErr *smithy.OperationError
	if errors.As(err, &opErr) {
		operation = opErr.OperationName
	}
	var requestErr interface{ ServiceRequestID() string }
	if errors.As(err, &requestErr) {
		requestID = requestErr.ServiceRequestID()
	}
	var dbErr interface{ SQLState() string }
	if errors.As(err, &dbErr) {
		sqlState = dbErr.SQLState()
	}
	log.Printf("managed domain check failed: team_id=%q domain_id=%q region=%q stage=%q error_type=%T provider_code=%q operation=%q request_id=%q sql_state=%q", team, domain, region, check.stage, check.cause, code, operation, requestID, sqlState)
}
