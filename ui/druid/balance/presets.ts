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
import Phase1AllianceGear from './gear_sets/p1_a.gear.json';
import Phase2AllianceGear from './gear_sets/p2_a.gear.json';
import Phase3Gear from './gear_sets/p3.gear.json';
import Phase3_5Gear from './gear_sets/p3_5.gear.json';
import Phase4Gear from './gear_sets/p4.gear.json';

export const PreraidPresetGear = PresetUtils.makePresetGear('Pre-raid', PreraidGear);
export const Phase1AlliancePresetGear = PresetUtils.makePresetGear('Phase 1 (A)', Phase1AllianceGear);
export const Phase2AlliancePresetGear = PresetUtils.makePresetGear('Phase 2 (A)', Phase2AllianceGear);
export const Phase3PresetGear = PresetUtils.makePresetGear('Phase 3', Phase3Gear);
export const Phase3_5PresetGear = PresetUtils.makePresetGear('Phase 3.5', Phase3_5Gear);
export const Phase4PresetGear = PresetUtils.makePresetGear('Phase 4', Phase4Gear);

export const StandardRotation = PresetUtils.makePresetAPLRotation('Default', DefaultAPL);

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
