import { OtherDefaults as SimUIOtherDefaults } from '../../core/individual_sim_ui';
import * as PresetUtils from '../../core/preset_utils.js';
import { ConsumesSpec, HealingModel, Profession, PseudoStat, Stat } from '../../core/proto/common.js';
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
			[Stat.StatStrength]: 0.61,
			[Stat.StatAgility]: 0.83,
			[Stat.StatStamina]: 1.15,
			[Stat.StatAttackPower]: 0.25,
			[Stat.StatMeleeHitRating]: 0.35,
			[Stat.StatMeleeCritRating]: 0.50,
			[Stat.StatMeleeHasteRating]: 0.41,
			[Stat.StatArmorPenetration]: 0.09,
			[Stat.StatExpertiseRating]: 2.01,
			[Stat.StatDefenseRating]: 0.41,
			[Stat.StatBlockRating]: 0.01,
			[Stat.StatBlockValue]: 0.57,
			[Stat.StatParryRating]: 0.51,
			[Stat.StatResilienceRating]: 0.02,
			[Stat.StatArmor]: 0.06,
			[Stat.StatBonusArmor]: 0.06,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 3.15,
		},
	),
);

// Default talents. Uses the wowhead calculator format, make the talents on
// https://wowhead.com/tbc/talent-calc and copy the numbers in the url.
export const DefaultTalents = {
	name: 'Default',
	data: SavedTalents.create({
		talentsString: '35000301302-03-0055511033001101501351',
	}),
};

export const DefaultOptions = ProtectionWarriorOptions.create({
	classOptions: {
		queueDelay: 250,
		startingRage: 100,
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

export const OtherDefaults: Partial<SimUIOtherDefaults> = {
	profession1: Profession.Engineering,
	profession2: Profession.Blacksmithing,
	distanceFromTarget: 0,
	healingModel: HealingModel.create({
		hps: 2200,
		cadenceSeconds: 0.4,
		cadenceVariation: 1.2,
		absorbFrac: 0.02,
		burstWindow: 6,
		inspirationUptime: 0.25,
	}),
};

export const P1_PRESET_BUILD = PresetUtils.makePresetBuild('P1', {
	gear: P1_PRESET,
	talents: DefaultTalents,
	epWeights: P1_EP_PRESET,
	rotation: ROTATION_DEFAULT,
});
