import * as PresetUtils from '../../core/preset_utils';
import { Debuffs, PseudoStat, RaidBuffs, Stat, ConsumesSpec } from '../../core/proto/common';
import { defaultRaidBuffMajorDamageCooldowns } from '../../core/proto_utils/utils';
import { Stats } from '../../core/proto_utils/stats';
import { SavedTalents } from '../../core/proto/ui';
import { Mage_Options as MageOptions } from '../../core/proto/mage';
import BlankAPL from './apls/blank.apl.json'
import BlankGear from './gear_sets/blank.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

export const BLANK_APL = PresetUtils.makePresetAPLRotation('Blank', BlankAPL)

export const BLANK_GEARSET = PresetUtils.makePresetGear('Blank', BlankGear);

// Preset options for EP weights
export const P1_EP_PRESET = PresetUtils.makePresetEpWeights(
	'A',
	Stats.fromMap(
		{
			[Stat.StatAgility]: 1.0,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 1.43,
			[PseudoStat.PseudoStatOffHandDps]: 0.26,
		},
	),
);

export const Talents = {
	name: 'A',
	data: SavedTalents.create({
		talentsString: '',
	}),
};

export const DefaultOptions = MageOptions.create({
	classOptions: {

	},
});

export const OtherDefaults = {
	distanceFromTarget: 20,
};

export const DefaultConsumables = ConsumesSpec.create({
	flaskId: 76084, // Flask of the Winds
	foodId: 74648, // Skewered Eel
	potId: 76089, // Potion of the Tol'vir
	prepotId: 76089, // Potion of the Tol'vir
});

export const DefaultRaidBuffs = RaidBuffs.create({
	...defaultRaidBuffMajorDamageCooldowns(),
});

export const DefaultDebuffs = Debuffs.create({

});
