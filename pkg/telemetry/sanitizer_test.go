package telemetry

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSanitizer(t *testing.T) {
	tests := []struct {
		name     string
		level    PIILevel
		tenantID string
	}{
		{"none level", PIILevelNone, "tenant-123"},
		{"hashed level", PIILevelHashed, "tenant-456"},
		{"full level", PIILevelFull, "tenant-789"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSanitizer(tt.level, tt.tenantID)
			require.NotNil(t, s)
			assert.Equal(t, tt.level, s.level)
			assert.Equal(t, tt.tenantID, s.tenantSalt)
		})
	}
}

func TestSanitizePrompt_None(t *testing.T) {
	s := NewSanitizer(PIILevelNone, "tenant-123")
	result := s.SanitizePrompt("My email is john@example.com")
	assert.Equal(t, "[REDACTED]", result)
}

func TestSanitizePrompt_Full(t *testing.T) {
	s := NewSanitizer(PIILevelFull, "tenant-123")
	input := "My email is john@example.com"
	result := s.SanitizePrompt(input)
	assert.Equal(t, input, result)
}

func TestSanitizePrompt_Hashed_Email(t *testing.T) {
	s := NewSanitizer(PIILevelHashed, "tenant-123")
	result := s.SanitizePrompt("Contact me at john.doe@example.com for details")

	assert.NotContains(t, result, "john.doe@example.com")
	assert.Contains(t, result, "[EMAIL:")
	assert.Contains(t, result, "for details")
}

func TestSanitizePrompt_Hashed_Phone(t *testing.T) {
	s := NewSanitizer(PIILevelHashed, "tenant-123")

	tests := []struct {
		name  string
		input string
	}{
		{"dashes", "Call me at 555-123-4567"},
		{"dots", "Call me at 555.123.4567"},
		{"spaces", "Call me at 555 123 4567"},
		{"no separator", "Call me at 5551234567"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.SanitizePrompt(tt.input)
			assert.NotContains(t, result, "555")
			assert.NotContains(t, result, "123")
			assert.NotContains(t, result, "4567")
			assert.Contains(t, result, "[PHONE:")
			assert.Contains(t, result, "Call me at")
		})
	}
}

func TestSanitizePrompt_Hashed_SSN(t *testing.T) {
	s := NewSanitizer(PIILevelHashed, "tenant-123")
	result := s.SanitizePrompt("My SSN is 123-45-6789")

	assert.NotContains(t, result, "123-45-6789")
	assert.Contains(t, result, "[SSN:REDACTED]")
}

func TestSanitizePrompt_Hashed_CreditCard(t *testing.T) {
	s := NewSanitizer(PIILevelHashed, "tenant-123")

	tests := []struct {
		name  string
		input string
	}{
		{"spaces", "Card: 4532 1234 5678 9010"},
		{"dashes", "Card: 4532-1234-5678-9010"},
		{"no separator", "Card: 4532123456789010"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.SanitizePrompt(tt.input)
			assert.NotContains(t, result, "4532")
			assert.Contains(t, result, "[CC:REDACTED]")
		})
	}
}

func TestSanitizePrompt_Hashed_IPv4(t *testing.T) {
	s := NewSanitizer(PIILevelHashed, "tenant-123")
	result := s.SanitizePrompt("Server IP: 192.168.1.100")

	assert.NotContains(t, result, "192.168.1.100")
	assert.Contains(t, result, "[IP:")
}

func TestSanitizePrompt_Hashed_IPv6(t *testing.T) {
	s := NewSanitizer(PIILevelHashed, "tenant-123")
	result := s.SanitizePrompt("IPv6: 2001:0db8:85a3:0000:0000:8a2e:0370:7334")

	assert.NotContains(t, result, "2001:0db8")
	assert.Contains(t, result, "[IP:")
}

func TestSanitizeUserID(t *testing.T) {
	tests := []struct {
		name     string
		level    PIILevel
		userID   string
		expected string
	}{
		{"none level", PIILevelNone, "user-123", "[REDACTED]"},
		{"full level", PIILevelFull, "user-123", "user-123"},
		{"empty string", PIILevelHashed, "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSanitizer(tt.level, "tenant-123")
			result := s.SanitizeUserID(tt.userID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeUserID_Hashed(t *testing.T) {
	s := NewSanitizer(PIILevelHashed, "tenant-123")
	result := s.SanitizeUserID("user-456")

	assert.NotEqual(t, "user-456", result)
	assert.Len(t, result, 8) // Hash is 8 characters
}

func TestHash_Deterministic(t *testing.T) {
	s := NewSanitizer(PIILevelHashed, "tenant-123")

	hash1 := s.hash("test@example.com")
	hash2 := s.hash("test@example.com")

	assert.Equal(t, hash1, hash2, "Same input should produce same hash")
}

func TestHash_TenantSpecific(t *testing.T) {
	s1 := NewSanitizer(PIILevelHashed, "tenant-123")
	s2 := NewSanitizer(PIILevelHashed, "tenant-456")

	hash1 := s1.hash("test@example.com")
	hash2 := s2.hash("test@example.com")

	assert.NotEqual(t, hash1, hash2, "Different tenants should produce different hashes")
}

func TestSanitizeMetadata(t *testing.T) {
	s := NewSanitizer(PIILevelHashed, "tenant-123")

	metadata := map[string]string{
		"email": "user@example.com",
		"phone": "555-123-4567",
		"note":  "This is a normal note",
	}

	result := s.SanitizeMetadata(metadata)

	assert.NotNil(t, result)
	assert.Len(t, result, 3)
	assert.NotContains(t, result["email"], "user@example.com")
	assert.Contains(t, result["email"], "[EMAIL:")
	assert.Contains(t, result["note"], "This is a normal note")
}

func TestSanitizeMetadata_Nil(t *testing.T) {
	s := NewSanitizer(PIILevelHashed, "tenant-123")
	result := s.SanitizeMetadata(nil)
	assert.Nil(t, result)
}

func TestSanitizePrompt_MultiplePII(t *testing.T) {
	s := NewSanitizer(PIILevelHashed, "tenant-123")
	input := "Contact John at john@example.com or 555-123-4567. His SSN is 123-45-6789."
	result := s.SanitizePrompt(input)

	assert.NotContains(t, result, "john@example.com")
	assert.NotContains(t, result, "555-123-4567")
	assert.NotContains(t, result, "123-45-6789")
	assert.Contains(t, result, "[EMAIL:")
	assert.Contains(t, result, "[PHONE:")
	assert.Contains(t, result, "[SSN:REDACTED]")
	assert.Contains(t, result, "Contact John at")
}

func TestSanitizePrompt_UTF8(t *testing.T) {
	s := NewSanitizer(PIILevelHashed, "tenant-123")
	input := "Email: test@example.com 你好世界 🌍"
	result := s.SanitizePrompt(input)

	assert.NotContains(t, result, "test@example.com")
	assert.Contains(t, result, "你好世界")
	assert.Contains(t, result, "🌍")
}

func TestSanitizePrompt_EmptyString(t *testing.T) {
	s := NewSanitizer(PIILevelHashed, "tenant-123")
	result := s.SanitizePrompt("")
	assert.Equal(t, "", result)
}

func TestSanitizePrompt_LongInput(t *testing.T) {
	s := NewSanitizer(PIILevelHashed, "tenant-123")

	// Create a 10KB prompt
	input := ""
	for i := 0; i < 100; i++ {
		input += "This is a long text with email test@example.com repeated many times. "
	}

	result := s.SanitizePrompt(input)
	assert.NotContains(t, result, "test@example.com")
	assert.Greater(t, len(result), 0)
}

func BenchmarkSanitizePrompt(b *testing.B) {
	s := NewSanitizer(PIILevelHashed, "tenant-123")
	input := "Contact me at john@example.com or call 555-123-4567"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.SanitizePrompt(input)
	}
}

func BenchmarkSanitizePrompt_1KB(b *testing.B) {
	s := NewSanitizer(PIILevelHashed, "tenant-123")

	input := ""
	for i := 0; i < 10; i++ {
		input += "Contact me at john@example.com or call 555-123-4567. My SSN is 123-45-6789. "
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.SanitizePrompt(input)
	}
}
