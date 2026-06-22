package game

import (
	"testing"
)

func TestDistance(t *testing.T) {
	tests := []struct {
		name string
		h1   Hex
		h2   Hex
		want int
	}{
		{"Center to Center", Hex{Q: 0, R: 0}, Hex{Q: 0, R: 0}, 0},
		{"Adjacent NE", Hex{Q: 0, R: 0}, Hex{Q: 1, R: -1}, 1},
		{"Adjacent E", Hex{Q: 0, R: 0}, Hex{Q: 1, R: 0}, 1},
		{"Opposite Ends", Hex{Q: -2, R: 2}, Hex{Q: 2, R: -2}, 4},
		{"Ring 2 to Ring 2 diagonal", Hex{Q: 0, R: -2}, Hex{Q: 0, R: 2}, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Distance(tt.h1, tt.h2)
			if got != tt.want {
				t.Errorf("Distance(%v, %v) = %d; want %d", tt.h1, tt.h2, got, tt.want)
			}
		})
	}
}

func TestGetNeighbors(t *testing.T) {
	h := Hex{Q: 0, R: 0}
	neighbors := GetNeighbors(h)

	if len(neighbors) != 6 {
		t.Fatalf("Expected 6 neighbors, got %d", len(neighbors))
	}

	expected := []Hex{
		{Q: 1, R: -1}, // 0: NE
		{Q: 1, R: 0},  // 1: E
		{Q: 0, R: 1},  // 2: SE
		{Q: -1, R: 1}, // 3: SW
		{Q: -1, R: 0}, // 4: W
		{Q: 0, R: -1}, // 5: NW
	}

	for i, want := range expected {
		if neighbors[i] != want {
			t.Errorf("Neighbor %d = %v; want %v", i, neighbors[i], want)
		}
	}
}

func TestIsValidBoardHex(t *testing.T) {
	tests := []struct {
		name string
		h    Hex
		want bool
	}{
		{"Center is valid", Hex{Q: 0, R: 0}, true},
		{"Ring 1 E is valid", Hex{Q: 1, R: 0}, true},
		{"Ring 2 NE is valid", Hex{Q: 2, R: -2}, true},
		{"Out of bounds Q", Hex{Q: 3, R: 0}, false},
		{"Out of bounds R", Hex{Q: 0, R: -3}, false},
		{"Out of bounds Q+R", Hex{Q: 2, R: 1}, false}, // 2 + 1 = 3 > 2
		{"Valid corner case", Hex{Q: -2, R: 0}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidBoardHex(tt.h)
			if got != tt.want {
				t.Errorf("IsValidBoardHex(%v) = %v; want %v", tt.h, got, tt.want)
			}
		})
	}
}

func TestRotateDirection(t *testing.T) {
	tests := []struct {
		name     string
		local    int
		rotation int
		want     int
	}{
		{"No rotation", 1, 0, 1},
		{"Rotate 1", 1, 1, 2},
		{"Rotate 5 wrap", 3, 4, 1}, // (3 + 4) % 6 = 1
		{"Wrap boundary", 5, 1, 0}, // (5 + 1) % 6 = 0
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RotateDirection(tt.local, tt.rotation)
			if got != tt.want {
				t.Errorf("RotateDirection(%d, %d) = %d; want %d", tt.local, tt.rotation, got, tt.want)
			}
		})
	}
}
