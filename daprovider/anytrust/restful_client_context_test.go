// Copyright 2021-2026, Offchain Labs, Inc.
// For license information, see https://github.com/OffchainLabs/nitro/blob/master/LICENSE.md

package anytrust

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRestfulClientCanceledContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"expirationPolicy":"keep-forever"}`))
	}))
	defer server.Close()
	client, err := NewRestfulClientFromURL(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := client.HealthCheck(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("HealthCheck = %v; want context.Canceled", err)
	}
	if _, err := client.ExpirationPolicy(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("ExpirationPolicy = %v; want context.Canceled", err)
	}
}
