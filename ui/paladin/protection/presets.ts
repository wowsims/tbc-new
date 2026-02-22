import * as PresetUtils from '../../core/preset_utils.js';
import { ConsumesSpec, Profession, PseudoStat, Stat } from '../../core/proto/common.js';
import { PaladinSeal, ProtectionPaladin_Options as ProtectionPaladinOptions } from '../../core/proto/paladin.js';
import { SavedTalents } from '../../core/proto/ui.js';
import { Stats } from '../../core/proto_utils/stats';
import DefaultApl from './apls/default.apl.json';
import P1_Balanced_Gear from './gear_sets/p1-balanced.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

export const P1_BALANCED_GEAR_PRESET = PresetUtils.makePresetGear('P1', P1_Balanced_Gear);

export const APL_PRESET = PresetUtils.makePresetAPLRotation('Sha of Fear', DefaultApl);

// Preset options for EP weights
export const P1_BALANCED_EP_PRESET = PresetUtils.makePresetEpWeights(
	'P1',
	Stats.fromMap(
		{
			[Stat.StatStamina]: 1.0,
			[Stat.StatStrength]: 1.0,
			[Stat.StatSpellDamage]: 1.0,
			[Stat.StatAgility]: 1.0,
			[Stat.StatAttackPower]: 1.0,
			[Stat.StatMeleeHitRating]: 1.0,
			[Stat.StatMeleeHasteRating]: 1.0,
			[Stat.StatMeleeCritRating]: 1.0,
			[Stat.StatArmorPenetration]: 1.0,
			[Stat.StatExpertiseRating]: 1.0,
			[Stat.StatResilienceRating]: 1.0,
			[Stat.StatDefenseRating]: 1.0,
			[Stat.StatDodgeRating]: 1.0,
			[Stat.StatParryRating]: 1.0,
			[Stat.StatArmor]: 1.0,
			[Stat.StatBonusArmor]: 1.0,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 1.0,
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

export const P1_BALANCED_BUILD_PRESET = PresetUtils.makePresetBuild('P1 Gear/EPs/Talents', {
	gear: P1_BALANCED_GEAR_PRESET,
	epWeights: P1_BALANCED_EP_PRESET,
	talents: DefaultTalents,
});

export const DefaultOptions = ProtectionPaladinOptions.create({
	classOptions: {
		seal: PaladinSeal.Insight,
	},
});

export const DefaultConsumables = ConsumesSpec.create({});

export const OtherDefaults = {
	profession1: Profession.Blacksmithing,
	profession2: Profession.Engineering,
	distanceFromTarget: 5,
	iterationCount: 25000,
};
