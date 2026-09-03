import * as PresetUtils from '../../core/preset_utils.js';
import {
	Class,
	ConsumesSpec,
	Debuffs,
	Drums,
	IndividualBuffs,
	PartyBuffs,
	Profession,
	Race,
	RaidBuffs,
	Stat,
	TristateEffect,
	UnitReference,
} from '../../core/proto/common.js';
import { BalanceDruid_Options as BalanceDruidOptions } from '../../core/proto/druid.js';
import { SavedTalents } from '../../core/proto/ui.js';
import { Stats } from '../../core/proto_utils/stats';
import { defaultRaidBuffMajorDamageCooldowns } from '../../core/proto_utils/utils';
import DefaultAPL from './apls/default.apl.json';
import PreraidGear from './gear_sets/preraid.gear.json';
import Phase1Gear from './gear_sets/p1_a.gear.json';
import Phase2Gear from './gear_sets/p2_a.gear.json';
import Phase3Gear from './gear_sets/p3.gear.json';
import Phase4Gear from './gear_sets/p4.gear.json';
import Phase5Gear from './gear_sets/p5.gear.json';

export const PreraidPresetGear = PresetUtils.makePresetGear('Pre-raid', PreraidGear);
export const Phase1PresetGear = PresetUtils.makePresetGear('Phase 1', Phase1Gear);
export const Phase2PresetGear = PresetUtils.makePresetGear('Phase 2', Phase2Gear);
export const Phase3PresetGear = PresetUtils.makePresetGear('Phase 3', Phase3Gear);
export const Phase4PresetGear = PresetUtils.makePresetGear('Phase 4', Phase4Gear);
export const Phase5PresetGear = PresetUtils.makePresetGear('Phase 5', Phase5Gear);

export const StandardRotation = PresetUtils.makePresetAPLRotation('Default', DefaultAPL);

// Pre-raid and Phase 1 weights are very close together, so in theory, both could be comined into one preset.
// But since all other phases have (just) distinct enough weights that they should be kept seperate,
// I decided keep pre-raid and P1 separate for consistency as well.

export const PreRaidEPWeights = PresetUtils.makePresetEpWeights(
	'Pre-raid',
	Stats.fromMap({
		[Stat.StatIntellect]: 0.61,
		[Stat.StatSpellDamage]: 1,
		[Stat.StatArcaneDamage]: 0.99,
		[Stat.StatNatureDamage]: 0.01,
		[Stat.StatSpellHitRating]: 1.76,
		[Stat.StatSpellCritRating]: 0.67,
		[Stat.StatSpellHasteRating]: 1.24,
		[Stat.StatSpirit]: 0.13,
		[Stat.StatMP5]: 0.05,
	}),
);

export const Phase1EPWeights = PresetUtils.makePresetEpWeights(
	'Phase 1',
	Stats.fromMap({
		[Stat.StatIntellect]: 0.65,
		[Stat.StatSpellDamage]: 1,
		[Stat.StatArcaneDamage]: 0.99,
		[Stat.StatNatureDamage]: 0.01,
		[Stat.StatSpellHitRating]: 1.85,
		[Stat.StatSpellCritRating]: 0.75,
		[Stat.StatSpellHasteRating]: 1.27,
		[Stat.StatSpirit]: 0.13,
		[Stat.StatMP5]: 0.06,
	}),
);

export const Phase2EPWeights = PresetUtils.makePresetEpWeights(
	'Phase 2',
	Stats.fromMap({
		[Stat.StatIntellect]: 0.56,
		[Stat.StatSpellDamage]: 1,
		[Stat.StatArcaneDamage]: 1,
		[Stat.StatNatureDamage]: 0.0,
		[Stat.StatSpellHitRating]: 1.86,
		[Stat.StatSpellCritRating]: 0.69,
		[Stat.StatSpellHasteRating]: 1.29,
		[Stat.StatSpirit]: 0.12,
		[Stat.StatMP5]: 0.04,
	}),
);

export const Phase3EPWeights = PresetUtils.makePresetEpWeights(
	'Phase 3',
	Stats.fromMap({
		[Stat.StatIntellect]: 0.57,
		[Stat.StatSpellDamage]: 1,
		[Stat.StatArcaneDamage]: 1,
		[Stat.StatNatureDamage]: 0,
		[Stat.StatSpellHitRating]: 1.91,
		[Stat.StatSpellCritRating]: 0.73,
		[Stat.StatSpellHasteRating]: 0.53,
		[Stat.StatSpirit]: 0.11,
		[Stat.StatMP5]: 0.02,
	}),
);

export const Phase3_5EPWeights = PresetUtils.makePresetEpWeights(
	'Phase 3.5',
	Stats.fromMap({
		[Stat.StatIntellect]: 0.58,
		[Stat.StatSpellDamage]: 1,
		[Stat.StatArcaneDamage]: 1,
		[Stat.StatNatureDamage]: 0,
		[Stat.StatSpellHitRating]: 1.46,
		[Stat.StatSpellCritRating]: 0.74,
		[Stat.StatSpellHasteRating]: 1.09,
		[Stat.StatSpirit]: 0.12,
		[Stat.StatMP5]: 0.05,
	}),
);

export const Phase4EPWeights = PresetUtils.makePresetEpWeights(
	'Phase 4',
	Stats.fromMap({
		[Stat.StatIntellect]: 0.59,
		[Stat.StatSpellDamage]: 1,
		[Stat.StatArcaneDamage]: 1,
		[Stat.StatNatureDamage]: 0,
		[Stat.StatSpellHitRating]: 2.03,
		[Stat.StatSpellCritRating]: 0.77,
		[Stat.StatSpellHasteRating]: 1.29,
		[Stat.StatSpirit]: 0.15,
		[Stat.StatMP5]: 0.11,
	}),
);

export const DefaultEPWeights = PresetUtils.makePresetEpWeights('Default (P2)', Phase2EPWeights.epWeights);
// Default talents. Uses the wowhead calculator format, make the talents on
// https://wowhead.com/tbc/talent-calc and copy the numbers in the url.
export const StandardTalents = {
	name: 'Standard',
	data: SavedTalents.create({
		talentsString: '510022312503135231351--520033',
	}),
};

export const DefaultOptions = BalanceDruidOptions.create({
	classOptions: {
		innervateTarget: UnitReference.create(),
	},
});

export const DefaultRaidBuffs = RaidBuffs.create({
	...defaultRaidBuffMajorDamageCooldowns(Class.ClassShaman),
	arcaneBrilliance: true,
	giftOfTheWild: TristateEffect.TristateEffectImproved,
	powerWordFortitude: TristateEffect.TristateEffectImproved,
	divineSpirit: TristateEffect.TristateEffectImproved,
});

export const DefaultPartyBuffs = PartyBuffs.create({
	chainOfTheTwilightOwl: true,
	draeneiRacialCaster: true,
	drums: Drums.LesserDrumsOfBattle,
	eyeOfTheNight: true,
	totemOfWrath: 1,
	wrathOfAirTotem: TristateEffect.TristateEffectImproved,
});

export const DefaultIndividualBuffs = IndividualBuffs.create({
	blessingOfKings: true,
	blessingOfWisdom: TristateEffect.TristateEffectImproved,
	shadowPriestDps: 800,
});

export const DefaultDebuffs = Debuffs.create({
	bloodFrenzy: true,
	curseOfElements: TristateEffect.TristateEffectImproved,
	curseOfRecklessness: true,
	exposeArmor: TristateEffect.TristateEffectImproved,
	giftOfArthas: true,
	huntersMark: TristateEffect.TristateEffectImproved,
	improvedSealOfTheCrusader: TristateEffect.TristateEffectImproved,
	judgementOfWisdom: true,
	mangle: true,
	misery: true,
	sunderArmor: true,
});

export const DefaultConsumables = ConsumesSpec.create({
	conjuredId: 12662, // Demonic Rune
	drumsId: Drums.LesserDrumsOfBattle,
	flaskId: 22861, // Flask of Blinding Light
	foodId: 27657, // Blackened Basilisk
	mhImbueId: 25122, // Brilliant Wizard Oil
	potId: 22832, // Super Mana Potion
});

export const OtherDefaults = {
	distanceFromTarget: 20,
	profession1: Profession.Enchanting,
	profession2: Profession.Tailoring,
	race: Race.RaceNightElf,
};
