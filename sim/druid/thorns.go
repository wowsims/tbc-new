package druid

import (
	"github.com/wowsims/tbc/sim/core"
)

// Self-cast Thorns (rank 7). Reuses the core raid-buff aura, passing the
// druid's own Brambles talent points. If the Thorns raid buff is selected it
// is already registered (buffs apply before Initialize) and wins; otherwise
// the self-cast version uses the druid's actual talent.
func (druid *Druid) registerThornsSpell() {
	thornsAura := druid.GetAura("Thorns")
	if thornsAura == nil {
		thornsAura = core.ThornsAura(druid.GetCharacter(), druid.Talents.Brambles)
	}

	druid.RegisterSpell(Humanoid, core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 26992},
		SpellSchool:    core.SpellSchoolNature,
		Flags:          core.SpellFlagAPL | core.SpellFlagHelpful,
		ClassSpellMask: DruidSpellThorns,
		ProcMask:       core.ProcMaskEmpty,

		ManaCost: core.ManaCostOptions{
			FlatCost: 400,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			thornsAura.Activate(sim)
		},

		RelatedSelfBuff: thornsAura,
	})
}
