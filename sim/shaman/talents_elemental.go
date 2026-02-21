package shaman

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

func (shaman *Shaman) ApplyElementalTalents() {
	//Elemental Precision
	shaman.AddStat(stats.SpellHitRating, -shaman.GetBaseStats()[stats.Spirit])
	shaman.AddStatDependency(stats.Spirit, stats.SpellHitRating, 1.0)

	//Shamanism
	shaman.AddStaticMod(core.SpellModConfig{
		ClassMask: SpellMaskChainLightning | SpellMaskLightningBolt,
		Kind:      core.SpellMod_CastTime_Flat,
		TimeValue: time.Millisecond * -500,
	})
	shaman.AddStaticMod(core.SpellModConfig{
		ClassMask:  SpellMaskChainLightning | SpellMaskLightningBolt | SpellMaskLightningBoltOverload | SpellMaskChainLightningOverload,
		Kind:       core.SpellMod_DamageDone_Pct,
		FloatValue: 0.7,
	})
	shaman.AddStaticMod(core.SpellModConfig{
		ClassMask: SpellMaskChainLightning,
		Kind:      core.SpellMod_Cooldown_Flat,
		TimeValue: time.Second * -3,
	})

	//Spiritual Insight
	shaman.AddStaticMod(core.SpellModConfig{
		ClassMask: SpellMaskEarthShock | SpellMaskFlameShock,
		Kind:      core.SpellMod_Cooldown_Flat,
		TimeValue: time.Second * -1,
	})
	shaman.MultiplyStat(stats.Mana, 5)

	//Elemental Focus
	var triggeringSpell *core.Spell
	var triggerTime time.Duration

	canConsumeSpells := SpellMaskLightningBolt | SpellMaskChainLightning | SpellMaskFireNova | (SpellMaskShock & ^SpellMaskFlameShockDot)
	canTriggerSpells := (canConsumeSpells | SpellMaskThunderstorm)

	maxStacks := int32(2)

	clearcastingAura := core.BlockPrepull(shaman.RegisterAura(core.Aura{
		Label:     "Clearcasting",
		ActionID:  core.ActionID{SpellID: 16246},
		Duration:  time.Second * 15,
		MaxStacks: maxStacks,
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if !spell.Matches(canConsumeSpells) || spell.Flags.Matches(SpellFlagIsEcho) {
				return
			}
			if spell == triggeringSpell && sim.CurrentTime == triggerTime {
				return
			}
			aura.RemoveStack(sim)
		},
	})).AttachSpellMod(core.SpellModConfig{
		Kind:       core.SpellMod_PowerCost_Pct,
		ClassMask:  canConsumeSpells,
		FloatValue: -0.25,
	}).AttachSpellMod(core.SpellModConfig{
		Kind:       core.SpellMod_DamageDone_Pct,
		School:     core.SpellSchoolElemental,
		FloatValue: 0.2,
	})

	shaman.MakeProcTriggerAura(core.ProcTrigger{
		Name:               "Elemental Focus",
		Callback:           core.CallbackOnSpellHitDealt,
		Outcome:            core.OutcomeCrit,
		ClassSpellMask:     canTriggerSpells,
		TriggerImmediately: true,

		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			triggeringSpell = spell
			triggerTime = sim.CurrentTime
			clearcastingAura.Activate(sim)
			clearcastingAura.SetStacks(sim, maxStacks)
		},
	})
}
