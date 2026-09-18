package priest

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

var shadowfiendRank = genRanks.Shadowfiend.BySpellID(34433)

func (priest *Priest) registerShadowfiendSpell() {
	actionID := core.ActionID{SpellID: shadowfiendRank.SpellID}

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
		Rank:           shadowfiendRank.Rank,

		ManaCost: core.ManaCostOptions{
			BaseCostPercent: 6,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: shadowfiendRank.GCD,
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
