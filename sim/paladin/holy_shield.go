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

var HolyShieldRankMap = genRanks.HolyShield

// Holy Shield (Talent)
// https://www.wowhead.com/tbc/spell=20925
//
// Increases chance to block by 30% for 10 sec, and deals Holy damage
// for each attack blocked while active. Damage caused by Holy Shield causes
// 35% additional threat. Each block expends a charge. 4 charges.
func (paladin *Paladin) registerHolyShield(rankConfig shared.SpellRank) {
	spellID := rankConfig.SpellID
	cost := rankConfig.Cost
	value := shared.SpellRankMin(rankConfig.Direct)
	coefficient := rankConfig.Direct.BonusCoefficient()

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
				Duration: rankConfig.Cooldown,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			holyShieldAura.Activate(sim)
			holyShieldAura.SetStacks(sim, maxStacks)
		},

		RelatedSelfBuff: holyShieldAura,
	})
}
