import * as PresetUtils from '../../core/preset_utils.js';
import { ConsumesSpec, Profession, PseudoStat, Stat } from '../../core/proto/common.js';
import { SavedTalents } from '../../core/proto/ui.js';
import { ProtectionWarrior_Options as ProtectionWarriorOptions, WarriorShout, WarriorStance } from '../../core/proto/warrior.js';
import { Stats } from '../../core/proto_utils/stats';
import * as WarriorPresets from '../presets';
import GenericApl from './apls/default.apl.json';
import P1BISGear from './gear_sets/p1_bis.gear.json';
import PreraidBISGear from './gear_sets/preraid.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

export const PRERAID_BALANCED_PRESET = PresetUtils.makePresetGear('Pre-raid', PreraidBISGear);
export const P1_PRESET = PresetUtils.makePresetGear('P1 - BIS', P1BISGear);

export const ROTATION_DEFAULT = PresetUtils.makePresetAPLRotation('Generic', GenericApl);

// Preset options for EP weights
export const P1_EP_PRESET = PresetUtils.makePresetEpWeights(
	'P1 - Default',
	Stats.fromMap(
		{
			[Stat.StatStrength]: 1,
			[Stat.StatStamina]: 1.07,
			[Stat.StatAttackPower]: 0.33,
			[Stat.StatArmor]: 0.55,
			[Stat.StatBonusArmor]: 0.55,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 0.96,
		},
	),
);

// Default talents. Uses the wowhead calculator format, make the talents on
// https://wowhead.com/tbc/talent-calc and copy the numbers in the url.
export const StandardTalents = {
	name: 'Standard',
	data: SavedTalents.create({
		talentsString: '350003011-05-0055511033001103501351',
	}),
};

export const DefaultOptions = ProtectionWarriorOptions.create({
	classOptions: {
		queueDelay: 250,
		startingRage: 0,
		defaultShout: WarriorShout.WarriorShoutCommanding,
		defaultStance: WarriorStance.WarriorStanceDefensive,
	},
});

export const DefaultConsumables = ConsumesSpec.create({
	...WarriorPresets.DefaultConsumables,
	flaskId: 22854,
	potId: 22828,
});

export const OtherDefaults = {
	profession1: Profession.Engineering,
	profession2: Profession.Blacksmithing,
	distanceFromTarget: 25,
};
