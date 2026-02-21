import * as PresetUtils from '../../core/preset_utils.js';
import { ConsumesSpec, Debuffs, IndividualBuffs, PartyBuffs, Profession, PseudoStat, RaidBuffs, Stat, UnitReference } from '../../core/proto/common.js';
import { BalanceDruid_Options as BalanceDruidOptions } from '../../core/proto/druid.js';
import { SavedTalents } from '../../core/proto/ui.js';
import { Stats, UnitStat, UnitStatPresets } from '../../core/proto_utils/stats';
import { defaultRaidBuffMajorDamageCooldowns } from '../../core/proto_utils/utils';
import StandardApl from './apls/standard.apl.json';
import PreraidGear from './gear_sets/preraid.gear.json';
import Phase1 from './gear_sets/p1.gear.json';

export const PreraidPresetGear = PresetUtils.makePresetGear('Pre-raid', PreraidGear);
export const Phase1PresetGear = PresetUtils.makePresetGear('P1', Phase1);

export const StandardRotation = PresetUtils.makePresetAPLRotation('Standard', StandardApl);

export const StandardEPWeights = PresetUtils.makePresetEpWeights(
	'Standard',
	Stats.fromMap({
		[Stat.StatIntellect]: 1,
		[Stat.StatSpirit]: 1,
		[Stat.StatSpellDamage]: 1,
		[Stat.StatNatureDamage]: 1,
		[Stat.StatArcaneDamage]: 1,
		[Stat.StatSpellHitRating]: 1,
		[Stat.StatSpellCritRating]: 1,
		[Stat.StatSpellHasteRating]: 1,
		[Stat.StatSpellPenetration]: 1,
		[Stat.StatMana]: 1,
	}),
);

// Default talents. Uses the wowhead calculator format, make the talents on
// https://wowhead.com/tbc/talent-calc and copy the numbers in the url.
export const StandardTalents = {
	name: 'Standard',
	data: SavedTalents.create({
		talentsString: '',
	}),
};

export const DefaultOptions = BalanceDruidOptions.create({
	classOptions: {
		innervateTarget: UnitReference.create(),
	},
});

export const DefaultConsumables = ConsumesSpec.create({});

export const DefaultRaidBuffs = RaidBuffs.create({
	...defaultRaidBuffMajorDamageCooldowns(),
});

export const DefaultIndividualBuffs = IndividualBuffs.create({});

export const DefaultPartyBuffs = PartyBuffs.create({});

export const DefaultDebuffs = Debuffs.create({});

export const OtherDefaults = {
	distanceFromTarget: 20,
	profession1: Profession.Engineering,
	profession2: Profession.Tailoring,
};

export const PresetPreraidBuild = PresetUtils.makePresetBuild('Pre-raid', {
	gear: PreraidPresetGear,
	talents: StandardTalents,
	rotation: StandardRotation,
	epWeights: StandardEPWeights,
});

export const Phase1PresetBuild = PresetUtils.makePresetBuild('P1', {
	gear: Phase1PresetGear,
	talents: StandardTalents,
	rotation: StandardRotation,
	epWeights: StandardEPWeights,
});
