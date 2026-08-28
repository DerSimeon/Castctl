// Package parent builds GCP resource names of the form
// projects/{project}/locations/{location}/... so command code can accept
// short IDs and let this package assemble the fully-qualified names.
package parent

import (
	"fmt"
	"strings"
)

// Location returns "projects/{project}/locations/{location}".
func Location(project, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, location)
}

// Resource joins a location parent with a collection and id.
// e.g. Resource(p, l, "channels", "my-ch") ->
//
//	projects/p/locations/l/channels/my-ch
//
// If id is already a fully-qualified name (contains "projects/"), it is
// returned unchanged so users may paste full names.
func Resource(project, location, collection, id string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/%s/%s", Location(project, location), collection, id)
}

// Child builds a name nested under a parent resource, e.g. events under a
// channel: Child(channelName, "events", "evt-1").
func Child(parentName, collection, id string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/%s/%s", parentName, collection, id)
}

// LastSegment returns the trailing id of a resource name.
// projects/p/locations/l/channels/my-ch -> my-ch
func LastSegment(name string) string {
	i := strings.LastIndex(name, "/")
	if i < 0 {
		return name
	}
	return name[i+1:]
}
