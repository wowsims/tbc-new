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
			[Stat.StatStamina]: 1.0,
			[Stat.StatStrength]: 1.0,
			[Stat.StatAgility]: 0.65,
			[Stat.StatAttackPower]: 0.46,
			[Stat.StatMeleeHitRating]: 0.57,
			[Stat.StatMeleeCritRating]: 0.88,
			[Stat.StatMeleeHasteRating]: 0.9,
			[Stat.StatArmorPenetration]: 0.15,
			[Stat.StatExpertiseRating]: 0.99,
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
	conjuredId: 22105,
	foodId: 27667,
	flaskId: undefined,
	battleElixirId: 22831,
	guardianElixirId: 9088,
	potId: 22849,
	nightmareSeed: true,
	scrollStr: true,
	scrollAgi: true,
	scrollArm: true,
});

export const OtherDefaults = {
	profession1: Profession.Engineering,
	profession2: Profession.Blacksmithing,
	distanceFromTarget: 0,
};
