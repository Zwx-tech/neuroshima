package game

import (
	"fmt"
	"neuroshima/internal/models"
)

// BattleEvent records a single action that occurs during combat.
type BattleEvent struct {
	Initiative  int         `json:"initiative"`
	Type        string      `json:"type"` // "ATTACK_MELEE", "ATTACK_RANGED", "BLOCK", "DAMAGE", "DEATH", "GAME_OVER"
	SourceHex   Hex         `json:"source_hex"`
	TargetHex   Hex         `json:"target_hex"`
	SourceName  string      `json:"source_name"`
	TargetName  string      `json:"target_name"`
	Damage      int         `json:"damage"`
	RemainingHP int         `json:"remaining_hp"`
}

// ResolveBattle simulates a full battle phase initiative-by-initiative.
func (e *Engine) ResolveBattle() []BattleEvent {
	var events []BattleEvent

	// Combat runs from Initiative 4 down to 0
	for initStep := 4; initStep >= 0; initStep-- {
		// 1. Gather all units active in this initiative step
		var activeUnits []*models.TileInstance
		for _, tile := range e.State.Board {
			if tile.IsNetted {
				continue // Netted units cannot act
			}
			// Check if this initiative step is in the tile's current initiative list
			for _, val := range tile.ModInitiative {
				if val == initStep {
					activeUnits = append(activeUnits, tile)
					break
				}
			}
		}

		if len(activeUnits) == 0 {
			continue
		}

		// 2. Process all attacks simultaneously
		damageQueue := make(map[Hex]int)
		var attackEvents []BattleEvent

		for _, attacker := range activeUnits {
			for _, attack := range attacker.Blueprint.Attacks {
				absDir := RotateDirection(attack.Direction, attacker.Rotation)

				if attack.Type == models.AttackMelee {
					targetHex := GetNeighbor(attacker.Hex, absDir)
					if targetTile, ok := e.State.Board[targetHex]; ok {
						e.processAttack(attacker, targetTile, absDir, attack.Power, initStep, "ATTACK_MELEE", &attackEvents, damageQueue)
					}
				} else if attack.Type == models.AttackRanged {
					if attacker.Blueprint.CardID == "mol_gauss" {
						// Gauss Cannon hits all units in the firing line
						currentHex := GetNeighbor(attacker.Hex, absDir)
						for IsValidBoardHex(currentHex) {
							if targetTile, ok := e.State.Board[currentHex]; ok {
								e.processAttack(attacker, targetTile, absDir, attack.Power, initStep, "ATTACK_RANGED", &attackEvents, damageQueue)
							}
							currentHex = GetNeighbor(currentHex, absDir)
						}
					} else {
						// Normal ranged attack hits the first unit in the line
						currentHex := GetNeighbor(attacker.Hex, absDir)
						for IsValidBoardHex(currentHex) {
							if targetTile, ok := e.State.Board[currentHex]; ok {
								e.processAttack(attacker, targetTile, absDir, attack.Power, initStep, "ATTACK_RANGED", &attackEvents, damageQueue)
								break
							}
							currentHex = GetNeighbor(currentHex, absDir)
						}
					}
				}
			}
		}

		// Append attack events
		events = append(events, attackEvents...)

		// 3. Apply damage at the end of the initiative step
		for hex, dmg := range damageQueue {
			targetTile, exists := e.State.Board[hex]
			if !exists {
				continue
			}

			targetTile.CurrentHP -= dmg
			if targetTile.CurrentHP < 0 {
				targetTile.CurrentHP = 0
			}

			events = append(events, BattleEvent{
				Initiative:  initStep,
				Type:        "DAMAGE",
				TargetHex:   hex,
				TargetName:  targetTile.Blueprint.Name,
				Damage:      dmg,
				RemainingHP: targetTile.CurrentHP,
			})
		}

		// 4. Remove dead units from the board
		for hex := range damageQueue {
			targetTile, exists := e.State.Board[hex]
			if !exists {
				continue
			}

			if targetTile.CurrentHP <= 0 {
				delete(e.State.Board, hex)
				events = append(events, BattleEvent{
					Initiative: initStep,
					Type:       "DEATH",
					TargetHex:  hex,
					TargetName: targetTile.Blueprint.Name,
				})

				// Send to owner's discard pile
				owner := e.State.Players[targetTile.OwnerID]
				owner.Discarded = append(owner.Discarded, targetTile)
			}
		}

		// 5. Re-evaluate board topology because nets or modules may have been destroyed
		e.UpdateTopology()

		// 6. Check game over conditions
		if e.checkGameOver(&events) {
			break
		}
	}

	return events
}

// processAttack computes shields and adds damage to the queue.
func (e *Engine) processAttack(
	attacker *models.TileInstance,
	target *models.TileInstance,
	attackDir int,
	power int,
	initStep int,
	atkType string,
	events *[]BattleEvent,
	damageQueue map[Hex]int,
) {
	// Armor/Shield check
	// The attack travels in direction `attackDir` from attacker to target.
	// The impact direction on target is the opposite: (attackDir + 3) % 6
	impactDir := (attackDir + 3) % 6

	// Translate target's relative armor faces to absolute directions
	hasArmor := false
	for _, relArmor := range target.Blueprint.Armor {
		absArmor := RotateDirection(relArmor, target.Rotation)
		if absArmor == impactDir {
			hasArmor = true
			break
		}
	}

	if hasArmor {
		// Attack blocked by shield
		*events = append(*events, BattleEvent{
			Initiative: initStep,
			Type:       "BLOCK",
			SourceHex:  attacker.Hex,
			TargetHex:  target.Hex,
			SourceName: attacker.Blueprint.Name,
			TargetName: target.Blueprint.Name,
		})
		return
	}

	// Queue the damage to resolve simultaneously
	damageQueue[target.Hex] += power

	*events = append(*events, BattleEvent{
		Initiative: initStep,
		Type:       atkType,
		SourceHex:  attacker.Hex,
		TargetHex:  target.Hex,
		SourceName: attacker.Blueprint.Name,
		TargetName: target.Blueprint.Name,
		Damage:     power,
	})
}

// checkGameOver checks if any HQs are destroyed and updates GameState accordingly.
func (e *Engine) checkGameOver(events *[]BattleEvent) bool {
	// Find HQs
	var p1HQ, p2HQ *models.TileInstance
	var p1ID, p2ID string

	// Collect player IDs
	for id := range e.State.Players {
		if p1ID == "" {
			p1ID = id
		} else {
			p2ID = id
		}
	}

	// Scan board for HQs
	for _, tile := range e.State.Board {
		if tile.Blueprint.Type == models.TileHQ {
			if tile.OwnerID == p1ID {
				p1HQ = tile
			} else if tile.OwnerID == p2ID {
				p2HQ = tile
			}
		}
	}

	p1Dead := p1HQ == nil || p1HQ.CurrentHP <= 0
	p2Dead := p2HQ == nil || p2HQ.CurrentHP <= 0

	if p1Dead || p2Dead {
		e.State.Phase = models.PhaseGameOver
		if p1Dead && p2Dead {
			e.State.Winner = "DRAW"
			*events = append(*events, BattleEvent{
				Type:       "GAME_OVER",
				TargetName: "DRAW",
			})
		} else if p1Dead {
			e.State.Winner = p2ID
			*events = append(*events, BattleEvent{
				Type:       "GAME_OVER",
				TargetName: fmt.Sprintf("WINNER: %s", p2ID),
			})
		} else {
			e.State.Winner = p1ID
			*events = append(*events, BattleEvent{
				Type:       "GAME_OVER",
				TargetName: fmt.Sprintf("WINNER: %s", p1ID),
			})
		}
		return true
	}

	return false
}
