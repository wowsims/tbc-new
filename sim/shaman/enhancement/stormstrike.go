package enhancement

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
	"github.com/wowsims/tbc/sim/shaman"
)

var StormstrikeActionID = core.ActionID{SpellID: 17364}

func (enh *EnhancementShaman) StormstrikeDebuffAura(target *core.Unit) *core.Aura {
	aura := target.GetOrRegisterAura(core.Aura{
		Label:     "Stormstrike-" + enh.Label,
		ActionID:  StormstrikeActionID,
		Duration:  time.Second * 12,
		MaxStacks: 2,
		OnSpellHitTaken: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !spell.SpellSchool.Matches(core.SpellSchoolNature) {
				return
			}
			if !result.Landed() || result.Damage == 0 {
				return
			}
			aura.RemoveStack(sim)
		},
	})
	return aura.AttachMultiplicativePseudoStatBuff(
		&target.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexNature],
		1.2,
	)
}

func (enh *EnhancementShaman) newStormstrikeHitSpellConfig(spellID int32, isMH bool) core.SpellConfig {
	var procMask core.ProcMask
	var actionTag int32

	procMask = core.Ternary(isMH, core.ProcMaskMeleeMHSpecial, core.ProcMaskMeleeOHSpecial)
	actionTag = core.TernaryInt32(isMH, 1, 2)

	return core.SpellConfig{
		ActionID:         core.ActionID{SpellID: spellID}.WithTag(actionTag),
		SpellSchool:      core.SpellSchoolPhysical,
		ProcMask:         procMask,
		Flags:            core.SpellFlagMeleeMetrics,
		ClassSpellMask:   shaman.SpellMaskStormstrikeDamage,
		ThreatMultiplier: 1,
		DamageMultiplier: 1,
		CritMultiplier:   enh.DefaultMeleeCritMultiplier(),
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			weaponDamage := core.Ternary(isMH, spell.Unit.MHWeaponDamage, spell.Unit.OHWeaponDamage)
			baseDamage := weaponDamage(sim, spell.MeleeAttackPower())
			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialBlockAndCrit)
		},
	}
}

func (enh *EnhancementShaman) newStormstrikeHitSpell(isMH bool) *core.Spell {
	return enh.RegisterSpell(enh.newStormstrikeHitSpellConfig(17364, isMH))
}

func (enh *EnhancementShaman) newStormstrikeSpellConfig(spellID int32, ssDebuffAuras *core.AuraArray, mhHit *core.Spell, ohHit *core.Spell) core.SpellConfig {
	stormstrikeSpellConfig := core.SpellConfig{
		ActionID:       core.ActionID{SpellID: spellID},
		SpellSchool:    core.SpellSchoolPhysical,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagAPL,
		ClassSpellMask: shaman.SpellMaskStormstrikeCast,
		ManaCost: core.ManaCostOptions{
			BaseCostPercent: 8,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    enh.NewTimer(),
				Duration: time.Second * 10,
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			enh.StormstrikeCastResult = spell.CalcOutcome(sim, target, spell.OutcomeMeleeSpecialHitNoHitCounter)
			if enh.StormstrikeCastResult.Landed() {
				ssDebuffAura := ssDebuffAuras.Get(target)
				ssDebuffAura.Activate(sim)
				ssDebuffAura.SetStacks(sim, 2)

				if enh.HasMHWeapon() {
					mhHit.Cast(sim, target)
				}

				if enh.AutoAttacks.IsDualWielding && enh.HasOHWeapon() {
					ohHit.Cast(sim, target)
				}
			}
			spell.DisposeResult(enh.StormstrikeCastResult)
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return (enh.HasMHWeapon() || enh.HasOHWeapon())
		},
	}
	return stormstrikeSpellConfig
}

func (enh *EnhancementShaman) registerStormstrikeSpell() {
	mhHit := enh.newStormstrikeHitSpell(true)
	ohHit := enh.newStormstrikeHitSpell(false)

	enh.StormStrikeDebuffAuras = enh.NewEnemyAuraArray(enh.StormstrikeDebuffAura)

	enh.Stormstrike = enh.RegisterSpell(enh.newStormstrikeSpellConfig(17364, &enh.StormStrikeDebuffAuras, mhHit, ohHit))
}
