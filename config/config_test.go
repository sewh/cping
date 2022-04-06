package config

import (
	"reflect"
	"testing"
)

func TestParseArgs(t *testing.T) {
	// test count
	countArgs := []string{"count", "10", "127.0.0.5"}
	c := Default()
	c.ParseArgs(countArgs)
	if c.Count != 10 {
		t.Fatalf("expected c.Count == 10, got %d", c.Count)
	}

	// test size
	sizeArgs := []string{"size", "60", "127.0.0.5"}
	c = Default()
	c.ParseArgs(sizeArgs)
	if c.Size != 60 {
		t.Fatalf("expected c.Size == 60, got %d", c.Size)
	}

	// test ipv4
	ipv4Args := []string{"ipv4", "127.0.0.5"}
	c = Default()
	c.IPVersion = 10
	c.ParseArgs(ipv4Args)
	if c.IPVersion != 4 {
		t.Fatalf("expected c.IPVersion == 4. got %d", c.IPVersion)
	}
	if c.DestIP != "127.0.0.5" {
		t.Fatalf("expected c.DestIP == 127.0.0.5, got %s", c.DestIP)
	}

	// test ipv6
	ipv6Args := []string{"ipv6", "::1"}
	c = Default()
	c.IPVersion = 10
	c.ParseArgs(ipv6Args)
	if c.IPVersion != 6 {
		t.Fatalf("expected c.IPVersion == 6. got %d", c.IPVersion)
	}
	if c.DestIP != "::1" {
		t.Fatalf("expected c.DestIP == ::1, got %s", c.DestIP)
	}

	// test payload
	payloadArgs := []string{"payload", "hello", "127.0.0.5"}
	c = Default()
	c.ParseArgs(payloadArgs)
	if !reflect.DeepEqual(c.Payload, []byte("hello")) {
		t.Fatalf("expected C.Payload == 'hello', got '%s'", c.Payload)
	}

	// test ttl
	ttlArgs := []string{"ttl", "60", "127.0.0.5"}
	c = Default()
	c.ParseArgs(ttlArgs)
	if c.TTL != 60 {
		t.Fatalf("expected c.TTL == 60, got %d", c.TTL)
	}

	// test ttl
	timeoutArgs := []string{"timeout", "20", "127.0.0.5"}
	c = Default()
	c.ParseArgs(timeoutArgs)
	if c.TimeoutSecs != 20 {
		t.Fatalf("expected c.TimeoutSecs == 20, got %d", c.TimeoutSecs)
	}

	// test full string
	fullArgs := []string{"count", "10", "size", "60", "payload", "hello", "127.0.0.5"}
	c = Default()
	c.ParseArgs(fullArgs)
	if c.Count != 10 {
		t.Fatalf("expected c.Count == 10, got %d", c.Count)
	}
	if c.Size != 60 {
		t.Fatalf("expected c.Size == 60, got %d", c.Size)
	}
	if !reflect.DeepEqual(c.Payload, []byte("hello")) {
		t.Fatalf("expected C.Payload == 'hello', got '%s'", c.Payload)
	}

	// test full string with abbreviations
	abbrvArgs := []string{"co", "10", "siz", "60", "payl", "hello", "127.0.0.5"}
	c = Default()
	c.ParseArgs(abbrvArgs)
	if c.Count != 10 {
		t.Fatalf("expected c.Count == 10, got %d", c.Count)
	}
	if c.Size != 60 {
		t.Fatalf("expected c.Size == 60, got %d", c.Size)
	}
	if !reflect.DeepEqual(c.Payload, []byte("hello")) {
		t.Fatalf("expected C.Payload == 'hello', got '%s'", c.Payload)
	}
}
