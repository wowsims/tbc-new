package warrior

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

func (war *Warrior) registerShieldBlock() {
	actionId := core.ActionID{SpellID: 2565}

	var spell *core.Spell
	aura := war.RegisterAura(core.Aura{
		Label:     "Shield Block",
		ActionID:  actionId,
		Duration:  time.Second * 5,
		MaxStacks: 1,
	}).AttachStatBuff(stats.BlockPercent, 0.75).AttachProcTrigger(core.ProcTrigger{
		Name:               "Shield Block - Consume",
		TriggerImmediately: true,
		Outcome:            core.OutcomeBlock,
		Callback:           core.CallbackOnSpellHitTaken,
		Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			spell.RelatedSelfBuff.RemoveStack(sim)
		},
	})

	spell = war.RegisterSpell(core.SpellConfig{
		ActionID:       actionId,
		SpellSchool:    core.SpellSchoolPhysical,
		ClassSpellMask: SpellMaskShieldBlock,
		Flags:          core.SpellFlagAPL | core.SpellFlagHelpful,

		RageCost: core.RageCostOptions{
			Cost: 10,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				NonEmpty: true,
			},
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    war.NewTimer(),
				Duration: time.Second * 5,
			},
		},

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return war.PseudoStats.CanBlock && war.StanceMatches(DefensiveStance)
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			spell.RelatedSelfBuff.Activate(sim)
		},

		RelatedSelfBuff: aura,
	})
}
