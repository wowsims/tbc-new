import * as PresetUtils from '../../core/preset_utils';
import { ConsumesSpec, PseudoStat, Stat } from '../../core/proto/common';
import { Rogue_Options as RogueOptions } from '../../core/proto/rogue';
import { SavedTalents } from '../../core/proto/ui';
import { Stats } from '../../core/proto_utils/stats';
import BlankAPL from './apls/blank.apl.json'
import PreraidSwordsGear from './gear_sets/preraid.gear.json';
import P1SwordsGear from './gear_sets/p1.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

export const BLANK_APL = PresetUtils.makePresetAPLRotation('Blank', BlankAPL)

export const P1_SWORDS_GEAR = PresetUtils.makePresetGear('P1 Combat Swords', P1SwordsGear);
export const PREARAID_SWORDS_GEAR = PresetUtils.makePresetGear('Preraid Combat Swords', PreraidSwordsGear);


// Preset options for EP weights
export const P1_EP_PRESET = PresetUtils.makePresetEpWeights(
	'Combat Swords',
	Stats.fromMap(
		{
			[Stat.StatAgility]: 1.0,
			[Stat.StatMeleeHitRating]: 1.0,
			[Stat.StatMeleeHasteRating]: 1.0,
			[Stat.StatMeleeCritRating]: 1.0,
			[Stat.StatArmorPenetration]: 1.0,
			[Stat.StatAttackPower]: 1.0,
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
	name: 'Combat Swords',
	data: SavedTalents.create({
		talentsString: '00532012502-023305200005015002321151',
	}),
};

export const DefaultOptions = RogueOptions.create({
	classOptions: {
	},
});

export const DefaultConsumables = ConsumesSpec.create({
	flaskId: 76084, // Flask of the Winds
	foodId: 74648, // Skewered Eel
	potId: 76089, // Potion of the Tol'vir
});

export const OtherDefaults = {
	distanceFromTarget: 5,
};
