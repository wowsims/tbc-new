package warrior

import (
	"github.com/wowsims/tbc/sim/core"
)

var thunderClapRank = genRanks.ThunderClap.BySpellID(25264)

func (war *Warrior) registerThunderClap() {
	auras := war.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return core.ThunderClapAura(target, war.Talents.ImprovedThunderClap)
	})

	war.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: thunderClapRank.SpellID},
		SpellSchool: core.SpellSchoolPhysical,
		// Thunder Clap is Physical but Magic in SpellCategories: it rolls on the spell hit table
		// (logs show full resists next to armor mitigation) and crits on spell crit chance for
		// 1.5x. Warriors have no base spell crit, so logs without Totem of Wrath show none
		// (0 of 799 landed hits from 6 prot warriors on fresh.warcraftlogs.com, 2026-09-14).
		DefenseType:    core.DefenseTypeMagic,
		ProcMask:       core.ProcMaskRangedSpecial,
		Flags:          core.SpellFlagAPL | core.SpellFlagBinary,
		ClassSpellMask: SpellMaskThunderClap,

		RageCost: core.RageCostOptions{
			Cost: thunderClapRank.Cost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: thunderClapRank.GCD,
			},
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    war.NewTimer(),
				Duration: thunderClapRank.Cooldown,
			},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1.75,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage, _ := thunderClapRank.Direct.Range()
			results := spell.CalcCleaveDamage(sim, target, 4, baseDamage, spell.OutcomeMagicHitAndCrit)
			war.CastNormalizedSweepingStrikesAttack(results, sim)

			for _, result := range results {
				if result.Landed() {
					auras.Get(result.Target).Activate(sim)
				}
				spell.DealDamage(sim, result)
			}
		},

		RelatedAuraArrays: auras.ToMap(),
	})
}
