package guardian

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/druid"
)

func (bear *GuardianDruid) registerBoneShieldSpell() {
	actionID := core.ActionID{SpellID: 122285}

	boneShieldAura := bear.RegisterAura(core.Aura{
		Label:     "Bone Shield",
		ActionID:  actionID,
		Duration:  time.Minute * 5,
		MaxStacks: 3,

		Icd: &core.Cooldown{
			Timer:    bear.NewTimer(),
			Duration: time.Second * 2,
		},

		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.PseudoStats.DamageTakenMultiplier *= 0.9
			aura.SetStacks(sim, 3)
			aura.Icd.Use(sim)
		},

		OnSpellHitTaken: func(aura *core.Aura, sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
			if (result.Damage > 0) && aura.Icd.IsReady(sim) {
				aura.RemoveStack(sim)
				aura.Icd.Use(sim)
			}
		},

		OnExpire: func(aura *core.Aura, _ *core.Simulation) {
			aura.Unit.PseudoStats.DamageTakenMultiplier /= 0.9
		},
	})

	boneShieldSpell := bear.RegisterSpell(druid.Any, core.SpellConfig{
		ActionID:        actionID,
		SpellSchool:     core.SpellSchoolShadow,
		ProcMask:        core.ProcMaskEmpty,
		Flags:           core.SpellFlagAPL,
		RelatedSelfBuff: boneShieldAura,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},

			CD: core.Cooldown{
				Timer:    bear.NewTimer(),
				Duration: time.Minute,
			},

			IgnoreHaste: true,
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			spell.RelatedSelfBuff.Activate(sim)
		},
	})

	bear.AddMajorCooldown(core.MajorCooldown{
		Spell: boneShieldSpell.Spell,
		Type:  core.CooldownTypeSurvival,
	})
}
