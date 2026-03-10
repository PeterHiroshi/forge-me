package cmd

import (
	"testing"
)

func TestDashboardCommandExists(t *testing.T) {
	cmd := rootCmd
	found := false
	for _, c := range cmd.Commands() {
		if c.Use == "dashboard [account-id]" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'dashboard' command to be registered")
	}
}

func TestDashboardCommandHasRefreshFlag(t *testing.T) {
	cmd := rootCmd
	for _, c := range cmd.Commands() {
		if c.Name() == "dashboard" {
			flag := c.Flags().Lookup("refresh")
			if flag == nil {
				t.Error("expected --refresh flag on dashboard command")
			}
			if flag != nil && flag.DefValue != "30s" {
				t.Errorf("--refresh default = %q, want %q", flag.DefValue, "30s")
			}
			return
		}
	}
	t.Error("dashboard command not found")
}
