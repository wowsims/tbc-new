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

var HolyShieldRankMap = shared.RankTable{
	{Rank: 1, SpellID: 20925, Cost: 135, Direct: &shared.Amount{Min: 59, Coef: 0.05}},
	{Rank: 2, SpellID: 20927, Cost: 175, Direct: &shared.Amount{Min: 86, Coef: 0.05}},
	{Rank: 3, SpellID: 20928, Cost: 215, Direct: &shared.Amount{Min: 117, Coef: 0.05}},
	{Rank: 4, SpellID: 27179, Cost: 280, Direct: &shared.Amount{Min: 155, Coef: 0.05}},
}

// Holy Shield (Talent)
// https://www.wowhead.com/tbc/spell=20925
//
// Increases chance to block by 30% for 10 sec, and deals Holy damage
// for each attack blocked while active. Damage caused by Holy Shield causes
// 35% additional threat. Each block expends a charge. 4 charges.
func (paladin *Paladin) registerHolyShield(rankConfig shared.RankRow) {
	spellID := rankConfig.SpellID
	cost := rankConfig.Cost
	value := rankConfig.Direct.Min
	coefficient := rankConfig.Direct.Coef

	actionID := core.ActionID{SpellID: spellID}

	procSpell := paladin.RegisterSpell(core.SpellConfig{
		ActionID:       actionID.WithTag(2),
		SpellSchool:    core.SpellSchoolHoly,
		DefenseType:    core.DefenseTypeMagic,
		ProcMask:       core.ProcMaskEmpty,
		ClassSpellMask: SpellMaskHolyShieldProc,
		Flags:          core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell | core.SpellFlagBinary,

		BonusCoefficient: coefficient,
		DamageMultiplier: 1,
		ThreatMultiplier: 1.35,

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

	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolHoly,
		DefenseType:    core.DefenseTypeMagic,
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
}
