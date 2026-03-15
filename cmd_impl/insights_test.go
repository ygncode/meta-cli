package cmd_impl

import "testing"

func TestInsightCommandsRegistered(t *testing.T) {
	insightCmd, _, err := rootCmd.Find([]string{"insight"})
	if err != nil {
		t.Fatalf("insight command not found: %v", err)
	}

	for _, name := range []string{"page", "post"} {
		found := false
		for _, sub := range insightCmd.Commands() {
			if sub.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected %q subcommand under insight", name)
		}
	}
}

func TestInsightPageHasFlags(t *testing.T) {
	cmd := insightPageCmd()

	for _, flag := range []string{"metric", "period"} {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("expected --%s flag on page command", flag)
		}
	}
}

func TestInsightPostHasFlags(t *testing.T) {
	cmd := insightPostCmd()

	if cmd.Flags().Lookup("metric") == nil {
		t.Error("expected --metric flag on post command")
	}

	// Verify it requires exactly 1 arg
	if err := cmd.Args(cmd, []string{}); err == nil {
		t.Error("expected error with no args")
	}
	if err := cmd.Args(cmd, []string{"post_id"}); err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}
	if err := cmd.Args(cmd, []string{"a", "b"}); err == nil {
		t.Error("expected error with 2 args")
	}
}
