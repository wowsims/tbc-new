package rogue

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

var stealthRank = genRanks.Stealth.BySpellID(1784)

func (rogue *Rogue) registerStealthAura() {
	rogue.StealthAura = rogue.RegisterAura(core.Aura{
		Label:    "Stealth",
		ActionID: core.ActionID{SpellID: stealthRank.SpellID},
		Duration: core.NeverExpires,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			// Stealth triggered auras
			if rogue.MasterOfSubtletyAura != nil {
				rogue.MasterOfSubtletyAura.Duration = core.NeverExpires
				rogue.MasterOfSubtletyAura.Activate(sim)
			}
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			if rogue.MasterOfSubtletyAura != nil {
				rogue.MasterOfSubtletyAura.Deactivate(sim)
				rogue.MasterOfSubtletyAura.Duration = time.Second * 6
				rogue.MasterOfSubtletyAura.Activate(sim)
			}
		},
		// Stealth breaks on damage taken (if not absorbed)
		// This may be desirable later, but not applicable currently
	})

	rogue.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: stealthRank.SpellID},
		SpellSchool:    core.SpellSchoolPhysical,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: RogueSpellStealth,

		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    rogue.NewTimer(),
				Duration: stealthRank.Cooldown,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return sim.CurrentTime < 0
		},
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.RelatedSelfBuff.Activate(sim)
		},
		RelatedSelfBuff: rogue.StealthAura,
	})
}
