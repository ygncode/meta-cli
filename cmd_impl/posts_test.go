package cmd_impl

import (
	"testing"
)

func TestPostModifyCmd(t *testing.T) {
	tests := []struct {
		use   string
		short string
	}{
		{"update", "Update a post's message"},
		{"edit", "Edit a post's message"},
	}

	for _, tt := range tests {
		t.Run(tt.use, func(t *testing.T) {
			cmd := postModifyCmd(tt.use, tt.short)

			if cmd.Use != tt.use+" <post-id>" {
				t.Errorf("expected Use=%q, got %q", tt.use+" <post-id>", cmd.Use)
			}
			if cmd.Short != tt.short {
				t.Errorf("expected Short=%q, got %q", tt.short, cmd.Short)
			}

			f := cmd.Flags().Lookup("message")
			if f == nil {
				t.Fatal("expected --message flag")
			}
			if f.Shorthand != "m" {
				t.Errorf("expected -m shorthand, got %q", f.Shorthand)
			}
		})
	}
}

func TestPostUpdateAndEditRegistered(t *testing.T) {
	postCmd, _, err := rootCmd.Find([]string{"post"})
	if err != nil {
		t.Fatalf("post command not found: %v", err)
	}

	for _, name := range []string{"update", "edit"} {
		found := false
		for _, sub := range postCmd.Commands() {
			if sub.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected %q subcommand under post", name)
		}
	}
}
