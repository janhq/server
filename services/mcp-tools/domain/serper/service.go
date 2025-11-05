package serper

import "context"

// SerperClient defines the Serper API operations required by the domain layer
type SerperClient interface {
	Search(ctx context.Context, query SearchRequest) (*SearchResponse, error)
	FetchWebpage(ctx context.Context, query FetchWebpageRequest) (*FetchWebpageResponse, error)
}

// SerperService orchestrates Serper MCP operations while remaining transport-agnostic
type SerperService struct {
	client SerperClient
}

// NewSerperService creates a new Serper service
func NewSerperService(client SerperClient) *SerperService {
	return &SerperService{
		client: client,
	}
}

// Search performs a web search using Serper API
func (s *SerperService) Search(ctx context.Context, query SearchRequest) (*SearchResponse, error) {
	return s.client.Search(ctx, query)
}

// FetchWebpage scrapes a webpage using Serper API
func (s *SerperService) FetchWebpage(ctx context.Context, query FetchWebpageRequest) (*FetchWebpageResponse, error) {
	return s.client.FetchWebpage(ctx, query)
}
