package warrior

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

func (war *Warrior) registerBloodrage() {
	actionID := core.ActionID{SpellID: 2687}
	rageMetrics := war.NewRageMetrics(actionID)
	healthCost := war.GetBaseStats()[stats.Health] * 0.16
	instantRage := 10.0 + 3*float64(war.Talents.ImprovedBloodrage)

	spell := war.RegisterSpell(core.SpellConfig{
		ActionID: actionID,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				NonEmpty: true,
			},
			CD: core.Cooldown{
				Timer:    war.NewTimer(),
				Duration: time.Minute,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			war.AddRage(sim, instantRage, rageMetrics)
			war.RemoveHealth(sim, healthCost)

			core.StartPeriodicAction(sim, core.PeriodicActionOptions{
				NumTicks: 10,
				Period:   time.Second * 1,
				OnAction: func(sim *core.Simulation) {
					war.AddRage(sim, 1, rageMetrics)
				},
			})
		},
	})

	war.AddMajorCooldown(core.MajorCooldown{
		Spell: spell,
		ShouldActivate: func(sim *core.Simulation, character *core.Character) bool {
			return war.CurrentRage() < 70
		},
	})
}
