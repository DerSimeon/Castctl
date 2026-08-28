package parent

import "testing"

func TestResource(t *testing.T) {
	tests := []struct {
		name                              string
		project, location, coll, id, want string
	}{
		{"short id", "p", "us-central1", "channels", "ch1",
			"projects/p/locations/us-central1/channels/ch1"},
		{"full name passthrough", "p", "us-central1", "channels",
			"projects/other/locations/eu/channels/x",
			"projects/other/locations/eu/channels/x"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Resource(tt.project, tt.location, tt.coll, tt.id); got != tt.want {
				t.Errorf("Resource() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestChild(t *testing.T) {
	parentName := "projects/p/locations/l/channels/ch"
	if got := Child(parentName, "events", "e1"); got != parentName+"/events/e1" {
		t.Errorf("Child() = %q", got)
	}
}

func TestLastSegment(t *testing.T) {
	if got := LastSegment("projects/p/locations/l/channels/ch"); got != "ch" {
		t.Errorf("LastSegment() = %q, want ch", got)
	}
	if got := LastSegment("bare"); got != "bare" {
		t.Errorf("LastSegment(bare) = %q", got)
	}
}
