package paladin

import (
	"time"

	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

func (paladin *Paladin) registerHealingSpells() {
	HolyLightRankMap.RegisterAll(paladin.registerHolyLight)
	FlashOfLightRankMap.RegisterAll(paladin.registerFlashOfLight)
	LayOnHandsRankMap.RegisterAll(paladin.registerLayOnHands)
}

var HolyLightRankMap = genRanks.HolyLight

// Holy Light
// https://www.wowhead.com/tbc/spell=27136
//
// Heals a friendly target for a large amount.
func (paladin *Paladin) registerHolyLight(rankConfig shared.RankRow) {
	spellID := rankConfig.SpellID
	cost := rankConfig.Cost
	minHealing := rankConfig.Heal.Min
	maxHealing := rankConfig.Heal.Max
	coefficient := rankConfig.Heal.Coef

	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: spellID},
		SpellSchool:    core.SpellSchoolHoly,
		DefenseType:    core.DefenseTypeMagic,
		ProcMask:       core.ProcMaskSpellHealing,
		Flags:          core.SpellFlagAPL | core.SpellFlagHelpful,
		Rank:           rankConfig.Rank,
		ClassSpellMask: SpellMaskHolyLight,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		MaxRange: 40,

		ManaCost: core.ManaCostOptions{
			FlatCost: cost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: time.Millisecond * 2500,
			},
		},

		BonusCoefficient: coefficient,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealHealing(sim, target, sim.Roll(minHealing, maxHealing), spell.OutcomeHealingCrit)
		},
	})
}

var FlashOfLightRankMap = genRanks.FlashOfLight

// Flash of Light
// https://www.wowhead.com/tbc/spell=27137
//
// Heals a friendly target for a small amount.
func (paladin *Paladin) registerFlashOfLight(rankConfig shared.RankRow) {
	spellID := rankConfig.SpellID
	cost := rankConfig.Cost
	minHealing := rankConfig.Heal.Min
	maxHealing := rankConfig.Heal.Max
	coefficient := rankConfig.Heal.Coef

	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: spellID},
		SpellSchool:    core.SpellSchoolHoly,
		DefenseType:    core.DefenseTypeMagic,
		ProcMask:       core.ProcMaskSpellHealing,
		Flags:          core.SpellFlagAPL | core.SpellFlagHelpful,
		Rank:           rankConfig.Rank,
		ClassSpellMask: SpellMaskFlashOfLight,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		MaxRange: 40,

		ManaCost: core.ManaCostOptions{
			FlatCost: cost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: time.Millisecond * 1500,
			},
		},

		BonusCoefficient: coefficient,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealHealing(sim, target, sim.Roll(minHealing, maxHealing), spell.OutcomeHealingCrit)
		},
	})
}

var LayOnHandsRankMap = genRanks.LayOnHands

// Lay on Hands
// https://www.wowhead.com/tbc/spell=27154
//
// Heals a friendly target for an amount equal to the Paladin's maximum health
// and restores mana to the target. Causes Forbearance for 1 min.
func (paladin *Paladin) registerLayOnHands(rankConfig shared.RankRow) {
	spellID := rankConfig.SpellID
	manaRestore := rankConfig.Energize

	cd := core.Cooldown{
		Timer:    paladin.NewTimer(),
		Duration: time.Hour,
	}

	manaMetrics := paladin.NewManaMetrics(core.ActionID{SpellID: spellID})

	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: spellID},
		SpellSchool:    core.SpellSchoolHoly,
		DefenseType:    core.DefenseTypeMagic,
		ProcMask:       core.ProcMaskSpellHealing,
		Flags:          core.SpellFlagAPL | core.SpellFlagHelpful,
		Rank:           rankConfig.Rank,
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
			target.AddMana(sim, manaRestore, manaMetrics)
			spell.CalcAndDealHealing(sim, target, spell.Unit.MaxHealth(), spell.OutcomeHealingCrit)
		},
	})
}
