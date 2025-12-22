package model

import "testing"

func TestParseEndpointsCommaSeparated(t *testing.T) {
	endpoints, err := ParseEndpoints("http://a:8101/v1, http://b:8101/v1/ ,http://c:8101/v1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(endpoints) != 3 {
		t.Fatalf("expected 3 endpoints, got %d", len(endpoints))
	}
	if endpoints[0].URL != "http://a:8101/v1" || endpoints[1].URL != "http://b:8101/v1" || endpoints[2].URL != "http://c:8101/v1" {
		t.Fatalf("unexpected endpoints: %+v", endpoints)
	}
}

func TestParseEndpointsJSONArray(t *testing.T) {
	input := `[{"url":"http://a:8101/v1","weight":2},{"url":"http://b:8101/v1"}]`
	endpoints, err := ParseEndpoints(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(endpoints) != 2 {
		t.Fatalf("expected 2 endpoints, got %d", len(endpoints))
	}
	if endpoints[0].Weight != 2 || endpoints[1].Weight != 1 {
		t.Fatalf("unexpected weights: %+v", endpoints)
	}
}

func TestParseEndpointsSkipsInvalid(t *testing.T) {
	endpoints, err := ParseEndpoints("http://valid:8101/v1,,invalid-url,ftp://bad")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(endpoints) != 1 {
		t.Fatalf("expected 1 endpoint after filtering, got %d", len(endpoints))
	}
	if endpoints[0].URL != "http://valid:8101/v1" {
		t.Fatalf("unexpected endpoint: %+v", endpoints[0])
	}
}
