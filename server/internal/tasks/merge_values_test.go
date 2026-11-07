package tasks

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBlankCampaignValuesKeepContactFields(t *testing.T) {
	vars := map[string]string{"first_name": "Ava", "company": ""}
	mergeCampaignValues(vars, map[string]string{"first_name": "", "company": " ", "promo": "", "plan": "Pro"})
	require.Equal(t, "Ava", vars["first_name"], "a blank campaign value keeps the contact's name")
	require.Equal(t, " ", vars["company"], "an empty contact field takes the campaign value")
	require.Equal(t, "", vars["promo"], "campaign-only variables are still set, even blank")
	require.Equal(t, "Pro", vars["plan"])

	vars = map[string]string{"first_name": "Ava"}
	mergeCampaignValues(vars, map[string]string{"first_name": "Friend"})
	require.Equal(t, "Friend", vars["first_name"], "a filled campaign value still overrides")
}
