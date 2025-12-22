// Package realtimeapi implements the realtime-api service which provides
// LiveKit-based real-time communication session management.
//
// The service provides:
//   - Session creation with LiveKit token generation
//   - Session lifecycle management (create, get, list, delete)
//   - LiveKit room synchronization via polling
//   - JWT authentication via Keycloak
//   - Optional API key authentication via Kong gateway
//
// For more information, see the README.md file.
package realtimeapi
