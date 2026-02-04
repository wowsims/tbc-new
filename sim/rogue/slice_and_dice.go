package rogue

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (rogue *Rogue) registerSliceAndDice() {
	actionID := core.ActionID{SpellID: 6774}

	rogue.SliceAndDiceBonusFlat = 0.3
	rogue.sliceAndDiceDurations = [6]time.Duration{
		0,
		time.Duration(time.Second * 9),
		time.Duration(time.Second * 12),
		time.Duration(time.Second * 15),
		time.Duration(time.Second * 18),
		time.Duration(time.Second * 21),
	}

	getDuration := func(comboPoints int32) time.Duration {
		duration := rogue.sliceAndDiceDurations[comboPoints]
		if rogue.Talents.ImprovedSliceAndDice > 0 {
			duration *= time.Duration(1 + 0.15*float64(rogue.Talents.ImprovedSliceAndDice))
		}
		return duration
	}

	var slideAndDiceMod float64
	rogue.SliceAndDiceAura = rogue.RegisterAura(core.Aura{
		Label:    "Slice and Dice",
		ActionID: actionID,
		// This will be overridden on cast, but set a non-zero default so it doesn't crash when used in APL prepull
		Duration: rogue.sliceAndDiceDurations[5],
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			slideAndDiceMod = 1 + rogue.SliceAndDiceBonusFlat
			rogue.MultiplyMeleeSpeed(sim, slideAndDiceMod)
			if sim.Log != nil {
				rogue.Log(sim, "[DEBUG]: Slice and Dice attack speed mod: %v", slideAndDiceMod)
			}
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			rogue.MultiplyMeleeSpeed(sim, 1/slideAndDiceMod)
		},
	})

	rogue.SliceAndDice = rogue.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		Flags:          SpellFlagFinisher | core.SpellFlagAPL,
		MetricSplits:   6,
		ClassSpellMask: RogueSpellSliceAndDice,

		EnergyCost: core.EnergyCostOptions{
			Cost: 25,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: time.Second,
			},
			IgnoreHaste: true,
			ModifyCast: func(sim *core.Simulation, spell *core.Spell, cast *core.Cast) {
				spell.SetMetricsSplit(rogue.ComboPoints())
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return rogue.ComboPoints() > 0
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			comboPoints := rogue.ComboPoints()
			rogue.ApplyFinisher(sim, spell)
			spell.RelatedSelfBuff.Deactivate(sim)
			spell.RelatedSelfBuff.Duration = getDuration(comboPoints)
			spell.RelatedSelfBuff.Activate(sim)
		},

		RelatedSelfBuff: rogue.SliceAndDiceAura,
	})
}
