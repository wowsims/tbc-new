package shaman

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

func (shaman *Shaman) ApplyTalents() {

	//"Hotfix (2013-09-23): Lightning Bolt's damage has been increased by 10%."
	// shaman.AddStaticMod(core.SpellModConfig{
	// 	ClassMask:  SpellMaskLightningBolt | SpellMaskLightningBoltOverload,
	// 	Kind:       core.SpellMod_DamageDone_Pct,
	// 	FloatValue: 0.1,
	// })
	// //"Hotfix (2013-09-23): Flametongue Weapon's Flametongue Attack effect now deals 50% more damage."
	// //"Hotfix (2013-09-23): Windfury Weapon's Windfury Attack effect now deals 50% more damage."
	// shaman.AddStaticMod(core.SpellModConfig{
	// 	ClassMask:  SpellMaskFlametongueWeapon | SpellMaskWindfuryWeapon,
	// 	Kind:       core.SpellMod_DamageDone_Pct,
	// 	FloatValue: 0.5,
	// })

	// shaman.ApplyElementalMastery()
	// shaman.ApplyAncestralSwiftness()
	// shaman.ApplyEchoOfTheElements()
}

func (shaman *Shaman) ApplyElementalMastery() {
	if !shaman.Talents.ElementalMastery {
		return
	}

	eleMasterActionID := core.ActionID{SpellID: 16166}

	buffAura := shaman.RegisterAura(core.Aura{
		Label:    "Elemental Mastery",
		ActionID: eleMasterActionID,
		Duration: time.Second * 20,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			shaman.MultiplyCastSpeed(sim, 1.3)
			shaman.MultiplyAttackSpeed(sim, 1.3)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			shaman.MultiplyCastSpeed(sim, 1/1.3)
			shaman.MultiplyAttackSpeed(sim, 1/1.3)
		},
	})

	eleMastSpell := shaman.RegisterSpell(core.SpellConfig{
		ActionID:       eleMasterActionID,
		ClassSpellMask: SpellMaskElementalMastery,
		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    shaman.NewTimer(),
				Duration: time.Second * 90,
			},
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			buffAura.Activate(sim)
		},
	})

	shaman.AddMajorCooldown(core.MajorCooldown{
		Spell: eleMastSpell,
		Type:  core.CooldownTypeDPS,
	})
}

func (shaman *Shaman) ApplyAncestralSwiftness() {
	if !shaman.Talents.AncestralSwiftness {
		return
	}

	core.MakePermanent(shaman.RegisterAura(core.Aura{
		Label:      "Ancestral Swiftness Passive",
		BuildPhase: core.CharacterBuildPhaseTalents,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			shaman.MultiplyMeleeSpeed(sim, 1.1)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			shaman.MultiplyMeleeSpeed(sim, 1/1.1)
		},
	}).AttachMultiplyCastSpeed(1.05))

	asCdTimer := shaman.NewTimer()
	asCd := time.Second * 90

	affectedSpells := SpellMaskLightningBolt | SpellMaskChainLightning
	shaman.AncestralSwiftnessInstantAura = shaman.RegisterAura(core.Aura{
		Label:    "Ancestral swiftness",
		ActionID: core.ActionID{SpellID: 16188},
		Duration: core.NeverExpires,
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if !spell.Matches(affectedSpells) || spell.Flags.Matches(SpellFlagIsEcho) {
				return
			}
			//If both AS and MW 5 stacks buff are active, only MW gets consumed.
			//As i don't know which OnCastComplete is going to be executed first, check here if MW has not just been consumed/is active
			if shaman.Spec == proto.Spec_SpecEnhancementShaman && (shaman.MaelstromWeaponAura.TimeInactive(sim) == 0 && (!shaman.MaelstromWeaponAura.IsActive() || shaman.MaelstromWeaponAura.GetStacks() == 5)) {
				return
			}
			asCdTimer.Set(sim.CurrentTime + asCd)
			shaman.UpdateMajorCooldowns()
			aura.Deactivate(sim)
		},
	}).AttachSpellMod(core.SpellModConfig{
		ClassMask:  affectedSpells,
		Kind:       core.SpellMod_CastTime_Pct,
		FloatValue: -100,
	})

	asSpell := shaman.RegisterSpell(core.SpellConfig{
		ActionID: core.ActionID{SpellID: 16188},
		Flags:    core.SpellFlagNoOnCastComplete,
		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    asCdTimer,
				Duration: asCd,
			},
		},
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			shaman.AncestralSwiftnessInstantAura.Activate(sim)
		},
	})

	shaman.AddMajorCooldown(core.MajorCooldown{
		Spell: asSpell,
		Type:  core.CooldownTypeDPS,
	})
}

// func (shaman *Shaman) ApplyEchoOfTheElements() {
// 	if !shaman.Talents.EchoOfTheElements {
// 		return
// 	}

// 	var copySpells = map[*core.Spell]*core.Spell{}
// 	var alreadyProcced = map[*core.Spell]bool{}
// 	var lastTimestamp time.Duration

// 	const cantProc int64 = SpellMaskTotem | SpellMaskLightningShield | SpellMaskImbue | SpellMaskFulmination | SpellMaskFlameShockDot

// 	core.MakePermanent(shaman.GetOrRegisterAura(core.Aura{
// 		Label: "Echo of The Elements Dummy",
// 		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
// 			if !result.Landed() || spell.Flags.Matches(SpellFlagIsEcho) || !spell.Flags.Matches(SpellFlagShamanSpell) || spell.Matches(cantProc) {
// 				return
// 			}
// 			if sim.CurrentTime == lastTimestamp && alreadyProcced[spell] {
// 				return
// 			} else if sim.CurrentTime != lastTimestamp {
// 				lastTimestamp = sim.CurrentTime
// 				alreadyProcced = map[*core.Spell]bool{}
// 			}
// 			procChance := core.TernaryFloat64(shaman.Spec == proto.Spec_SpecElementalShaman, 0.06, 0.3)
// 			if spell.Matches(SpellMaskElementalBlast | SpellMaskElementalBlastOverload) {
// 				procChance = 0.06
// 			}
// 			if !sim.Proc(procChance, "Echo of The Elements") {
// 				return
// 			}
// 			alreadyProcced[spell] = true
// 			if copySpells[spell] == nil {
// 				copySpells[spell] = spell.Unit.RegisterSpell(core.SpellConfig{
// 					ActionID:                 core.ActionID{SpellID: spell.SpellID, Tag: core.TernaryInt32(spell.Tag == CastTagLightningOverload, 8, 7)},
// 					SpellSchool:              spell.SpellSchool,
// 					ProcMask:                 core.ProcMaskSpellProc,
// 					ApplyEffects:             spell.ApplyEffects,
// 					ManaCost:                 core.ManaCostOptions{},
// 					CritMultiplier:           shaman.DefaultCritMultiplier(),
// 					BonusCritPercent:         spell.BonusCritPercent,
// 					DamageMultiplier:         core.TernaryFloat64(spell.Tag == CastTagLightningOverload, 0.75, 1),
// 					DamageMultiplierAdditive: 1,
// 					MissileSpeed:             spell.MissileSpeed,
// 					ClassSpellMask:           spell.ClassSpellMask,
// 					BonusCoefficient:         spell.BonusCoefficient,
// 					Flags:                    spell.Flags & ^core.SpellFlagAPL | SpellFlagIsEcho,
// 					RelatedDotSpell:          spell.RelatedDotSpell,
// 				})
// 			}
// 			copySpell := copySpells[spell]
// 			copySpell.SpellMetrics[result.Target.UnitIndex].Casts--
// 			copySpell.Cast(sim, result.Target)
// 		},
// 	}))
// }
