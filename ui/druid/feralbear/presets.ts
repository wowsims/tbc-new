import * as Mechanics from '../../core/constants/mechanics.js';
import * as PresetUtils from '../../core/preset_utils.js';
import { ConsumesSpec, Profession, PseudoStat, Stat } from '../../core/proto/common';
import { FeralBearDruid_Options as DruidOptions, FeralBearDruid_Rotation as DruidRotation } from '../../core/proto/druid.js';
import { SavedTalents } from '../../core/proto/ui.js';
import PreraidGear from './gear_sets/preraid.gear.json';
import P1Gear from './gear_sets/p1.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.
export const PRERAID_PRESET = PresetUtils.makePresetGear('Pre-Raid', PreraidGear);
export const P1_PRESET = PresetUtils.makePresetGear('P1', P1Gear);

export const DefaultSimpleRotation = DruidRotation.create({
	maintainFaerieFire: true,
	maintainDemoralizingRoar: true,
	demoTime: 4.0,
	pulverizeTime: 4.0,
	prepullStampede: true,
});

import { Stats } from '../../core/proto_utils/stats';
import DefaultApl from './apls/default.apl.json';
export const ROTATION_DEFAULT = PresetUtils.makePresetAPLRotation('Default', DefaultApl);

//export const ROTATION_PRESET_SIMPLE = PresetUtils.makePresetSimpleRotation('Simple Default', Spec.SpecGuardianDruid, DefaultSimpleRotation);

// Preset options for EP weights
export const SURVIVAL_EP_PRESET = PresetUtils.makePresetEpWeights(
	'Survival',
	Stats.fromMap(
		{
			[Stat.StatHealth]: 0.17,
			[Stat.StatStamina]: 3.93,
			[Stat.StatAgility]: 1.0,
			[Stat.StatArmor]: 4.81,
			[Stat.StatBonusArmor]: 1.1,
			[Stat.StatStrength]: 0.02,
			[Stat.StatAttackPower]: 0.02,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 0.0,
			[PseudoStat.PseudoStatMeleeHitPercent]: 1.07 * Mechanics.PHYSICAL_HIT_RATING_PER_HIT_PERCENT,
			[PseudoStat.PseudoStatSpellHitPercent]: 0.01 * Mechanics.SPELL_HIT_RATING_PER_HIT_PERCENT,
		},
	),
);

export const BALANCED_EP_PRESET = PresetUtils.makePresetEpWeights(
	'Balanced',
	Stats.fromMap(
		{
			[Stat.StatHealth]: 0.06,
			[Stat.StatStamina]: 1.41,
			[Stat.StatAgility]: 1.0,
			[Stat.StatArmor]: 1.7,
			[Stat.StatBonusArmor]: 0.39,
			[Stat.StatStrength]: 0.18,
			[Stat.StatAttackPower]: 0.18,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 0.84,
			[PseudoStat.PseudoStatMeleeHitPercent]: 1.5 * Mechanics.PHYSICAL_HIT_RATING_PER_HIT_PERCENT,
			[PseudoStat.PseudoStatSpellHitPercent]: 0.0 * Mechanics.SPELL_HIT_RATING_PER_HIT_PERCENT,
		},
	),
);

export const OFFENSIVE_EP_PRESET = PresetUtils.makePresetEpWeights(
	'Offensive',
	Stats.fromMap(
		{
			[Stat.StatHealth]: 0.03,
			[Stat.StatStamina]: 0.64,
			[Stat.StatAgility]: 1.0,
			[Stat.StatArmor]: 0.76,
			[Stat.StatBonusArmor]: 0.17,
			[Stat.StatStrength]: 0.23,
			[Stat.StatAttackPower]: 0.22,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 1.09,
			[PseudoStat.PseudoStatMeleeHitPercent]: 1.64 * Mechanics.PHYSICAL_HIT_RATING_PER_HIT_PERCENT,
			[PseudoStat.PseudoStatSpellHitPercent]: 0.0 * Mechanics.SPELL_HIT_RATING_PER_HIT_PERCENT,
		},
	),
);

// Default talents. Uses the wowhead calculator format, make the talents on
// https://wowhead.com/tbc/talent-calc and copy the numbers in the url.
export const DefaultTalents = {
	name: 'Default',
	data: SavedTalents.create({
		talentsString: '',
	}),
};

export const DefaultOptions = DruidOptions.create({});

export const DefaultConsumables = ConsumesSpec.create({});
export const OtherDefaults = {
	iterationCount: 50000,
	profession1: Profession.Engineering,
	profession2: Profession.ProfessionUnknown,
};
