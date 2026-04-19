import { OtherDefaults as SimUIOtherDefaults } from '../../core/individual_sim_ui';
import * as PresetUtils from '../../core/preset_utils.js';
import { ConsumesSpec, HealingModel, Profession, PseudoStat, Race, Spec, Stat } from '../../core/proto/common.js';
import { SavedTalents } from '../../core/proto/ui.js';
import { ProtectionWarrior_Options as ProtectionWarriorOptions, WarriorShout, WarriorStance } from '../../core/proto/warrior.js';
import { Stats } from '../../core/proto_utils/stats';
import * as WarriorPresets from '../presets';
import GenericApl from './apls/default.apl.json';
import P1BISGear from './gear_sets/p1_bis.gear.json';
import P2BISGear from './gear_sets/p2_bis.gear.json';
import P2HydrossGear from './gear_sets/p2_hydross.gear.json';
import P3BISGear from './gear_sets/p3_bis.gear.json';
import P35BISGear from './gear_sets/p35_bis.gear.json';
import P4BISGear from './gear_sets/p4_bis.gear.json';
import PreraidBISGear from './gear_sets/preraid.gear.json';
import DefaultBuild from './builds/default_encounter_only.build.json';
import MagtheridonBuild from './builds/magtheridon_encounter_only.build.json';
import KarazhanBuild from './builds/karazhan_encounter_only.build.json';
import MorogrimBuild from './builds/morogrim_encounter_only.build.json';
import HydrossBuild from './builds/hydross_encounter_only.build.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

export const PRERAID_BALANCED_PRESET = PresetUtils.makePresetGear('Pre-raid', PreraidBISGear, { group: 'Default' });
export const P1_PRESET = PresetUtils.makePresetGear('P1 - BIS', P1BISGear, { group: 'Default' });
export const P2_PRESET = PresetUtils.makePresetGear('P2 - BIS', P2BISGear, { group: 'Default' });
export const P2_HYDROSS_PRESET = PresetUtils.makePresetGear('P2 - Hydross (Frost Resist)', P2HydrossGear, { group: 'Encounter specific' });
export const P3_PRESET = PresetUtils.makePresetGear('P3 - BIS', P3BISGear, { group: 'Default' });
export const P35_PRESET = PresetUtils.makePresetGear('P3.5 - BIS', P35BISGear, { group: 'Default' });
export const P4_PRESET = PresetUtils.makePresetGear('P4 - BIS', P4BISGear, { group: 'Default' });

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
			[Stat.StatMeleeCritRating]: 0.5,
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
		hasBsT2: true,
		stanceSnapshot: true,
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
	race: Race.RaceOrc,
	distanceFromTarget: 0,
	healingModel: HealingModel.create({
		hps: 2200,
		cadenceSeconds: 0.4,
		cadenceVariation: 1.2,
		absorbFrac: 0.02,
		burstWindow: 6,
		inspirationUptime: 0.25,
	}),
	// Morogrim
	// healingModel: HealingModel.create({
	// 	hps: 3300,
	// 	cadenceSeconds: 1.5,
	// 	cadenceVariation: 1.0,
	// 	absorbFrac: 0.02,
	// 	burstWindow: 6,
	// 	inspirationUptime: 0.12,
	// }),
};

export const P1_PRESET_BUILD = PresetUtils.makePresetBuild('P1', {
	gear: P1_PRESET,
	talents: DefaultTalents,
	epWeights: P1_EP_PRESET,
	rotation: ROTATION_DEFAULT,
});

export const DEFAULT_PRESET_BUILD = PresetUtils.makePresetBuildFromJSON('Default', Spec.SpecProtectionWarrior, DefaultBuild);
export const MAGTHERIDON_PRESET_BUILD = PresetUtils.makePresetBuildFromJSON('Magtheridon', Spec.SpecProtectionWarrior, MagtheridonBuild);
export const KARAZHAN_PRESET_BUILD = PresetUtils.makePresetBuildFromJSON('Karazhan (Boss Average)', Spec.SpecProtectionWarrior, KarazhanBuild);
export const MOROGRIM_PRESET_BUILD = PresetUtils.makePresetBuildFromJSON('Morogrim', Spec.SpecProtectionWarrior, MorogrimBuild);
export const HYDROSS_PRESET_BUILD = PresetUtils.makePresetBuildFromJSON('Hydross', Spec.SpecProtectionWarrior, HydrossBuild);
