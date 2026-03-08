package hunter

import (
	"math"
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

func (hunter *Hunter) registerAspectOfTheHawkSpell() {
	actionID := core.ActionID{SpellID: 27044}

	hunter.AspectOfTheHawkAura = hunter.applySharedAspectConfig(hunter.RegisterAura(core.Aura{
		Label:      "Aspect of the Hawk",
		ActionID:   actionID,
		BuildPhase: core.CharacterBuildPhaseBase,
	}).AttachStatBuff(stats.RangedAttackPower, 155))

	hunter.AspectOfTheHawk = hunter.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolNature,
		ClassSpellMask: HunterSpellAspectOfTheHawk,
		Flags:          core.SpellFlagAPL,

		ManaCost: core.ManaCostOptions{
			FlatCost: 140,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true,
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			spell.RelatedSelfBuff.Activate(sim)
		},

		RelatedSelfBuff: hunter.AspectOfTheHawkAura,
	})
}

func (hunter *Hunter) registerAspectOfTheViper() {
	actionID := core.ActionID{SpellID: 34074}

	var pa *core.PendingAction
	hunter.AspectOfTheViperAura = hunter.applySharedAspectConfig(hunter.RegisterAura(core.Aura{
		Label:    "Aspect of the Viper",
		ActionID: actionID,

		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			pa = core.StartPeriodicAction(sim, core.PeriodicActionOptions{
				Period:   time.Second * 5,
				Priority: core.ActionPriorityRegen,

				OnAction: func(sim *core.Simulation) {
					percentMana := math.Max(0.2, math.Min(0.9, hunter.CurrentManaPercent()))
					scaling := 22.0/35.0*(0.9-percentMana) + 0.11
					if hunter.GronnStalker2PcAura.IsActive() {
						scaling += 0.05
					}

					bonusPer5Seconds := hunter.GetStat(stats.Intellect)*scaling + 0.35*70
					manaGain := bonusPer5Seconds * 2 / 5
					hunter.AddMana(sim, manaGain, hunter.AspectOfTheViper.ResourceMetrics)
				},
			})
		},

		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			pa.Cancel(sim)
			pa = nil
		},
	}))

	hunter.AspectOfTheViper = hunter.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolNature,
		ClassSpellMask: HunterSpellAspectOfTheViper,
		Flags:          core.SpellFlagAPL,

		ManaCost: core.ManaCostOptions{
			FlatCost: 40,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true,
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			spell.RelatedSelfBuff.Activate(sim)
		},

		RelatedSelfBuff: hunter.AspectOfTheViperAura,
	})
}

func (hunter *Hunter) registerAspects() {
	hunter.registerAspectOfTheHawkSpell()
	hunter.registerAspectOfTheViper()
}

func (hunter *Hunter) applySharedAspectConfig(aura *core.Aura) *core.Aura {
	aura.Duration = core.NeverExpires
	aura.NewExclusiveEffect("Aspect", true, core.ExclusiveEffect{})
	return aura
}
