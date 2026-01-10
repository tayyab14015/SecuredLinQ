package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateRoomID(t *testing.T) {
	// Test that generateRoomID produces a 12-character alphanumeric string
	roomID := generateRoomID()

	assert.Len(t, roomID, 12)
	// Should be alphanumeric (no dashes)
	for _, char := range roomID {
		assert.True(t, (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9'))
	}

	// Test uniqueness - generate multiple IDs and ensure they're different
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := generateRoomID()
		assert.False(t, ids[id], "Generated duplicate room ID")
		ids[id] = true
	}
}

// Note: Tests for GetOrCreateMeetingRoom, GetMeetingRoomByRoomID, etc.
// require database integration. For integration tests, see TESTING.md
