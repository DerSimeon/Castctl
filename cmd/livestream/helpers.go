package livestream

import (
	"fmt"
	"sort"
	"strings"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// errNeedSpecOr is returned when create is invoked with neither -f nor an
// alternative convenience flag.
func errNeedSpecOr(alt string) error {
	return fmt.Errorf("provide a spec with -f/--file (or %s)", alt)
}

// errBadEnum reports an invalid enum value and lists the valid names.
func errBadEnum(field, got string, names map[int32]string) error {
	valid := make([]string, 0, len(names))
	for k, v := range names {
		if k == 0 {
			continue // skip *_UNSPECIFIED
		}
		valid = append(valid, v)
	}
	sort.Strings(valid)
	return fmt.Errorf("invalid %s %q; valid: %s", field, got, strings.Join(valid, ", "))
}

// splitCSV splits a comma-separated flag value, trimming spaces and empties.
func splitCSV(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// ts formats a proto timestamp as RFC3339, or "-" when nil.
func ts(t *timestamppb.Timestamp) string {
	if t == nil {
		return "-"
	}
	return t.AsTime().Format("2006-01-02T15:04:05Z07:00")
}
