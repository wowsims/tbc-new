package frost

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/mage"
)

func (frost *FrostMage) registerFrozenOrbSpell() {

	frozenOrbCoefficient := 0.51099997759
	frozenOrbScaling := 0.65200001001
	frozenOrbVariance := 0.25
	frozenOrbTicks := 0

	frozenOrbTickSpell := frost.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 84721},
		SpellSchool:    core.SpellSchoolFrost,
		ProcMask:       core.ProcMaskSpellDamage,
		ClassSpellMask: mage.MageSpellFrozenOrbTick,
		Flags:          core.SpellFlagAoE,

		DamageMultiplier: 1,
		CritMultiplier:   frost.DefaultCritMultiplier(),
		BonusCoefficient: frozenOrbCoefficient,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			results := spell.CalcAndDealAoeDamageWithVariance(sim, spell.OutcomeMagicHitAndCrit, func(sim *core.Simulation, spell *core.Spell) float64 {
				return frost.CalcAndRollDamageRange(sim, frozenOrbScaling, frozenOrbVariance)
			})

			if results.AnyLanded() && (frozenOrbTicks == 0 || sim.Proc(0.15, "FingersOfFrostProc")) {
				frost.FingersOfFrostAura.Activate(sim)
				frost.FingersOfFrostAura.AddStack(sim)
			}
			frozenOrbTicks++
		},
	})

	frozenOrb := frost.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 84714},
		SpellSchool:    core.SpellSchoolFrost,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: mage.MageSpellFrozenOrb,

		MissileSpeed: 8,

		ManaCost: core.ManaCostOptions{
			BaseCostPercent: 10,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    frost.NewTimer(),
				Duration: time.Minute,
			},
		},

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "Frozen Orb",
			},
			NumberOfTicks: 9,
			TickLength:    time.Second * 1,

			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				frozenOrbTickSpell.Cast(sim, target)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			frozenOrbTicks = 0

			result := spell.CalcOutcome(sim, target, spell.OutcomeAlwaysHit)
			spell.WaitTravelTime(sim, func(s *core.Simulation) {
				spell.DealOutcome(sim, result)
				dot := spell.Dot(target)
				dot.Apply(sim)
				dot.TickOnce(sim)
			})
		},
	})

	frost.AddMajorCooldown(core.MajorCooldown{
		Spell: frozenOrb,
		Type:  core.CooldownTypeDPS,
	})
}
