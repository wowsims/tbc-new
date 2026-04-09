import { Phase } from '../../core/constants/other';
import * as PresetUtils from '../../core/preset_utils';
import {
	Class,
	ConsumesSpec,
	Debuffs,
	Drums,
	IndividualBuffs,
	PartyBuffs,
	Profession,
	PseudoStat,
	Race,
	RaidBuffs,
	Spec,
	Stat,
	TristateEffect,
} from '../../core/proto/common';
import {
	HunterOptions_PetType as PetType,
	Hunter_Options as HunterOptions,
	HunterOptions_Ammo,
	HunterOptions_QuiverBonus,
	Hunter_Rotation,
} from '../../core/proto/hunter';
import { SavedTalents } from '../../core/proto/ui';
import { Stats } from '../../core/proto_utils/stats';
import { defaultExposeWeaknessSettings, defaultRaidBuffMajorDamageCooldowns } from '../../core/proto_utils/utils';
import DefaultAPL from './apls/default.apl.json';

const allBuilds = import.meta.glob('./builds/**/*.build.json', { eager: true, import: 'default' });
const allGearSets = import.meta.glob('./gear_sets/**/*.gear.json', { eager: true, import: 'default' });

function getBuild(phase: Phase, spec: 'bm' | 'sv', variant: string) {
	return allBuilds[`./builds/phase_${phase}/${spec}/${variant}.build.json`];
}

function getGear(phase: Phase, spec: 'bm' | 'sv', variant: string) {
	return allGearSets[`./gear_sets/phase_${phase}/${spec}/${variant}.gear.json`];
}

export const DefaultRotation = PresetUtils.makePresetAPLRotation('APL', DefaultAPL);

export const TurretRotation = Hunter_Rotation.create({
	viperStartManaPercent: 0.05,
	viperStopManaPercent: 0.25,
	meleeWeave: false,
	weaveOnlyRaptor: false,
	timeToWeave: 400,
	useMulti: true,
	useArcane: true,
});
export const TurretSimple = PresetUtils.makePresetSimpleRotation('Turret', Spec.SpecHunter, TurretRotation);

export const WeaveRotation = Hunter_Rotation.create({
	viperStartManaPercent: 0.05,
	viperStopManaPercent: 0.25,
	meleeWeave: true,
	weaveOnlyRaptor: false,
	timeToWeave: 400,
	useMulti: true,
	useArcane: true,
});
export const WeaveSimple = PresetUtils.makePresetSimpleRotation('Weave', Spec.SpecHunter, WeaveRotation);

// Phase 1
export const P1_PreRaid_GEARSET = PresetUtils.makePresetGear('Pre-Raid', getGear(Phase.Phase1, 'bm', 'pre_raid'), { phase: Phase.Phase1, group: 'Beast Mastery' });
export const P1_BM_DW_6P_GEARSET = PresetUtils.makePresetGear('DW - 6% hit', getGear(Phase.Phase1, 'bm', 'dw_6p'), { phase: Phase.Phase1, group: 'Beast Mastery' });
export const P1_BM_DW_9P_GEARSET = PresetUtils.makePresetGear('DW - 9% hit', getGear(Phase.Phase1, 'bm', 'dw_9p'), { phase: Phase.Phase1, group: 'Beast Mastery' });
export const P1_BM_2H_6P_GEARSET = PresetUtils.makePresetGear('2H - 6% hit', getGear(Phase.Phase1, 'bm', '2h_6p'), { phase: Phase.Phase1, group: 'Beast Mastery' });
export const P1_BM_2H_9P_GEARSET = PresetUtils.makePresetGear('2H - 9% hit', getGear(Phase.Phase1, 'bm', '2h_9p'), { phase: Phase.Phase1, group: 'Beast Mastery' });
export const P1_SV_DW_3P_GEARSET = PresetUtils.makePresetGear('DW - 6% hit', getGear(Phase.Phase1, 'sv', 'dw_3p'), { phase: Phase.Phase1, group: 'Survival' });
export const P1_SV_DW_6P_GEARSET = PresetUtils.makePresetGear('DW - 9% hit', getGear(Phase.Phase1, 'sv', 'dw_6p'), { phase: Phase.Phase1, group: 'Survival' });
export const P1_SV_2H_3P_GEARSET = PresetUtils.makePresetGear('2H - 6% hit', getGear(Phase.Phase1, 'sv', '2h_3p'), { phase: Phase.Phase1, group: 'Survival' });
export const P1_SV_2H_6P_GEARSET = PresetUtils.makePresetGear('2H - 9% hit', getGear(Phase.Phase1, 'sv', '2h_6p'), { phase: Phase.Phase1, group: 'Survival' });

// Phase 2
export const P2_BM_DW_6P_GEARSET = PresetUtils.makePresetGear('DW - 6% hit', getGear(Phase.Phase2, 'bm', 'dw_6p'), { phase: Phase.Phase2, group: 'Beast Mastery' });
export const P2_BM_DW_9P_GEARSET = PresetUtils.makePresetGear('DW - 9% hit', getGear(Phase.Phase2, 'bm', 'dw_9p'), { phase: Phase.Phase2, group: 'Beast Mastery' });
export const P2_BM_2H_6P_GEARSET = PresetUtils.makePresetGear('2H - 6% hit', getGear(Phase.Phase2, 'bm', '2h_6p'), { phase: Phase.Phase2, group: 'Beast Mastery' });
export const P2_BM_2H_9P_GEARSET = PresetUtils.makePresetGear('2H - 9% hit', getGear(Phase.Phase2, 'bm', '2h_9p'), { phase: Phase.Phase2, group: 'Beast Mastery' });
export const P2_SV_DW_6P_GEARSET = PresetUtils.makePresetGear('DW - 6% hit', getGear(Phase.Phase2, 'sv', 'dw_6p'), { phase: Phase.Phase2, group: 'Survival' });
export const P2_SV_2H_6P_GEARSET = PresetUtils.makePresetGear('2H - 6% hit', getGear(Phase.Phase2, 'sv', '2h_6p'), { phase: Phase.Phase2, group: 'Survival' });

// Phase 3
export const P3_BM_DW_6P_GEARSET = PresetUtils.makePresetGear('DW - 6% hit', getGear(Phase.Phase3, 'bm', 'dw_6p'), { phase: Phase.Phase3, group: 'Beast Mastery' });
export const P3_BM_DW_9P_GEARSET = PresetUtils.makePresetGear('DW - 9% hit', getGear(Phase.Phase3, 'bm', 'dw_9p'), { phase: Phase.Phase3, group: 'Beast Mastery' });
export const P3_BM_2H_6P_GEARSET = PresetUtils.makePresetGear('2H - 6% hit', getGear(Phase.Phase3, 'bm', '2h_6p'), { phase: Phase.Phase3, group: 'Beast Mastery' });
export const P3_BM_2H_9P_GEARSET = PresetUtils.makePresetGear('2H - 9% hit', getGear(Phase.Phase3, 'bm', '2h_9p'), { phase: Phase.Phase3, group: 'Beast Mastery' });
export const P3_SV_DW_6P_GEARSET = PresetUtils.makePresetGear('DW - 6% hit', getGear(Phase.Phase3, 'sv', 'dw_6p'), { phase: Phase.Phase3, group: 'Survival' });
export const P3_SV_DW_9P_GEARSET = PresetUtils.makePresetGear('DW - 9% hit', getGear(Phase.Phase3, 'sv', 'dw_9p'), { phase: Phase.Phase3, group: 'Survival' });
export const P3_SV_2H_6P_GEARSET = PresetUtils.makePresetGear('2H - 6% hit', getGear(Phase.Phase3, 'sv', '2h_6p'), { phase: Phase.Phase3, group: 'Survival' });
export const P3_SV_2H_9P_GEARSET = PresetUtils.makePresetGear('2H - 9% hit', getGear(Phase.Phase3, 'sv', '2h_9p'), { phase: Phase.Phase3, group: 'Survival' });

// Phase 4
export const P4_BM_DW_6P_GEARSET = PresetUtils.makePresetGear('DW - 6% hit', getGear(Phase.Phase4, 'bm', 'dw_6p'), { phase: Phase.Phase4, group: 'Beast Mastery' });
export const P4_BM_DW_9P_GEARSET = PresetUtils.makePresetGear('DW - 9% hit', getGear(Phase.Phase4, 'bm', 'dw_9p'), { phase: Phase.Phase4, group: 'Beast Mastery' });
export const P4_BM_2H_6P_GEARSET = PresetUtils.makePresetGear('2H - 6% hit', getGear(Phase.Phase4, 'bm', '2h_6p'), { phase: Phase.Phase4, group: 'Beast Mastery' });
export const P4_BM_2H_9P_GEARSET = PresetUtils.makePresetGear('2H - 9% hit', getGear(Phase.Phase4, 'bm', '2h_9p'), { phase: Phase.Phase4, group: 'Beast Mastery' });
export const P4_SV_DW_6P_GEARSET = PresetUtils.makePresetGear('DW - 6% hit', getGear(Phase.Phase4, 'sv', 'dw_6p'), { phase: Phase.Phase4, group: 'Survival' });
export const P4_SV_DW_9P_GEARSET = PresetUtils.makePresetGear('DW - 9% hit', getGear(Phase.Phase4, 'sv', 'dw_9p'), { phase: Phase.Phase4, group: 'Survival' });
export const P4_SV_2H_6P_GEARSET = PresetUtils.makePresetGear('2H - 6% hit', getGear(Phase.Phase4, 'sv', '2h_6p'), { phase: Phase.Phase4, group: 'Survival' });
export const P4_SV_2H_9P_GEARSET = PresetUtils.makePresetGear('2H - 9% hit', getGear(Phase.Phase4, 'sv', '2h_9p'), { phase: Phase.Phase4, group: 'Survival' });

export const P1_BM_EP_PRESET = PresetUtils.makePresetEpWeights(
	'P1 BM',
	Stats.fromMap(
		{
			[Stat.StatAgility]: 1,
			[Stat.StatStrength]: 0.06,
			[Stat.StatIntellect]: 0.01,
			[Stat.StatAttackPower]: 0.06,
			[Stat.StatRangedAttackPower]: 0.4,
			[Stat.StatMeleeHitRating]: 0.12,
			[Stat.StatMeleeCritRating]: 0.92,
			[Stat.StatMeleeHasteRating]: 0.788,
			[Stat.StatArmorPenetration]: 0.16,
		},
		{
			[PseudoStat.PseudoStatRangedDps]: 1.75,
		},
	),
);

export const P1_SV_EP_PRESET = PresetUtils.makePresetEpWeights(
	'P1 SV',
	Stats.fromMap(
		{
			[Stat.StatAgility]: 1,
			[Stat.StatStrength]: 0.06,
			[Stat.StatIntellect]: 0.01,
			[Stat.StatAttackPower]: 0.06,
			[Stat.StatRangedAttackPower]: 0.4,
			[Stat.StatMeleeHitRating]: 0.12,
			[Stat.StatMeleeCritRating]: 0.92,
			[Stat.StatMeleeHasteRating]: 0.788,
			[Stat.StatArmorPenetration]: 0.16,
		},
		{
			[PseudoStat.PseudoStatRangedDps]: 1.75,
		},
	),
);

// Default talents. Uses the wowhead calculator format, make the talents on
// https://wowhead.com/wotlk/talent-calc and copy the numbers in the url.

export const BMTalents = {
	name: 'BM',
	data: SavedTalents.create({
		talentsString: '522002005150122431051-0505201205',
	}),
};
export const SVTalents = {
	name: 'SV',
	data: SavedTalents.create({
		talentsString: '502-0550201205-333200022003223005103',
	}),
};

export const DefaultOptions = HunterOptions.create({
	classOptions: {
		ammo: HunterOptions_Ammo.WardensArrow,
		quiverBonus: HunterOptions_QuiverBonus.Speed15,
		petType: PetType.Ravager,
		petUptime: 1,
		petSingleAbility: false,
	},
});

export const DefaultIndividualBuffs = IndividualBuffs.create({
	blessingOfKings: true,
	blessingOfMight: TristateEffect.TristateEffectImproved,
	blessingOfWisdom: TristateEffect.TristateEffectImproved,
	unleashedRage: true,
});

export const DefaultPartyBuffs = PartyBuffs.create({
	battleShout: TristateEffect.TristateEffectImproved,
	braidedEterniumChain: true,
	ferociousInspiration: 1,
	graceOfAirTotem: TristateEffect.TristateEffectImproved,
	leaderOfThePack: TristateEffect.TristateEffectImproved,
	strengthOfEarthTotem: TristateEffect.TristateEffectImproved,
	totemTwisting: true,
	windfuryTotem: TristateEffect.TristateEffectImproved,
	drums: Drums.LesserDrumsOfBattle,
});

export const DefaultRaidBuffs = RaidBuffs.create({
	...defaultRaidBuffMajorDamageCooldowns(Class.ClassWarrior),
	arcaneBrilliance: true,
	divineSpirit: TristateEffect.TristateEffectImproved,
	giftOfTheWild: TristateEffect.TristateEffectImproved,
	powerWordFortitude: TristateEffect.TristateEffectImproved,
	shadowProtection: true,
});

export const DefaultDebuffs = Debuffs.create({
	bloodFrenzy: true,
	curseOfRecklessness: true,
	exposeArmor: TristateEffect.TristateEffectImproved,
	...defaultExposeWeaknessSettings(),
	faerieFire: TristateEffect.TristateEffectImproved,
	giftOfArthas: true,
	huntersMark: TristateEffect.TristateEffectImproved,
	improvedSealOfTheCrusader: TristateEffect.TristateEffectImproved,
	insectSwarm: true,
	judgementOfLight: true,
	judgementOfWisdom: true,
	mangle: true,
	misery: true,
	sunderArmor: true,
});

export const DefaultConsumables = ConsumesSpec.create({
	battleElixirId: 22831, // Elixir of Major Agility
	guardianElixirId: 22840, // Elixir of Major Mageblood
	foodId: 27659, // Warp Burger
	potId: 22838, // Haste Potion
	conjuredId: 12662,
	explosiveId: 30217,
	petFoodId: 33874, // Kibler's Bits
	petScrollAgi: true,
	petScrollStr: true,
	superSapper: true,
	goblinSapper: true,
	scrollAgi: true,
	scrollStr: true,
});

export const OtherDefaults = {
	distanceFromTarget: 7,
	iterationCount: 25000,
	profession1: Profession.Engineering,
	profession2: Profession.Blacksmithing,
	race: Race.RaceOrc,
};

// Phase 1
export const P1_PRESET_BUILD_PRE_RAID = PresetUtils.makePresetBuildFromJSON('Pre-Raid', Spec.SpecHunter, getBuild(Phase.Phase1, 'bm', 'pre_raid'), { phase: Phase.Phase1, group: 'Beast Mastery' });
export const P1_PRESET_BUILD_BM_2H_6P = PresetUtils.makePresetBuildFromJSON('2H - 6% hit', Spec.SpecHunter, getBuild(Phase.Phase1, 'bm', '2h_6p'), { phase: Phase.Phase1, group: 'Beast Mastery' });
export const P1_PRESET_BUILD_BM_2H_9P = PresetUtils.makePresetBuildFromJSON('2H - 9% hit', Spec.SpecHunter, getBuild(Phase.Phase1, 'bm', '2h_9p'), { phase: Phase.Phase1, group: 'Beast Mastery' });
export const P1_PRESET_BUILD_BM_DW_6P = PresetUtils.makePresetBuildFromJSON('DW - 6% hit', Spec.SpecHunter, getBuild(Phase.Phase1, 'bm', 'dw_6p'), { phase: Phase.Phase1, group: 'Beast Mastery' });
export const P1_PRESET_BUILD_BM_DW_9P = PresetUtils.makePresetBuildFromJSON('DW - 9% hit', Spec.SpecHunter, getBuild(Phase.Phase1, 'bm', 'dw_9p'), { phase: Phase.Phase1, group: 'Beast Mastery' });
export const P1_PRESET_BUILD_SV_2H_6P = PresetUtils.makePresetBuildFromJSON('2H - 6% hit', Spec.SpecHunter, getBuild(Phase.Phase1, 'sv', '2h_6p'), { phase: Phase.Phase1, group: 'Survival' });
export const P1_PRESET_BUILD_SV_2H_9P = PresetUtils.makePresetBuildFromJSON('2H - 9% hit', Spec.SpecHunter, getBuild(Phase.Phase1, 'sv', '2h_9p'), { phase: Phase.Phase1, group: 'Survival' });
export const P1_PRESET_BUILD_SV_DW_6P = PresetUtils.makePresetBuildFromJSON('DW - 6% hit', Spec.SpecHunter, getBuild(Phase.Phase1, 'sv', 'dw_6p'), { phase: Phase.Phase1, group: 'Survival' });
export const P1_PRESET_BUILD_SV_DW_9P = PresetUtils.makePresetBuildFromJSON('DW - 9% hit', Spec.SpecHunter, getBuild(Phase.Phase1, 'sv', 'dw_9p'), { phase: Phase.Phase1, group: 'Survival' });

// Phase 2
export const P2_PRESET_BUILD_BM_2H_6P = PresetUtils.makePresetBuildFromJSON('2H - 6% hit', Spec.SpecHunter, getBuild(Phase.Phase2, 'bm', '2h_6p'), { phase: Phase.Phase2, group: 'Beast Mastery' });
export const P2_PRESET_BUILD_BM_2H_9P = PresetUtils.makePresetBuildFromJSON('2H - 9% hit', Spec.SpecHunter, getBuild(Phase.Phase2, 'bm', '2h_9p'), { phase: Phase.Phase2, group: 'Beast Mastery' });
export const P2_PRESET_BUILD_BM_DW_6P = PresetUtils.makePresetBuildFromJSON('DW - 6% hit', Spec.SpecHunter, getBuild(Phase.Phase2, 'bm', 'dw_6p'), { phase: Phase.Phase2, group: 'Beast Mastery' });
export const P2_PRESET_BUILD_BM_DW_9P = PresetUtils.makePresetBuildFromJSON('DW - 9% hit', Spec.SpecHunter, getBuild(Phase.Phase2, 'bm', 'dw_9p'), { phase: Phase.Phase2, group: 'Beast Mastery' });
export const P2_PRESET_BUILD_SV_2H_6P = PresetUtils.makePresetBuildFromJSON('2H - 6% hit', Spec.SpecHunter, getBuild(Phase.Phase2, 'sv', '2h_6p'), { phase: Phase.Phase2, group: 'Survival' });
export const P2_PRESET_BUILD_SV_DW_6P = PresetUtils.makePresetBuildFromJSON('DW - 6% hit', Spec.SpecHunter, getBuild(Phase.Phase2, 'sv', 'dw_6p'), { phase: Phase.Phase2, group: 'Survival' });

// Phase 3
export const P3_PRESET_BUILD_BM_2H_6P = PresetUtils.makePresetBuildFromJSON('2H - 6% hit', Spec.SpecHunter, getBuild(Phase.Phase3, 'bm', '2h_6p'), { phase: Phase.Phase3, group: 'Beast Mastery' });
export const P3_PRESET_BUILD_BM_2H_9P = PresetUtils.makePresetBuildFromJSON('2H - 9% hit', Spec.SpecHunter, getBuild(Phase.Phase3, 'bm', '2h_9p'), { phase: Phase.Phase3, group: 'Beast Mastery' });
export const P3_PRESET_BUILD_BM_DW_6P = PresetUtils.makePresetBuildFromJSON('DW - 6% hit', Spec.SpecHunter, getBuild(Phase.Phase3, 'bm', 'dw_6p'), { phase: Phase.Phase3, group: 'Beast Mastery' });
export const P3_PRESET_BUILD_BM_DW_9P = PresetUtils.makePresetBuildFromJSON('DW - 9% hit', Spec.SpecHunter, getBuild(Phase.Phase3, 'bm', 'dw_9p'), { phase: Phase.Phase3, group: 'Beast Mastery' });
export const P3_PRESET_BUILD_SV_2H_6P = PresetUtils.makePresetBuildFromJSON('2H - 6% hit', Spec.SpecHunter, getBuild(Phase.Phase3, 'sv', '2h_6p'), { phase: Phase.Phase3, group: 'Survival' });
export const P3_PRESET_BUILD_SV_2H_9P = PresetUtils.makePresetBuildFromJSON('2H - 9% hit', Spec.SpecHunter, getBuild(Phase.Phase3, 'sv', '2h_9p'), { phase: Phase.Phase3, group: 'Survival' });
export const P3_PRESET_BUILD_SV_DW_6P = PresetUtils.makePresetBuildFromJSON('DW - 6% hit', Spec.SpecHunter, getBuild(Phase.Phase3, 'sv', 'dw_6p'), { phase: Phase.Phase3, group: 'Survival' });
export const P3_PRESET_BUILD_SV_DW_9P = PresetUtils.makePresetBuildFromJSON('DW - 9% hit', Spec.SpecHunter, getBuild(Phase.Phase3, 'sv', 'dw_9p'), { phase: Phase.Phase3, group: 'Survival' });

// Phase 4
export const P4_PRESET_BUILD_BM_2H_6P = PresetUtils.makePresetBuildFromJSON('2H - 6% hit', Spec.SpecHunter, getBuild(Phase.Phase4, 'bm', '2h_6p'), { phase: Phase.Phase4, group: 'Beast Mastery' });
export const P4_PRESET_BUILD_BM_2H_9P = PresetUtils.makePresetBuildFromJSON('2H - 9% hit', Spec.SpecHunter, getBuild(Phase.Phase4, 'bm', '2h_9p'), { phase: Phase.Phase4, group: 'Beast Mastery' });
export const P4_PRESET_BUILD_BM_DW_6P = PresetUtils.makePresetBuildFromJSON('DW - 6% hit', Spec.SpecHunter, getBuild(Phase.Phase4, 'bm', 'dw_6p'), { phase: Phase.Phase4, group: 'Beast Mastery' });
export const P4_PRESET_BUILD_BM_DW_9P = PresetUtils.makePresetBuildFromJSON('DW - 9% hit', Spec.SpecHunter, getBuild(Phase.Phase4, 'bm', 'dw_9p'), { phase: Phase.Phase4, group: 'Beast Mastery' });
export const P4_PRESET_BUILD_SV_2H_6P = PresetUtils.makePresetBuildFromJSON('2H - 6% hit', Spec.SpecHunter, getBuild(Phase.Phase4, 'sv', '2h_6p'), { phase: Phase.Phase4, group: 'Survival' });
export const P4_PRESET_BUILD_SV_2H_9P = PresetUtils.makePresetBuildFromJSON('2H - 9% hit', Spec.SpecHunter, getBuild(Phase.Phase4, 'sv', '2h_9p'), { phase: Phase.Phase4, group: 'Survival' });
export const P4_PRESET_BUILD_SV_DW_6P = PresetUtils.makePresetBuildFromJSON('DW - 6% hit', Spec.SpecHunter, getBuild(Phase.Phase4, 'sv', 'dw_6p'), { phase: Phase.Phase4, group: 'Survival' });
export const P4_PRESET_BUILD_SV_DW_9P = PresetUtils.makePresetBuildFromJSON('DW - 9% hit', Spec.SpecHunter, getBuild(Phase.Phase4, 'sv', 'dw_9p'), { phase: Phase.Phase4, group: 'Survival' });
