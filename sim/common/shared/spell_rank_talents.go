package shared

import "fmt"

// The Misc value of an A_ADD_PCT_MODIFIER or A_ADD_FLAT_MODIFIER effect: which part of the spell it
// changes. Untyped - Misc is a stat under A_MOD_TOTAL_STAT_PERCENTAGE, a school mask under
// A_MOD_DAMAGE_DONE. Hand-written; the client ships no name list, so each carries its evidence.
const (
	SPELLMOD_DAMAGE             = 0  // Fire Power, Piercing Ice, Contagion
	SPELLMOD_DURATION           = 1  // Permafrost, Improved Gouge, Brutal Impact
	SPELLMOD_THREAT             = 2  // Subtlety, Improved Drain Soul
	SPELLMOD_EFFECT1            = 3  // Arcane Potency, Improved Concentration Aura
	SPELLMOD_CHARGES            = 4  // Improved Shield Block, Improved Holy Shield
	SPELLMOD_RANGE              = 5  // Arctic Reach, Flame Throwing, Grim Reach
	SPELLMOD_RADIUS             = 6  // Arctic Reach, Holy Reach, Booming Voice
	SPELLMOD_CRITICAL_CHANCE    = 7  // Arcane Impact, Improved Flamestrike, Incineration
	SPELLMOD_ALL_EFFECTS        = 8  // Frost Warding, Magic Attunement, Demonic Aegis
	SPELLMOD_CASTING_TIME       = 10 // Improved Fireball, Improved Frostbolt
	SPELLMOD_COOLDOWN           = 11 // Improved Fire Blast, Improved Frost Nova, Ice Floes
	SPELLMOD_EFFECT2            = 12 // Malediction, Mana Feed
	SPELLMOD_COST               = 14 // Frost Channeling, Cataclysm
	SPELLMOD_CRIT_DAMAGE_BONUS  = 15 // Ice Shards, Ruin, Vengeance
	SPELLMOD_RESIST_MISS_CHANCE = 16 // Arcane Focus, Elemental Precision, Suppression
	SPELLMOD_ACTIVATION_TIME    = 19 // Improved Fire Totems
	SPELLMOD_DOT                = 22 // Emberstorm, Contagion, Fire Power
	SPELLMOD_EFFECT3            = 23 // Improved Faerie Fire, Savage Fury
	SPELLMOD_BONUS_MULTIPLIER   = 24 // Empowered Arcane Missiles / Fireball / Frostbolt / Corruption

	// Read off one or two talents each, so likelier to be wrong than the ones above.
	SPELLMOD_NOT_LOSE_CASTING_TIME = 9  // Burning Soul, Fel Concentration, Intensity
	SPELLMOD_CHANCE_OF_SUCCESS     = 18 // Improved Poisons, Improved Nature's Grasp
	SPELLMOD_VALUE_MULTIPLIER      = 27 // Improved Mana Shield
	SPELLMOD_RESIST_DISPEL_CHANCE  = 28 // Vile Poisons, Sanctified Seals
)

// Reads the ladder by points spent. Rank 0 is untaken and answers 0, where ByRank would panic.
func (t SpellRankTableOf[T]) ValueAt(rank int32) float64 {
	return ladderValue(t, rank, nil)
}

// The client states a percentage as an integer: 16, not 0.16.
func (t SpellRankTableOf[T]) FractionAt(rank int32) float64 {
	return t.ValueAt(rank) / 100
}

// The sign comes from the data: Improved Righteous Fury states -2/-4/-6, so rank 3 gives 0.94.
func (t SpellRankTableOf[T]) MultiplierAt(rank int32) float64 {
	return 1 + t.FractionAt(rank)
}

// For the case Effect cannot serve: two effects sharing an aura and misc value, as Tactical Mastery's
// two threat modifiers do. The index is the client's EffectIndex, not the slice position.
func (t SpellRankTableOf[T]) EffectAt(index int32) SpellRankEffectLadder[T] {
	pick := func(row SpellRank) float64 {
		for _, e := range row.Effects {
			if e.Index == index {
				return e.Value
			}
		}
		panic(fmt.Sprintf("spell %d rank %d has no effect at index %d", row.SpellID, row.Rank, index))
	}
	return SpellRankEffectLadder[T]{table: t, pick: pick}
}

func (t SpellRankTableOf[T]) Effect(aura SpellRankAura, misc int32) SpellRankEffectLadder[T] {
	pick := func(row SpellRank) float64 { return row.Effect(aura, misc).Value }
	return SpellRankEffectLadder[T]{table: t, pick: pick}
}

type SpellRankEffectLadder[T SpellRanked] struct {
	table SpellRankTableOf[T]
	pick  func(SpellRank) float64
}

func (l SpellRankEffectLadder[T]) ValueAt(rank int32) float64 {
	return ladderValue(l.table, rank, l.pick)
}

func (l SpellRankEffectLadder[T]) FractionAt(rank int32) float64 {
	return l.ValueAt(rank) / 100
}

func (l SpellRankEffectLadder[T]) MultiplierAt(rank int32) float64 {
	return 1 + l.FractionAt(rank)
}

func ladderValue[T SpellRanked](table SpellRankTableOf[T], rank int32, pick func(SpellRank) float64) float64 {
	if rank <= 0 {
		return 0
	}

	row, ok := any(table.ByRank(rank)).(SpellRank)
	if !ok {
		panic(fmt.Sprintf("rank %d is not a SpellRank, so it carries no effects", rank))
	}
	if pick != nil {
		return pick(row)
	}

	// Nothing named means nothing to choose between - reading the first of several silently is the
	// bug this shape exists to prevent.
	if len(row.Effects) != 1 {
		panic(fmt.Sprintf("spell %d rank %d has %d effects - name the one you mean with Effect(aura, misc)",
			row.SpellID, rank, len(row.Effects)))
	}
	return row.Effects[0].Value
}
