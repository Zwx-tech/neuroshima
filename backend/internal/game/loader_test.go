package game

import (
	"testing"
)

func TestEmbeddedRegistry(t *testing.T) {
	// Verify armies exist in registry
	armies := []string{"moloch", "borgo", "posterunek", "hegemonia"}
	for _, army := range armies {
		blueprints, ok := registry.Armies[army]
		if !ok {
			t.Errorf("Expected army %q to exist in registry", army)
			continue
		}
		if len(blueprints) == 0 {
			t.Errorf("Expected army %q to have blueprints, found 0", army)
		}
	}
}

func TestBuildDeck(t *testing.T) {
	owner := "player-1"
	deck, err := BuildDeck("borgo", owner)
	if err != nil {
		t.Fatalf("Failed to build deck for Borgo: %v", err)
	}

	// Calculate expected counts based on our armies.json for borgo:
	// bor_hq: 1
	// bor_mutek: 6
	// bor_sieciarz: 2
	// bor_zabojca: 2
	// bor_battle: 6
	// bor_move: 4
	// bor_grenade: 1
	// Total = 1+6+2+2+6+4+1 = 22
	expectedCount := 22
	if len(deck) != expectedCount {
		t.Errorf("Expected deck size for Borgo to be %d, got %d", expectedCount, len(deck))
	}

	// Verify instances are set correctly
	for _, inst := range deck {
		if inst.OwnerID != owner {
			t.Errorf("Expected OwnerID to be %q, got %q", owner, inst.OwnerID)
		}
		if inst.CurrentHP != inst.Blueprint.BaseHP {
			t.Errorf("Expected CurrentHP to be %d, got %d", inst.Blueprint.BaseHP, inst.CurrentHP)
		}
	}
}
