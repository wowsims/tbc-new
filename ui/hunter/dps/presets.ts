import * as PresetUtils from '../../core/preset_utils';
import { APLRotation_Type as APLRotationType } from '../../core/proto/apl.js';
import { ConsumesSpec, Profession, PseudoStat, Spec, Stat } from '../../core/proto/common';
import { HunterOptions_PetType as PetType, Hunter_Options as HunterOptions } from '../../core/proto/hunter';
import { SavedTalents } from '../../core/proto/ui';
import { Stats } from '../../core/proto_utils/stats';
import BlankAPL from './apls/blank.apl.json'
import BlankGear from './gear_sets/blank.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

export const BLANK_APL = PresetUtils.makePresetAPLRotation('Blank', BlankAPL)

export const BLANK_GEARSET = PresetUtils.makePresetGear('Blank', BlankGear);

export const P2_EP_PRESET = PresetUtils.makePresetEpWeights(
	'P2',
	Stats.fromMap(
		{
			[Stat.StatAgility]: 1,
			[Stat.StatRangedAttackPower]: 0.35,
		},
		{
			[PseudoStat.PseudoStatRangedDps]: 1.75,
		},
	),
);
export const P3_EP_PRESET = PresetUtils.makePresetEpWeights(
	'P3',
	Stats.fromMap(
		{
			[Stat.StatAgility]: 1,
			[Stat.StatRangedAttackPower]: 0.33,
		},
		{
			[PseudoStat.PseudoStatRangedDps]: 1.72,
		},
	),
);

// Default talents. Uses the wowhead calculator format, make the talents on
// https://wowhead.com/wotlk/talent-calc and copy the numbers in the url.

export const Talents = {
	name: 'A',
	data: SavedTalents.create({
		talentsString: '',
	}),
};

// export const PRERAID_PRESET = PresetUtils.makePresetBuildFromJSON('Pre-raid', Spec.SpecHunter, PreRaidBuild, {
// 	epWeights: P2_EP_PRESET,
// 	rotationType: APLRotationType.TypeAuto,
// });

export const MMDefaultOptions = HunterOptions.create({
	classOptions: {
		useHuntersMark: true,
		petType: PetType.Tallstrider,
		petUptime: 1,
		glaiveTossSuccess: 0.8,
	},
});

export const DefaultConsumables = ConsumesSpec.create({
	flaskId: 76084, // Flask of the Winds
	foodId: 74648, // Seafood Magnifique Feast
	potId: 76089, // Potion of the Tol'vir
	prepotId: 76089, // Potion of the Tol'vir
});

export const OtherDefaults = {
	distanceFromTarget: 24,
	iterationCount: 25000,
	profession1: Profession.Engineering,
	profession2: Profession.Tailoring,
};
