package paladin

import (
	"strconv"
	"time"

	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

var HolyShieldRankMap = shared.SpellRankMap{
	{Rank: 1, SpellID: 20925, Cost: 135, MinDamage: 59, Coefficient: 0.05, ThreatMultiplier: 1.35},
	{Rank: 2, SpellID: 20927, Cost: 175, MinDamage: 86, Coefficient: 0.05, ThreatMultiplier: 1.35},
	{Rank: 3, SpellID: 20928, Cost: 215, MinDamage: 117, Coefficient: 0.05, ThreatMultiplier: 1.35},
	{Rank: 4, SpellID: 27179, Cost: 280, MinDamage: 155, Coefficient: 0.05, ThreatMultiplier: 1.35},
}

// Holy Shield (Talent)
// https://www.wowhead.com/tbc/spell=20925
//
// Increases chance to block by 30% for 10 sec, and deals Holy damage
// for each attack blocked while active. Damage caused by Holy Shield causes
// 35% additional threat. Each block expends a charge. 4 charges.
func (paladin *Paladin) registerHolyShield(rankConfig shared.SpellRankConfig) {
	rank := rankConfig.Rank
	spellID := rankConfig.SpellID
	cost := rankConfig.Cost
	value := rankConfig.MinDamage
	coefficient := rankConfig.Coefficient
	threatMultiplier := rankConfig.ThreatMultiplier

	cd := core.Cooldown{
		Timer:    paladin.NewTimer(),
		Duration: time.Second * 10,
	}

	blockRating := 30 * core.BlockRatingPerBlockPercent // 30% block chance

	actionID := core.ActionID{SpellID: spellID}

	procSpell := paladin.RegisterSpell(core.SpellConfig{
		ActionID:    actionID,
		SpellSchool: core.SpellSchoolHoly,
		ProcMask:    core.ProcMaskEmpty,
		Flags:       core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell,

		BonusCoefficient: coefficient,
		DamageMultiplier: 1,
		ThreatMultiplier: threatMultiplier,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealDamage(sim, target, value, spell.OutcomeAlwaysHit)
		},
	})

	holyShieldAura := paladin.RegisterAura(core.Aura{
		Label:     "Holy Shield" + paladin.Label + "Rank" + strconv.Itoa(int(rank)),
		ActionID:  actionID,
		Duration:  time.Second * 10,
		MaxStacks: 4,

		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.AddStatDynamic(sim, stats.BlockRating, blockRating)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.AddStatDynamic(sim, stats.BlockRating, -blockRating)
		},

		OnSpellHitTaken: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if result.Outcome.Matches(core.OutcomeBlock) {
				procSpell.Cast(sim, spell.Unit)
				aura.RemoveStack(sim)
			}
		},
	})

	holyShieldSpell := paladin.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: SpellMaskHolyShield,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		ManaCost: core.ManaCostOptions{
			FlatCost: cost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: cd,
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			holyShieldAura.SetStacks(sim, 4)
			holyShieldAura.Activate(sim)
		},
	})

	paladin.HolyShieldAuras = append(paladin.HolyShieldAuras, holyShieldAura)
	paladin.HolyShields = append(paladin.HolyShields, holyShieldSpell)
}
