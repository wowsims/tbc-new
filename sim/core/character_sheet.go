package core

import (
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

var ownExposeWeaknessByClass = map[proto.Class]func(talentsString string) bool{}

// RegisterCharacterSheetExposeWeakness lets a class report whether a talent string applies Expose
// Weakness itself, in which case the sheet credits the character's own agility.
func RegisterCharacterSheetExposeWeakness(class proto.Class, appliesOwn func(talentsString string) bool) {
	ownExposeWeaknessByClass[class] = appliesOwn
}

// CharacterSheetExposeWeaknessAgility returns the agility the stats panel credits Expose
// Weakness with: the character's own for one who applies it with their own talent (see
// RegisterCharacterSheetExposeWeakness), else the agility configured for the debuff.
// characterAgility is the character's final agility, or 0 when it is not known, which falls back
// to the configured value as the panel does. Mirrors characterSheetExposeWeaknessAgility in
// ui/sim/player/debuff_stats.ts.
func CharacterSheetExposeWeaknessAgility(debuffs *proto.Debuffs, player *proto.Player, characterAgility float64) float64 {
	appliesOwn := ownExposeWeaknessByClass[player.GetClass()]
	if appliesOwn == nil || characterAgility <= 0 || !appliesOwn(player.GetTalentsString()) {
		return debuffs.GetExposeWeaknessHunterAgility()
	}
	return characterAgility
}

// CharacterSheetDebuffStats returns what the stats panel adds to a character's final stats for the
// raid's debuffs. They lower the target's defences rather than raising the character's stats, so
// FinalStats leaves them out, but the panel shows them as the character's own, and batch stat
// constraints and the gem optimizer's caps are judged on the panel's values. exposeWeaknessAgility
// is what CharacterSheetExposeWeaknessAgility returns. Mirrors characterSheetDebuffStats in
// ui/sim/player/debuff_stats.ts.
func CharacterSheetDebuffStats(debuffs *proto.Debuffs, exposeWeaknessAgility float64) UnitStats {
	result := NewUnitStats()
	if debuffs.GetFaerieFire() == proto.TristateEffect_TristateEffectImproved {
		result.AddStat(stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatMeleeHitPercent), 3)
		result.AddStat(stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatRangedHitPercent), 3)
	}
	if debuffs.GetImprovedSealOfTheCrusader() != proto.TristateEffect_TristateEffectMissing {
		result.AddStat(stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatMeleeCritPercent), 3)
		result.AddStat(stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatRangedCritPercent), 3)
		result.AddStat(stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatSpellCritPercent), 3)
	}
	if debuffs.GetExposeWeaknessUptime() != 0 && debuffs.GetExposeWeaknessHunterAgility() != 0 {
		attackPower := exposeWeaknessAgility * 0.25
		result.AddStat(stats.UnitStatFromStat(stats.AttackPower), attackPower)
		result.AddStat(stats.UnitStatFromStat(stats.RangedAttackPower), attackPower)
	}
	if debuffs.GetHuntersMark() != proto.TristateEffect_TristateEffectMissing {
		result.AddStat(stats.UnitStatFromStat(stats.RangedAttackPower), 440)
		if debuffs.GetHuntersMark() == proto.TristateEffect_TristateEffectImproved {
			result.AddStat(stats.UnitStatFromStat(stats.AttackPower), 110)
		}
	}
	return result
}

// WithCharacterSheetDebuffs returns the player's final stats as the stats panel shows them: with
// the raid's debuffs added (see CharacterSheetDebuffStats).
func WithCharacterSheetDebuffs(finalStats *proto.UnitStats, debuffs *proto.Debuffs, player *proto.Player) *proto.UnitStats {
	characterAgility := 0.0
	if int(stats.Agility) < len(finalStats.GetStats()) {
		characterAgility = finalStats.GetStats()[stats.Agility]
	}
	sheet := CharacterSheetDebuffStats(debuffs, CharacterSheetExposeWeaknessAgility(debuffs, player, characterAgility))
	result := &proto.UnitStats{
		Stats:       make([]float64, max(len(finalStats.GetStats()), len(sheet.Stats))),
		PseudoStats: make([]float64, max(len(finalStats.GetPseudoStats()), len(sheet.PseudoStats))),
	}
	copy(result.Stats, finalStats.GetStats())
	copy(result.PseudoStats, finalStats.GetPseudoStats())
	for i, value := range sheet.Stats {
		result.Stats[i] += value
	}
	for i, value := range sheet.PseudoStats {
		result.PseudoStats[i] += value
	}
	return result
}

func (character *Character) GetPseudoStatsProto() []float64 {
	return []float64{
		proto.PseudoStat_PseudoStatMainHandDps: character.AutoAttacks.MH().DPS(),
		proto.PseudoStat_PseudoStatOffHandDps:  character.AutoAttacks.OH().DPS(),
		proto.PseudoStat_PseudoStatRangedDps:   character.AutoAttacks.Ranged().DPS(),

		// Base values are modified by Enemy attackTables, but we display for LVL 70 enemy as paperdoll default
		proto.PseudoStat_PseudoStatDodgePercent:            (character.PseudoStats.BaseDodgeChance + character.GetDodgeFromRating() + character.GetDefenseReduction()) * 100,
		proto.PseudoStat_PseudoStatParryPercent:            Ternary(character.PseudoStats.CanParry, (character.PseudoStats.BaseParryChance+character.GetParryFromRating()+character.GetDefenseReduction())*100, 0),
		proto.PseudoStat_PseudoStatBlockPercent:            Ternary(character.PseudoStats.CanBlock, (character.PseudoStats.BaseBlockChance+character.GetBlockFromRating()+character.GetDefenseReduction())*100, 0),
		proto.PseudoStat_PseudoStatBlockValueMultiplier:    character.PseudoStats.BlockValueMultiplier,
		proto.PseudoStat_PseudoStatReducedCritTakenPercent: character.PseudoStats.ReducedCritTakenPercent * 100,

		// Used by UI to incorporate multiplicative Haste buffs into final character stats display.
		proto.PseudoStat_PseudoStatRangedSpeedMultiplier: character.PseudoStats.RangedSpeedMultiplier * character.PseudoStats.AttackSpeedMultiplier,
		proto.PseudoStat_PseudoStatMeleeSpeedMultiplier:  character.PseudoStats.MeleeSpeedMultiplier * character.PseudoStats.AttackSpeedMultiplier,
		proto.PseudoStat_PseudoStatCastSpeedMultiplier:   character.PseudoStats.CastSpeedMultiplier,
		proto.PseudoStat_PseudoStatMeleeHastePercent:     (character.TotalMeleeHasteMultiplier() - 1) * 100,
		proto.PseudoStat_PseudoStatRangedHastePercent:    (character.TotalRangedHasteMultiplier() - 1) * 100,
		proto.PseudoStat_PseudoStatSpellHastePercent:     (character.TotalSpellHasteMultiplier() - 1) * 100,

		// School-specific fully buffed Hit/Crit are represented as proper Stats in the back-end so
		// that stat dependencies will work correctly, but are stored as PseudoStats in proto
		// messages. This is done so that the stats arrays embedded in database files and saved
		// Encounter settings can omit these extraneous fields.
		proto.PseudoStat_PseudoStatMeleeHitPercent:        character.GetStat(stats.PhysicalHitPercent),
		proto.PseudoStat_PseudoStatSpellHitPercent:        character.GetStat(stats.SpellHitPercent),
		proto.PseudoStat_PseudoStatSchoolHitPercentArcane: character.GetStat(stats.SpellHitPercent) + character.PseudoStats.SchoolBonusHitChance[stats.SchoolIndexArcane],
		proto.PseudoStat_PseudoStatSchoolHitPercentFire:   character.GetStat(stats.SpellHitPercent) + character.PseudoStats.SchoolBonusHitChance[stats.SchoolIndexFire],
		proto.PseudoStat_PseudoStatSchoolHitPercentFrost:  character.GetStat(stats.SpellHitPercent) + character.PseudoStats.SchoolBonusHitChance[stats.SchoolIndexFrost],
		proto.PseudoStat_PseudoStatSchoolHitPercentHoly:   character.GetStat(stats.SpellHitPercent) + character.PseudoStats.SchoolBonusHitChance[stats.SchoolIndexHoly],
		proto.PseudoStat_PseudoStatSchoolHitPercentNature: character.GetStat(stats.SpellHitPercent) + character.PseudoStats.SchoolBonusHitChance[stats.SchoolIndexNature],
		proto.PseudoStat_PseudoStatSchoolHitPercentShadow: character.GetStat(stats.SpellHitPercent) + character.PseudoStats.SchoolBonusHitChance[stats.SchoolIndexShadow],
		proto.PseudoStat_PseudoStatRangedHitPercent:       character.GetStat(stats.RangedHitPercent) + character.GetStat(stats.PhysicalHitPercent),
		proto.PseudoStat_PseudoStatMeleeCritPercent:       character.GetStat(stats.PhysicalCritPercent),
		proto.PseudoStat_PseudoStatSpellCritPercent:       character.GetStat(stats.SpellCritPercent),
		proto.PseudoStat_PseudoStatRangedCritPercent:      character.GetStat(stats.RangedCritPercent) + character.GetStat(stats.PhysicalCritPercent),
	}
}
