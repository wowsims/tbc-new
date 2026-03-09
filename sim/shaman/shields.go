package shaman

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

func (shaman *Shaman) registerShieldsSpells() {
	shaman.registerWaterShieldSpell()
	shaman.registerLightningShieldSpell()
	shaman.registerShieldEffectTriggerSpell()
}

func (shaman *Shaman) registerShieldEffectTriggerSpell() {
	shaman.ShieldSelfProcSpell = shaman.RegisterSpell(core.SpellConfig{
		Flags:          core.SpellFlagNoMetrics | core.SpellFlagNoLogs,
		ClassSpellMask: SpellMaskShieldSelfProc,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealDamage(sim, target, 0, spell.OutcomeAlwaysHit)
		},
	})
}

func (shaman *Shaman) startShieldProcPeriodicAction(sim *core.Simulation) {
	if shaman.SelfBuffs.ShieldProcrate > 0 {
		core.StartPeriodicAction(sim, core.PeriodicActionOptions{
			Period:   60 * time.Second / time.Duration(shaman.SelfBuffs.ShieldProcrate),
			Priority: core.ActionPriorityGCD,
			OnAction: func(sim *core.Simulation) {
				shaman.ShieldSelfProcSpell.Cast(sim, &shaman.Unit)
			},
		})
	}
}

func (shaman *Shaman) registerWaterShieldSpell() {
	manaReturned := 204.0
	mp5 := 50.0
	if shaman.CouldHaveSetBonus(ItemSetTidefuryRaiment, 4) {
		manaReturned += 56
	}

	actionID := core.ActionID{SpellID: 33736}
	waterShieldManaMetrics := shaman.NewManaMetrics(actionID)

	shaman.WaterShieldAura = shaman.RegisterAura(core.Aura{
		Label:     "Water Shield",
		ActionID:  actionID,
		Duration:  10 * time.Minute,
		MaxStacks: 3,
	}).AttachProcTrigger(core.ProcTrigger{
		Name:           "Water Shield Trigger",
		Callback:       core.CallbackOnSpellHitTaken,
		ICD:            3500 * time.Millisecond,
		ClassSpellMask: SpellMaskShieldSelfProc,
		Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			shaman.WaterShieldAura.RemoveStack(sim)
			shaman.AddMana(sim, manaReturned, waterShieldManaMetrics)
		},
	}).AttachStatBuff(stats.MP5, mp5)

	shaman.RegisterSpell(core.SpellConfig{
		ActionID:    actionID,
		SpellSchool: core.SpellSchoolNature,
		Flags:       core.SpellFlagAPL | SpellFlagInstant,
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			shaman.LightningShieldAura.Deactivate(sim)
			shaman.WaterShieldAura.Activate(sim)
			shaman.WaterShieldAura.SetStacks(sim, 3)
		},
		RelatedSelfBuff: shaman.WaterShieldAura,
	})
}

func (shaman *Shaman) registerLightningShieldSpell() {
	actionID := core.ActionID{SpellID: 25472}

	lsDamage := shaman.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{SpellID: 25472},
		SpellSchool:      core.SpellSchoolNature,
		ProcMask:         core.ProcMaskEmpty,
		Flags:            SpellFlagShamanSpell,
		ClassSpellMask:   SpellMaskLightningShield,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		CritMultiplier:   shaman.DefaultSpellCritMultiplier(),
		BonusCoefficient: 0.26699998975,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := 287.0
			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
		},
	})

	shaman.LightningShieldAura = shaman.RegisterAura(core.Aura{
		Label:     "Lightning Shield",
		ActionID:  actionID,
		Duration:  10 * time.Minute,
		MaxStacks: 3,
	}).AttachProcTrigger(core.ProcTrigger{
		Name:           "Lightning Shield Trigger",
		Callback:       core.CallbackOnSpellHitTaken,
		ICD:            3500 * time.Millisecond,
		ClassSpellMask: SpellMaskShieldSelfProc,
		Handler: func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
			shaman.LightningShieldAura.RemoveStack(sim)
			lsDamage.Cast(sim, shaman.CurrentTarget)
		},
	})

	shaman.RegisterSpell(core.SpellConfig{
		ActionID:    actionID,
		SpellSchool: core.SpellSchoolNature,
		Flags:       core.SpellFlagAPL | SpellFlagInstant,
		ManaCost: core.ManaCostOptions{
			FlatCost: 400,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			shaman.WaterShieldAura.Deactivate(sim)
			shaman.LightningShieldAura.Activate(sim)
			shaman.LightningShieldAura.SetStacks(sim, 3)
		},
		RelatedSelfBuff: shaman.LightningShieldAura,
	})
}
