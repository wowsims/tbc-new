package mage

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

var ItemSetAldorRegalia = core.NewItemSet(core.ItemSet{
	ID:                      648,
	Name:                    "Aldor Regalia",
	DisabledInChallengeMode: true,
	Bonuses: map[int32]core.ApplySetBonus{
		4: func(_ core.Agent, setBonusAura *core.Aura) {
			setBonusAura.AttachSpellMod(core.SpellModConfig{
				Kind:      core.SpellMod_CastTime_Flat,
				TimeValue: time.Second * -24,
				ClassMask: MageSpellPresenceOfMind,
			}).AttachSpellMod(core.SpellModConfig{
				Kind:      core.SpellMod_CastTime_Flat,
				TimeValue: time.Second * -4,
				ClassMask: MageSpellBlastWave,
			}).AttachSpellMod(core.SpellModConfig{
				Kind:      core.SpellMod_CastTime_Flat,
				TimeValue: time.Second * -40,
				ClassMask: MageSpellIceBlock,
			})
		},
	},
})

var ItemSetTirisfalRegalia = core.NewItemSet(core.ItemSet{
	ID:                      649,
	Name:                    "Tirisfal Regalia",
	DisabledInChallengeMode: true,
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.AttachSpellMod(core.SpellModConfig{
				Kind:       core.SpellMod_DamageDone_Pct,
				FloatValue: .20,
				ClassMask:  MageSpellArcaneBlast,
			}).AttachSpellMod(core.SpellModConfig{
				Kind:       core.SpellMod_PowerCost_Pct,
				FloatValue: .20,
				ClassMask:  MageSpellArcaneBlast,
			})
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			mage := agent.(MageAgent).GetMage()

			madnessAura := mage.NewTemporaryStatsAura(
				"Arcane Madness",
				core.ActionID{SpellID: 37444},
				stats.Stats{stats.SpellDamage: 70},
				time.Second*6,
			)

			setBonusAura.AttachProcTrigger(core.ProcTrigger{
				Name:     "Tirisfal 4PC",
				Callback: core.CallbackOnSpellHitDealt,
				ProcMask: core.ProcMaskSpellDamage,
				Outcome:  core.OutcomeCrit,

				Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
					madnessAura.Activate(sim)
				},
			})
		},
	},
})

var ItemSetTempestRegalia = core.NewItemSet(core.ItemSet{
	ID:                      671,
	Name:                    "Tempest Regalia",
	DisabledInChallengeMode: true,
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.AttachSpellMod(core.SpellModConfig{
				Kind:      core.SpellMod_DotNumberOfTicks_Flat,
				IntValue:  1,
				ClassMask: MageSpellEvocation,
			})
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.AttachSpellMod(core.SpellModConfig{
				Kind:       core.SpellMod_DamageDone_Pct,
				FloatValue: .05,
				ClassMask:  MageSpellFireball,
			}).AttachSpellMod(core.SpellModConfig{
				Kind:       core.SpellMod_DamageDone_Pct,
				FloatValue: .05,
				ClassMask:  MageSpellFrostbolt,
			}).AttachSpellMod(core.SpellModConfig{
				Kind:       core.SpellMod_DamageDone_Pct,
				FloatValue: .05,
				ClassMask:  MageSpellArcaneMissilesTick,
			})
		},
	},
})
