import { Class, Debuffs, PseudoStat, Stat, TristateEffect } from '@generated/proto/common';
import { HunterTalents } from '@generated/proto/hunter';

import { Stats } from '../proto/stats';
import { talentStringToProto } from '../talents/factory';
import { hunterTalentsConfig } from '../talents/hunter';

/**
 * The agility the stats panel credits Expose Weakness with: the character's own for a hunter who
 * applies it with their own talent, else the agility configured for the debuff. `characterAgility`
 * is the character's final agility, or 0 when it is not known, which falls back to the configured
 * value. Mirrored in Go by core.CharacterSheetExposeWeaknessAgility.
 */
export const characterSheetExposeWeaknessAgility = (debuffs: Debuffs, playerClass: Class, talentsString: string, characterAgility: number): number => {
	const configured = debuffs.exposeWeaknessHunterAgility;
	if (playerClass != Class.ClassHunter || characterAgility <= 0 || !hunterTakesExposeWeakness(talentsString)) return configured;
	return characterAgility;
};

const hunterTakesExposeWeakness = (talentsString: string): boolean => {
	try {
		return talentStringToProto(HunterTalents.create(), talentsString, hunterTalentsConfig).exposeWeakness > 0;
	} catch {
		// A malformed string is the sim's problem to report, not the sheet's.
		return false;
	}
};

/**
 * What the stats panel adds to a character's final stats for the raid's debuffs. They lower the
 * target's defences rather than raising the character's stats, so final stats leave them out, but
 * the panel shows them as the character's own, and batch stat constraints are judged on the
 * panel's values. Expose Weakness is credited `exposeWeaknessAgility`, which is what
 * characterSheetExposeWeaknessAgility returns. Mirrored in Go by core.CharacterSheetDebuffStats
 * (sim/core/character_sheet.go).
 */
export const characterSheetDebuffStats = (debuffs: Debuffs, exposeWeaknessAgility = debuffs.exposeWeaknessHunterAgility): Stats => {
	let debuffStats = new Stats();

	if (debuffs.faerieFire == TristateEffect.TristateEffectImproved) {
		debuffStats = debuffStats.addPseudoStat(PseudoStat.PseudoStatMeleeHitPercent, 3);
		debuffStats = debuffStats.addPseudoStat(PseudoStat.PseudoStatRangedHitPercent, 3);
	}

	if (debuffs.improvedSealOfTheCrusader) {
		debuffStats = debuffStats.addPseudoStat(PseudoStat.PseudoStatMeleeCritPercent, 3);
		debuffStats = debuffStats.addPseudoStat(PseudoStat.PseudoStatRangedCritPercent, 3);
		debuffStats = debuffStats.addPseudoStat(PseudoStat.PseudoStatSpellCritPercent, 3);
	}

	if (debuffs.exposeWeaknessUptime && debuffs.exposeWeaknessHunterAgility) {
		debuffStats = debuffStats.addStat(Stat.StatAttackPower, exposeWeaknessAgility * 0.25);
		debuffStats = debuffStats.addStat(Stat.StatRangedAttackPower, exposeWeaknessAgility * 0.25);
	}

	if (debuffs.huntersMark != TristateEffect.TristateEffectMissing) {
		debuffStats = debuffStats.addStat(Stat.StatRangedAttackPower, 440);

		if (debuffs.huntersMark == TristateEffect.TristateEffectImproved) {
			debuffStats = debuffStats.addStat(Stat.StatAttackPower, 110);
		}
	}

	return debuffStats;
};
