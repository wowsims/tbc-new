package druid

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (druid *Druid) registerMaulSpell() {
	// The actual Maul spell that fires on the next auto-attack swing.
	maulSpell := druid.RegisterSpell(Bear, core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 26996},
		SpellSchool:    core.SpellSchoolPhysical,
		ProcMask:       core.ProcMaskMeleeMHSpecial,
		ClassSpellMask: DruidSpellMaul,
		Flags:          core.SpellFlagMeleeMetrics,

		RageCost: core.RageCostOptions{
			Cost:   15,
			Refund: 0.8,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				NonEmpty: true,
			},
		},

		DamageMultiplier: 1,
		CritMultiplier:   druid.FeralCritMultiplier(),
		ThreatMultiplier: 1.75,
		FlatThreatBonus:  176,
		MaxRange:         core.MaxMeleeRange,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := 176 + druid.IdolMaulBonus + spell.Unit.MHWeaponDamage(sim, spell.MeleeAttackPower(target))
			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
			if !result.Landed() {
				spell.IssueRefund(sim)
			}
			if druid.maulQueueAura != nil {
				druid.maulQueueAura.Deactivate(sim)
			}
		},

		ExpectedInitialDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
			baseDamage := 176 + druid.IdolMaulBonus + spell.Unit.AutoAttacks.MH().CalculateAverageWeaponDamage(spell.MeleeAttackPower(target))
			return spell.CalcDamage(sim, target, baseDamage, spell.OutcomeExpectedMeleeWeaponSpecialHitAndCrit)
		},
	})

	druid.Maul = druid.makeMaulQueueSpellAndAura(maulSpell)
}

// makeMaulQueueSpellAndAura creates the APL-visible queue spell and the aura
// that signals ReplaceMHSwing to fire Maul on the next auto-attack, mirroring
// the warrior Heroic Strike queue pattern.
func (druid *Druid) makeMaulQueueSpellAndAura(maulSpell *DruidSpell) *DruidSpell {
	isMaulQueued := false

	druid.maulQueueAura = druid.RegisterAura(core.Aura{
		Label:    "Maul Queue Aura",
		ActionID: maulSpell.ActionID.WithTag(1),
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			isMaulQueued = false
		},
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			druid.maulQueueSpell = maulSpell.Spell
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			druid.maulQueueSpell = nil
		},
	})

	druid.maulRealismICD = &core.Cooldown{
		Timer:    druid.NewTimer(),
		Duration: time.Millisecond * 50,
	}

	queueSpell := druid.RegisterSpell(Bear, core.SpellConfig{
		ActionID:    maulSpell.ActionID.WithTag(1),
		SpellSchool: core.SpellSchoolPhysical,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagAPL,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				NonEmpty: true,
			},
		},

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return druid.maulQueueAura != nil &&
				!druid.maulQueueAura.IsActive() &&
				!isMaulQueued &&
				druid.CurrentRage() >= maulSpell.Cost.GetCurrentCost() &&
				druid.maulRealismICD.IsReady(sim)
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			if druid.maulRealismICD.IsReady(sim) {
				isMaulQueued = true
				druid.maulRealismICD.Use(sim)
				sim.AddPendingAction(&core.PendingAction{
					NextActionAt: sim.CurrentTime + druid.maulRealismICD.Duration,
					OnAction: func(sim *core.Simulation) {
						druid.maulQueueAura.Activate(sim)
						isMaulQueued = false
					},
				})
			}
		},
	})

	return queueSpell
}

// TryMaul returns the Maul spell if the queue aura is active and Maul can
// fire, otherwise returns the normal auto-attack swing spell.
func (druid *Druid) TryMaul(sim *core.Simulation, mhSwingSpell *core.Spell) *core.Spell {
	if !druid.maulQueueAura.IsActive() {
		return mhSwingSpell
	}
	if !druid.maulQueueSpell.CanCast(sim, druid.CurrentTarget) {
		druid.maulQueueAura.Deactivate(sim)
		return mhSwingSpell
	}
	return druid.maulQueueSpell
}
