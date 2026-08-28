package transcoder

import "google.golang.org/protobuf/types/known/timestamppb"

// ts formats a proto timestamp as RFC3339, or "-" when nil.
func ts(t *timestamppb.Timestamp) string {
	if t == nil {
		return "-"
	}
	return t.AsTime().Format("2006-01-02T15:04:05Z07:00")
}
