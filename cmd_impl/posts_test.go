package cmd_impl

import (
	"testing"
	"time"
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

func TestPostListScheduledRegistered(t *testing.T) {
	postCmd, _, err := rootCmd.Find([]string{"post"})
	if err != nil {
		t.Fatalf("post command not found: %v", err)
	}

	found := false
	for _, sub := range postCmd.Commands() {
		if sub.Name() == "list-scheduled" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected list-scheduled subcommand under post")
	}
}

func TestPostCreateHasScheduleFlags(t *testing.T) {
	cmd := postCreateCmd()

	for _, flag := range []string{"schedule", "tz"} {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("expected --%s flag on create command", flag)
		}
	}
}

func TestPostCreateHasVideoFlags(t *testing.T) {
	cmd := postCreateCmd()

	for _, flag := range []string{"video", "title", "thumbnail"} {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("expected --%s flag on create command", flag)
		}
	}
}

func TestParseScheduleTime(t *testing.T) {
	t.Run("valid with timezone", func(t *testing.T) {
		result, err := parseScheduleTime("2026-03-20 14:00", "Asia/Yangon")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		loc, _ := time.LoadLocation("Asia/Yangon")
		expected := time.Date(2026, 3, 20, 14, 0, 0, 0, loc)
		if !result.Equal(expected) {
			t.Errorf("expected %v, got %v", expected, result)
		}
	})

	t.Run("valid with local timezone", func(t *testing.T) {
		result, err := parseScheduleTime("2026-03-20 14:00", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := time.Date(2026, 3, 20, 14, 0, 0, 0, time.Local)
		if !result.Equal(expected) {
			t.Errorf("expected %v, got %v", expected, result)
		}
	})

	t.Run("invalid timezone", func(t *testing.T) {
		_, err := parseScheduleTime("2026-03-20 14:00", "Invalid/Zone")
		if err == nil {
			t.Error("expected error for invalid timezone")
		}
	})

	t.Run("invalid datetime format", func(t *testing.T) {
		_, err := parseScheduleTime("not-a-date", "")
		if err == nil {
			t.Error("expected error for invalid datetime")
		}
	})
}
