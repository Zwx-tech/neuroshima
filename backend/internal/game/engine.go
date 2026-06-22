package game

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"neuroshima/internal/models"
)

type Engine struct {
	State *models.GameState
}

// NewEngine initializes a new game session with 2 players and shuffled decks.
func NewEngine(gameID string, player1ID, player1Army, player2ID, player2Army string) (*Engine, error) {
	p1Deck, err := BuildDeck(player1Army, player1ID)
	if err != nil {
		return nil, fmt.Errorf("failed to build deck for player 1: %w", err)
	}
	p2Deck, err := BuildDeck(player2Army, player2ID)
	if err != nil {
		return nil, fmt.Errorf("failed to build deck for player 2: %w", err)
	}

	p1 := &models.Player{
		ID:        player1ID,
		Army:      player1Army,
		Hand:      make([]*models.TileInstance, 0),
		Deck:      p1Deck,
		Discarded: make([]*models.TileInstance, 0),
	}

	p2 := &models.Player{
		ID:        player2ID,
		Army:      player2Army,
		Hand:      make([]*models.TileInstance, 0),
		Deck:      p2Deck,
		Discarded: make([]*models.TileInstance, 0),
	}

	shuffleDeck(p1.Deck)
	shuffleDeck(p2.Deck)

	state := &models.GameState{
		GameID:         gameID,
		ActivePlayer:   player1ID,
		StartingPlayer: player1ID,
		TurnNumber:     1,
		Phase:          models.PhaseHQPlacement,
		Board:          make(map[Hex]*models.TileInstance),
		Players: map[string]*models.Player{
			player1ID: p1,
			player2ID: p2,
		},
	}

	// Pull HQs out of decks and place them into the players' hands
	if err := moveHQToHand(p1); err != nil {
		return nil, err
	}
	if err := moveHQToHand(p2); err != nil {
		return nil, err
	}

	return &Engine{State: state}, nil
}

// PlaceHQ places a player's HQ from their hand onto the board.
func (e *Engine) PlaceHQ(playerID string, hex Hex, rotation int) error {
	if e.State.Phase != models.PhaseHQPlacement {
		return errors.New("not in HQ placement phase")
	}
	if e.State.ActivePlayer != playerID {
		return errors.New("not your turn")
	}
	player := e.State.Players[playerID]
	if player.HQPlaced {
		return errors.New("HQ already placed")
	}

	// Find HQ in hand
	var hq *models.TileInstance
	hqIdx := -1
	for i, inst := range player.Hand {
		if inst.Blueprint.Type == models.TileHQ {
			hq = inst
			hqIdx = i
			break
		}
	}
	if hq == nil {
		return errors.New("HQ not found in hand")
	}

	if !IsValidBoardHex(hex) {
		return errors.New("invalid board coordinates")
	}
	if e.State.Board[hex] != nil {
		return errors.New("board hex already occupied")
	}

	// Place HQ on board
	hq.Hex = hex
	hq.Rotation = rotation
	hq.IsPlaced = true
	e.State.Board[hex] = hq

	// Remove from hand
	player.Hand = append(player.Hand[:hqIdx], player.Hand[hqIdx+1:]...)
	player.HQPlaced = true

	opponentID := e.getOpponentID(playerID)
	opponent := e.State.Players[opponentID]

	if !opponent.HQPlaced {
		// Switch to opponent to place their HQ
		e.State.ActivePlayer = opponentID
	} else {
		// Both placed HQs! Transition to Turn 1.
		e.State.Phase = models.PhaseMain
		e.State.ActivePlayer = e.State.StartingPlayer
		e.startTurn(e.State.ActivePlayer)
	}

	return nil
}

// DiscardTile removes a tile from the player's hand and sends it to the discard pile.
func (e *Engine) DiscardTile(playerID string, instanceID string) error {
	if e.State.ActivePlayer != playerID {
		return errors.New("not your turn")
	}
	if e.State.Phase != models.PhaseDiscardMandatory && e.State.Phase != models.PhaseMain {
		return errors.New("cannot discard in current phase")
	}

	player := e.State.Players[playerID]
	tileIdx := -1
	for i, inst := range player.Hand {
		if inst.InstanceID == instanceID {
			tileIdx = i
			break
		}
	}
	if tileIdx == -1 {
		return errors.New("tile not found in hand")
	}

	tile := player.Hand[tileIdx]
	// Remove from hand
	player.Hand = append(player.Hand[:tileIdx], player.Hand[tileIdx+1:]...)
	// Add to discard pile
	player.Discarded = append(player.Discarded, tile)

	// If we successfully discarded the mandatory tile, transition to Main
	if e.State.Phase == models.PhaseDiscardMandatory {
		e.State.Phase = models.PhaseMain
	}

	return nil
}

// PlayTile places a regular board tile (Soldier/Module) onto the hex grid.
func (e *Engine) PlayTile(playerID string, instanceID string, hex Hex, rotation int) error {
	if e.State.ActivePlayer != playerID {
		return errors.New("not your turn")
	}
	if e.State.Phase != models.PhaseMain {
		return errors.New("cannot play tiles in current phase (must discard 1 mandatory tile first)")
	}

	player := e.State.Players[playerID]
	tileIdx := -1
	for i, inst := range player.Hand {
		if inst.InstanceID == instanceID {
			tileIdx = i
			break
		}
	}
	if tileIdx == -1 {
		return errors.New("tile not found in hand")
	}

	tile := player.Hand[tileIdx]
	if tile.Blueprint.Type == models.TileInstant {
		return errors.New("cannot place an instant tile on the board")
	}

	if !IsValidBoardHex(hex) {
		return errors.New("invalid board coordinates")
	}
	if e.State.Board[hex] != nil {
		return errors.New("board hex already occupied")
	}

	// Place on board
	tile.Hex = hex
	tile.Rotation = rotation
	tile.IsPlaced = true
	e.State.Board[hex] = tile

	// Remove from hand
	player.Hand = append(player.Hand[:tileIdx], player.Hand[tileIdx+1:]...)

	// Run topological update
	e.UpdateTopology()

	// Check if board is full (19 hexes filled)
	if e.IsBoardFull() {
		e.ResolveBattle()
	}

	return nil
}

// EndTurn completes the active player's turn and starts the next player's turn.
func (e *Engine) EndTurn(playerID string) error {
	if e.State.ActivePlayer != playerID {
		return errors.New("not your turn")
	}
	if e.State.Phase != models.PhaseMain {
		return errors.New("cannot end turn in current phase (must discard 1 mandatory tile first)")
	}

	opponentID := e.getOpponentID(playerID)
	e.State.ActivePlayer = opponentID

	if opponentID == e.State.StartingPlayer {
		e.State.TurnNumber++
	}

	e.startTurn(opponentID)
	return nil
}

// startTurn draws new tiles for the player and determines if mandatory discard is active.
func (e *Engine) startTurn(playerID string) {
	player := e.State.Players[playerID]

	// Determine how many tiles to draw
	drawCount := 3 - len(player.Hand)
	if e.State.TurnNumber == 1 {
		if playerID == e.State.StartingPlayer {
			drawCount = 1
		} else {
			drawCount = 2
		}
	}

	// Draw tiles from deck
	for i := 0; i < drawCount; i++ {
		if len(player.Deck) == 0 {
			break // deck exhausted
		}
		tile := player.Deck[0]
		player.Deck = player.Deck[1:]
		player.Hand = append(player.Hand, tile)
	}

	// Check if mandatory discard is needed (only if hand size is exactly 3)
	if len(player.Hand) == 3 {
		e.State.Phase = models.PhaseDiscardMandatory
	} else {
		e.State.Phase = models.PhaseMain
	}
}

func (e *Engine) getOpponentID(playerID string) string {
	for id := range e.State.Players {
		if id != playerID {
			return id
		}
	}
	return ""
}

func (e *Engine) IsBoardFull() bool {
	return len(e.State.Board) == 19
}

func (e *Engine) UpdateTopology() {
	// Stub for Task 4.2 (topological pass)
}



func shuffleDeck(deck []*models.TileInstance) {
	n := len(deck)
	for i := n - 1; i > 0; i-- {
		res, _ := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		j := int(res.Int64())
		deck[i], deck[j] = deck[j], deck[i]
	}
}

func moveHQToHand(p *models.Player) error {
	hqIdx := -1
	for i, inst := range p.Deck {
		if inst.Blueprint.Type == models.TileHQ {
			hqIdx = i
			break
		}
	}
	if hqIdx == -1 {
		return fmt.Errorf("HQ not found in deck for player %s", p.ID)
	}

	hq := p.Deck[hqIdx]
	p.Deck = append(p.Deck[:hqIdx], p.Deck[hqIdx+1:]...)
	p.Hand = append(p.Hand, hq)
	return nil
}
