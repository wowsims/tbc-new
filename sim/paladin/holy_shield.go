package paladin

import (
	"time"

	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

func (paladin *Paladin) getHolyShieldTimer() *core.Timer {
	if paladin.holyShieldTimer == nil {
		paladin.holyShieldTimer = paladin.NewTimer()
	}
	return paladin.holyShieldTimer
}

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
	spellID := rankConfig.SpellID
	cost := rankConfig.Cost
	value := rankConfig.MinDamage
	coefficient := rankConfig.Coefficient
	threatMultiplier := rankConfig.ThreatMultiplier

	actionID := core.ActionID{SpellID: spellID}

	procSpell := paladin.RegisterSpell(core.SpellConfig{
		ActionID:       actionID.WithTag(2),
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskEmpty,
		ClassSpellMask: SpellMaskHolyShieldProc,
		Flags:          core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell | core.SpellFlagBinary,

		BonusCoefficient: coefficient,
		DamageMultiplier: 1,
		ThreatMultiplier: threatMultiplier,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealDamage(sim, target, value, spell.OutcomeMagicHit)
		},
	})

	maxStacks := []int32{4, 6, 8}[paladin.Talents.ImprovedHolyShield]

	var holyShieldAura *core.Aura
	holyShieldAura = paladin.RegisterAura(core.Aura{
		Label:     "Holy Shield" + paladin.Label + " " + rankConfig.GetRankLabel(),
		ActionID:  actionID,
		Duration:  time.Second * 10,
		MaxStacks: maxStacks,
	}).AttachProcTrigger(core.ProcTrigger{
		Callback: core.CallbackOnSpellHitTaken,
		Outcome:  core.OutcomeBlock,

		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			procSpell.Cast(sim, spell.Unit)
			holyShieldAura.RemoveStack(sim)
		},
	}).AttachStatBuff(stats.BlockPercent, 0.3)

	holyShieldSpell := paladin.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagAPL | core.SpellFlagMeleeMetrics,
		ClassSpellMask: SpellMaskHolyShield,
		Rank:           rankConfig.Rank,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		ManaCost: core.ManaCostOptions{
			FlatCost: cost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    paladin.getHolyShieldTimer(),
				Duration: time.Second * 10,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			holyShieldAura.Activate(sim)
			holyShieldAura.SetStacks(sim, maxStacks)
		},

		RelatedSelfBuff: holyShieldAura,
	})

	paladin.HolyShieldAuras = append(paladin.HolyShieldAuras, holyShieldAura)
	paladin.HolyShields = append(paladin.HolyShields, holyShieldSpell)
}
