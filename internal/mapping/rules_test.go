package mapping

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		content := `
splitwise:
  group_id: 12345
  my_user_id: 111
members:
  - splitwise_id: 111
    default_entity: SCOTIA
  - splitwise_id: 222
    default_entity: AMEX
`
		cfg, err := LoadConfig(writeTempConfig(t, content))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Splitwise.GroupID != 12345 {
			t.Errorf("GroupID = %d, want 12345", cfg.Splitwise.GroupID)
		}
		if cfg.Splitwise.MyUserID != 111 {
			t.Errorf("MyUserID = %d, want 111", cfg.Splitwise.MyUserID)
		}
		if len(cfg.Members) != 2 {
			t.Fatalf("len(Members) = %d, want 2", len(cfg.Members))
		}
		if cfg.Members[0].SplitwiseID != 111 || cfg.Members[0].DefaultEntity != "SCOTIA" {
			t.Errorf("Members[0] = %+v", cfg.Members[0])
		}
		if cfg.Members[1].SplitwiseID != 222 || cfg.Members[1].DefaultEntity != "AMEX" {
			t.Errorf("Members[1] = %+v", cfg.Members[1])
		}
	})

	t.Run("missing file returns error", func(t *testing.T) {
		_, err := LoadConfig("/nonexistent/path/config.yaml")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("invalid YAML returns error", func(t *testing.T) {
		_, err := LoadConfig(writeTempConfig(t, "{invalid yaml ["))
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("group_id not a number returns error", func(t *testing.T) {
		content := `
splitwise:
  group_id: not_a_number
  my_user_id: 111
`
		_, err := LoadConfig(writeTempConfig(t, content))
		if err == nil {
			t.Error("expected error for non-numeric group_id, got nil")
		}
	})

	t.Run("empty members list is valid", func(t *testing.T) {
		content := `
splitwise:
  group_id: 1
  my_user_id: 1
members: []
`
		cfg, err := LoadConfig(writeTempConfig(t, content))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(cfg.Members) != 0 {
			t.Errorf("expected empty members, got %d", len(cfg.Members))
		}
	})
}

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "config*.yaml")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	defer f.Close()
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	return f.Name()
}
