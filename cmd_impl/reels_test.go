package cmd_impl

import (
	"testing"
)

func TestReelCreateHasFlags(t *testing.T) {
	cmd := reelCreateCmd()

	for _, flag := range []string{"video", "message", "title", "schedule", "tz"} {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("expected --%s flag on reel create command", flag)
		}
	}
}

func TestReelCommandRegistered(t *testing.T) {
	reelCmd, _, err := rootCmd.Find([]string{"reel"})
	if err != nil {
		t.Fatalf("reel command not found: %v", err)
	}

	found := false
	for _, sub := range reelCmd.Commands() {
		if sub.Name() == "create" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected create subcommand under reel")
	}
}
