package paladin

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (paladin *Paladin) registerHealingSpells() {
	paladin.registerHolyLight()
	paladin.registerFlashOfLight()
	paladin.registerLayOnHands()
}

// Holy Light
// https://www.wowhead.com/tbc/spell=27136
//
// Heals a friendly target for a large amount.
func (paladin *Paladin) registerHolyLight() {
	var ranks = []struct {
		level        int32
		spellID      int32
		manaCost     int32
		minValue     float64
		maxValue     float64
		coeff        float64
		scaleLevel   int32
		scalingCoeff float64
	}{
		{},
		{level: 1, spellID: 635, manaCost: 35, minValue: 39, maxValue: 47, coeff: 0.205, scaleLevel: 5, scalingCoeff: 0.80},
		{level: 6, spellID: 639, manaCost: 60, minValue: 76, maxValue: 90, coeff: 0.339, scaleLevel: 11, scalingCoeff: 1.10},
		{level: 14, spellID: 647, manaCost: 110, minValue: 159, maxValue: 187, coeff: 0.554, scaleLevel: 19, scalingCoeff: 1.70},
		{level: 22, spellID: 1026, manaCost: 190, minValue: 310, maxValue: 356, coeff: 0.714, scaleLevel: 27, scalingCoeff: 2.40},
		{level: 30, spellID: 1042, manaCost: 275, minValue: 491, maxValue: 553, coeff: 0.714, scaleLevel: 35, scalingCoeff: 3.10},
		{level: 38, spellID: 3472, manaCost: 365, minValue: 698, maxValue: 780, coeff: 0.714, scaleLevel: 43, scalingCoeff: 3.80},
		{level: 46, spellID: 10328, manaCost: 465, minValue: 945, maxValue: 1053, coeff: 0.714, scaleLevel: 51, scalingCoeff: 4.60},
		{level: 54, spellID: 10329, manaCost: 580, minValue: 1246, maxValue: 1388, coeff: 0.714, scaleLevel: 59, scalingCoeff: 5.20},
		{level: 60, spellID: 25292, manaCost: 660, minValue: 1590, maxValue: 1770, coeff: 0.714, scaleLevel: 65, scalingCoeff: 5.80},
		{level: 62, spellID: 27135, manaCost: 710, minValue: 1741, maxValue: 1939, coeff: 0.714, scaleLevel: 67, scalingCoeff: 6.40},
		{level: 70, spellID: 27136, manaCost: 840, minValue: 2196, maxValue: 2446, coeff: 0.714, scaleLevel: 75, scalingCoeff: 7.00},
	}

	for rank := 1; rank < len(ranks); rank++ {
		if paladin.Level < ranks[rank].level {
			break
		}

		minHealing := ranks[rank].minValue + ranks[rank].scalingCoeff*float64(min(paladin.Level, ranks[rank].scaleLevel)-ranks[rank].level)
		maxHealing := ranks[rank].maxValue + ranks[rank].scalingCoeff*float64(min(paladin.Level, ranks[rank].scaleLevel)-ranks[rank].level)

		holyLight := paladin.RegisterSpell(core.SpellConfig{
			ActionID:       core.ActionID{SpellID: ranks[rank].spellID},
			SpellSchool:    core.SpellSchoolHoly,
			ProcMask:       core.ProcMaskSpellHealing,
			Flags:          core.SpellFlagAPL | core.SpellFlagHelpful,
			ClassSpellMask: SpellMaskHolyLight,

			DamageMultiplier: 1,
			ThreatMultiplier: 1,

			MaxRange: 40,

			ManaCost: core.ManaCostOptions{
				FlatCost: ranks[rank].manaCost,
			},
			Cast: core.CastConfig{
				DefaultCast: core.Cast{
					GCD:      core.GCDDefault,
					CastTime: time.Millisecond * 2500,
				},
			},

			BonusCoefficient: ranks[rank].coeff,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.CalcAndDealHealing(sim, target, sim.Roll(minHealing, maxHealing), spell.OutcomeHealingCrit)
			},
		})

		paladin.HolyLights = append(paladin.HolyLights, holyLight)
	}
}

// Flash of Light
// https://www.wowhead.com/tbc/spell=27137
//
// Heals a friendly target for a small amount.
func (paladin *Paladin) registerFlashOfLight() {
	var ranks = []struct {
		level        int32
		spellID      int32
		manaCost     int32
		minValue     float64
		maxValue     float64
		coeff        float64
		scaleLevel   int32
		scalingCoeff float64
	}{
		{},
		{level: 20, spellID: 19750, manaCost: 35, minValue: 62, maxValue: 72, coeff: 0.429, scaleLevel: 25, scalingCoeff: 1.00},
		{level: 26, spellID: 19939, manaCost: 50, minValue: 96, maxValue: 110, coeff: 0.429, scaleLevel: 31, scalingCoeff: 1.30},
		{level: 34, spellID: 19940, manaCost: 70, minValue: 145, maxValue: 163, coeff: 0.429, scaleLevel: 39, scalingCoeff: 1.60},
		{level: 42, spellID: 19941, manaCost: 90, minValue: 197, maxValue: 221, coeff: 0.429, scaleLevel: 47, scalingCoeff: 1.90},
		{level: 50, spellID: 19942, manaCost: 115, minValue: 267, maxValue: 299, coeff: 0.429, scaleLevel: 55, scalingCoeff: 2.20},
		{level: 58, spellID: 19943, manaCost: 140, minValue: 343, maxValue: 383, coeff: 0.429, scaleLevel: 63, scalingCoeff: 2.60},
		{level: 66, spellID: 27137, manaCost: 180, minValue: 448, maxValue: 502, coeff: 0.429, scaleLevel: 71, scalingCoeff: 2.60},
	}

	for rank := 1; rank < len(ranks); rank++ {
		if paladin.Level < ranks[rank].level {
			break
		}

		minHealing := ranks[rank].minValue + ranks[rank].scalingCoeff*float64(min(paladin.Level, ranks[rank].scaleLevel)-ranks[rank].level)
		maxHealing := ranks[rank].maxValue + ranks[rank].scalingCoeff*float64(min(paladin.Level, ranks[rank].scaleLevel)-ranks[rank].level)

		flashOfLight := paladin.RegisterSpell(core.SpellConfig{
			ActionID:       core.ActionID{SpellID: ranks[rank].spellID},
			SpellSchool:    core.SpellSchoolHoly,
			ProcMask:       core.ProcMaskSpellHealing,
			Flags:          core.SpellFlagAPL | core.SpellFlagHelpful,
			ClassSpellMask: SpellMaskFlashOfLight,

			DamageMultiplier: 1,
			ThreatMultiplier: 1,

			MaxRange: 40,

			ManaCost: core.ManaCostOptions{
				FlatCost: ranks[rank].manaCost,
			},
			Cast: core.CastConfig{
				DefaultCast: core.Cast{
					GCD:      core.GCDDefault,
					CastTime: time.Millisecond * 1500,
				},
			},

			BonusCoefficient: ranks[rank].coeff,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.CalcAndDealHealing(sim, target, sim.Roll(minHealing, maxHealing), spell.OutcomeHealingCrit)
			},
		})

		paladin.FlashOfLights = append(paladin.FlashOfLights, flashOfLight)
	}
}

// Lay on Hands
// https://www.wowhead.com/tbc/spell=27154
//
// Heals a friendly target for an amount equal to the Paladin's maximum health
// and restores mana to the target. Causes Forbearance for 1 min.
func (paladin *Paladin) registerLayOnHands() {
	var ranks = []struct {
		level        int32
		spellID      int32
		manaRestore  float64
	}{
		{},
		{level: 10, spellID: 633, manaRestore: 0},
		{level: 30, spellID: 2800, manaRestore: 250},
		{level: 50, spellID: 10310, manaRestore: 550},
		{level: 69, spellID: 27154, manaRestore: 900},
	}

	cd := core.Cooldown{
		Timer:    paladin.NewTimer(),
		Duration: time.Hour,
	}

	for rank := 1; rank < len(ranks); rank++ {
		if paladin.Level < ranks[rank].level {
			break
		}
		manaMetrics := paladin.NewManaMetrics(core.ActionID{SpellID: ranks[rank].spellID})
		paladin.LayOnHands = append(paladin.LayOnHands, paladin.RegisterSpell(core.SpellConfig{
			ActionID:       core.ActionID{SpellID: ranks[rank].spellID},
			SpellSchool:    core.SpellSchoolHoly,
			ProcMask:       core.ProcMaskSpellHealing,
			Flags:          core.SpellFlagAPL | core.SpellFlagHelpful,
			ClassSpellMask: SpellMaskLayOnHands,

			DamageMultiplier: 1,
			ThreatMultiplier: 1,

			MaxRange: 40,
	
			Cast: core.CastConfig{
				DefaultCast: core.Cast{
					GCD: core.GCDDefault,
				},
				CD: cd,
			},
	
			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				// Drain all of the caster's mana
				spell.Unit.AddMana(sim, -spell.Unit.CurrentMana(), manaMetrics)

				// Restore mana and health to the target
				target.AddMana(sim, ranks[rank].manaRestore, manaMetrics)
				spell.CalcAndDealHealing(sim, target, spell.Unit.MaxHealth(), spell.OutcomeHealingCrit)
			},
		}))
	}
}
