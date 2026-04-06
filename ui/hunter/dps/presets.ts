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
import P1PreRaidBuild from './builds/p1_pre_raid.build.json';
import P1BM2HBuild from './builds/p1_bm_2h.build.json';
import P1BMDWBuild from './builds/p1_bm_dw.build.json';
import P1SV2HBuild from './builds/p1_sv_2h.build.json';
import P1SVDWBuild from './builds/p1_sv_dw.build.json';
import P1PreRaidGear from './gear_sets/p1_pre_raid.gear.json';
import P1BMDW6PGear from './gear_sets/p1_bm_dw_6p.gear.json';
import P1BMDW9PGear from './gear_sets/p1_bm_dw_9p.gear.json';
import P1BM2H6PGear from './gear_sets/p1_bm_2h_6p.gear.json';
import P1BM2H9PGear from './gear_sets/p1_bm_2h_9p.gear.json';
import P1SVDW3PGear from './gear_sets/p1_sv_dw_3p.gear.json';
import P1SVDW6PGear from './gear_sets/p1_sv_dw_6p.gear.json';
import P1SV2H3PGear from './gear_sets/p1_sv_2h_3p.gear.json';
import P1SV2H6PGear from './gear_sets/p1_sv_2h_6p.gear.json';

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

export const P1_PreRaid_GEARSET = PresetUtils.makePresetGear('P1 - Pre-raid', P1PreRaidGear);
export const P1_BM_DW_6P_GEARSET = PresetUtils.makePresetGear('P1 - BM - DW (imp. FF)', P1BMDW6PGear);
export const P1_BM_DW_9P_GEARSET = PresetUtils.makePresetGear('P1 - BM - DW (no imp. FF)', P1BMDW9PGear);
export const P1_BM_2H_6P_GEARSET = PresetUtils.makePresetGear('P1 - BM - 2H (imp. FF)', P1BM2H6PGear);
export const P1_BM_2H_9P_GEARSET = PresetUtils.makePresetGear('P1 - BM - 2H (no imp. FF)', P1BM2H9PGear);
export const P1_SV_DW_3P_GEARSET = PresetUtils.makePresetGear('P1 - SV - DW (imp. FF)', P1SVDW3PGear);
export const P1_SV_DW_6P_GEARSET = PresetUtils.makePresetGear('P1 - SV - DW (no imp. FF)', P1SVDW6PGear);
export const P1_SV_2H_3P_GEARSET = PresetUtils.makePresetGear('P1 - SV - 2H (imp. FF)', P1SV2H3PGear);
export const P1_SV_2H_6P_GEARSET = PresetUtils.makePresetGear('P1 - SV - 2H (no imp. FF)', P1SV2H6PGear);

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

export const P1_PRESET_BUILD_PRE_RAID = PresetUtils.makePresetBuildFromJSON('P1 - Pre-Raid', Spec.SpecHunter, P1PreRaidBuild);
export const P1_PRESET_BUILD_BM_2H = PresetUtils.makePresetBuildFromJSON('P1 - BM - 2H', Spec.SpecHunter, P1BM2HBuild);
export const P1_PRESET_BUILD_BM_DW = PresetUtils.makePresetBuildFromJSON('P1 - BM - DW', Spec.SpecHunter, P1BMDWBuild);
export const P1_PRESET_BUILD_SV_2H = PresetUtils.makePresetBuildFromJSON('P1 - SV - 2H', Spec.SpecHunter, P1SV2HBuild);
export const P1_PRESET_BUILD_SV_DW = PresetUtils.makePresetBuildFromJSON('P1 - SV - DW', Spec.SpecHunter, P1SVDWBuild);
