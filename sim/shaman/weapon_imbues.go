package shaman

import (
	"fmt"
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

const (
	frostbrandEnchantID  int32 = 2
	flametongueEnchantID int32 = 5
	windfuryEnchantID    int32 = 283
	earthlivingEnchantID int32 = 3345
	rockbiterEnchantID   int32 = 3021
)

func (shaman *Shaman) RegisterOnItemSwapWithImbue(effectID int32, procMask *core.ProcMask, aura *core.Aura) {
	shaman.RegisterItemSwapCallback(core.AllWeaponSlots(), func(sim *core.Simulation, slot proto.ItemSlot) {
		mask := core.ProcMaskUnknown
		if shaman.MainHand().TempEnchant == effectID {
			mask |= core.ProcMaskMeleeMH
		}
		if shaman.OffHand().TempEnchant == effectID {
			mask |= core.ProcMaskMeleeOH
		}
		*procMask = mask

		if mask == core.ProcMaskUnknown {
			aura.Deactivate(sim)
		} else {
			aura.Activate(sim)
		}
	})
}

func (shaman *Shaman) setupItemSwapImbue(imbue proto.ShamanImbue, imbueID int32) {
	if shaman.ItemSwap.IsEnabled() {
		if mhSwap := shaman.ItemSwap.GetUnequippedItemBySlot(proto.ItemSlot_ItemSlotMainHand); mhSwap != nil && shaman.SelfBuffs.ImbueMHSwap == imbue {
			mhSwap.TempEnchant = imbueID
			shaman.ItemSwap.AddTempEnchant(imbueID, proto.ItemSlot_ItemSlotMainHand, true)
		}
		if ohSwap := shaman.ItemSwap.GetUnequippedItemBySlot(proto.ItemSlot_ItemSlotOffHand); ohSwap != nil && shaman.SelfBuffs.ImbueOHSwap == imbue {
			ohSwap.TempEnchant = imbueID
			shaman.ItemSwap.AddTempEnchant(imbueID, proto.ItemSlot_ItemSlotOffHand, true)
		}
	}
}

func (shaman *Shaman) newWindfuryImbueSpell(isMH bool) *core.Spell {
	apBonus := 475.0

	tag := 1
	procMask := core.ProcMaskMeleeMHSpecial
	weaponDamageFunc := shaman.MHWeaponDamage
	if !isMH {
		tag = 2
		procMask = core.ProcMaskMeleeOHSpecial
		weaponDamageFunc = shaman.OHWeaponDamage
		apBonus *= 2 // applied after 50% offhand penalty
	}

	spellConfig := core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 25505, Tag: int32(tag)},
		SpellSchool:    core.SpellSchoolPhysical,
		ProcMask:       procMask,
		ClassSpellMask: SpellMaskWindfuryWeapon,
		Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagPassiveSpell,

		DamageMultiplier: 1,
		CritMultiplier:   shaman.DefaultMeleeCritMultiplier(),
		ThreatMultiplier: 1,
		BonusCoefficient: 1,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			mAP := spell.MeleeAttackPower() + apBonus

			baseDamage1 := weaponDamageFunc(sim, mAP)
			baseDamage2 := weaponDamageFunc(sim, mAP)
			result1 := spell.CalcDamage(sim, target, baseDamage1, spell.OutcomeMeleeSpecialHitAndCrit)
			result2 := spell.CalcDamage(sim, target, baseDamage2, spell.OutcomeMeleeSpecialHitAndCrit)
			spell.DealDamage(sim, result1)
			spell.DealDamage(sim, result2)
		},
	}

	return shaman.RegisterSpell(spellConfig)
}

func (shaman *Shaman) makeWFProcTriggerAura(dpm *core.DynamicProcManager, procMask *core.ProcMask, mhSpell *core.Spell, ohSpell *core.Spell) *core.Aura {
	icd := &core.Cooldown{
		Timer:    shaman.NewTimer(),
		Duration: time.Second * 3,
	}
	aura := shaman.RegisterAura(core.Aura{
		Label:    "Windfury Imbue",
		Icd:      icd,
		Dpm:      dpm,
		Duration: core.NeverExpires,
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !aura.Icd.IsReady(sim) || !result.Landed() || !spell.ProcMask.Matches(*procMask) || !dpm.Proc(sim, *procMask, "Windfury Imbue") {
				return
			}
			aura.Icd.Use(sim)
			if spell.IsMH() {
				mhSpell.Cast(sim, result.Target)
			} else {
				ohSpell.Cast(sim, result.Target)
			}
		},
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
	})
	return aura
}

func (shaman *Shaman) getWindfuryFixedProcChance(procMask core.ProcMask) float64 {
	return core.TernaryFloat64(procMask == core.ProcMaskMelee, 0.36, 0.2)
}

func (shaman *Shaman) RegisterWindfuryImbue(procMask core.ProcMask) {
	if procMask == core.ProcMaskUnknown && !shaman.ItemSwap.IsEnabled() {
		return
	}

	mask := core.ProcMaskUnknown

	mH := shaman.MainHand()
	if mH != nil && shaman.SelfBuffs.ImbueMH == proto.ShamanImbue_WindfuryWeapon {
		mH.TempEnchant = windfuryEnchantID
		if shaman.ItemSwap.IsEnabled() {
			shaman.ItemSwap.AddTempEnchant(windfuryEnchantID, proto.ItemSlot_ItemSlotMainHand, false)
		}
		mask |= core.ProcMaskMeleeMH
	}
	oH := shaman.OffHand()
	if oH != nil && shaman.SelfBuffs.ImbueOH == proto.ShamanImbue_WindfuryWeapon {
		oH.TempEnchant = windfuryEnchantID
		if shaman.ItemSwap.IsEnabled() {
			shaman.ItemSwap.AddTempEnchant(windfuryEnchantID, proto.ItemSlot_ItemSlotOffHand, false)
		}
		mask |= core.ProcMaskMeleeOH
	}

	shaman.setupItemSwapImbue(proto.ShamanImbue_WindfuryWeapon, windfuryEnchantID)

	dpm := shaman.NewDynamicLegacyProcForTempEnchant(windfuryEnchantID, 0, shaman.getWindfuryFixedProcChance)

	mhSpell := shaman.newWindfuryImbueSpell(true)
	ohSpell := shaman.newWindfuryImbueSpell(false)

	aura := shaman.makeWFProcTriggerAura(dpm, &mask, mhSpell, ohSpell)

	shaman.RegisterOnItemSwapWithImbue(windfuryEnchantID, &mask, aura)
}

func (shaman *Shaman) newFlametongueImbueSpell(weapon *core.Item) *core.Spell {
	return shaman.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{SpellID: 25489},
		SpellSchool:      core.SpellSchoolFire,
		ProcMask:         core.ProcMaskSpellDamageProc,
		ClassSpellMask:   SpellMaskFlametongueWeapon,
		Flags:            core.SpellFlagPassiveSpell | SpellFlagShamanSpell,
		DamageMultiplier: 1,
		CritMultiplier:   shaman.DefaultMeleeCritMultiplier(),
		ThreatMultiplier: 1,
		BonusCoefficient: 0.10000000149,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			if weapon.SwingSpeed != 0 {
				baseDamage := weapon.SwingSpeed * 35 // from old tbc sim
				spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
			}
		},
	})
}

func (shaman *Shaman) makeFTProcTriggerAura(itemSlot proto.ItemSlot, triggerProcMask core.ProcMask, flameTongueSpell *core.Spell) *core.Aura {
	aura := shaman.MakeProcTriggerAura(core.ProcTrigger{
		Name:               fmt.Sprintf("Flametongue Imbue %s", itemSlot),
		ProcMask:           triggerProcMask,
		Outcome:            core.OutcomeLanded,
		Callback:           core.CallbackOnSpellHitDealt,
		TriggerImmediately: true,

		Handler: func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
			flameTongueSpell.Cast(sim, result.Target)
		},
	})

	shaman.RegisterItemSwapCallback([]proto.ItemSlot{itemSlot}, func(sim *core.Simulation, is proto.ItemSlot) {
		if is == proto.ItemSlot_ItemSlotMainHand {
			mh := shaman.MainHand()
			mhSwap := shaman.ItemSwap.GetUnequippedItemBySlot(is)
			if mh.TempEnchant != flametongueEnchantID {
				// The new main hand does not have flametongue on, so deactivate
				aura.Deactivate(sim)
				return
			}
			if mhSwap.TempEnchant != flametongueEnchantID {
				// The new main hand has flametongue on and the swapped one does not, so need to activate
				aura.Activate(sim)
				return
			}
		}
		if is == proto.ItemSlot_ItemSlotOffHand {
			oh := shaman.OffHand()
			ohSwap := shaman.ItemSwap.GetUnequippedItemBySlot(is)
			if oh.TempEnchant != flametongueEnchantID {
				// The new offhand does not have flametongue on, so deactivate
				aura.Deactivate(sim)
				return
			}
			if ohSwap.TempEnchant != flametongueEnchantID {
				// The new offhand has flametongue on and the swapped one does not, so need to activate
				aura.Activate(sim)
				return
			}

		}
	})

	return aura
}

func (shaman *Shaman) RegisterFlametongueImbue(procMask core.ProcMask) {
	if procMask == core.ProcMaskUnknown && !shaman.ItemSwap.IsEnabled() {
		return
	}

	for _, itemSlot := range core.AllWeaponSlots() {
		var weapon *core.Item
		var triggerProcMask core.ProcMask
		switch {
		case shaman.SelfBuffs.ImbueMH == proto.ShamanImbue_FlametongueWeapon && itemSlot == proto.ItemSlot_ItemSlotMainHand:
			weapon = shaman.MainHand()
			triggerProcMask = core.ProcMaskMeleeMH | core.ProcMaskMeleeProc
		case shaman.SelfBuffs.ImbueOH == proto.ShamanImbue_FlametongueWeapon && itemSlot == proto.ItemSlot_ItemSlotOffHand:
			weapon = shaman.OffHand()
			triggerProcMask = core.ProcMaskMeleeOH
		}

		if weapon == nil {
			continue
		}

		weapon.TempEnchant = flametongueEnchantID

		if shaman.ItemSwap.IsEnabled() {
			shaman.ItemSwap.AddTempEnchant(flametongueEnchantID, itemSlot, false)
		}

		flameTongueSpell := shaman.newFlametongueImbueSpell(weapon)
		shaman.makeFTProcTriggerAura(itemSlot, triggerProcMask, flameTongueSpell)
	}

	shaman.setupItemSwapImbue(proto.ShamanImbue_FlametongueWeapon, flametongueEnchantID)
}

func (shaman *Shaman) newFrostbrandImbueSpell() *core.Spell {
	return shaman.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 8033},
		SpellSchool:    core.SpellSchoolFrost,
		ClassSpellMask: SpellMaskFrostbrandWeapon,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagPassiveSpell | SpellFlagShamanSpell,

		DamageMultiplier: 1,
		CritMultiplier:   shaman.DefaultMeleeCritMultiplier(),
		ThreatMultiplier: 1,
		BonusCoefficient: 0.10000000149,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := 0.0 //spell id 8034
			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
		},
	})
}

func (shaman *Shaman) RegisterFrostbrandImbue(procMask core.ProcMask) {
	if procMask == core.ProcMaskUnknown && !shaman.ItemSwap.IsEnabled() {
		return
	}

	mH := shaman.MainHand()
	if mH != nil && shaman.SelfBuffs.ImbueMH == proto.ShamanImbue_FrostbrandWeapon {
		mH.TempEnchant = frostbrandEnchantID
		if shaman.ItemSwap.IsEnabled() {
			shaman.ItemSwap.AddTempEnchant(frostbrandEnchantID, proto.ItemSlot_ItemSlotMainHand, false)
		}
	}
	oH := shaman.OffHand()
	if oH != nil && shaman.SelfBuffs.ImbueOH == proto.ShamanImbue_FrostbrandWeapon {
		oH.TempEnchant = frostbrandEnchantID
		if shaman.ItemSwap.IsEnabled() {
			shaman.ItemSwap.AddTempEnchant(frostbrandEnchantID, proto.ItemSlot_ItemSlotOffHand, false)
		}
	}

	shaman.setupItemSwapImbue(proto.ShamanImbue_FrostbrandWeapon, frostbrandEnchantID)

	dpm := shaman.NewDynamicLegacyProcForTempEnchant(frostbrandEnchantID, 9.0, func(pm core.ProcMask) float64 { return 0 })

	fbSpell := shaman.newFrostbrandImbueSpell()

	aura := shaman.MakeProcTriggerAura(core.ProcTrigger{
		Name:               "Frostbrand Imbue",
		Callback:           core.CallbackOnSpellHitDealt,
		Outcome:            core.OutcomeLanded,
		DPM:                dpm,
		TriggerImmediately: true,

		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			fbSpell.Cast(sim, result.Target)
		},
	})

	shaman.RegisterOnItemSwapWithImbue(frostbrandEnchantID, &procMask, aura)
}
