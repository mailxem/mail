package sending

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Config struct {
	Enabled                                            bool
	Region, AccountID, TopicARN                        string
	SMTPAddr, SMTPHost, TLSCert, TLSKey                string
	SMTPEnabled                                        bool
	MaxBytes                                           int64
	DailyLimit, MonthlyLimit, BudgetMicros, CostMicros int64
	QueueLimit                                         int64
}

func LoadConfig() (Config, error) {
	c := Config{Enabled: os.Getenv("MANAGED_SENDING_ENABLED") == "true", Region: os.Getenv("MANAGED_SES_REGION"), AccountID: os.Getenv("MANAGED_AWS_ACCOUNT_ID"), TopicARN: os.Getenv("MANAGED_SNS_TOPIC_ARN"), SMTPEnabled: os.Getenv("MANAGED_SMTP_ENABLED") == "true", SMTPAddr: os.Getenv("MANAGED_SMTP_ADDR"), SMTPHost: os.Getenv("MANAGED_SMTP_HOST"), TLSCert: os.Getenv("MANAGED_SMTP_TLS_CERT"), TLSKey: os.Getenv("MANAGED_SMTP_TLS_KEY")}
	if c.SMTPAddr == "" {
		c.SMTPAddr = ":587"
	}
	values := []struct {
		k   string
		dst *int64
		def int64
	}{{"MANAGED_MAX_MESSAGE_BYTES", &c.MaxBytes, 5 * 1024 * 1024}, {"MANAGED_DAILY_LIMIT", &c.DailyLimit, autoApprovalDailyLimit}, {"MANAGED_MONTHLY_LIMIT", &c.MonthlyLimit, autoApprovalMonthlyLimit}, {"MANAGED_MONTHLY_BUDGET_MICROS", &c.BudgetMicros, autoApprovalBudgetMicros}, {"MANAGED_COST_PER_RECIPIENT_MICROS", &c.CostMicros, 1000}, {"MANAGED_QUEUE_LIMIT", &c.QueueLimit, 1000}}
	for _, v := range values {
		*v.dst = v.def
		if raw := os.Getenv(v.k); raw != "" {
			n, e := strconv.ParseInt(raw, 10, 64)
			if e != nil || n <= 0 || n > 1000000000 {
				return c, fmt.Errorf("invalid %s", v.k)
			}
			*v.dst = n
		}
	}
	if !c.Enabled {
		if c.SMTPEnabled {
			return c, fmt.Errorf("managed SMTP requires managed sending")
		}
		return c, nil
	}
	if !regexp.MustCompile(`^[a-z]{2}-[a-z]+-[0-9]+$`).MatchString(c.Region) || !regexp.MustCompile(`^[0-9]{12}$`).MatchString(c.AccountID) {
		return c, fmt.Errorf("managed sending requires a commercial AWS region and account ID")
	}
	if !strings.HasPrefix(c.TopicARN, "arn:aws:sns:"+c.Region+":"+c.AccountID+":") {
		return c, fmt.Errorf("SNS topic must belong to the configured AWS account and region")
	}
	if c.MaxBytes > 10*1024*1024 {
		return c, fmt.Errorf("managed message size must not exceed 10 MiB")
	}
	if c.SMTPEnabled {
		if !validDomain(c.SMTPHost) || c.TLSCert == "" || c.TLSKey == "" {
			return c, fmt.Errorf("SMTP requires hostname and certificate/key files")
		}
		if _, _, e := net.SplitHostPort(c.SMTPAddr); e != nil {
			return c, e
		}
	}
	return c, nil
}
