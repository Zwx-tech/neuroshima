package game

import (
	"neuroshima/internal/models"
	"testing"
)

func TestBattleSimultaneousKills(t *testing.T) {
	p1 := "player-1"
	p2 := "player-2"

	engine, err := NewEngine("test-game", p1, "borgo", p2, "moloch")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// Move out of HQ Placement phase by placing HQs
	_ = engine.PlaceHQ(p1, Hex{Q: 0, R: -2}, 0)
	_ = engine.PlaceHQ(p2, Hex{Q: 0, R: 2}, 0)

	// Place two adjacent units attacking each other at Initiative 2:
	// Borgo Mutek at (0, 0) rotated 1 (Melee attack in local 0, now absolute 1 pointing East to (1, 0))
	mutek, err := GetBlueprint("borgo", "bor_mutek")
	if err != nil {
		t.Fatalf("Failed to get blueprint: %v", err)
	}
	mutekInst := &models.TileInstance{
		InstanceID:    "t-mutek",
		Blueprint:     mutek,
		OwnerID:       p1,
		Hex:           Hex{Q: 0, R: 0},
		Rotation:      1,
		IsPlaced:      true,
		CurrentHP:     mutek.BaseHP,
		ModInitiative: mutek.Initiative,
	}
	engine.State.Board[Hex{Q: 0, R: 0}] = mutekInst

	// Moloch Cyborg at (1, 0) rotated 4 (Ranged attack in local 0, now absolute 4 pointing West to (0, 0))
	cyborg, err := GetBlueprint("moloch", "mol_cyborg")
	if err != nil {
		t.Fatalf("Failed to get blueprint: %v", err)
	}
	cyborgInst := &models.TileInstance{
		InstanceID:    "t-cyborg",
		Blueprint:     cyborg,
		OwnerID:       p2,
		Hex:           Hex{Q: 1, R: 0},
		Rotation:      4,
		IsPlaced:      true,
		CurrentHP:     cyborg.BaseHP,
		ModInitiative: cyborg.Initiative,
	}
	engine.State.Board[Hex{Q: 1, R: 0}] = cyborgInst

	// Resolve Battle
	events := engine.ResolveBattle()

	// Both have 1 HP and Initiative 2, and attack each other.
	// They should deal damage to each other simultaneously and both die.
	if engine.State.Board[Hex{Q: 0, R: 0}] != nil {
		t.Error("Expected Mutek to be killed and removed from board")
	}
	if engine.State.Board[Hex{Q: 1, R: 0}] != nil {
		t.Error("Expected Cyborg to be killed and removed from board")
	}

	// Verify events log contains damage and deaths
	hasBlock := false
	hasDamageMutek := false
	hasDamageCyborg := false
	hasDeathMutek := false
	hasDeathCyborg := false

	for _, ev := range events {
		if ev.Type == "BLOCK" {
			hasBlock = true
		}
		if ev.Type == "DAMAGE" && ev.TargetHex == (Hex{Q: 0, R: 0}) {
			hasDamageMutek = true
		}
		if ev.Type == "DAMAGE" && ev.TargetHex == (Hex{Q: 1, R: 0}) {
			hasDamageCyborg = true
		}
		if ev.Type == "DEATH" && ev.TargetHex == (Hex{Q: 0, R: 0}) {
			hasDeathMutek = true
		}
		if ev.Type == "DEATH" && ev.TargetHex == (Hex{Q: 1, R: 0}) {
			hasDeathCyborg = true
		}
	}

	if hasBlock {
		t.Error("Did not expect any attacks to be blocked by armor")
	}
	if !hasDamageMutek || !hasDamageCyborg {
		t.Error("Expected both units to receive damage")
	}
	if !hasDeathMutek || !hasDeathCyborg {
		t.Error("Expected both units to die")
	}
}

func TestBattleArmorDefense(t *testing.T) {
	p1 := "player-1"
	p2 := "player-2"

	engine, err := NewEngine("test-game", p1, "borgo", p2, "moloch")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	_ = engine.PlaceHQ(p1, Hex{Q: 0, R: -2}, 0)
	_ = engine.PlaceHQ(p2, Hex{Q: 0, R: 2}, 0)

	// Borgo Mutek at (0, 0) rotated 1 (Attacks East to (1, 0) at Initiative 2)
	mutek, _ := GetBlueprint("borgo", "bor_mutek")
	mutekInst := &models.TileInstance{
		InstanceID:    "t-mutek",
		Blueprint:     mutek,
		OwnerID:       p1,
		Hex:           Hex{Q: 0, R: 0},
		Rotation:      1,
		IsPlaced:      true,
		CurrentHP:     mutek.BaseHP,
		ModInitiative: mutek.Initiative,
	}
	engine.State.Board[Hex{Q: 0, R: 0}] = mutekInst

	// Moloch Bloker at (1, 0) (Has 360-degree armor in base blueprint, Initiative list is empty so it won't attack)
	bloker, _ := GetBlueprint("moloch", "mol_bloker")
	blokerInst := &models.TileInstance{
		InstanceID:    "t-bloker",
		Blueprint:     bloker,
		OwnerID:       p2,
		Hex:           Hex{Q: 1, R: 0},
		Rotation:      0,
		IsPlaced:      true,
		CurrentHP:     bloker.BaseHP,
		ModInitiative: bloker.Initiative,
	}
	engine.State.Board[Hex{Q: 1, R: 0}] = blokerInst

	// Resolve Battle
	events := engine.ResolveBattle()

	// Bloker has armor, so the Mutek attack should be blocked. Bloker should take 0 damage.
	if blokerInst.CurrentHP != bloker.BaseHP {
		t.Errorf("Expected Bloker HP to remain %d, got %d", bloker.BaseHP, blokerInst.CurrentHP)
	}

	hasBlock := false
	for _, ev := range events {
		if ev.Type == "BLOCK" && ev.TargetHex == (Hex{Q: 1, R: 0}) {
			hasBlock = true
		}
	}
	if !hasBlock {
		t.Error("Expected to find a BLOCK event for the Bloker tile")
	}
}

func TestGaussCannonFiringLineDamage(t *testing.T) {
	p1 := "player-1"
	p2 := "player-2"

	engine, err := NewEngine("test-game", p1, "moloch", p2, "borgo")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	_ = engine.PlaceHQ(p1, Hex{Q: 0, R: -2}, 0)
	_ = engine.PlaceHQ(p2, Hex{Q: 2, R: -2}, 0)

	// Moloch Gauss Cannon at (0, 0) rotated 2 (Ranged attack in absolute direction 2 pointing South-East)
	gauss, _ := GetBlueprint("moloch", "mol_gauss")
	gaussInst := &models.TileInstance{
		InstanceID:    "t-gauss",
		Blueprint:     gauss,
		OwnerID:       p1,
		Hex:           Hex{Q: 0, R: 0},
		Rotation:      2,
		IsPlaced:      true,
		CurrentHP:     gauss.BaseHP,
		ModInitiative: gauss.Initiative,
	}
	engine.State.Board[Hex{Q: 0, R: 0}] = gaussInst

	// Enemy Mutek 1 at adjacent cell (0, 1) along SE line
	mutek, _ := GetBlueprint("borgo", "bor_mutek")
	mutek1 := &models.TileInstance{
		InstanceID:    "t-mutek-1",
		Blueprint:     mutek,
		OwnerID:       p2,
		Hex:           Hex{Q: 0, R: 1},
		Rotation:      0,
		IsPlaced:      true,
		CurrentHP:     mutek.BaseHP,
		ModInitiative: mutek.Initiative,
	}
	engine.State.Board[Hex{Q: 0, R: 1}] = mutek1

	// Enemy Mutek 2 at cell (0, 2) which is also the Borgo HQ!
	// In this test, we place a regular Mutek at (0, 2) to see if the Gauss Cannon penetrates both
	mutek2 := &models.TileInstance{
		InstanceID:    "t-mutek-2",
		Blueprint:     mutek,
		OwnerID:       p2,
		Hex:           Hex{Q: 0, R: 2},
		Rotation:      0,
		IsPlaced:      true,
		CurrentHP:     mutek.BaseHP,
		ModInitiative: mutek.Initiative,
	}
	engine.State.Board[Hex{Q: 0, R: 2}] = mutek2

	// Resolve Battle
	events := engine.ResolveBattle()
	t.Logf("Board size: %d", len(engine.State.Board))
	for k, v := range engine.State.Board {
		t.Logf("Board cell %v: ID=%s, HP=%d", k, v.InstanceID, v.CurrentHP)
	}
	for _, ev := range events {
		t.Logf("Event: %+v", ev)
	}

	// Gauss Cannon deals 1 damage to ALL targets in its line. Both Muteks should be killed.
	if engine.State.Board[Hex{Q: 0, R: 1}] != nil {
		t.Error("Expected first Mutek in firing line to be killed")
	}
	if engine.State.Board[Hex{Q: 0, R: 2}] != nil {
		t.Error("Expected second Mutek in firing line to be killed")
	}
}

func TestHQDestructionAndGameOver(t *testing.T) {
	p1 := "player-1"
	p2 := "player-2"

	engine, err := NewEngine("test-game", p1, "borgo", p2, "moloch")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// Place HQs
	_ = engine.PlaceHQ(p1, Hex{Q: 0, R: 0}, 0) // Borgo HQ at (0, 0)
	_ = engine.PlaceHQ(p2, Hex{Q: 1, R: 0}, 0) // Moloch HQ at (1, 0)

	// Set Borgo HQ health directly to 1 HP
	p1HQ := engine.State.Board[Hex{Q: 0, R: 0}]
	p1HQ.CurrentHP = 1

	// Place Moloch Cyborg at (0, 1) rotated 5 (Ranged attack absolute direction 5 pointing North to (0, 0))
	cyborg, _ := GetBlueprint("moloch", "mol_cyborg")
	cyborgInst := &models.TileInstance{
		InstanceID:    "t-cyborg",
		Blueprint:     cyborg,
		OwnerID:       p2,
		Hex:           Hex{Q: 0, R: 1},
		Rotation:      5,
		IsPlaced:      true,
		CurrentHP:     cyborg.BaseHP,
		ModInitiative: cyborg.Initiative,
	}
	engine.State.Board[Hex{Q: 0, R: 1}] = cyborgInst

	// Resolve Battle
	events := engine.ResolveBattle()

	// Cyborg deals 1 damage to Borgo HQ at Initiative 2.
	// Since Borgo HQ had 1 HP, it should be destroyed (HP drops to 0).
	// The game should end, phase transitions to GameOver, and winner should be player-2.
	if p1HQ.CurrentHP != 0 {
		t.Errorf("Expected HQ HP to be 0, got %d", p1HQ.CurrentHP)
	}
	if engine.State.Phase != models.PhaseGameOver {
		t.Errorf("Expected phase to transition to %s, got %s", models.PhaseGameOver, engine.State.Phase)
	}
	if engine.State.Winner != p2 {
		t.Errorf("Expected winner to be %s, got %s", p2, engine.State.Winner)
	}

	// Verify events log records GAME_OVER
	hasGameOver := false
	for _, ev := range events {
		if ev.Type == "GAME_OVER" {
			hasGameOver = true
		}
	}
	if !hasGameOver {
		t.Error("Expected log to record GAME_OVER event")
	}
}
