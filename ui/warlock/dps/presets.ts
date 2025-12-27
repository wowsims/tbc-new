import * as PresetUtils from '../../core/preset_utils';
import { ConsumesSpec, PseudoStat, Stat } from '../../core/proto/common';
import { Warlock_Options as WarlockOptions, WarlockOptions_Armor, WarlockOptions_Summon } from '../../core/proto/warlock';
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
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 1.43,
			[PseudoStat.PseudoStatOffHandDps]: 0.26,
		},
	),
);

// Default talents. Uses the wowhead calculator format, make the talents on
// https://wowhead.com/wotlk/talent-calc and copy the numbers in the url.

export const Talents = {
	name: 'Destruction',
	data: SavedTalents.create({
		talentsString: '-20501301332001-50500051220051053105',
	}),
};

export const DefaultOptions = WarlockOptions.create({
	classOptions: {
		armor: WarlockOptions_Armor.FelArmor,
		summon: WarlockOptions_Summon.Succubus,
		detonateSeed: false,
		useItemSwapBonusStats: false,
		sacrificeSummon: true,
	},
});

export const DefaultConsumables = ConsumesSpec.create({
	flaskId: 22866, // Flask of Pure Death
	foodId: 33825, // Poached Bluefish
	potId: 22839, // Destructive Potion
	prepotId: 22839, // Destructive Potion
	mhImbueId: 20749 // Brilliant Wizard Oil
});

export const OtherDefaults = {
	distanceFromTarget: 5,
};
