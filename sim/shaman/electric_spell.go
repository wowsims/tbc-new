package shaman

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

const (
	// This could be value or bitflag if we ended up needing multiple flags at the same time.
	//1 to 5 are used by MaelstromWeapon Stacks
	CastTagLightningOverload int32 = 6
)

type ShamSpellConfig struct {
	ActionID            core.ActionID
	Rank                int32
	BaseFlatCost        int32
	BaseCastTime        time.Duration
	IsElementalOverload bool
	BonusCoefficient    float64
	BounceReduction     float64
	SpellSchool         core.SpellSchool
	Overloads           map[int32][]*core.Spell
	ClassSpellMask      int64
}

// Shared precomputation logic for LB and CL.
// Needs isElementalOverload, actionID, BaseFlatCost, baseCastTime, bonusCoefficient fields of the shamSpellConfig
func (shaman *Shaman) newElectricSpellConfig(config ShamSpellConfig) core.SpellConfig {
	flags := SpellFlagShamanSpell | SpellFlagFocusable
	if config.IsElementalOverload {
		flags |= core.SpellFlagPassiveSpell | core.SpellFlagProc
	} else {
		flags |= core.SpellFlagAPL
	}

	spell := core.SpellConfig{
		ActionID:       config.ActionID,
		SpellSchool:    core.SpellSchoolNature,
		DefenseType:    core.DefenseTypeMagic,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          flags,
		ClassSpellMask: config.ClassSpellMask,
		Rank:           config.Rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: core.TernaryInt32(config.IsElementalOverload, 0, config.BaseFlatCost),
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				CastTime: config.BaseCastTime,
				GCD:      core.GCDDefault,
			},
			ModifyCast: func(sim *core.Simulation, spell *core.Spell, cast *core.Cast) {
				castTime := shaman.ApplyCastSpeedForSpell(cast.CastTime, spell)
				if sim.CurrentTime+castTime > shaman.AutoAttacks.NextAttackAt() {
					shaman.AutoAttacks.StopMeleeUntil(sim, sim.CurrentTime+castTime)
				}
			},
		},

		DamageMultiplier: 1,
		BonusCoefficient: config.BonusCoefficient,
		ThreatMultiplier: 1,
	}

	if config.IsElementalOverload {
		spell.ActionID.Tag = CastTagLightningOverload
		spell.ManaCost.FlatCost = 0
		spell.Cast.DefaultCast.CastTime = 0
		spell.Cast.DefaultCast.GCD = 0
		spell.Cast.DefaultCast.Cost = 0
		spell.Cast.ModifyCast = nil
		spell.MetricSplits = 0
		spell.DamageMultiplier *= 0.5
		spell.ThreatMultiplier = 0
	}

	return spell
}
