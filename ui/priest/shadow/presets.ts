import * as PresetUtils from '../../core/preset_utils';
import { ConsumesSpec, Debuffs, IndividualBuffs, Profession, PseudoStat, RaidBuffs, Stat } from '../../core/proto/common';
import { PriestOptions_Armor, ShadowPriest_Options as Options } from '../../core/proto/priest';
import { SavedTalents } from '../../core/proto/ui';
import { Stats, UnitStat, UnitStatPresets } from '../../core/proto_utils/stats';
import { defaultRaidBuffMajorDamageCooldowns } from '../../core/proto_utils/utils';
import DefaultApl from './apls/default.apl.json';
import P1Gear from './gear_sets/p1.gear.json';
import PreRaidGear from './gear_sets/pre_raid.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.
export const PRE_RAID_PRESET = PresetUtils.makePresetGear('Pre Raid Preset', PreRaidGear);
export const P1_PRESET = PresetUtils.makePresetGear('P1 Preset', P1Gear);
export const ROTATION_PRESET_DEFAULT = PresetUtils.makePresetAPLRotation('Default', DefaultApl);

// Preset options for EP weights
export const P1_EP_PRESET = PresetUtils.makePresetEpWeights(
	'P1',
	Stats.fromMap({
		[Stat.StatIntellect]: 1.0,
		[Stat.StatSpirit]: 0.9,
		[Stat.StatSpellDamage]: 0.98,
	}),
);

// Default talents. Uses the wowhead calculator format, make the talents on
// https://www.wowhead.com/tbc/talent-calc/priest and copy the numbers in the url.
export const StandardTalents = {
	name: 'Standard',
	data: SavedTalents.create({
		talentsString: '',
	}),
};

export const DefaultOptions = Options.create({
	classOptions: {
		armor: PriestOptions_Armor.InnerFire,
	},
});

export const DefaultConsumables = ConsumesSpec.create({});

export const DefaultRaidBuffs = RaidBuffs.create({
	...defaultRaidBuffMajorDamageCooldowns(),
});

export const DefaultIndividualBuffs = IndividualBuffs.create({});

export const DefaultDebuffs = Debuffs.create({});

export const OtherDefaults = {
	channelClipDelay: 100,
	distanceFromTarget: 28,
	profession1: Profession.Engineering,
	profession2: Profession.Tailoring,
};
