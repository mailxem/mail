package onboardingemails

// ServiceTemplates share the website brand layout with managed onboarding.
// Runtime values are escaped by the common renderer, never interpreted as HTML.
type ServiceTemplate struct {
	Template
	Label, Action, Path string
	Account             bool
}

type ServiceData struct{ Name, Email, Workspace, Time, Count string }

var ServiceTemplates = []ServiceTemplate{
	{Template{"welcome", "Welcome to Xem", "Your next chapter starts here.", "Welcome to Xem. Your account is ready, and there’s room to make your email feel like you.", "Explore the template library, connect a sender when you’re ready, and send a test to yourself."}, "WELCOME TO XEM", "Open your workspace", "/dashboard", true},
	{Template{"password_reset", "Reset your Xem password", "Let’s get you back in.", "We received a request to reset your Xem password. Use the button below to choose a new one. This link expires 15 minutes after it was requested and can be used once.", "If you didn’t request this, you can ignore this email. Your password stays the same until you choose a new one."}, "ACCOUNT SECURITY", "Reset your password", "/auth/reset-password/{{reset_token}}", true},
	{Template{"password_changed", "Your Xem password has changed", "Your password is updated.", "The password for your Xem account was successfully changed. Existing sign-in sessions have been revoked; sign in again with your new password.", "If this wasn’t you, reset your password immediately and contact your workspace administrator or support."}, "ACCOUNT SECURITY", "Secure your account", "/auth/forgot-password", true},
	{Template{"sending_failed", "Some emails need your attention", "A send needs a second look.", "One or more emails could not be submitted successfully. This can happen when a sender is unavailable, credentials need attention, or a provider rejects a message.", "Review the sending logs and sender settings. Check the latest status before retrying; some failures may already be scheduled for another attempt."}, "SENDING ALERT", "Review sending activity", "/analytics/delivery", false},
	{Template{"sending_delayed", "An email delivery is delayed", "Taking a little longer.", "The email provider reported a delivery delay. A delay is not a confirmed failure, and the provider may still be retrying delivery.", "Check the latest delivery events before sending another copy. Give the provider’s retry process time to finish."}, "SENDING ALERT", "Review delivery events", "/settings/sending", false},
	{Template{"sending_unknown", "An email’s delivery outcome is uncertain", "Let’s check before resending.", "A sending acknowledgement was interrupted or could not be confirmed. The provider or receiving server may already have accepted the email.", "Review provider events and sending logs before resending. Retrying an uncertain message can create a duplicate."}, "SENDING ALERT", "Review sending activity", "/analytics/delivery", false},
	{Template{"sending_bounced", "An email bounced", "An email couldn’t get through.", "The email provider reported a bounce. Permanent bounces are added to suppression so future sends do not keep targeting an unreachable address.", "Review the bounce details and check your audience data. Don’t remove a suppression until you have verified the address and the reason for the bounce."}, "DELIVERY HEALTH", "Review delivery events", "/settings/sending", false},
	{Template{"sending_complaint", "A recipient reported your email as spam", "A signal worth listening to.", "The email provider reported a spam complaint. The affected recipient has been suppressed, and managed sending may be suspended for review.", "Review audience consent, recent campaigns, and your managed sending status. Do not send again to the complaining recipient."}, "DELIVERY HEALTH", "Review sending status", "/settings/sending", false},
	{Template{"sending_suspended", "Managed sending has been suspended", "Time to review your sending.", "Managed sending for your workspace was suspended. Review the latest account status before attempting new sends.", "Open managed sending, review recent delivery feedback, and contact support or your administrator before resuming."}, "SENDING STATUS", "Review sending status", "/settings/sending", false},
}

func FindService(key string) (ServiceTemplate, bool) {
	for _, t := range ServiceTemplates {
		if t.Key == key {
			return t, true
		}
	}
	return ServiceTemplate{}, false
}
func (t ServiceTemplate) presentation(d ServiceData) presentation {
	p := presentation{Label: t.Label, Action: t.Action, Footer: "You received this service update as the workspace owner.", Details: []detail{{"Workspace", d.Workspace}, {"Events recorded", d.Count}, {"Recorded at", d.Time}}}
	if t.Account {
		p.Footer = "You received this account email because you use Xem."
		p.Details = []detail{{"Account", d.Email}, {"Requested at", d.Time}}
		if t.Key == "welcome" {
			p.Details = []detail{{"Name", d.Name}, {"Workspace", d.Workspace}}
		}
		if t.Key == "password_changed" {
			p.Details[1].Label = "Changed at"
		}
	}
	return p
}
func (t ServiceTemplate) Render(data ServiceData, link string) (string, string) {
	return t.renderWith(t.presentation(data), link)
}
func (t ServiceTemplate) Design() []byte {
	return t.designWith(t.presentation(ServiceData{Name: "{{first_name}}", Email: "{{account_email}}", Workspace: "{{workspace_name}}", Time: "{{event_time}}", Count: "{{event_count}}"}), "https://app.xem.email"+t.Path)
}
