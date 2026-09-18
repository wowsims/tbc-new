package paladin

import (
	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

func (paladin *Paladin) getAvengersShieldTimer() *core.Timer {
	if paladin.avengersShieldTimer == nil {
		paladin.avengersShieldTimer = paladin.NewTimer()
	}
	return paladin.avengersShieldTimer
}

var AvengersShieldRankMap = genRanks.AvengersShield

// Avenger's Shield (Talent)
// https://www.wowhead.com/tbc/spell=31935
//
// Hurls a holy shield at the enemy, dealing Holy damage, dazing them and
// then jumping to additional nearby enemies. Affects 3 total targets.
func (paladin *Paladin) registerAvengersShield(rankConfig shared.SpellRank) {
	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: rankConfig.SpellID},
		SpellSchool:    core.SpellSchoolHoly,
		DefenseType:    core.DefenseTypeRanged,
		ProcMask:       core.ProcMaskRangedSpecial,
		Flags:          core.SpellFlagAPL | core.SpellFlagBinary,
		ClassSpellMask: SpellMaskAvengersShield,
		Rank:           rankConfig.Rank,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		MaxRange:     rankConfig.MaxRange,
		MissileSpeed: 35,

		ManaCost: core.ManaCostOptions{
			FlatCost: rankConfig.Cost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      rankConfig.GCD,
				CastTime: rankConfig.CastTime,
			},
			CD: core.Cooldown{
				Timer:    paladin.getAvengersShieldTimer(),
				Duration: rankConfig.Cooldown,
			},
			ModifyCast: func(sim *core.Simulation, spell *core.Spell, cast *core.Cast) {
				castTime := paladin.ApplyCastSpeedForSpell(cast.CastTime, spell)
				paladin.AutoAttacks.StopMeleeUntil(sim, sim.CurrentTime+castTime)
			},
		},

		BonusCoefficient: rankConfig.Direct.BonusCoefficient(),

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			damage := rankConfig.Direct.Damage(sim)
			results := spell.CalcCleaveDamage(sim, target, 3, damage, spell.OutcomeRangedHitAndCrit)
			spell.WaitTravelTime(sim, func(sim *core.Simulation) {
				for _, result := range results {
					spell.DealDamage(sim, result)
				}
			})
		},
	})
}
