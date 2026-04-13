package paladin

import (
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

func (paladin *Paladin) registerAuras() {
	paladin.registerDevotionAura()
	paladin.registerRetributionAura()
	paladin.registerConcentrationAura()
	paladin.registerFireResistanceAura()
	paladin.registerFrostResistanceAura()
	paladin.registerShadowResistanceAura()
}

// registerAuraSpell wires up a castable paladin aura spell that toggles the
// given self-cast aura (which must already be in the PaladinAuraCategory).
func (paladin *Paladin) registerAuraSpell(aura *core.Aura, classSpellMask int64) {
	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       aura.ActionID,
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagAPL | core.SpellFlagHelpful,
		ClassSpellMask: classSpellMask,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			aura.Activate(sim)
		},
	})
}

// Devotion Aura
// https://www.wowhead.com/tbc/spell=27149
//
// Gives 861 additional armor to party members within 30 yards.
// Improved Devotion Aura talent increases the armor bonus by up to 40%.
// Players may only have one Aura on them per Paladin at any one time.
func (paladin *Paladin) registerDevotionAura() {
	aura := core.DevotionAuraBuff(&paladin.Character, true, paladin.Talents.ImprovedDevotionAura)
	paladin.registerAuraSpell(aura, SpellMaskDevotionAura)
}

// Retribution Aura
// https://www.wowhead.com/tbc/spell=27150
//
// Causes 26 Holy damage to any creature that strikes a party member within 30 yards.
// Improved Retribution Aura talent increases damage by up to 50%.
// Players may only have one Aura on them per Paladin at any one time.
func (paladin *Paladin) registerRetributionAura() {
	aura := core.RetributionAuraBuff(&paladin.Character, true, paladin.Talents.ImprovedRetributionAura)
	paladin.registerAuraSpell(aura, SpellMaskRetributionAura)
}

// registerSelfCastAura creates a self-cast paladin aura with the standard
// PaladinAuraCategory exclusivity and returns the aura for further configuration.
func (paladin *Paladin) registerSelfCastAura(label string, actionID core.ActionID) *core.Aura {
	aura := paladin.RegisterAura(core.Aura{
		Label:    label + " (Player)" + paladin.Label,
		ActionID: actionID,
		Duration: core.NeverExpires,
	})
	aura.NewExclusiveEffect(core.PaladinAuraCategory, true, core.ExclusiveEffect{})
	return aura
}

// Concentration Aura
// https://www.wowhead.com/tbc/spell=19746
//
// All party members within 30 yards lose 35% less casting or channeling time
// when damaged. Players may only have one Aura on them per Paladin at any one time.
func (paladin *Paladin) registerConcentrationAura() {
	aura := paladin.registerSelfCastAura("Concentration Aura", core.ActionID{SpellID: 19746})
	paladin.registerAuraSpell(aura, SpellMaskConcentrationAura)
}

// Fire Resistance Aura
// https://www.wowhead.com/tbc/spell=27153
func (paladin *Paladin) registerFireResistanceAura() {
	aura := paladin.registerSelfCastAura("Fire Resistance Aura", core.ActionID{SpellID: 27153}).
		AttachStatBuff(stats.FireResistance, 70)
	paladin.registerAuraSpell(aura, SpellMaskFireResistanceAura)
}

// Frost Resistance Aura
// https://www.wowhead.com/tbc/spell=27152
func (paladin *Paladin) registerFrostResistanceAura() {
	aura := paladin.registerSelfCastAura("Frost Resistance Aura", core.ActionID{SpellID: 27152}).
		AttachStatBuff(stats.FrostResistance, 70)
	paladin.registerAuraSpell(aura, SpellMaskFrostResistanceAura)
}

// Shadow Resistance Aura
// https://www.wowhead.com/tbc/spell=27151
func (paladin *Paladin) registerShadowResistanceAura() {
	aura := paladin.registerSelfCastAura("Shadow Resistance Aura", core.ActionID{SpellID: 27151}).
		AttachStatBuff(stats.ShadowResistance, 70)
	paladin.registerAuraSpell(aura, SpellMaskShadowResistanceAura)
}

// Sanctity Aura (Talent)
// https://www.wowhead.com/tbc/spell=20218
//
// Increases Holy damage done by party members within 30 yards by 10%.
// Players may only have one Aura on them per Paladin at any one time.
func (paladin *Paladin) registerSanctityAura() {
	aura := paladin.registerSelfCastAura("Sanctity Aura", core.ActionID{SpellID: 20218}).
		AttachMultiplicativePseudoStatBuff(
			&paladin.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexHoly], 1.1,
		)
	paladin.registerAuraSpell(aura, SpellMaskSanctityAura)
}
