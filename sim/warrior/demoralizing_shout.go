package warrior

import (
	"github.com/wowsims/tbc/sim/core"
)

var demoralizingShoutRank = genRanks.DemoralizingShout.BySpellID(25203)

func (war *Warrior) registerDemoralizingShout() {
	war.DemoralizingShoutAuras = war.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return core.DemoralizingShoutAura(target, war.Talents.BoomingVoice, war.Talents.ImprovedDemoralizingShout)
	})

	war.DemoralizingShout = war.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: demoralizingShoutRank.SpellID},
		SpellSchool:    core.SpellSchoolPhysical,
		DefenseType:    core.DefenseTypeMagic,
		ClassSpellMask: SpellMaskDemoralizingShout,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagAPL,

		RageCost: core.RageCostOptions{
			Cost: demoralizingShoutRank.Cost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: demoralizingShoutRank.GCD,
			},
			IgnoreHaste: true,
		},

		ThreatMultiplier: 1,
		FlatThreatBonus:  56,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			for _, aoeTarget := range sim.Encounter.ActiveTargetUnits {
				result := spell.CalcAndDealOutcome(sim, aoeTarget, spell.OutcomeMagicHit)
				if result.Landed() {
					war.DemoralizingShoutAuras.Get(aoeTarget).Activate(sim)
				}
			}
		},

		RelatedAuraArrays: war.DemoralizingShoutAuras.ToMap(),
	})
}
