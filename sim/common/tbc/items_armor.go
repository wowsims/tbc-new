package tbc

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

func init() {
	// Bulwark of Azzinoth
	core.NewItemEffect(32375, func(agent core.Agent) {
		character := agent.GetCharacter()

		aura := character.NewTemporaryStatsAura(
			"Unbreakable",
			core.ActionID{SpellID: 40407},
			stats.Stats{stats.Armor: 2000},
			time.Second*10,
		)

		procAura := character.MakeProcTriggerAura(core.ProcTrigger{
			Name:               "Illidan Tank Shield",
			ActionID:           core.ActionID{ItemID: 32375},
			ProcMask:           core.ProcMaskDirect,
			ProcChance:         0.02,
			ICD:                time.Minute * 1,
			RequireDamageDealt: true,
			Outcome:            core.OutcomeLanded,
			Callback:           core.CallbackOnSpellHitTaken,
			Handler: func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
				aura.Activate(sim)
			},
		})

		character.ItemSwap.RegisterProc(32375, procAura)
	})

	// Eye of the Night
	core.NewItemEffect(24116, func(agent core.Agent) {
		character := agent.GetCharacter()
		core.MakePermanent(core.EyeOfTheNightAura(character))
	})

	// Chain of the Twilight Owl
	core.NewItemEffect(24121, func(agent core.Agent) {
		character := agent.GetCharacter()
		core.MakePermanent(core.ChainOfTheTwilightOwlAura(character))
	})

	// Braided Eternium Chain
	core.NewItemEffect(24114, func(agent core.Agent) {
		character := agent.GetCharacter()
		core.MakePermanent(core.BraidedEterniumChainAura(character))
	})

	// Pendants of [School] — on-use absorb shields. Each rolls a random absorb
	// amount (900..2700) of the given school, lasts 5 minutes, and has a 1 hour
	// cooldown.
	registerPendantAbsorb(24092, "Pendant of Frozen Flame", 30997, core.SpellSchoolFire)
	registerPendantAbsorb(24093, "Pendant of Thawing", 30994, core.SpellSchoolFrost)
	registerPendantAbsorb(24095, "Pendant of Withering", 30999, core.SpellSchoolNature)
	registerPendantAbsorb(24097, "Pendant of Shadow's End", 31000, core.SpellSchoolShadow)
	registerPendantAbsorb(24098, "Pendant of the Null Rune", 31002, core.SpellSchoolArcane)
}

// registerPendantAbsorb registers an on-use damage absorption shield for a
// single-school resistance pendant (24092 / 24093 / 24095 / 24097 / 24098).
// Each cast rolls a random absorb amount (900..2700) of the given school,
// lasts 5 minutes, and has a 1 hour cooldown.
func registerPendantAbsorb(itemID int32, label string, buffSpellID int32, school core.SpellSchool) {
	core.NewItemEffect(itemID, func(agent core.Agent) {
		character := agent.GetCharacter()

		var rolledStrength float64
		var absorbAura *core.DamageAbsorptionAura

		spell := character.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{ItemID: itemID},
			SpellSchool: core.SpellSchoolPhysical,
			ProcMask:    core.ProcMaskEmpty,

			Cast: core.CastConfig{
				CD: core.Cooldown{
					Timer:    character.NewTimer(),
					Duration: time.Hour,
				},
			},

			ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
				rolledStrength = sim.RollWithLabel(900, 2700, label)
				absorbAura.Activate(sim)
			},
		})

		absorbAura = character.NewDamageAbsorptionAura(core.AbsorptionAuraConfig{
			Aura: core.Aura{
				Label:    label,
				ActionID: core.ActionID{SpellID: buffSpellID},
				Duration: time.Minute * 5,
			},
			ShieldStrengthCalculator: func(_ *core.Unit) float64 {
				return rolledStrength
			},
			ShouldApplyToResult: func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult, _ bool) bool {
				return spell.SpellSchool.Matches(school)
			},
			OnDamageAbsorbed: func(_ *core.Simulation, _ *core.DamageAbsorptionAura, _ *core.SpellResult, absorbedDamage float64) {
				spell.SpellMetrics[character.UnitIndex].TotalShielding += absorbedDamage
				spell.SpellMetrics[character.UnitIndex].Hits++
			},
		})

		character.AddMajorCooldown(core.MajorCooldown{
			Spell: spell,
			Type:  core.CooldownTypeSurvival,
			BuffAura: &core.StatBuffAura{
				Aura:            absorbAura.Aura,
				BuffedStatTypes: []stats.Stat{stats.Health},
			},
			ShouldActivate: func(_ *core.Simulation, character *core.Character) bool {
				return false
			},
		})
	})
}
