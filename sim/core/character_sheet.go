package core

import (
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

// HunterTalentTreeSizes is the hunter's talent tree layout, here because the character sheet
// reads one hunter talent; sim/hunter takes its copy from this.
var HunterTalentTreeSizes = [3]int{21, 20, 24}

// CharacterSheetExposeWeaknessAgility returns the agility the stats panel credits Expose
// Weakness with: the character's own for a hunter who applies it with their own talent, else the
// agility configured for the debuff. characterAgility is the character's final agility, or 0 when
// it is not known, which falls back to the configured value as the panel does. Mirrors
// characterSheetExposeWeaknessAgility in ui/sim/player/debuff_stats.ts.
func CharacterSheetExposeWeaknessAgility(debuffs *proto.Debuffs, player *proto.Player, characterAgility float64) float64 {
	configured := debuffs.GetExposeWeaknessHunterAgility()
	if player.GetClass() != proto.Class_ClassHunter || characterAgility <= 0 || !hunterTakesExposeWeakness(player.GetTalentsString()) {
		return configured
	}
	return characterAgility
}

func hunterTakesExposeWeakness(talentsString string) (taken bool) {
	// A malformed string is the sim's problem to report, not the sheet's: it just credits the
	// configured value.
	defer func() {
		if recover() != nil {
			taken = false
		}
	}()
	talents := &proto.HunterTalents{}
	FillTalentsProto(talents.ProtoReflect(), talentsString, HunterTalentTreeSizes)
	return talents.ExposeWeakness > 0
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
