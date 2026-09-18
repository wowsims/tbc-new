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

var HolyLightRankMap = shared.RankTable{
	{Rank: 1, SpellID: 635, Cost: 35, Heal: &shared.Amount{Min: 42, Max: 51, Coef: 0.205}},
	{Rank: 2, SpellID: 639, Cost: 60, Heal: &shared.Amount{Min: 81, Max: 96, Coef: 0.339}},
	{Rank: 3, SpellID: 647, Cost: 110, Heal: &shared.Amount{Min: 167, Max: 196, Coef: 0.554}},
	{Rank: 4, SpellID: 1026, Cost: 190, Heal: &shared.Amount{Min: 322, Max: 368, Coef: 0.714}},
	{Rank: 5, SpellID: 1042, Cost: 275, Heal: &shared.Amount{Min: 506, Max: 569, Coef: 0.714}},
	{Rank: 6, SpellID: 3472, Cost: 365, Heal: &shared.Amount{Min: 717, Max: 799, Coef: 0.714}},
	{Rank: 7, SpellID: 10328, Cost: 465, Heal: &shared.Amount{Min: 968, Max: 1076, Coef: 0.714}},
	{Rank: 8, SpellID: 10329, Cost: 580, Heal: &shared.Amount{Min: 1272, Max: 1414, Coef: 0.714}},
	{Rank: 9, SpellID: 25292, Cost: 660, Heal: &shared.Amount{Min: 1619, Max: 1799, Coef: 0.714}},
	{Rank: 10, SpellID: 27135, Cost: 710, Heal: &shared.Amount{Min: 1773, Max: 1971, Coef: 0.714}},
	{Rank: 11, SpellID: 27136, Cost: 840, Heal: &shared.Amount{Min: 2196, Max: 2446, Coef: 0.714}},
}

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

var FlashOfLightRankMap = shared.RankTable{
	{Rank: 1, SpellID: 19750, Cost: 35, Heal: &shared.Amount{Min: 67, Max: 77, Coef: 0.429}},
	{Rank: 2, SpellID: 19939, Cost: 50, Heal: &shared.Amount{Min: 102, Max: 117, Coef: 0.429}},
	{Rank: 3, SpellID: 19940, Cost: 70, Heal: &shared.Amount{Min: 153, Max: 171, Coef: 0.429}},
	{Rank: 4, SpellID: 19941, Cost: 90, Heal: &shared.Amount{Min: 206, Max: 231, Coef: 0.429}},
	{Rank: 5, SpellID: 19942, Cost: 115, Heal: &shared.Amount{Min: 278, Max: 310, Coef: 0.429}},
	{Rank: 6, SpellID: 19943, Cost: 140, Heal: &shared.Amount{Min: 356, Max: 396, Coef: 0.429}},
	{Rank: 7, SpellID: 27137, Cost: 180, Heal: &shared.Amount{Min: 458, Max: 513, Coef: 0.429}},
}

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

var LayOnHandsRankMap = shared.RankTable{
	{Rank: 1, SpellID: 633, Cost: 0, Energize: 0},
	{Rank: 2, SpellID: 2800, Cost: 0, Energize: 250},
	{Rank: 3, SpellID: 10310, Cost: 0, Energize: 550},
	{Rank: 4, SpellID: 27154, Cost: 0, Energize: 900},
}

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
