package tbc

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

func init() {

	// Mongoose
	// EffectID: 2673, Proc SpellID: 28093
	// PPM: 1, ICD: 0
	// Permanently enchant a Melee Weapon to occasionally increase Agility by 120 and attack speed slightly (2%).
	core.NewEnchantEffect(2673, func(agent core.Agent) {
		character := agent.GetCharacter()
		duration := time.Second * 15

		createMongooseAuras := func(tag int32) *core.StatBuffAura {
			labelSuffix := core.Ternary(tag == 1, " (MH)", " (OH)")
			slot := core.Ternary(tag == 1, proto.ItemSlot_ItemSlotMainHand, proto.ItemSlot_ItemSlotOffHand)
			aura := character.NewTemporaryStatsAuraWrapped(
				"Lightning Speed"+labelSuffix,
				core.ActionID{SpellID: 28093}.WithTag(tag),
				stats.Stats{stats.Agility: 120},
				duration,
				func(aura *core.Aura) {
					aura.ApplyOnGain(func(aura *core.Aura, sim *core.Simulation) {
						character.MultiplyAttackSpeed(sim, 1.2)
					})
					aura.ApplyOnExpire(func(aura *core.Aura, sim *core.Simulation) {
						character.MultiplyAttackSpeed(sim, 1/1.2)
					})
				},
			)
			character.AddStatProcBuff(2673, aura, true, []proto.ItemSlot{slot})
			character.ItemSwap.RegisterWeaponEnchantBuff(aura.Aura, 2673)
			return aura
		}

		mhAuras := createMongooseAuras(1)
		ohAuras := createMongooseAuras(2)

		character.MakeProcTriggerAura(core.ProcTrigger{
			Name:     "Enchant Weapon - Mongoose",
			Callback: core.CallbackOnSpellHitDealt,
			ActionID: core.ActionID{SpellID: 28093},
			DPM:      character.NewDynamicLegacyProcForEnchant(2673, 1.0, 0),
			Outcome:  core.OutcomeLanded,
			Handler: func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
				core.Ternary(spell.IsOH(), ohAuras, mhAuras).Activate(sim)
			},
		})
	})

	// Executioner
	// EffectID: 3225, Proc SpellID: 42976
	// PPM: ?, ICD: 0
	// Permanently enchant a Melee Weapon to occasionally ignore 840 of your enemy's armor.  Requires a level 60 or higher item.
	core.NewEnchantEffect(3225, func(agent core.Agent) {
		character := agent.GetCharacter()
		duration := time.Second * 15

		aura := character.NewTemporaryStatsAura(
			"Executioner",
			core.ActionID{SpellID: 42976},
			stats.Stats{stats.ArmorPenetration: 840},
			duration,
		)
		character.AddStatProcBuff(3225, aura, true, core.AllMeleeWeaponSlots())
		character.ItemSwap.RegisterWeaponEnchantBuff(aura.Aura, 3225)

		character.MakeProcTriggerAura(core.ProcTrigger{
			Name:     "Enchant Weapon - Executioner",
			Callback: core.CallbackOnSpellHitDealt,
			ActionID: core.ActionID{SpellID: 28093},
			DPM:      character.NewDynamicLegacyProcForEnchant(3225, 1.0, 0),
			Outcome:  core.OutcomeLanded,
			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				aura.Activate(sim)
			},
		})
	})

	// Deathfrost
	// EffectID: 3273, Proc SpellID: 46579, Damage SpellID: 46579, Debuff SpellID: 46629
	// Proc Chance: 50%, ICD: 25s
	// Permanently enchant a weapon so your damaging spells and melee weapon hits occasionally inflict an additional 150 Frost damage
	// and reduce the target's melee, ranged, and casting speed by 15% for 8 sec.  Requires a level 60 or higher item.
	core.NewEnchantEffect(3273, func(agent core.Agent) {
		character := agent.GetCharacter()
		duration := time.Second * 8
		effect := 0.85

		debuffArray := character.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
			return target.GetOrRegisterAura(core.Aura{
				Label:    "Deathfrost",
				Duration: duration,
				ActionID: core.ActionID{SpellID: 46629},
				OnGain: func(aura *core.Aura, sim *core.Simulation) {
					aura.Unit.MultiplyAttackSpeed(sim, effect)
					aura.Unit.MultiplyRangedSpeed(sim, effect)
					aura.Unit.MultiplyCastSpeed(sim, effect)
				},
				OnExpire: func(aura *core.Aura, sim *core.Simulation) {
					aura.Unit.MultiplyAttackSpeed(sim, 1/effect)
					aura.Unit.MultiplyRangedSpeed(sim, 1/effect)
					aura.Unit.MultiplyCastSpeed(sim, 1/effect)
				},
			})
		})

		dfSpell := character.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: 46579},
			SpellSchool: core.SpellSchoolFrost,
			Flags:       core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell,
			ProcMask:    core.ProcMaskSpellProc,

			DamageMultiplier: 1,
			CritMultiplier:   character.DefaultSpellCritMultiplier(),
			ThreatMultiplier: 1,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.CalcAndDealDamage(sim, target, 150, spell.OutcomeMagicCrit)
			},
		})

		character.MakeProcTriggerAura(core.ProcTrigger{
			Name:     "Enchant Weapon - Deathfrost",
			Callback: core.CallbackOnSpellHitDealt,
			ActionID: core.ActionID{SpellID: 46579},
			DPM:      character.NewFixedProcChanceManager(0.5, core.ProcMaskMeleeOrMeleeProc|core.ProcMaskSpellOrSpellProc),
			Outcome:  core.OutcomeLanded,
			ICD:      time.Second * 25,
			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				debuffArray.Get(result.Target).Activate(sim)
				dfSpell.Cast(sim, result.Target)
			},
		})
	})

	// Scopes
	// The ratings for these don't exist, so just apply a spellmod for Ranged-flagged things

}
