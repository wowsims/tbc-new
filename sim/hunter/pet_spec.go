package hunter

import (
	"github.com/wowsims/mop/sim/core"
	"github.com/wowsims/mop/sim/core/proto"
)

func (hp *HunterPet) ApplySpikedCollar() {
	if hp.hunterOwner.Options.PetSpec != proto.PetSpec_Ferocity {
		return
	}

	core.MakePermanent(hp.RegisterAura(core.Aura{
		Label:    "Spiked Collar",
		ActionID: core.ActionID{SpellID: 53184},
	})).AttachSpellMod(core.SpellModConfig{
		Kind:       core.SpellMod_DamageDone_Pct,
		ClassMask:  HunterPetFocusDump,
		FloatValue: 0.1,
	}).AttachSpellMod(core.SpellModConfig{
		Kind:       core.SpellMod_BonusCrit_Percent,
		FloatValue: 10,
	}).AttachMultiplyMeleeSpeed(1.1)
}

func (hp *HunterPet) ApplyCombatExperience() {
	core.MakePermanent(hp.RegisterAura(core.Aura{
		Label:    "Combat Experience",
		ActionID: core.ActionID{SpellID: 20782},
	})).AttachMultiplicativePseudoStatBuff(
		&hp.PseudoStats.DamageDealtMultiplier, 1.5,
	)
}

func (hp *HunterPet) ApplyBoarsSpeed() {
	if !hp.isPrimary {
		return
	}

	hp.BoarsSpeedAura = core.MakePermanent(hp.RegisterAura(core.Aura{
		Label:    "Boar's Speed",
		ActionID: core.ActionID{SpellID: 19596},
	}))
	hp.BoarsSpeedAura.NewActiveMovementSpeedEffect(0.3)
}
