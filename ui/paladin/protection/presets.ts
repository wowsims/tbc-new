import * as Mechanics from '../../core/constants/mechanics.js';
import * as PresetUtils from '../../core/preset_utils.js';
import { APLRotation_Type as APLRotationType } from '../../core/proto/apl.js';
import { ConsumesSpec, Glyphs, Profession, PseudoStat, Spec, Stat } from '../../core/proto/common.js';
import { PaladinMajorGlyph, PaladinMinorGlyph, PaladinSeal, ProtectionPaladin_Options as ProtectionPaladinOptions } from '../../core/proto/paladin.js';
import { SavedTalents } from '../../core/proto/ui.js';
import { Stats } from '../../core/proto_utils/stats';
import DefaultApl from './apls/default.apl.json';
import P1_Balanced_Gear from './gear_sets/p1-balanced.gear.json';
import P1_Offensive_Gear from './gear_sets/p1-offensive.gear.json';
import DefaultBuild from './builds/sha_default.build.json';
import ShaBuild from './builds/sha_encounter_only.build.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

export const P1_BALANCED_GEAR_PRESET = PresetUtils.makePresetGear('P1 BiS - Defensive', P1_Balanced_Gear);
export const P1_OFFENSIVE_GEAR_PRESET = PresetUtils.makePresetGear('P1 BiS - Offensive', P1_Offensive_Gear);

export const APL_PRESET = PresetUtils.makePresetAPLRotation('Sha of Fear', DefaultApl);

// Preset options for EP weights
export const P1_BALANCED_EP_PRESET = PresetUtils.makePresetEpWeights(
	'P1 Balanced',
	Stats.fromMap(
		{
			[Stat.StatStrength]: 1.00,
			[Stat.StatStamina]: 0.91,
			[Stat.StatHitRating]: 1.30,
			[Stat.StatCritRating]: 0.44,
			[Stat.StatHasteRating]: 0.89,
			[Stat.StatExpertiseRating]: 0.90,
			[Stat.StatDodgeRating]: 0.40,
			[Stat.StatParryRating]: 0.35,
			[Stat.StatMasteryRating]: 0.86,
			[Stat.StatArmor]: 0.48,
			[Stat.StatBonusArmor]: 0.48,
			[Stat.StatAttackPower]: 0.31,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 0.44,
		},
	),
);

// Default talents. Uses the wowhead calculator format, make the talents on
// https://wowhead.com/tbc/talent-calc and copy the numbers in the url.

export const DefaultTalents = {
	name: 'Default',
	data: SavedTalents.create({
		talentsString: '313213',
		glyphs: Glyphs.create({
			major1: PaladinMajorGlyph.GlyphOfFocusedShield,
			major2: PaladinMajorGlyph.GlyphOfTheAlabasterShield,
			major3: PaladinMajorGlyph.GlyphOfDivineProtection,

			minor1: PaladinMinorGlyph.GlyphOfFocusedWrath,
		}),
	}),
};

export const P1_BALANCED_BUILD_PRESET = PresetUtils.makePresetBuild('P1 Gear/EPs/Talents', {
	gear: P1_BALANCED_GEAR_PRESET,
	epWeights: P1_BALANCED_EP_PRESET,
	talents: DefaultTalents,
});
export const PRESET_BUILD_DEFAULT = PresetUtils.makePresetBuildFromJSON("Default", Spec.SpecProtectionPaladin, DefaultBuild);
export const PRESET_BUILD_SHA = PresetUtils.makePresetBuildFromJSON("Sha of Fear Model", Spec.SpecProtectionPaladin, ShaBuild);

export const DefaultOptions = ProtectionPaladinOptions.create({
	classOptions: {
		seal: PaladinSeal.Insight,
	},
});

export const DefaultConsumables = ConsumesSpec.create({
	flaskId: 76087, // Flask of the Earth
	foodId: 74656, // Chun Tian Spring Rolls
	potId: 76095, // Potion of Mogu Power
	prepotId: 76095, // Potion of Mogu Power
});

export const OtherDefaults = {
	profession1: Profession.Blacksmithing,
	profession2: Profession.Engineering,
	distanceFromTarget: 5,
	iterationCount: 25000,
};
