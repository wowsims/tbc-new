package shaman

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (shaman *Shaman) newTotemSpellConfig(flatCost int32, spellID int32, spellMask int64) core.SpellConfig {
	return core.SpellConfig{
		ActionID:       core.ActionID{SpellID: spellID},
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: spellMask,

		ManaCost: core.ManaCostOptions{
			FlatCost: flatCost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: time.Second,
			},
		},
	}
}

func (shaman *Shaman) registerWindfuryTotemSpell() {
	config := shaman.newTotemSpellConfig(325, 25587, SpellMaskBasicTotem)
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		shaman.TotemExpirations[AirTotem] = sim.CurrentTime + time.Second*120
	}
	shaman.RegisterSpell(config)
}

func (shaman *Shaman) registerStrengthOfEarthTotemSpell() {
	config := shaman.newTotemSpellConfig(300, 25528, SpellMaskBasicTotem)
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		shaman.TotemExpirations[EarthTotem] = sim.CurrentTime + time.Second*120
	}
	shaman.RegisterSpell(config)
}

func (shaman *Shaman) registerGraceOfAirTotemSpell() {
	config := shaman.newTotemSpellConfig(310, 25359, SpellMaskBasicTotem)
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		shaman.TotemExpirations[AirTotem] = sim.CurrentTime + time.Second*120
	}
	shaman.RegisterSpell(config)
}
func (shaman *Shaman) registerHealingStreamTotemSpell() {
	config := shaman.newTotemSpellConfig(3, 5394, SpellMaskBasicTotem)
	hsHeal := shaman.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{SpellID: 5394},
		SpellSchool:      core.SpellSchoolNature,
		ProcMask:         core.ProcMaskEmpty,
		Flags:            core.SpellFlagHelpful | core.SpellFlagNoOnCastComplete | SpellFlagInstant,
		DamageMultiplier: 1,
		CritMultiplier:   1,
		ThreatMultiplier: 1,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			healing := 28 + spell.HealingPower(target)*0.08272
			spell.CalcAndDealHealing(sim, target, healing, spell.OutcomeHealing)
		},
	})
	config.Hot = core.DotConfig{
		Aura: core.Aura{
			Label: "HealingStreamHot",
		},
		NumberOfTicks: 150,
		TickLength:    time.Second * 2,
		OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
			hsHeal.Cast(sim, target)
		},
	}
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		shaman.TotemExpirations[WaterTotem] = sim.CurrentTime + time.Second*300
		for _, agent := range shaman.Party.Players {
			spell.Hot(&agent.GetCharacter().Unit).Activate(sim)
		}
	}
	shaman.HealingStreamTotem = shaman.RegisterSpell(config)
}
