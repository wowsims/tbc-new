package core

import (
	"math"

	"github.com/wowsims/tbc/sim/core/stats"
)

func (result *SpellResult) applyArmor(spell *Spell, isPeriodic bool, attackTable *AttackTable) {
	armorMitigationMultiplier := spell.armorMultiplier(isPeriodic, attackTable)

	result.Damage *= armorMitigationMultiplier

	result.ArmorAndResistanceMultiplier = armorMitigationMultiplier
	result.PostArmorAndResistanceMultiplier = result.Damage
}

// Returns Armor mitigation fraction for the spell
func (spell *Spell) armorMultiplier(isPeriodic bool, attackTable *AttackTable) float64 {
	if spell.Flags.Matches(SpellFlagIgnoreResists) {
		return 1
	}

	// There are no (partial) resists in MoP
	if !spell.SpellSchool.Matches(SpellSchoolPhysical) {
		return 1
	}

	// All physical dots (Bleeds) ignore armor.
	if isPeriodic && !spell.Flags.Matches(SpellFlagApplyArmorReduction) {
		return 1
	}

	// return armor mitigation fraction
	return attackTable.getArmorDamageModifier()
}

// https://web.archive.org/web/20130511200023/http://elitistjerks.com/f15/t29453-combat_ratings_level_85_cataclysm/p40/#post2171306
func (at *AttackTable) getArmorDamageModifier() float64 {
	if at.IgnoreArmor {
		return 1.0
	}

	ignoreArmorFactor := Clamp(at.ArmorIgnoreFactor, 0.0, 1.0)

	// Assume target > 80
	armorConstant := float64(at.Attacker.Level)*467.5 - 22167.5
	defenderArmor := at.Defender.Armor() - (at.Defender.Armor() * ignoreArmorFactor)
	// TBC ANNI: Apply flat ArP
	defenderArmor = max(defenderArmor-math.Abs(at.Attacker.stats[stats.ArmorPenetration]), 0)
	return 1 - defenderArmor/(defenderArmor+armorConstant)
}
