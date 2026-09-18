package hunter

import (
	"math"

	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

var aspectOfTheHawkRank = genRanks.AspectOfTheHawk.BySpellID(27044)

func (hunter *Hunter) registerAspectOfTheHawkSpell() {
	actionID := core.ActionID{SpellID: 27044}

	hunter.AspectOfTheHawkAura = hunter.applySharedAspectConfig(hunter.RegisterAura(core.Aura{
		Label:      "Aspect of the Hawk",
		ActionID:   actionID,
		BuildPhase: core.CharacterBuildPhaseBase,
	}).AttachStatBuff(stats.RangedAttackPower, shared.SpellRankMin(aspectOfTheHawkRank.Direct)))

	hunter.AspectOfTheHawk = hunter.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolNature,
		DefenseType:    core.DefenseTypeMagic,
		ClassSpellMask: HunterSpellAspectOfTheHawk,
		Flags:          core.SpellFlagAPL,

		ManaCost: core.ManaCostOptions{
			FlatCost: aspectOfTheHawkRank.Cost,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: aspectOfTheHawkRank.GCD,
			},
			IgnoreHaste: true,
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			spell.RelatedSelfBuff.Activate(sim)
		},

		RelatedSelfBuff: hunter.AspectOfTheHawkAura,
	})
}

func (hunter *Hunter) registerAspectOfTheViper() {
	actionID := core.ActionID{SpellID: 34074}

	hunter.AspectOfTheViperAura = hunter.applySharedAspectConfig(hunter.RegisterAura(core.Aura{
		Label:    "Aspect of the Viper",
		ActionID: actionID,
	}))

	hunter.AspectOfTheViper = hunter.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolNature,
		DefenseType:    core.DefenseTypeMagic,
		ClassSpellMask: HunterSpellAspectOfTheViper,
		Flags:          core.SpellFlagAPL,

		ManaCost: core.ManaCostOptions{
			FlatCost: 40,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				// Aspect of the Viper has no generated row - it carries no "Rank N" subtext, so the
				// ladder discovery never sees it. See the not-generated list at the head of
				// spell_ranks_auto_gen.go.
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true,
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			spell.RelatedSelfBuff.Activate(sim)
		},

		RelatedSelfBuff: hunter.AspectOfTheViperAura,
	})
}

func (hunter *Hunter) registerAspects() {
	hunter.registerAspectOfTheHawkSpell()
	hunter.registerAspectOfTheViper()
}

func (hunter *Hunter) applySharedAspectConfig(aura *core.Aura) *core.Aura {
	aura.Duration = core.NeverExpires
	aura.NewExclusiveEffect("Aspect", true, core.ExclusiveEffect{})
	return aura
}

func (hunter *Hunter) OnManaTick(sim *core.Simulation) {
	// https://wowpedia.fandom.com/wiki/Aspect_of_the_Viper?oldid=1458832
	if hunter.AspectOfTheViperAura.IsActive() {
		currentMana := hunter.CurrentManaPercent()
		if currentMana >= 100 {
			return
		}

		percentMana := math.Max(0.2, math.Min(0.9, currentMana))
		scaling := 22.0/35.0*(0.9-percentMana) + 0.11
		if hunter.GronnStalker2PcAura.IsActive() {
			scaling += 0.05
		}

		bonusPer5Seconds := hunter.GetStat(stats.Intellect)*scaling + 0.35*70
		manaGain := bonusPer5Seconds * 2 / 5
		hunter.AddMana(sim, manaGain, hunter.AspectOfTheViper.Cost.ResourceCostImpl.(*core.ManaCost).ResourceMetrics)
	}
}
