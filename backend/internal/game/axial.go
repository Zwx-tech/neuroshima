package game

import (
	"neuroshima/internal/models"
)

type Hex = models.Hex

// Distance returns the Manhattan distance between two hexes.
func Distance(h1, h2 Hex) int {
	return (abs(h1.Q-h2.Q) + abs(h1.Q+h1.R-h2.Q-h2.R) + abs(h1.R-h2.R)) / 2
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// HexDirections defines the 6 neighbors offsets in clockwise order:
// 0: NE (Up-Right), 1: E (Right), 2: SE (Down-Right),
// 3: SW (Down-Left), 4: W (Left), 5: NW (Up-Left)
var HexDirections = [6]Hex{
	{Q: 1, R: -1}, // 0: NE
	{Q: 1, R: 0},  // 1: E
	{Q: 0, R: 1},  // 2: SE
	{Q: -1, R: 1}, // 3: SW
	{Q: -1, R: 0}, // 4: W
	{Q: 0, R: -1}, // 5: NW
}

// GetNeighbor returns the neighbor of a hex in a given direction (0-5).
func GetNeighbor(h Hex, direction int) Hex {
	dir := HexDirections[direction%6]
	return Hex{Q: h.Q + dir.Q, R: h.R + dir.R}
}

// GetNeighbors returns all 6 neighbors of a hex.
func GetNeighbors(h Hex) []Hex {
	neighbors := make([]Hex, 6)
	for i := 0; i < 6; i++ {
		neighbors[i] = GetNeighbor(h, i)
	}
	return neighbors
}

// IsValidBoardHex checks if the hex is within the 19-hex board radius of 2.
func IsValidBoardHex(h Hex) bool {
	return max(abs(h.Q), abs(h.R), abs(h.Q+h.R)) <= 2
}

func max(a, b, c int) int {
	m := a
	if b > m {
		m = b
	}
	if c > m {
		m = c
	}
	return m
}

// RotateDirection converts a local direction index (0-5) to an absolute direction (0-5)
// based on the tile's current clockwise rotation (0-5).
func RotateDirection(localDir, rotation int) int {
	return (localDir + rotation) % 6
}
