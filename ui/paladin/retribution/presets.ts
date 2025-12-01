import * as PresetUtils from '../../core/preset_utils.js';
import { APLRotation_Type as APLRotationType } from '../../core/proto/apl.js';
import { ConsumesSpec, Profession, PseudoStat, Race, Spec, Stat } from '../../core/proto/common.js';
import { PaladinSeal, RetributionPaladin_Options as RetributionPaladinOptions } from '../../core/proto/paladin.js';
import { SavedTalents } from '../../core/proto/ui.js';
import { Stats } from '../../core/proto_utils/stats';
import DefaultApl from './apls/default.apl.json';
import P2_Gear from './gear_sets/p2.gear.json';
import P3_Gear from './gear_sets/p3.gear.json';
import Preraid_Gear from './gear_sets/preraid.gear.json';
import P2RetBuild from './builds/p2.build.json';
import P3RetBuild from './builds/p3.build.json';
import PreraidRetBuild from './builds/preraid.build.json';

export const P2_GEAR_PRESET = PresetUtils.makePresetGear('P2', P2_Gear);
export const P3_GEAR_PRESET = PresetUtils.makePresetGear('P3 (WiP)', P3_Gear);
export const PRERAID_GEAR_PRESET = PresetUtils.makePresetGear('Pre-raid', Preraid_Gear);

export const APL_PRESET = PresetUtils.makePresetAPLRotation('Default', DefaultApl);

export const P1_P2_EP_PRESET = PresetUtils.makePresetEpWeights(
	'P2',
	Stats.fromMap(
		{
			[Stat.StatStrength]: 1.0,
			[Stat.StatAttackPower]: 0.44,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 1.91,
		},
	),
);

export const P3_EP_PRESET = PresetUtils.makePresetEpWeights(
	'P3',
	Stats.fromMap(
		{
			[Stat.StatStrength]: 1.0,
			[Stat.StatAttackPower]: 0.44,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 1.86,
		},
	),
);

export const PRERAID_EP_PRESET = PresetUtils.makePresetEpWeights(
	'Pre-raid',
	Stats.fromMap(
		{
			[Stat.StatStrength]: 1.0,
			[Stat.StatAttackPower]: 0.44,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 1.77,
		},
	),
);

export const DefaultTalents = {
	name: 'Default',
	data: SavedTalents.create({
		talentsString: '221223',
	}),
};

export const P2_BUILD_PRESET = PresetUtils.makePresetBuildFromJSON('P2', Spec.SpecRetributionPaladin, P2RetBuild, {
	epWeights: P1_P2_EP_PRESET,
	rotationType: APLRotationType.TypeAuto,
});

export const P3_BUILD_PRESET = PresetUtils.makePresetBuildFromJSON('P3 (WiP)', Spec.SpecRetributionPaladin, P3RetBuild, {
	epWeights: P3_EP_PRESET,
	rotationType: APLRotationType.TypeAuto,
});

export const PRERAID_BUILD_PRESET = PresetUtils.makePresetBuildFromJSON('Pre-raid', Spec.SpecRetributionPaladin, PreraidRetBuild, {
	epWeights: PRERAID_EP_PRESET,
	rotationType: APLRotationType.TypeAuto,
});

export const DefaultOptions = RetributionPaladinOptions.create({
	classOptions: {
		seal: PaladinSeal.Truth,
	},
});

export const DefaultConsumables = ConsumesSpec.create({
	flaskId: 76088, // Flask of Winter's Bite
	foodId: 74646, // Black Pepper Ribs and Shrimp
	potId: 76095, // Potion of Mogu Power
	prepotId: 76095, // Potion of Mogu Power
});

export const OtherDefaults = {
	profession1: Profession.Engineering,
	profession2: Profession.Blacksmithing,
	distanceFromTarget: 5,
	iterationCount: 25000,
	race: Race.RaceBloodElf,
};
