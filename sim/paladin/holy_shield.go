package paladin

import (
	"strconv"
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

// Holy Shield (Talent)
// https://www.wowhead.com/tbc/spell=20925
//
// Increases chance to block by 30% for 10 sec, and deals Holy damage
// for each attack blocked while active. Damage caused by Holy Shield causes
// 35% additional threat. Each block expends a charge. 4 charges.
func (paladin *Paladin) registerHolyShield() {
	var ranks = []struct {
		level    int32
		spellID  int32
		manaCost int32
		value    float64
		coeff    float64
	}{
		{},
		{level: 40, spellID: 20925, manaCost: 135, value: 59, coeff: 0.05},
		{level: 50, spellID: 20927, manaCost: 175, value: 86, coeff: 0.05},
		{level: 60, spellID: 20928, manaCost: 215, value: 117, coeff: 0.05},
		{level: 70, spellID: 27179, manaCost: 280, value: 155, coeff: 0.05},
	}

	cd := core.Cooldown{
		Timer:    paladin.NewTimer(),
		Duration: time.Second * 10,
	}

	blockRating := 30 * core.BlockRatingPerBlockPercent // 30% block chance

	for rank := 1; rank < len(ranks); rank++ {
		if paladin.Level < ranks[rank].level {
			break
		}

		actionID := core.ActionID{SpellID: ranks[rank].spellID}
		manaCost := ranks[rank].manaCost
		value := ranks[rank].value
		coeff := ranks[rank].coeff

		procSpell := paladin.RegisterSpell(core.SpellConfig{
			ActionID:    actionID,
			SpellSchool: core.SpellSchoolHoly,
			ProcMask:    core.ProcMaskEmpty,
			Flags:       core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell,

			BonusCoefficient: coeff,
			DamageMultiplier: 1,
			ThreatMultiplier: 1.35,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.CalcAndDealDamage(sim, target, value, spell.OutcomeAlwaysHit)
			},
		})

		holyShieldAura := paladin.RegisterAura(core.Aura{
			Label:     "Holy Shield" + paladin.Label + "Rank" + strconv.Itoa(rank),
			ActionID:  actionID,
			Duration:  time.Second * 10,
			MaxStacks: 4,

			OnGain: func(aura *core.Aura, sim *core.Simulation) {
				aura.Unit.AddStatDynamic(sim, stats.BlockRating, blockRating)
			},
			OnExpire: func(aura *core.Aura, sim *core.Simulation) {
				aura.Unit.AddStatDynamic(sim, stats.BlockRating, -blockRating)
			},

			OnSpellHitTaken: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				if result.Outcome.Matches(core.OutcomeBlock) {
					procSpell.Cast(sim, spell.Unit)
					aura.RemoveStack(sim)
				}
			},
		})

		holyShieldSpell := paladin.RegisterSpell(core.SpellConfig{
			ActionID:       actionID,
			SpellSchool:    core.SpellSchoolHoly,
			ProcMask:       core.ProcMaskEmpty,
			Flags:          core.SpellFlagAPL,
			ClassSpellMask: SpellMaskHolyShield,

			DamageMultiplier: 1,
			ThreatMultiplier: 1,

			ManaCost: core.ManaCostOptions{
				FlatCost: manaCost,
			},
			Cast: core.CastConfig{
				DefaultCast: core.Cast{
					GCD: core.GCDDefault,
				},
				CD: cd,
			},

			ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
				holyShieldAura.SetStacks(sim, 4)
				holyShieldAura.Activate(sim)
			},
		})

		paladin.HolyShieldAuras = append(paladin.HolyShieldAuras, holyShieldAura)
		paladin.HolyShields = append(paladin.HolyShields, holyShieldSpell)
	}
}
