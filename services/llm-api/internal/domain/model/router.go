package model

import "errors"

var (
	ErrNoEndpoints        = errors.New("no endpoints available")
	ErrNoHealthyEndpoints = errors.New("no healthy endpoints available")
)

// EndpointRouter selects the next endpoint for a request.
type EndpointRouter interface {
	// NextEndpoint returns the next endpoint URL for the given provider.
	// Returns ErrNoEndpoints if the slice is empty, ErrNoHealthyEndpoints if none are healthy.
	NextEndpoint(providerID string, endpoints EndpointList) (string, error)

	// Reset clears router state.
	Reset()
}
