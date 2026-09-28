package utils

import "testing"

func TestSMTPDestinationsExcludePrivateAndMetadataNetworks(t *testing.T) {
	for _, host := range []string{"127.0.0.1", "::1", "10.0.0.1", "192.168.0.1", "172.16.0.1", "169.254.169.254", "100.100.100.200", "198.18.0.1", "0.0.0.0", "::ffff:127.0.0.1", "fd00::1", "fe80::1", "not-an-ip"} {
		if publicSMTPAddress(host) {
			t.Errorf("allowed non-public destination %s", host)
		}
	}
	for _, host := range []string{"8.8.8.8", "2606:4700:4700::1111"} {
		if !publicSMTPAddress(host) {
			t.Errorf("rejected public destination %s", host)
		}
	}
}
