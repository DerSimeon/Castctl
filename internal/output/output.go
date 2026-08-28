// Package output renders results either as human-readable tables or as JSON.
// Proto messages are marshaled with protojson so JSON output round-trips back
// into create/update inputs.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var protoJSON = protojson.MarshalOptions{
	Multiline:       true,
	Indent:          "  ",
	UseProtoNames:   true,
	EmitUnpopulated: false,
}

// Column defines a table column header and a per-row value extractor.
type Column[T any] struct {
	Header string
	Value  func(T) string
}

// JSONProto writes a single proto message as indented JSON to stdout.
func JSONProto(m proto.Message) error {
	b, err := protoJSON.Marshal(m)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(os.Stdout, string(b))
	return err
}

// JSONProtoList writes a slice of proto messages as a JSON array.
func JSONProtoList[T proto.Message](items []T) error {
	parts := make([]json.RawMessage, 0, len(items))
	for _, it := range items {
		b, err := protoJSON.Marshal(it)
		if err != nil {
			return err
		}
		parts = append(parts, b)
	}
	b, err := json.MarshalIndent(parts, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(os.Stdout, string(b))
	return err
}

// JSONValue writes any Go value as indented JSON (for non-proto results).
func JSONValue(v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(os.Stdout, string(b))
	return err
}

// Table renders rows using the provided columns to stdout.
func Table[T any](items []T, cols []Column[T]) error {
	return tableTo(os.Stdout, items, cols)
}

func tableTo[T any](w io.Writer, items []T, cols []Column[T]) error {
	tw := tabwriter.NewWriter(w, 0, 2, 3, ' ', 0)
	headers := make([]string, len(cols))
	for i, c := range cols {
		headers[i] = strings.ToUpper(c.Header)
	}
	if _, err := fmt.Fprintln(tw, strings.Join(headers, "\t")); err != nil {
		return err
	}
	for _, it := range items {
		cells := make([]string, len(cols))
		for i, c := range cols {
			cells[i] = c.Value(it)
		}
		if _, err := fmt.Fprintln(tw, strings.Join(cells, "\t")); err != nil {
			return err
		}
	}
	return tw.Flush()
}

// Empty prints a friendly "no resources" line unless JSON mode is requested.
func Empty(kind string, jsonMode bool) error {
	if jsonMode {
		_, err := fmt.Fprintln(os.Stdout, "[]")
		return err
	}
	_, err := fmt.Fprintf(os.Stdout, "No %s found.\n", kind)
	return err
}
