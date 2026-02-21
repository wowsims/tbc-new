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
import DefaultApl from './apls/default.apl.json';
import P1Gear from './gear_sets/p1.gear.json';
import P2Gear from './gear_sets/p2.gear.json';
import P3Gear from './gear_sets/p3.gear.json';
import P3_5Gear from './gear_sets/p3_5.gear.json';
import P4Gear from './gear_sets/p4.gear.json';
import PreraidGear from './gear_sets/preraid.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

export const PRERAID_PRESET = PresetUtils.makePresetGear('Pre-Raid', PreraidGear);
export const P1_PRESET = PresetUtils.makePresetGear('Phase 1', P1Gear);
export const P2_PRESET = PresetUtils.makePresetGear('Phase 2', P2Gear);
export const P3_PRESET = PresetUtils.makePresetGear('Phase 3', P3Gear);
export const P3_5_PRESET = PresetUtils.makePresetGear('Phase 3.5', P3_5Gear);
export const P4_PRESET = PresetUtils.makePresetGear('Phase 4', P4Gear);

export const ROTATION_PRESET_DEFAULT = PresetUtils.makePresetAPLRotation('Default', DefaultApl);

// Preset options for EP weights
export const EP_PRESET_DEFAULT = PresetUtils.makePresetEpWeights(
	'Default',
	Stats.fromMap({
		[Stat.StatIntellect]: 0.33,
		[Stat.StatSpellDamage]: 1.0,
		[Stat.StatNatureDamage]: 1.0,
		[Stat.StatSpellCritRating]: 0.78,
		[Stat.StatSpellHasteRating]: 1.22,
		[Stat.StatSpellHitRating]: 0.33,
		[Stat.StatMP5]: 0.08,
	}),
);

// https://wowhead.com/tbc/talent-calc and copy the numbers in the url.
export const StandardTalents = {
	name: 'Default',
	data: SavedTalents.create({
		talentsString: '55003105100213351051--05105301005',
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
});

const ENCOUNTER_SINGLE_TARGET = PresetUtils.makePresetEncounter('Single Target Dummy', Encounter.defaultEncounterProto());

export const P1_PRESET_BUILD_DEFAULT = PresetUtils.makePresetBuild('Default', {
	talents: StandardTalents,
	rotation: ROTATION_PRESET_DEFAULT,
	encounter: ENCOUNTER_SINGLE_TARGET,
	epWeights: EP_PRESET_DEFAULT,
});
