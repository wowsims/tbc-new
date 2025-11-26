import * as PresetUtils from '../../core/preset_utils';
import { ConsumesSpec, Glyphs, Profession, PseudoStat, Race, Spec, Stat } from '../../core/proto/common';
import {
	DruidMajorGlyph,
	DruidMinorGlyph,
	FeralDruid_Options as FeralDruidOptions,
	FeralDruid_Rotation as FeralDruidRotation,
	FeralDruid_Rotation_AplType,
	FeralDruid_Rotation_HotwStrategy,
} from '../../core/proto/druid';
import { SavedTalents } from '../../core/proto/ui';
// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.
import PreraidGear from './gear_sets/preraid.gear.json';
export const PRERAID_PRESET = PresetUtils.makePresetGear('Pre-Raid', PreraidGear);
import P1Gear from './gear_sets/p1.gear.json';
export const P1_PRESET = PresetUtils.makePresetGear('P1', P1Gear);
import P2Gear from './gear_sets/p2.gear.json';
export const P2_PRESET = PresetUtils.makePresetGear('P2', P2Gear);
import P3Gear from './gear_sets/p3.gear.json';
export const P3_PRESET = PresetUtils.makePresetGear('P3 (Tentative)', P3Gear);
import P4Gear from './gear_sets/p4.gear.json';
export const P4_PRESET = PresetUtils.makePresetGear('P4', P4Gear);
import ItemSwapGear from './gear_sets/p1_item_swap.gear.json';
export const ITEM_SWAP_PRESET = PresetUtils.makePresetItemSwapGear('HotW Caster Weapon Swap', ItemSwapGear);

import DefaultApl from './apls/default.apl.json';
export const APL_ROTATION_DEFAULT = PresetUtils.makePresetAPLRotation('APL List View', DefaultApl);
import SingleTargetBuild from './builds/single_target.build.json';
export const PRESET_BUILD_ST = PresetUtils.makePresetBuildFromJSON("Single-Target Patchwerk", Spec.SpecFeralDruid, SingleTargetBuild);
import SustainedCleaveBuild from './builds/sustained_cleave.build.json';
export const PRESET_BUILD_CLEAVE = PresetUtils.makePresetBuildFromJSON("4-Target Cleave", Spec.SpecFeralDruid, SustainedCleaveBuild);

import { Stats } from '../../core/proto_utils/stats';

// Preset options for EP weights
export const DOC_EP_PRESET = PresetUtils.makePresetEpWeights(
	'DoC Bear-Weave',
	Stats.fromMap(
		{
			[Stat.StatStrength]: 0.39,
			[Stat.StatAgility]: 1.0,
			[Stat.StatAttackPower]: 0.37,
			[Stat.StatHitRating]: 0.45,
			[Stat.StatExpertiseRating]: 0.45,
			[Stat.StatCritRating]: 0.36,
			[Stat.StatHasteRating]: 0.27,
			[Stat.StatMasteryRating]: 0.42,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 0.73,
		},
	),
);

export const DOC_RORO_PRESET = PresetUtils.makePresetEpWeights(
	'DoC RoRo',
	Stats.fromMap(
		{
			[Stat.StatStrength]: 0.39,
			[Stat.StatAgility]: 1.0,
			[Stat.StatAttackPower]: 0.37,
			[Stat.StatHitRating]: 0.44,
			[Stat.StatExpertiseRating]: 0.44,
			[Stat.StatCritRating]: 0.49,
			[Stat.StatHasteRating]: 0.42,
			[Stat.StatMasteryRating]: 0.39,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 0.74,
		},
	),
);

export const HOTW_EP_PRESET = PresetUtils.makePresetEpWeights(
	'HotW Wrath-Weave',
	Stats.fromMap(
		{
			[Stat.StatStrength]: 0.34,
			[Stat.StatAgility]: 1.0,
			[Stat.StatAttackPower]: 0.32,
			[Stat.StatHitRating]: 0.37,
			[Stat.StatExpertiseRating]: 0.37,
			[Stat.StatCritRating]: 0.36,
			[Stat.StatHasteRating]: 0.25,
			[Stat.StatMasteryRating]: 0.37,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 0.72,
		},
	),
);

export const HOTW_RORO_PRESET = PresetUtils.makePresetEpWeights(
	'HotW RoRo',
	Stats.fromMap(
		{
			[Stat.StatStrength]: 0.34,
			[Stat.StatAgility]: 1.0,
			[Stat.StatAttackPower]: 0.32,
			[Stat.StatHitRating]: 0.36,
			[Stat.StatExpertiseRating]: 0.36,
			[Stat.StatCritRating]: 0.54,
			[Stat.StatHasteRating]: 0.35,
			[Stat.StatMasteryRating]: 0.35,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 0.72,
		},
	),
);

export const DefaultRotation = FeralDruidRotation.create({
	rotationType: FeralDruid_Rotation_AplType.SingleTarget,
	bearWeave: true,
	snekWeave: true,
	useNs: true,
	allowAoeBerserk: false,
	manualParams: false,
	minRoarOffset: 40,
	ripLeeway: 4,
	useBite: true,
	biteTime: 6,
	berserkBiteTime: 5,
	hotwStrategy: FeralDruid_Rotation_HotwStrategy.Wrath,
});

export const SIMPLE_ROTATION_DEFAULT = PresetUtils.makePresetSimpleRotation('Single Target Default', Spec.SpecFeralDruid, DefaultRotation);

//export const AoeRotation = FeralDruidRotation.create({
//	rotationType: FeralDruid_Rotation_AplType.Aoe,
//	bearWeave: true,
//	maintainFaerieFire: false,
//	snekWeave: true,
//	allowAoeBerserk: false,
//	cancelPrimalMadness: false,
//});
//
//export const AOE_ROTATION_DEFAULT = PresetUtils.makePresetSimpleRotation('AoE Default', Spec.SpecFeralDruid, AoeRotation);

// Default talents. Uses the wowhead calculator format, make the talents on
// https://wowhead.com/tbc/talent-calc and copy the numbers in the url.
export const StandardTalents = {
	name: 'DoC',
	data: SavedTalents.create({
		talentsString: "100302",
		glyphs: Glyphs.create({
			major1: 40923,
			major2: 40914,
			major3: 40897,
		}),
	}),
};

export const HotWTalents = {
	name: 'HotW',
	data: SavedTalents.create({
		talentsString: "100301",
		glyphs: Glyphs.create({
			major1: 40923,
			major2: 40914,
			major3: 40897,
		}),
	}),
};

export const DefaultOptions = FeralDruidOptions.create({
	assumeBleedActive: true,
});

export const DefaultConsumables = ConsumesSpec.create({
	flaskId: 76084, // Flask of Spring Blossoms
	foodId: 74648, // Sea Mist Rice Noodles
	potId: 76089, // Virmen's Bite
	prepotId: 76089, // Virmen's Bite
});

export const OtherDefaults = {
	distanceFromTarget: 24,
	highHpThreshold: 0.8,
	iterationCount: 25000,
	profession1: Profession.Engineering,
	profession2: Profession.ProfessionUnknown,
	race: Race.RaceWorgen,
};
