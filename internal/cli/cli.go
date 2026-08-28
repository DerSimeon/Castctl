// Package cli holds shared runtime state and helpers used across command
// packages: the resolved Settings for the current invocation, input-spec
// reading (-f / stdin), and a small spinner for long-running operations.
package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/simeon/castctl/internal/config"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// Current holds the settings resolved in the root PersistentPreRunE, so leaf
// commands need not thread them through every call.
var Current config.Settings

// ReadSpec returns the bytes of an input spec: from file when path != "",
// otherwise from stdin. Used by create/update commands that take -f.
func ReadSpec(path string) ([]byte, error) {
	if path == "" || path == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(path)
}

// Infof prints a status line to stderr (kept off stdout so --json stays clean).
func Infof(format string, a ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", a...)
}

// UnmarshalSpec parses a resource spec (protojson) from -f/stdin into m.
// It tolerates unknown fields so specs produced by other tools still load.
func UnmarshalSpec(path string, m proto.Message) error {
	data, err := ReadSpec(path)
	if err != nil {
		return err
	}
	opts := protojson.UnmarshalOptions{DiscardUnknown: true}
	return opts.Unmarshal(data, m)
}
