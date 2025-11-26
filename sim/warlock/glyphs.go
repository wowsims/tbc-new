package warlock

import (
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

func (warlock *Warlock) registerGlyphs() {

	if warlock.HasMajorGlyph(proto.WarlockMajorGlyph_GlyphOfSiphonLife) {
		warlock.SiphonLife = warlock.RegisterSpell(core.SpellConfig{
			ActionID:       core.ActionID{SpellID: 63106},
			SpellSchool:    core.SpellSchoolShadow,
			ProcMask:       core.ProcMaskSpellHealing,
			Flags:          core.SpellFlagHelpful | core.SpellFlagPassiveSpell,
			ClassSpellMask: WarlockSpellSiphonLife,

			DamageMultiplier: 1,
			CritMultiplier:   warlock.DefaultCritMultiplier(),
			ThreatMultiplier: 1,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.CalcAndDealPeriodicHealing(sim, target, warlock.MaxHealth()*0.005, spell.OutcomeHealing)
			},
		})
	}

	if warlock.HasMajorGlyph(proto.WarlockMajorGlyph_GlyphOfEternalResolve) {
		warlock.AddStaticMod(core.SpellModConfig{
			ClassMask:  WarlockSpellAgony | WarlockSpellCorruption | WarlockSpellUnstableAffliction | WarlockSpellDoom,
			Kind:       core.SpellMod_DotBaseDuration_Pct,
			FloatValue: 0.5,
		})

		warlock.AddStaticMod(core.SpellModConfig{
			ClassMask:  WarlockSpellAgony | WarlockSpellCorruption | WarlockSpellUnstableAffliction | WarlockSpellDoom,
			Kind:       core.SpellMod_DotDamageDone_Pct,
			FloatValue: -0.2,
		})
	}
}
