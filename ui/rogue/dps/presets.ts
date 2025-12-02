import * as PresetUtils from '../../core/preset_utils';
import { ConsumesSpec, PseudoStat, Stat } from '../../core/proto/common';
import { RogueOptions_PoisonOptions, Rogue_Options as RogueOptions } from '../../core/proto/rogue';
import { SavedTalents } from '../../core/proto/ui';
import { Stats } from '../../core/proto_utils/stats';
import BlankAPL from './apls/blank.apl.json'
import BlankGear from './gear_sets/blank.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

export const BLANK_APL = PresetUtils.makePresetAPLRotation('Blank', BlankAPL)

export const BLANK_GEARSET = PresetUtils.makePresetGear('Blank', BlankGear);

// Preset options for EP weights
export const P1_EP_PRESET = PresetUtils.makePresetEpWeights(
	'Sub',
	Stats.fromMap(
		{
			[Stat.StatAgility]: 1.0,
			[Stat.StatAllCritRating]: 1.0,
			[Stat.StatAllHasteRating]: 1.0,
			[Stat.StatAllHitRating]: 1.0,
			[Stat.StatArmorPenetration]: 1.0,
			[Stat.StatAttackPower]: 1.0,
			[Stat.StatMeleeCritRating]: 1.0,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 1,
			[PseudoStat.PseudoStatOffHandDps]: 1,
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

export const DefaultOptions = RogueOptions.create({
	classOptions: {
		lethalPoison: RogueOptions_PoisonOptions.DeadlyPoison,
		applyPoisonsManually: false,
		startingOverkillDuration: 20,
		vanishBreakTime: 0.1,
	},
});

export const DefaultConsumables = ConsumesSpec.create({
	flaskId: 76084, // Flask of the Winds
	foodId: 74648, // Skewered Eel
	potId: 76089, // Potion of the Tol'vir
	prepotId: 76089, // Potion of the Tol'vir
});

export const OtherDefaults = {
	distanceFromTarget: 5,
};
