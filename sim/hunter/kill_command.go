package hunter

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (hunter *Hunter) registerKillCommandSpell() {
	if hunter.Pet == nil {
		return
	}

	hunter.MakeProcTriggerAura(core.ProcTrigger{
		Name:     "Kill Command Trigger",
		Callback: core.CallbackOnSpellHitDealt,
		Outcome:  core.OutcomeCrit,

		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			hunter.killCommandEnabledUntil = sim.CurrentTime + time.Second*5
		},
	})

	hunter.KillCommand = hunter.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 34026},
		SpellSchool:    core.SpellSchoolPhysical,
		ProcMask:       core.ProcMaskMelee,
		ClassSpellMask: HunterSpellKillCommand,
		Flags:          core.SpellFlagAPL | core.SpellFlagMeleeMetrics,

		MaxRange: 45,

		ManaCost: core.ManaCostOptions{
			FlatCost: 75,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				NonEmpty: true,
			},
			CD: core.Cooldown{
				Timer:    hunter.NewTimer(),
				Duration: time.Second * 5,
			},
		},

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return hunter.killCommandEnabledUntil >= sim.CurrentTime && hunter.Pet != nil && hunter.Pet.KillCommand.CanCast(sim, target)
		},

		DamageMultiplier: 1,
		CritMultiplier:   hunter.DefaultMeleeCritMultiplier(),
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			hunter.killCommandEnabledUntil = 0
			hunter.Pet.KillCommand.Cast(sim, target)
		},
	})
}
