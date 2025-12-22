package router

import (
	"testing"

	"jan-server/services/llm-api/internal/domain/model"
)

func TestRoundRobinRouterEmpty(t *testing.T) {
	r := NewRoundRobinRouter()
	if _, err := r.NextEndpoint("prov-1", nil); err != model.ErrNoEndpoints {
		t.Fatalf("expected ErrNoEndpoints, got %v", err)
	}
}

func TestRoundRobinRouterDistribution(t *testing.T) {
	r := NewRoundRobinRouter()
	endpoints := model.EndpointList{
		{URL: "http://a:8101/v1", Healthy: true},
		{URL: "http://b:8101/v1", Healthy: true},
		{URL: "http://c:8101/v1", Healthy: true},
	}

	counts := map[string]int{}
	for i := 0; i < 9; i++ {
		url, err := r.NextEndpoint("prov-1", endpoints)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		counts[url]++
	}

	for url, count := range counts {
		if count != 3 {
			t.Fatalf("expected 3 selections for %s, got %d", url, count)
		}
	}
}

func TestRoundRobinRouterUnhealthyFallback(t *testing.T) {
	r := NewRoundRobinRouter()
	endpoints := model.EndpointList{
		{URL: "http://a:8101/v1", Healthy: false},
		{URL: "http://b:8101/v1", Healthy: false},
	}

	url, err := r.NextEndpoint("prov-1", endpoints)
	if err != model.ErrNoHealthyEndpoints {
		t.Fatalf("expected ErrNoHealthyEndpoints, got %v", err)
	}
	if url != endpoints[0].URL {
		t.Fatalf("expected fallback to first endpoint, got %s", url)
	}
}
