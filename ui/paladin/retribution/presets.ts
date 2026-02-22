import * as PresetUtils from '../../core/preset_utils.js';
import { ConsumesSpec, Profession, PseudoStat, Race, Stat } from '../../core/proto/common.js';
import { RetributionPaladin_Options as RetributionPaladinOptions } from '../../core/proto/paladin.js';
import { SavedTalents } from '../../core/proto/ui.js';
import { Stats } from '../../core/proto_utils/stats';
import DefaultApl from './apls/default.apl.json';
import P1_Gear from './gear_sets/p1.gear.json';
import Preraid_Gear from './gear_sets/preraid.gear.json';

export const P1_GEAR_PRESET = PresetUtils.makePresetGear('P1', P1_Gear);
export const PRERAID_GEAR_PRESET = PresetUtils.makePresetGear('Pre-raid', Preraid_Gear);

export const APL_PRESET = PresetUtils.makePresetAPLRotation('Default', DefaultApl);

export const P1_EP_PRESET = PresetUtils.makePresetEpWeights(
	'P1',
	Stats.fromMap(
		{
			[Stat.StatStrength]: 1.0,
			[Stat.StatAgility]: 1.0,
			[Stat.StatAttackPower]: 1.0,
			[Stat.StatMeleeHitRating]: 1.0,
			[Stat.StatMeleeCritRating]: 1.0,
			[Stat.StatMeleeHasteRating]: 1.0,
			[Stat.StatArmorPenetration]: 1.0,
			[Stat.StatExpertiseRating]: 1.0,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 1.0,
		},
	),
);

export const DefaultTalents = {
	name: 'Default',
	data: SavedTalents.create({
		talentsString: '',
	}),
};

export const DefaultOptions = RetributionPaladinOptions.create({
	classOptions: {},
});

export const DefaultConsumables = ConsumesSpec.create({});

export const OtherDefaults = {
	profession1: Profession.Engineering,
	profession2: Profession.Blacksmithing,
	distanceFromTarget: 5,
	iterationCount: 25000,
	race: Race.RaceBloodElf,
};
