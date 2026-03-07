package shaman

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

func (shaman *Shaman) newTotemSpellConfig(flatCost int32, spellID int32, spellMask int64) core.SpellConfig {
	return core.SpellConfig{
		ActionID:       core.ActionID{SpellID: spellID},
		Flags:          core.SpellFlagAPL | SpellFlagInstant,
		ClassSpellMask: spellMask,

		ManaCost: core.ManaCostOptions{
			FlatCost: flatCost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: time.Second,
			},
		},
	}
}

func (shaman *Shaman) registerWindfuryTotemSpell() {
	duration := time.Second * 120
	value := 445 * (1 + 0.15*float64(shaman.Talents.ImprovedWeaponTotems))

	wfProcAura := shaman.NewTemporaryStatsAura("Windfury Totem Proc (Self)", core.ActionID{SpellID: 25584}, stats.Stats{stats.AttackPower: value}, time.Millisecond*1500)
	wfProcAura.MaxStacks = 2
	wfProcAura.AttachProcTrigger(core.ProcTrigger{
		Name:     "Windfury Attack (Self)",
		Callback: core.CallbackOnSpellHitDealt,
		ProcMask: core.ProcMaskMeleeMHAuto | core.ProcMaskMeleeOHAuto,
		// TriggerImmediately ommited for improved UI clarity (the timeline tick would be near invisible for MHAuto procs)
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if wfProcAura.IsActive() && !spell.ProcMask.Matches(core.ProcMaskMeleeSpecial) {
				wfProcAura.RemoveStack(sim)
				if wfProcAura.GetStacks() == 0 {
					wfProcAura.Deactivate(sim)
				}
			}
		},
	})

	config := shaman.newTotemSpellConfig(325, 25587, SpellMaskBasicTotem)

	var windfurySpell *core.Spell
	wfProcTrigger := shaman.MakeProcTriggerAura(core.ProcTrigger{
		Name:               "Windfury Totem Trigger (Self)",
		MetricsActionID:    core.ActionID{SpellID: 25580},
		ProcChance:         0.2,
		Duration:           core.NeverExpires,
		Outcome:            core.OutcomeLanded,
		Callback:           core.CallbackOnSpellHitDealt,
		ProcMask:           core.ProcMaskMeleeMHAuto,
		ICD:                time.Millisecond * 1500,
		TriggerImmediately: true,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			wfProcAura.Activate(sim)
			if spell.ProcMask == core.ProcMaskMeleeMHAuto {
				wfProcAura.SetStacks(sim, 1)
			} else {
				wfProcAura.SetStacks(sim, 2)
			}
			shaman.AutoAttacks.MaybeReplaceMHSwing(sim, windfurySpell).Cast(sim, result.Target)
		},
	})

	wfIntermediateAuraForExclusitivity := shaman.RegisterAura(core.Aura{
		Label:    "Windfury Dummy Aura (self)",
		Duration: time.Second * 10,
	})

	wfAura := shaman.RegisterAura(core.Aura{
		Label:    "Windfury Totem (Self)",
		ActionID: config.ActionID,
		Duration: duration,
	}).ApplyOnInit(func(aura *core.Aura, sim *core.Simulation) {
		mhConfig := *shaman.AutoAttacks.MHConfig()
		mhConfig.ActionID = mhConfig.ActionID.WithTag(25584)
		windfurySpell = shaman.GetOrRegisterSpell(mhConfig)
	}).AttachPeriodicAction(core.PeriodicActionOptions{
		Period:          time.Second * 5,
		TickImmediately: true,
		Priority:        core.ActionPriorityAuto,
		OnAction: func(sim *core.Simulation) {
			wfIntermediateAuraForExclusitivity.Activate(sim)
		},
	})

	wfIntermediateAuraForExclusitivity.NewExclusiveEffect(core.WindfuryTotemCategory, false, core.ExclusiveEffect{
		Priority: value,
		OnGain: func(_ *core.ExclusiveEffect, sim *core.Simulation) {
			wfProcTrigger.Activate(sim)
		},
		OnExpire: func(_ *core.ExclusiveEffect, sim *core.Simulation) {
			wfProcTrigger.Deactivate(sim)
			wfIntermediateAuraForExclusitivity.Deactivate(sim)
		},
	})

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		if shaman.AirTotemAura != nil {
			shaman.AirTotemAura.Deactivate(sim)
		}
		shaman.TotemExpirations[AirTotem] = sim.CurrentTime + duration
		shaman.AirTotemAura = wfAura
		wfAura.Activate(sim)
	}

	shaman.RegisterItemSwapCallback([]proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand}, func(sim *core.Simulation, slot proto.ItemSlot) {
		wfIntermediateAuraForExclusitivity.Deactivate(sim)
	})

	shaman.RegisterSpell(config)
}

func (shaman *Shaman) registerStrengthOfEarthTotemSpell() {
	duration := time.Second * 120
	value := 86 * []float64{1, 1.08, 1.15}[shaman.Talents.EnhancingTotems]
	config := shaman.newTotemSpellConfig(300, 25528, SpellMaskBasicTotem)
	buffAura := shaman.RegisterAura(core.Aura{
		Label:    "Strength Of Earth Totem (Self)",
		ActionID: config.ActionID,
		Duration: duration,
	})
	buffAura.NewExclusiveEffect(core.StrengthOfEarthTotemCategory+stats.Strength.StatName()+"Add", false, core.ExclusiveEffect{
		Priority: value,
		OnGain: func(ee *core.ExclusiveEffect, sim *core.Simulation) {
			ee.Aura.Unit.AddStatDynamic(sim, stats.Strength, value)
		},
		OnExpire: func(ee *core.ExclusiveEffect, sim *core.Simulation) {
			ee.Aura.Unit.AddStatDynamic(sim, stats.Strength, -value)
		},
	})
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		if shaman.EarthTotemAura != nil {
			shaman.EarthTotemAura.Deactivate(sim)
		}
		shaman.TotemExpirations[EarthTotem] = sim.CurrentTime + duration
		shaman.EarthTotemAura = buffAura
		buffAura.Activate(sim)
	}
	shaman.RegisterSpell(config)
}

func (shaman *Shaman) registerGraceOfAirTotemSpell() {
	duration := time.Second * 120
	value := 77 * []float64{1, 1.08, 1.15}[shaman.Talents.EnhancingTotems]
	config := shaman.newTotemSpellConfig(310, 25359, SpellMaskBasicTotem)
	buffAura := shaman.RegisterAura(core.Aura{
		Label:    "Grace Of Air Totem (Self)",
		ActionID: config.ActionID,
		Duration: duration,
	})
	buffAura.NewExclusiveEffect(core.GraceOfAirTotemCategory+stats.Agility.StatName()+"Add", false, core.ExclusiveEffect{
		Priority: value,
		OnGain: func(ee *core.ExclusiveEffect, sim *core.Simulation) {
			ee.Aura.Unit.AddStatDynamic(sim, stats.Agility, value)
		},
		OnExpire: func(ee *core.ExclusiveEffect, sim *core.Simulation) {
			ee.Aura.Unit.AddStatDynamic(sim, stats.Agility, -value)
		},
	})
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		if shaman.AirTotemAura != nil {
			shaman.AirTotemAura.Deactivate(sim)
		}
		shaman.TotemExpirations[AirTotem] = sim.CurrentTime + duration
		shaman.AirTotemAura = buffAura
		buffAura.Activate(sim)
	}
	shaman.RegisterSpell(config)
}

func (shaman *Shaman) registerWrathOfAirTotemSpell() {
	duration := time.Second * 120
	config := shaman.newTotemSpellConfig(320, 3738, SpellMaskBasicTotem)
	buffAura := shaman.RegisterAura(core.Aura{
		Label:    "Wrath Of Air Totem (Self)",
		ActionID: config.ActionID,
		Duration: duration,
	})

	statsAura := core.WrathOfAirTotemAura(&shaman.Character, shaman.Character.CouldHaveSetBonus(ItemSetCycloneRegalia, 2))
	buffAura.AttachDependentAura(statsAura)
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		if shaman.AirTotemAura != nil {
			shaman.AirTotemAura.Deactivate(sim)
		}
		shaman.TotemExpirations[AirTotem] = sim.CurrentTime + duration
		shaman.AirTotemAura = buffAura
		buffAura.Activate(sim)
	}
	shaman.RegisterSpell(config)
}

func (shaman *Shaman) registerManaSpringTotemSpell() {
	duration := time.Second * 120
	value := 50 * (1 + 0.05*float64(shaman.Talents.RestorativeTotems))
	config := shaman.newTotemSpellConfig(120, 25570, SpellMaskBasicTotem)
	buffAura := shaman.RegisterAura(core.Aura{
		Label:    "Mana Spring Totem (Self)",
		ActionID: config.ActionID,
		Duration: duration,
	})
	buffAura.NewExclusiveEffect(core.ManaSpringTotemCategory+stats.MP5.StatName()+"Add", false, core.ExclusiveEffect{
		Priority: value,
		OnGain: func(ee *core.ExclusiveEffect, sim *core.Simulation) {
			ee.Aura.Unit.AddStatDynamic(sim, stats.MP5, value)
		},
		OnExpire: func(ee *core.ExclusiveEffect, sim *core.Simulation) {
			ee.Aura.Unit.AddStatDynamic(sim, stats.MP5, -value)
		},
	})
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		if shaman.WaterTotemAura != nil {
			shaman.WaterTotemAura.Deactivate(sim)
		}
		shaman.TotemExpirations[WaterTotem] = sim.CurrentTime + duration
		shaman.WaterTotemAura = buffAura
		buffAura.Activate(sim)
	}
	shaman.RegisterSpell(config)
}

/* func (shaman *Shaman) registerHealingStreamTotemSpell() {
	config := shaman.newTotemSpellConfig(3, 5394, SpellMaskBasicTotem)
	hsHeal := shaman.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{SpellID: 5394},
		SpellSchool:      core.SpellSchoolNature,
		ProcMask:         core.ProcMaskEmpty,
		Flags:            core.SpellFlagHelpful | core.SpellFlagNoOnCastComplete | SpellFlagInstant,
		DamageMultiplier: 1,
		CritMultiplier:   1,
		ThreatMultiplier: 1,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			healing := 28 + spell.HealingPower(target)*0.08272
			spell.CalcAndDealHealing(sim, target, healing, spell.OutcomeHealing)
		},
	})
	config.Hot = core.DotConfig{
		Aura: core.Aura{
			Label: "HealingStreamHot",
		},
		NumberOfTicks: 150,
		TickLength:    time.Second * 2,
		OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
			hsHeal.Cast(sim, target)
		},
	}
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		shaman.TotemExpirations[WaterTotem] = sim.CurrentTime + time.Second*300
		for _, agent := range shaman.Party.Players {
			spell.Hot(&agent.GetCharacter().Unit).Activate(sim)
		}
	}
	shaman.HealingStreamTotem = shaman.RegisterSpell(config)
} */
