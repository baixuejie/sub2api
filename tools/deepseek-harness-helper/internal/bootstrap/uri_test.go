package bootstrap

import "testing"

func TestParseLaunchURI(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		raw  string
		ok   bool
	}{
		{"https", "sub2api-harness://bootstrap?server=https%3A%2F%2Fexample.com&ticket=t-1&operation_id=op-1", true},
		{"localhost http", "sub2api-harness://bootstrap?server=http%3A%2F%2Flocalhost%3A8080&ticket=t-1&operation_id=op-1", true},
		{"loopback v4", "sub2api-harness://bootstrap?server=http%3A%2F%2F127.0.0.1%3A8080&ticket=t-1&operation_id=op-1", true},
		{"custom extension", "sub2api-harness://bootstrap?server=https%3A%2F%2Fexample.com&ticket=t-1&operation_id=op-1&extension_id=hermes", true},
		{"remote http rejected", "sub2api-harness://bootstrap?server=http%3A%2F%2Fexample.com&ticket=t-1&operation_id=op-1", false},
		{"wrong action", "sub2api-harness://other?server=https%3A%2F%2Fexample.com&ticket=t-1&operation_id=op-1", false},
		{"extra query", "sub2api-harness://bootstrap?server=https%3A%2F%2Fexample.com&ticket=t-1&operation_id=op-1&x=1", false},
		{"invalid extension", "sub2api-harness://bootstrap?server=https%3A%2F%2Fexample.com&ticket=t-1&operation_id=op-1&extension_id=Hermes%2Fshell", false},
		{"userinfo", "sub2api-harness://bootstrap?server=https%3A%2F%2Fu%3Ap%40example.com&ticket=t-1&operation_id=op-1", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ParseLaunchURI(test.raw)
			if (err == nil) != test.ok {
				t.Fatalf("ParseLaunchURI() error = %v, want ok=%v", err, test.ok)
			}
		})
	}
}

func TestValidateStatusURL(t *testing.T) {
	t.Parallel()
	good := "https://example.com/api/v1/deepseek-harness/sessions/op-1/events"
	if _, err := ValidateStatusURL(good, "https://example.com", "op-1", DefaultExtensionID); err != nil {
		t.Fatal(err)
	}
	for _, raw := range []string{
		"https://evil.example/api/v1/deepseek-harness/sessions/op-1/events",
		"https://example.com/api/v1/deepseek-harness/sessions/op-2/events",
		"https://example.com/api/v1/deepseek-harness/sessions/op-1/events?x=1",
	} {
		if _, err := ValidateStatusURL(raw, "https://example.com", "op-1", DefaultExtensionID); err == nil {
			t.Fatalf("expected rejection for %s", raw)
		}
	}
	if _, err := ValidateStatusURL("https://example.com/api/v1/hermes/sessions/op-1/events", "https://example.com", "op-1", "hermes"); err != nil {
		t.Fatal(err)
	}
}
