package priest

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (priest *Priest) registerShadowfiendSpell() {
	actionID := core.ActionID{SpellID: 34433}

	// Timeline aura
	priest.ShadowfiendAura = priest.RegisterAura(core.Aura{
		ActionID: actionID,
		Label:    "Shadowfiend",
		Duration: time.Second * 15,
	})

	priest.Shadowfiend = priest.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolShadow,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: PriestSpellShadowFiend,

		ManaCost: core.ManaCostOptions{
			BaseCostPercent: 6,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    priest.NewTimer(),
				Duration: time.Minute * 5,
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			priest.ShadowfiendPet.EnableWithTimeout(sim, priest.ShadowfiendPet, spell.RelatedSelfBuff.Duration)
			spell.RelatedSelfBuff.Activate(sim)
		},

		RelatedSelfBuff: priest.ShadowfiendAura,
	})
}
