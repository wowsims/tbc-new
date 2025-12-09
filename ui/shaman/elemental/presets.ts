import { Encounter } from '../../core/encounter';
import * as PresetUtils from '../../core/preset_utils.js';
import { Class, ConsumesSpec, Debuffs, Profession, Race, RaidBuffs, Stat } from '../../core/proto/common.js';
import {
	ElementalShaman_Options as ElementalShamanOptions,
	FeleAutocastSettings,
	ShamanShield,
} from '../../core/proto/shaman.js';
import { SavedTalents } from '../../core/proto/ui.js';
import { Stats } from '../../core/proto_utils/stats';
import { defaultRaidBuffMajorDamageCooldowns } from '../../core/proto_utils/utils';
import AoEApl from './apls/aoe.apl.json';
import CleaveApl from './apls/cleave.apl.json';
import DefaultApl from './apls/default.apl.json';
import P3Apl from './apls/p3.apl.json';
import P1Gear from './gear_sets/p1.gear.json';
import P2Gear from './gear_sets/p2.gear.json';
import P3Gear from './gear_sets/p3.gear.json';
import PreraidGear from './gear_sets/preraid.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

export const PRERAID_PRESET = PresetUtils.makePresetGear('Pre-raid', PreraidGear);
export const P1_PRESET = PresetUtils.makePresetGear('P1 - Default', P1Gear);
export const P2_PRESET = PresetUtils.makePresetGear('P2 - Default', P2Gear);
export const P3_PRESET = PresetUtils.makePresetGear('P3 - (WiP)', P3Gear);

export const ROTATION_PRESET_DEFAULT = PresetUtils.makePresetAPLRotation('Default', DefaultApl);
export const ROTATION_PRESET_P3 = PresetUtils.makePresetAPLRotation('P3 (WiP)', P3Apl);
export const ROTATION_PRESET_CLEAVE = PresetUtils.makePresetAPLRotation('Cleave', CleaveApl);
export const ROTATION_PRESET_AOE = PresetUtils.makePresetAPLRotation('AoE (3+)', AoEApl);

// Preset options for EP weights
export const EP_PRESET_DEFAULT = PresetUtils.makePresetEpWeights(
	'Default',
	Stats.fromMap({
		[Stat.StatIntellect]: 1.0,
		[Stat.StatSpellDamage]: 0.82,
		[Stat.StatSpirit]: 1.1,
	}),
);

export const EP_PRESET_P3 = PresetUtils.makePresetEpWeights(
	'P3 (WiP)',
	Stats.fromMap({
		[Stat.StatIntellect]: 1.0,
		[Stat.StatSpellDamage]: 0.82,
		[Stat.StatSpirit]: 1.25,
	}),
);

export const EP_PRESET_AOE = PresetUtils.makePresetEpWeights(
	'AoE (4+)',
	Stats.fromMap({
		[Stat.StatIntellect]: 1.0,
		[Stat.StatSpellDamage]: 0.74,
		[Stat.StatSpirit]: 1.18,
	}),
);

// Default talents. Uses the wowhead calculator format, make the talents on
// https://wowhead.com/tbc/talent-calc and copy the numbers in the url.
export const StandardTalents = {
	name: 'Default',
	data: SavedTalents.create({
		talentsString: '333121',
	}),
};

export const P3_TALENTS = {
	name: 'P3 (WiP)',
	data: SavedTalents.create({
		talentsString: '333322',
	}),
};

export const TalentsCleave = {
	name: 'Cleave',
	data: SavedTalents.create({
		talentsString: '333322',
	}),
};

export const TalentsAoE = {
	name: 'AoE (4+)',
	data: SavedTalents.create({
		...TalentsCleave.data,
	}),
};

export const DefaultOptions = ElementalShamanOptions.create({
	classOptions: {
		shield: ShamanShield.LightningShield,
		feleAutocast: FeleAutocastSettings.create({
			autocastFireblast: true,
			autocastFirenova: true,
			autocastImmolate: true,
			autocastEmpower: false,
		}),
	},
});

export const OtherDefaults = {
	distanceFromTarget: 20,
	profession1: Profession.Engineering,
	profession2: Profession.Tailoring,
	race: Race.RaceTroll,
};

export const DefaultRaidBuffs = RaidBuffs.create({
	...defaultRaidBuffMajorDamageCooldowns(Class.ClassShaman),
});

export const DefaultDebuffs = Debuffs.create({

});

export const DefaultConsumables = ConsumesSpec.create({
	flaskId: 76085, // Flask of the Warm Sun
	foodId: 74650, // Mogu Fish Stew
	potId: 76093, // Potion of the Jade Serpent
	prepotId: 76093, // Potion of the Jade Serpent
});

const ENCOUNTER_SINGLE_TARGET = PresetUtils.makePresetEncounter('Single Target Dummy', Encounter.defaultEncounterProto());
const ENCOUNTER_CLEAVE = PresetUtils.makePresetEncounter('Cleave', Encounter.defaultEncounterProto(2));
const ENCOUNTER_AOE = PresetUtils.makePresetEncounter('AOE (4+)', Encounter.defaultEncounterProto(4));

export const P1_PRESET_BUILD_DEFAULT = PresetUtils.makePresetBuild('Default', {
	talents: StandardTalents,
	rotation: ROTATION_PRESET_DEFAULT,
	encounter: ENCOUNTER_SINGLE_TARGET,
	epWeights: EP_PRESET_DEFAULT,
});

export const P1_PRESET_BUILD_CLEAVE = PresetUtils.makePresetBuild('Cleave', {
	talents: TalentsCleave,
	rotation: ROTATION_PRESET_CLEAVE,
	encounter: ENCOUNTER_CLEAVE,
	epWeights: EP_PRESET_DEFAULT,
});

export const P1_PRESET_BUILD_AOE = PresetUtils.makePresetBuild('AoE (4+)', {
	talents: TalentsAoE,
	rotation: ROTATION_PRESET_AOE,
	encounter: ENCOUNTER_AOE,
	epWeights: EP_PRESET_AOE,
});

export const P3_PRESET_BUILD_DEFAULT = PresetUtils.makePresetBuild('P3 (WiP)', {
	talents: P3_TALENTS,
	rotation: ROTATION_PRESET_P3,
	encounter: ENCOUNTER_SINGLE_TARGET,
	epWeights: EP_PRESET_P3,
	gear: P3_PRESET,
});
