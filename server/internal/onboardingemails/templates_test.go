package onboardingemails

import (
	"html"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMilestoneRenderingEscapesWorkspaceData(t *testing.T) {
	workspace := `Studio <img src=x onerror="alert(1)"> & Friends`
	domain := strings.Repeat("long-label-", 10) + "example.net"
	link := "https://mail.company.example:8443/settings/sending?view=setup&step=next"
	for _, template := range Templates {
		t.Run(template.Key, func(t *testing.T) {
			body, plain := template.Render(workspace, domain, link)
			require.Contains(t, body, html.EscapeString(workspace))
			require.NotContains(t, body, "<img src=x")
			require.Contains(t, body, `href="`+html.EscapeString(link)+`"`)
			require.Contains(t, body, `src="https://mail.company.example:8443/assets/template-starters/brand/xem-mark.png"`)
			require.NotContains(t, body, "https://app.xem.email")
			require.Contains(t, plain, workspace)
			require.Contains(t, plain, domain)
			require.Contains(t, plain, link)
		})
	}
}

func TestBrandAssetURLsUseOnlyDashboardOrigin(t *testing.T) {
	require.Equal(t, "https://mail.company.example/assets/template-starters/brand/fonts.css", brandAsset("https://user:pass@mail.company.example/settings?token=private#step", "fonts.css"))
	for _, input := range []string{"javascript:alert(1)", "//example.net", "/settings/sending", "https://"} {
		require.Empty(t, brandAsset(input, "fonts.css"))
	}
}

func TestServiceTemplatesEscapeValuesAndKeepResetURLSeparate(t *testing.T) {
	for _, template := range ServiceTemplates {
		t.Run(template.Key, func(t *testing.T) {
			body, plain := template.Render(ServiceData{Name: `<img onerror="bad">`, Email: `<script>bad</script>`, Workspace: `Studio <b>bad</b>`, Time: "30 Sep 2026, 12:00 UTC", Count: "3"}, "https://mail.example.net/auth/reset-password/private-token")
			require.NotContains(t, body, "<script>")
			require.NotContains(t, body, "<img onerror")
			require.NotContains(t, body, "Studio <b>")
			require.Contains(t, body, "Built with Xem")
			require.Contains(t, plain, "https://mail.example.net/")
			require.NotContains(t, string(template.Design()), "private-token")
			require.Contains(t, body, `src="https://mail.example.net/assets/template-starters/brand/xem-mark.png"`)
		})
	}
}
