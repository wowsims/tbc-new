package warrior

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (war *Warrior) registerRetaliation() {
	actionID := core.ActionID{SpellID: 20230}

	// The hits will proc in any stance
	attackSpell := war.RegisterSpell(core.SpellConfig{
		ClassSpellMask: SpellMaskRetaliation,
		ActionID:       core.ActionID{SpellID: 20240},
		SpellSchool:    core.SpellSchoolPhysical,
		ProcMask:       core.ProcMaskMeleeMH,
		Flags:          core.SpellFlagMeleeMetrics,

		DamageMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := war.MHWeaponDamage(sim, spell.MeleeAttackPower())
			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
		},
	})

	buffAura := war.RegisterAura(core.Aura{
		ActionID:  actionID,
		Label:     "Retaliation",
		Duration:  time.Second * 15,
		MaxStacks: 30,
		OnSpellHitTaken: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if spell.ProcMask.Matches(core.ProcMaskMelee) && result.Landed() && result.Damage > 0 {
				attackSpell.Cast(sim, spell.Unit)
				aura.RemoveStack(sim)
			}
		},
	})

	spell := war.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		ClassSpellMask: SpellMaskRetaliation,
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    war.NewTimer(),
				Duration: time.Minute * 30,
			},
			SharedCD: core.Cooldown{
				Timer:    war.sharedMCD,
				Duration: time.Minute * 30,
			},
		},

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return war.StanceMatches(BattleStance)
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			buffAura.Activate(sim)
			buffAura.SetStacks(sim, 30)
		},

		RelatedSelfBuff: buffAura,
	})

	war.AddMajorCooldown(core.MajorCooldown{
		Spell: spell,
		Type:  core.CooldownTypeDPS,
		// Require manual CD usage
		ShouldActivate: func(sim *core.Simulation, character *core.Character) bool {
			return false
		},
	})
}
