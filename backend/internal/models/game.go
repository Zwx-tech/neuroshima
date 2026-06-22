package models

// TileType represents the functional category of a tile.
type TileType string

const (
	TileHQ      TileType = "HQ"
	TileSoldier TileType = "Soldier"
	TileModule  TileType = "Module"
	TileInstant TileType = "Instant"
)

// AttackType represents Melee or Ranged combat types.
type AttackType string

const (
	AttackMelee  AttackType = "Melee"
	AttackRanged AttackType = "Ranged"
)

// Attack represents an attack capability in a relative direction.
type Attack struct {
	Direction int        `json:"direction"` // 0 to 5 relative direction
	Type      AttackType `json:"type"`
	Power     int        `json:"power"`
}

// BuffType represents statistical modifiers from modules/HQs.
type BuffType string

const (
	BuffInitiative   BuffType = "Initiative"
	BuffMeleePower   BuffType = "MeleePower"
	BuffRangedPower  BuffType = "RangedPower"
	BuffDoubleAction BuffType = "DoubleAction" // Posterunek HQ
)

// Buff represents a modular modifier.
type Buff struct {
	Type       BuffType `json:"type"`
	Value      int      `json:"value"`
	Directions []int    `json:"directions"` // relative directions the module applies buffs to
}

// StaticTileDef defines the base stats of a card loaded from the catalog.
type StaticTileDef struct {
	CardID     string   `json:"card_id"`
	Name       string   `json:"name"`
	Type       TileType `json:"type"`
	BaseHP     int      `json:"base_hp"`
	Initiative []int    `json:"initiative"` // e.g., [2, 1] for double attack or [0] for HQ
	Attacks    []Attack `json:"attacks,omitempty"`
	Armor      []int    `json:"armor,omitempty"` // directions with armor (relative 0-5)
	Nets       []int    `json:"nets,omitempty"`  // directions with nets (relative 0-5)
	Buffs      []Buff   `json:"buffs,omitempty"` // buffs provided (if type == Module)
	Effect     string   `json:"effect,omitempty"` // effect identifier (if type == Instant)
}

// Hex represents a cell in axial coordinates.
type Hex struct {
	Q int `json:"q"`
	R int `json:"r"`
}

// TileInstance represents a live tile in the game.
type TileInstance struct {
	InstanceID string        `json:"instance_id"`
	Blueprint  StaticTileDef `json:"blueprint"`
	OwnerID    string        `json:"owner_id"`

	// Positioning
	Hex      Hex  `json:"hex"`
	Rotation int  `json:"rotation"` // 0-5 clockwise
	IsPlaced bool `json:"is_placed"`

	// Runtime Stats
	CurrentHP int  `json:"current_hp"`
	IsNetted  bool `json:"is_netted"`

	// Modifiers calculated each topological pass
	ModInitiative  []int       `json:"mod_initiative"`
	ModMeleePower  map[int]int `json:"mod_melee_power"`  // absolute direction -> power
	ModRangedPower map[int]int `json:"mod_ranged_power"` // absolute direction -> power
}

// Player represents a participant in the match.
type Player struct {
	ID        string          `json:"id"`
	Army      string          `json:"army"` // "moloch", "borgo", etc.
	Hand      []*TileInstance `json:"hand"`
	Deck      []*TileInstance `json:"deck"`
	Discarded []*TileInstance `json:"discarded"`
	HQPlaced  bool            `json:"hq_placed"`
}

// GamePhase represents the current turn/interaction state.
type GamePhase string

const (
	PhaseHQPlacement     GamePhase = "HQ_PLACEMENT"
	PhaseDiscardMandatory GamePhase = "DISCARD_MANDATORY"
	PhaseMain            GamePhase = "MAIN"
	PhaseGameOver        GamePhase = "GAME_OVER"
)

// GameState is the root session model.
type GameState struct {
	GameID         string                `json:"game_id"`
	ActivePlayer   string                `json:"active_player"`
	StartingPlayer string                `json:"starting_player"`
	TurnNumber     int                   `json:"turn_number"`
	Phase          GamePhase             `json:"phase"`
	Board          map[Hex]*TileInstance `json:"board"`
	Players        map[string]*Player    `json:"players"`
	Winner         string                `json:"winner,omitempty"`
}
