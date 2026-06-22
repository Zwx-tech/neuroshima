package game

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"neuroshima/internal/models"
)

//go:embed armies.json
var armiesJSON []byte

// TileBlueprint wraps the StaticTileDef model and adds a count property.
type TileBlueprint struct {
	models.StaticTileDef
	Count int `json:"count"`
}

// ArmyData represents the parsed list of blueprints per army.
type ArmyData struct {
	Armies map[string][]TileBlueprint `json:"armies"`
}

var registry ArmyData

func init() {
	if err := json.Unmarshal(armiesJSON, &registry); err != nil {
		panic(fmt.Sprintf("failed to parse embedded armies.json: %v", err))
	}
}

// GetBlueprint returns the blueprint definition for a given army and card ID.
func GetBlueprint(army string, cardID string) (models.StaticTileDef, error) {
	blueprints, ok := registry.Armies[army]
	if !ok {
		return models.StaticTileDef{}, fmt.Errorf("army %q not found in registry", army)
	}

	for _, bp := range blueprints {
		if bp.CardID == cardID {
			return bp.StaticTileDef, nil
		}
	}

	return models.StaticTileDef{}, fmt.Errorf("card %q not found in army %q", cardID, army)
}

// BuildDeck instantiates a full deck of 35 tiles for the given army.
func BuildDeck(army string, ownerID string) ([]*models.TileInstance, error) {
	blueprints, ok := registry.Armies[army]
	if !ok {
		return nil, fmt.Errorf("army %q not found in registry", army)
	}

	var deck []*models.TileInstance
	instanceCounter := 1

	for _, bp := range blueprints {
		for i := 0; i < bp.Count; i++ {
			instanceID := fmt.Sprintf("%s-%s-%03d", ownerID, bp.CardID, instanceCounter)
			instanceCounter++

			// Create a copy of the blueprint
			inst := &models.TileInstance{
				InstanceID:     instanceID,
				Blueprint:      bp.StaticTileDef,
				OwnerID:        ownerID,
				CurrentHP:      bp.BaseHP,
				ModInitiative:  bp.Initiative,
				ModMeleePower:  make(map[int]int),
				ModRangedPower: make(map[int]int),
			}

			// Pre-populate attacks into runtime modifiers map
			for _, atk := range bp.Attacks {
				if atk.Type == models.AttackMelee {
					inst.ModMeleePower[atk.Direction] = atk.Power
				} else if atk.Type == models.AttackRanged {
					inst.ModRangedPower[atk.Direction] = atk.Power
				}
			}

			deck = append(deck, inst)
		}
	}

	return deck, nil
}
