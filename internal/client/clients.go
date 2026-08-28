// Package client lazily constructs the GCP Live Stream and Transcoder API
// clients using Application Default Credentials (ADC). No credentials are
// handled explicitly: the underlying libraries discover them from
// GOOGLE_APPLICATION_CREDENTIALS, gcloud's application-default login, or the
// attached service account on GCP.
package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	livestream "cloud.google.com/go/video/livestream/apiv1"
	transcoder "cloud.google.com/go/video/transcoder/apiv1"
	"google.golang.org/api/transport"
)

// LiveStream returns an ADC-authenticated Live Stream API client.
func LiveStream(ctx context.Context) (*livestream.Client, error) {
	c, err := livestream.NewClient(ctx)
	if err != nil {
		return nil, wrapAuth(err)
	}
	return c, nil
}

// Transcoder returns an ADC-authenticated Transcoder API client.
func Transcoder(ctx context.Context) (*transcoder.Client, error) {
	c, err := transcoder.NewClient(ctx)
	if err != nil {
		return nil, wrapAuth(err)
	}
	return c, nil
}

// ADCEmail resolves the identity behind the current ADC, best-effort.
// Returns the credential's associated email or an empty string when it cannot
// be determined (e.g. user credentials without an email claim).
func ADCEmail(ctx context.Context) (string, error) {
	creds, err := transport.Creds(ctx)
	if err != nil {
		return "", wrapAuth(err)
	}
	if creds == nil {
		return "", errors.New("no application default credentials found")
	}
	// JSON key files expose the client_email; user creds usually do not.
	if len(creds.JSON) > 0 {
		if email := extractJSONField(creds.JSON, "client_email"); email != "" {
			return email, nil
		}
	}
	return "", nil
}

// wrapAuth adds a clear hint when credentials are missing.
func wrapAuth(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	if strings.Contains(msg, "could not find default credentials") ||
		strings.Contains(msg, "default credentials") {
		return fmt.Errorf("%w\n\nNo Application Default Credentials found. Run:\n  gcloud auth application-default login\nor set GOOGLE_APPLICATION_CREDENTIALS to a service-account key file", err)
	}
	return err
}

// extractJSONField pulls a top-level string field out of a small JSON blob
// without a full unmarshal target struct.
func extractJSONField(b []byte, field string) string {
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return ""
	}
	if v, ok := m[field].(string); ok {
		return v
	}
	return ""
}
