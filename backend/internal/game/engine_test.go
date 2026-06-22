package game

import (
	"neuroshima/internal/models"
	"testing"
)

func TestEngineHQPlacementAndTurnFlow(t *testing.T) {
	p1 := "player-1"
	p2 := "player-2"

	// 1. Initialize Engine
	engine, err := NewEngine("test-game", p1, "borgo", p2, "moloch")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	state := engine.State

	// Verify starting phase and active player
	if state.Phase != models.PhaseHQPlacement {
		t.Errorf("Expected phase %s, got %s", models.PhaseHQPlacement, state.Phase)
	}
	if state.ActivePlayer != p1 {
		t.Errorf("Expected active player to be %s, got %s", p1, state.ActivePlayer)
	}

	// Verify both players have exactly 1 card in hand (their HQ)
	if len(state.Players[p1].Hand) != 1 || state.Players[p1].Hand[0].Blueprint.Type != models.TileHQ {
		t.Error("Player 1 should start with only their HQ in hand")
	}
	if len(state.Players[p2].Hand) != 1 || state.Players[p2].Hand[0].Blueprint.Type != models.TileHQ {
		t.Error("Player 2 should start with only their HQ in hand")
	}

	// 2. Player 1 places HQ
	err = engine.PlaceHQ(p1, Hex{Q: 0, R: 0}, 0)
	if err != nil {
		t.Fatalf("Player 1 failed to place HQ: %v", err)
	}

	// Verify active player switched to Player 2 to place their HQ
	if state.ActivePlayer != p2 {
		t.Errorf("Expected active player to switch to %s, got %s", p2, state.ActivePlayer)
	}
	if len(state.Players[p1].Hand) != 0 {
		t.Error("Player 1's hand should be empty after placing HQ")
	}

	// Attempting to place HQ again for Player 1 should fail
	err = engine.PlaceHQ(p1, Hex{Q: 0, R: 1}, 0)
	if err == nil {
		t.Error("Player 1 should not be able to place HQ twice")
	}

	// 3. Player 2 places HQ
	err = engine.PlaceHQ(p2, Hex{Q: 1, R: 0}, 0)
	if err != nil {
		t.Fatalf("Player 2 failed to place HQ: %v", err)
	}

	// Verify phase transitioned to PhaseMain and turn count is 1
	if state.Phase != models.PhaseMain {
		t.Errorf("Expected phase to transition to %s, got %s", models.PhaseMain, state.Phase)
	}
	if state.TurnNumber != 1 {
		t.Errorf("Expected TurnNumber to be 1, got %d", state.TurnNumber)
	}

	// Verify active player is back to Player 1 (starting player)
	if state.ActivePlayer != p1 {
		t.Errorf("Expected active player to return to %s, got %s", p1, state.ActivePlayer)
	}

	// Turn 1 hand draws: Player 1 should draw exactly 1 tile
	if len(state.Players[p1].Hand) != 1 {
		t.Errorf("Expected Player 1 hand size to be 1 on Turn 1, got %d", len(state.Players[p1].Hand))
	}

	// 4. Player 1 ends Turn 1
	err = engine.EndTurn(p1)
	if err != nil {
		t.Fatalf("Player 1 failed to end turn: %v", err)
	}

	// Active player is now Player 2
	if state.ActivePlayer != p2 {
		t.Errorf("Expected active player to switch to %s, got %s", p2, state.ActivePlayer)
	}

	// Player 2 draws up to 2 tiles on Turn 1 (drawn count is 2 - current hand size which is 0)
	if len(state.Players[p2].Hand) != 2 {
		t.Errorf("Expected Player 2 hand size to be 2 on Turn 1, got %d", len(state.Players[p2].Hand))
	}

	// 5. Player 2 ends their turn
	err = engine.EndTurn(p2)
	if err != nil {
		t.Fatalf("Player 2 failed to end turn: %v", err)
	}

	// Turn transitions to Turn 2. Player 1 starts and draws up to 3 tiles.
	if state.TurnNumber != 2 {
		t.Errorf("Expected turn to increment to 2, got %d", state.TurnNumber)
	}
	if state.ActivePlayer != p1 {
		t.Errorf("Expected active player to switch to %s, got %s", p1, state.ActivePlayer)
	}

	// Player 1 has 1 tile kept from Turn 1 + draws 2 tiles = 3 tiles in hand.
	if len(state.Players[p1].Hand) != 3 {
		t.Errorf("Expected Player 1 hand size to be 3 on Turn 2, got %d", len(state.Players[p1].Hand))
	}

	// Verify phase is PhaseDiscardMandatory since Player 1 has exactly 3 tiles in hand
	if state.Phase != models.PhaseDiscardMandatory {
		t.Errorf("Expected phase to be %s, got %s", models.PhaseDiscardMandatory, state.Phase)
	}

	// Player 1 attempts to play a tile while in DiscardMandatory phase -> should fail
	// Ensure Hand[0] is a Soldier so it doesn't fail on "cannot place instant tile" validation
	state.Players[p1].Hand[0] = &models.TileInstance{
		InstanceID: "test-soldier-1",
		Blueprint: models.StaticTileDef{
			CardID: "bor_mutek",
			Name:   "Mutek",
			Type:   models.TileSoldier,
			BaseHP: 1,
		},
		OwnerID:   p1,
		CurrentHP: 1,
	}
	playTileID := state.Players[p1].Hand[0].InstanceID
	err = engine.PlayTile(p1, playTileID, Hex{Q: -1, R: 0}, 0)
	if err == nil {
		t.Error("Expected error when trying to play tile without completing mandatory discard")
	}

	// Player 1 discards 1 tile
	err = engine.DiscardTile(p1, playTileID)
	if err != nil {
		t.Fatalf("Failed to discard tile: %v", err)
	}

	// Verify hand size is now 2 and phase transitioned back to PhaseMain
	if len(state.Players[p1].Hand) != 2 {
		t.Errorf("Expected Player 1 hand size to be 2, got %d", len(state.Players[p1].Hand))
	}
	if state.Phase != models.PhaseMain {
		t.Errorf("Expected phase to transition to %s, got %s", models.PhaseMain, state.Phase)
	}

	// Player 1 plays a tile successfully now that mandatory discard is completed
	state.Players[p1].Hand[0] = &models.TileInstance{
		InstanceID: "test-soldier-2",
		Blueprint: models.StaticTileDef{
			CardID: "bor_mutek",
			Name:   "Mutek",
			Type:   models.TileSoldier,
			BaseHP: 1,
		},
		OwnerID:   p1,
		CurrentHP: 1,
	}
	playTileID2 := state.Players[p1].Hand[0].InstanceID
	err = engine.PlayTile(p1, playTileID2, Hex{Q: -1, R: 0}, 0)
	if err != nil {
		t.Fatalf("Player 1 failed to play tile: %v", err)
	}

	// Verify tile is placed on board
	placedTile := state.Board[Hex{Q: -1, R: 0}]
	if placedTile == nil || placedTile.InstanceID != playTileID2 {
		t.Error("Expected tile to be placed at coordinates (-1, 0)")
	}
	if !placedTile.IsPlaced {
		t.Error("Expected placed tile to have IsPlaced = true")
	}
	if len(state.Players[p1].Hand) != 1 {
		t.Errorf("Expected hand size to decrease to 1, got %d", len(state.Players[p1].Hand))
	}
}
