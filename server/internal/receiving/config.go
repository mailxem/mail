package receiving

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Config struct {
	Enabled                                                        bool
	Region, AccountID, Bucket, TopicARN, QueueURL, RuleSet, Prefix string
	MaxMailboxesPerDomain                                          int
	MaxMessageBytes, MailboxQuotaBytes                             int64
}

func LoadConfig() (Config, error) {
	c := Config{Enabled: os.Getenv("MANAGED_RECEIVING_ENABLED") == "true", Region: os.Getenv("MANAGED_SES_REGION"), AccountID: os.Getenv("MANAGED_AWS_ACCOUNT_ID"), Bucket: os.Getenv("MANAGED_RECEIVING_BUCKET"), TopicARN: os.Getenv("MANAGED_RECEIVING_TOPIC_ARN"), QueueURL: os.Getenv("MANAGED_RECEIVING_QUEUE_URL"), RuleSet: os.Getenv("MANAGED_RECEIVING_RULE_SET"), Prefix: "incoming/", MaxMailboxesPerDomain: 100, MaxMessageBytes: 10 * 1024 * 1024, MailboxQuotaBytes: 1024 * 1024 * 1024}
	for _, v := range []struct {
		name string
		dst  *int64
	}{{"MANAGED_RECEIVING_MAX_MESSAGE_BYTES", &c.MaxMessageBytes}, {"MANAGED_RECEIVING_MAILBOX_QUOTA_BYTES", &c.MailboxQuotaBytes}} {
		if raw := os.Getenv(v.name); raw != "" {
			n, e := strconv.ParseInt(raw, 10, 64)
			if e != nil || n <= 0 {
				return c, fmt.Errorf("invalid %s", v.name)
			}
			*v.dst = n
		}
	}
	if !c.Enabled {
		return c, nil
	}
	if !regexp.MustCompile(`^[a-z]{2}-[a-z]+-[0-9]+$`).MatchString(c.Region) || !regexp.MustCompile(`^[0-9]{12}$`).MatchString(c.AccountID) {
		return c, fmt.Errorf("managed receiving requires AWS region and account ID")
	}
	bucketPattern := regexp.MustCompile(`^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$`)
	rulePattern := regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)
	queue, queueErr := url.Parse(c.QueueURL)
	queueOK := queueErr == nil && queue.Scheme == "https" && queue.Host == "sqs."+c.Region+".amazonaws.com" && queue.RawQuery == "" && queue.Fragment == "" && queue.User == nil && strings.HasPrefix(queue.EscapedPath(), "/"+c.AccountID+"/")
	topicPattern := regexp.MustCompile(`^arn:aws:sns:` + regexp.QuoteMeta(c.Region) + `:` + regexp.QuoteMeta(c.AccountID) + `:[A-Za-z0-9_-]{1,256}$`)
	if !bucketPattern.MatchString(c.Bucket) || !queueOK || !rulePattern.MatchString(c.RuleSet) || !topicPattern.MatchString(c.TopicARN) {
		return c, fmt.Errorf("managed receiving infrastructure is incomplete")
	}
	if c.MaxMessageBytes > 10*1024*1024 || c.MailboxQuotaBytes < c.MaxMessageBytes {
		return c, fmt.Errorf("managed receiving size limits are invalid")
	}
	return c, nil
}
