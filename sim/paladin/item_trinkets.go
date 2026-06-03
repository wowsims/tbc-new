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
}
