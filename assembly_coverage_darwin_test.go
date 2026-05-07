//go:build darwin

package assembly

import (
	"os"
	"testing"
)

// ============================================================================
// goString — cover empty string (ptr == 0)
// ============================================================================

func TestGoString_Empty(t *testing.T) {
	if goString(0) != "" {
		t.Fatal("goString(0) should return empty string")
	}
}

// ============================================================================
// getEntrypoint — verify darwin path
// ============================================================================

func TestGetEntrypoint_CurrentProcess(t *testing.T) {
	path, err := os.Executable()
	if err != nil {
		t.Fatalf("Executable: %v", err)
	}
	entry, err := getEntrypoint(path)
	if err != nil {
		t.Fatalf("getEntrypoint: %v", err)
	}
	if entry == 0 {
		t.Fatal("getEntrypoint: expected non-zero entry")
	}
}
