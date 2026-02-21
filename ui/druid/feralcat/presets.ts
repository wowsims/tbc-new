import * as PresetUtils from '../../core/preset_utils';
import { Stats } from '../../core/proto_utils/stats';
import { ConsumesSpec, Profession, PseudoStat, Race, Stat } from '../../core/proto/common';
import { FeralCatDruid_Options as FeralDruidOptions } from '../../core/proto/druid';
import { SavedTalents } from '../../core/proto/ui';
import PreraidGear from './gear_sets/preraid.gear.json';
import P1Gear from './gear_sets/p1.gear.json';
import DefaultApl from './apls/default.apl.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.
export const PRERAID_PRESET = PresetUtils.makePresetGear('Pre-Raid', PreraidGear);
export const P1_PRESET = PresetUtils.makePresetGear('P1', P1Gear);

export const APL_ROTATION_DEFAULT = PresetUtils.makePresetAPLRotation('Default', DefaultApl);

// Preset options for EP weights
export const EP_PRESET = PresetUtils.makePresetEpWeights(
	'Default',
	Stats.fromMap(
		{
			[Stat.StatStrength]: 0.39,
			[Stat.StatAgility]: 1.0,
			[Stat.StatAttackPower]: 0.37,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 0.73,
		},
	),
);

// Default talents. Uses the wowhead calculator format, make the talents on
// https://wowhead.com/tbc/talent-calc and copy the numbers in the url.
export const StandardTalents = {
	name: 'Default',
	data: SavedTalents.create({
		talentsString: '',
	}),
};

export const DefaultOptions = FeralDruidOptions.create({});

export const DefaultConsumables = ConsumesSpec.create({});

export const OtherDefaults = {
	distanceFromTarget: 24,
	profession1: Profession.Engineering,
	profession2: Profession.ProfessionUnknown,
	race: Race.RaceTauren,
};
