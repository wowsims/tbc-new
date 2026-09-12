package paladin

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

func init() {
	// Tome of Fiery Redemption
	core.NewItemEffect(30447, func(agent core.Agent) {
		paladin := agent.(PaladinAgent).GetPaladin()

		procAura := paladin.NewTemporaryStatsAura(
			"Blessing of Righteousness",
			core.ActionID{SpellID: 37198},
			stats.Stats{stats.SpellDamage: 290},
			time.Second*15)

		paladin.MakeProcTriggerAura(core.ProcTrigger{
			Name:            "Tome of Fiery Redemption",
			MetricsActionID: core.ActionID{SpellID: 37197},
			Callback:        core.CallbackOnCastComplete,
			ClassSpellMask:  SpellMaskCanProcTome,
			ProcChance:      0.15,
			ICD:             time.Second * 45,

			ExtraCondition: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) bool {
				return sim.CurrentTime >= 0 &&
					!spell.SpellSchool.Matches(core.SpellSchoolPhysical) &&
					(!spell.Matches(SpellMaskAllSeals) || spell.Flags.Matches(core.SpellFlagAPL))
			},

			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				procAura.Activate(sim)
			},
		})
	})

	// Ashtongue Talisman of Zeal
	// https://www.wowhead.com/tbc/item=32489/ashtongue-talisman-of-zeal
	//
	// Your Judgements have a 50% chance to inflict 480 damage on their target over 8 sec.
	// (The Flash of Light / Holy Light healing proc is intentionally not implemented.)
	core.NewItemEffect(32489, func(agent core.Agent) {
		paladin := agent.(PaladinAgent).GetPaladin()

		// Enduring Judgement
		// https://www.wowhead.com/tbc/spell=40472/enduring-judgement
		dotSpell := paladin.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: 40472},
			SpellSchool: core.SpellSchoolHoly,
			ProcMask:    core.ProcMaskEmpty,
			Flags:       core.SpellFlagPassiveSpell,

			DamageMultiplier: 1,
			ThreatMultiplier: 1,

			Dot: core.DotConfig{
				Aura: core.Aura{
					Label:    "Enduring Judgement",
					ActionID: core.ActionID{SpellID: 40472},
				},
				NumberOfTicks: 4,
				TickLength:    2 * time.Second,
				OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
					dot.Spell.CalcAndDealPeriodicDamage(sim, target, 120, dot.OutcomeTick)
				},
			},

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.Dot(target).Apply(sim)
			},
		})

		paladin.MakeProcTriggerAura(core.ProcTrigger{
			Name:            "Ashtongue Talisman of Zeal",
			MetricsActionID: core.ActionID{SpellID: 40472},
			Callback:        core.CallbackOnSpellHitDealt,
			ClassSpellMask:  SpellMaskAllJudgements,
			Outcome:         core.OutcomeLanded,
			ProcChance:      0.5,

			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				dotSpell.Cast(sim, result.Target)
			},
		})
	})
}
