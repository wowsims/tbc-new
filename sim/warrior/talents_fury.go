package warrior

import (
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

func (war *Warrior) registerFuryTalents() {
	// Tier 1
	// Booming Voice implemented in shouts.go
	war.registerCruelty()

	// Tier 2
	// Improved Demoralizing Shout implemented in demoralizing_shout.go
	war.registerUnbridledWrath()

	// Tier 3
	war.registerImprovedCleave()
	// Piercing Howl not implemented
	// Blood Craze not implemented
	// Commanding Presence implemented in shouts.go

	// Tier 4
	war.registerDualWieldSpecialization()
}

func (war *Warrior) registerCruelty() {
	if war.Talents.Cruelty == 0 {
		return
	}

	war.AddStat(stats.PhysicalCritPercent, 1+0.01*float64(war.Talents.Cruelty))
}

func (war *Warrior) registerUnbridledWrath() {
	if war.Talents.UnbridledWrath == 0 {
		return
	}

	rageMetrics := war.NewRageMetrics(core.ActionID{SpellID: 13002})

	war.MakeProcTriggerAura(core.ProcTrigger{
		Name:               "Unbridled Wrath",
		DPM:                war.NewStaticLegacyPPMManager(3*float64(war.Talents.UnbridledWrath), core.ProcMaskMeleeWhiteHit),
		RequireDamageDealt: true,
		Outcome:            core.OutcomeLanded,
		Callback:           core.CallbackOnSpellHitDealt,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			war.AddRage(sim, 1, rageMetrics)
		},
	})
}

func (war *Warrior) registerImprovedCleave() {
	if war.Talents.ImprovedCleave == 0 {
		return
	}

	war.AddStaticMod(core.SpellModConfig{
		ClassMask:  SpellMaskCleave,
		Kind:       core.SpellMod_DamageDone_Flat,
		FloatValue: 0.4 * float64(war.Talents.ImprovedCleave),
	})
}

func (war *Warrior) registerDualWieldSpecialization() {
	if war.Talents.DualWieldSpecialization == 0 {
		return
	}

	war.AddStaticMod(core.SpellModConfig{
		ProcMask:   core.ProcMaskMeleeOH,
		Kind:       core.SpellMod_DamageDone_Flat,
		FloatValue: 0.05 * float64(war.Talents.DualWieldSpecialization),
	})
}
