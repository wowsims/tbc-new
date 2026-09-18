package paladin

import (
	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

var ExorcismRankMap = genRanks.Exorcism

func (paladin *Paladin) getExorcismTimer() *core.Timer {
	if paladin.exorcismTimer == nil {
		paladin.exorcismTimer = paladin.NewTimer()
	}
	return paladin.exorcismTimer
}

// Exorcism
// https://www.wowhead.com/tbc/spell=10314
//
// Causes X to Y Holy damage to an Undead or Demon target.
func (paladin *Paladin) registerExorcism(rankConfig shared.SpellRank) {
	spellID := rankConfig.SpellID
	cost := rankConfig.Cost
	minDamage := shared.SpellRankMin(rankConfig.Direct)
	maxDamage := shared.SpellRankMax(rankConfig.Direct)
	coefficient := rankConfig.Direct.BonusCoefficient()

	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: spellID},
		SpellSchool:    core.SpellSchoolHoly,
		DefenseType:    core.DefenseTypeMagic,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		Rank:           rankConfig.Rank,
		ClassSpellMask: SpellMaskExorcism,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		MaxRange: rankConfig.MaxRange,

		ManaCost: core.ManaCostOptions{
			FlatCost: cost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    paladin.getExorcismTimer(),
				Duration: rankConfig.Cooldown,
			},
		},

		BonusCoefficient: coefficient,

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return target.MobType == proto.MobType_MobTypeUndead || target.MobType == proto.MobType_MobTypeDemon
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealDamage(sim, target, sim.Roll(minDamage, maxDamage), spell.OutcomeMagicHitAndCrit)
		},
	})
}
