package beast_mastery

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/hunter"
)

func (bmHunter *BeastMasteryHunter) registerBestialWrathCD() {
	if bmHunter.Pet == nil {
		return
	}

	duration := core.TernaryDuration(bmHunter.CouldHaveSetBonus(hunter.YaungolSlayersBattlegear, 4), 16, 10) * time.Second

	actionID := core.ActionID{SpellID: 19574}

	bmHunter.Pet.BestialWrathAura = bmHunter.Pet.RegisterAura(core.Aura{
		Label:    "Bestial Wrath Pet",
		ActionID: actionID,
		Duration: duration,
	}).AttachMultiplicativePseudoStatBuff(
		&bmHunter.Pet.PseudoStats.DamageDealtMultiplier, 1.2,
	)

	bmHunter.BestialWrathAura = bmHunter.RegisterAura(core.Aura{
		Label:    "Bestial Wrath",
		ActionID: actionID,
		Duration: duration,
	}).AttachSpellMod(core.SpellModConfig{
		Kind:       core.SpellMod_PowerCost_Pct,
		ClassMask:  hunter.HunterSpellsAll | hunter.HunterSpellsTalents,
		FloatValue: -0.5,
	}).AttachMultiplicativePseudoStatBuff(
		&bmHunter.PseudoStats.DamageDealtMultiplier, 1.1,
	).AttachDependentAura(
		bmHunter.Pet.BestialWrathAura,
	)
	core.RegisterPercentDamageModifierEffect(bmHunter.BestialWrathAura, 1.1)

	bestialWrath := bmHunter.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		ClassSpellMask: hunter.HunterSpellBestialWrath,
		Flags:          core.SpellFlagReadinessTrinket,
		FocusCost: core.FocusCostOptions{
			Cost: 0,
		},
		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    bmHunter.NewTimer(),
				Duration: time.Minute * 1,
			},
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			bmHunter.BestialWrathAura.Activate(sim)
		},
	})

	bmHunter.AddMajorCooldown(core.MajorCooldown{
		Spell: bestialWrath,
		Type:  core.CooldownTypeDPS,
	})
}
