package config

import "testing"

func TestResolvePrecedence(t *testing.T) {
	// Flag beats env; empty flag falls through to env.
	t.Setenv("GOOGLE_CLOUD_PROJECT", "env-proj")
	t.Setenv("CASTCTL_LOCATION", "env-loc")

	// Point HOME at an empty dir so no config file interferes.
	t.Setenv("HOME", t.TempDir())
	t.Setenv("USERPROFILE", t.TempDir())

	s, err := Resolve("flag-proj", "", true, false)
	if err != nil {
		t.Fatal(err)
	}
	if s.Project != "flag-proj" {
		t.Errorf("project: flag should win, got %q", s.Project)
	}
	if s.Location != "env-loc" {
		t.Errorf("location: env should win when no flag, got %q", s.Location)
	}
	if !s.JSON {
		t.Error("json flag not carried")
	}
}

func TestRequireProjectLocation(t *testing.T) {
	if err := (Settings{}).RequireProjectLocation(); err == nil {
		t.Error("expected error when project missing")
	}
	if err := (Settings{Project: "p"}).RequireProjectLocation(); err == nil {
		t.Error("expected error when location missing")
	}
	if err := (Settings{Project: "p", Location: "l"}).RequireProjectLocation(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
