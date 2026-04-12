import * as PresetUtils from '../../core/preset_utils.js';
import {
	ConsumesSpec,
	Debuffs,
	Drums,
	HealingModel,
	IndividualBuffs,
	PartyBuffs,
	Profession,
	PseudoStat,
	RaidBuffs,
	Spec,
	Stat,
	TristateEffect,
} from '../../core/proto/common.js';
import { ProtectionPaladin_Options as ProtectionPaladinOptions, ProtectionPaladin_Rotation as ProtectionPaladinRotation } from '../../core/proto/paladin.js';
import { SavedTalents } from '../../core/proto/ui.js';
import { Stats } from '../../core/proto_utils/stats';
import { defaultExposeWeaknessSettings } from '../../core/proto_utils/utils';
import DefaultApl from './apls/default.apl.json';
import P1_Gear from './gear_sets/p1.gear.json';
import P2_Gear from './gear_sets/p2.gear.json';
import P3_Gear from './gear_sets/p3.gear.json';
import P4_Gear from './gear_sets/p4.gear.json';
import P5_Gear from './gear_sets/p5.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

export const P1_GEAR_PRESET = PresetUtils.makePresetGear('P1', P1_Gear);
export const P2_GEAR_PRESET = PresetUtils.makePresetGear('P2', P2_Gear);
export const P3_GEAR_PRESET = PresetUtils.makePresetGear('P3', P3_Gear);
export const P4_GEAR_PRESET = PresetUtils.makePresetGear('P4', P4_Gear);
export const P5_GEAR_PRESET = PresetUtils.makePresetGear('P5', P5_Gear);

export const APL_PRESET = PresetUtils.makePresetAPLRotation('Default', DefaultApl);

export const DefaultSimpleRotation = ProtectionPaladinRotation.create({
	prioritizeHolyShield: true,
	useConsecrate: true,
	useExorcism: false,
	useAvengersShield: true,
	maintainJudgementOfWisdom: true,
});

export const APL_SIMPLE = PresetUtils.makePresetSimpleRotation('Simple', Spec.SpecProtectionPaladin, DefaultSimpleRotation);

// Preset options for EP weights
export const P4_EP_PRESET = PresetUtils.makePresetEpWeights(
	'P4',
	Stats.fromMap(
		{
			[Stat.StatStamina]: 1.5,
			[Stat.StatStrength]: 0.33,
			[Stat.StatSpellDamage]: 0.67,
			[Stat.StatAgility]: 0.6,
			[Stat.StatAttackPower]: 0.17,
			[Stat.StatMeleeHitRating]: 0.54,
			[Stat.StatMeleeHasteRating]: 0.18,
			[Stat.StatMeleeCritRating]: 0.32,
			[Stat.StatExpertiseRating]: 0.67,
			[Stat.StatDefenseRating]: 0.8,
			[Stat.StatDodgeRating]: 0.7,
			[Stat.StatParryRating]: 0.65,
			[Stat.StatArmor]: 0.05,
			[Stat.StatBonusArmor]: 0.05,
			[Stat.StatBlockRating]: 0.5,
			[Stat.StatBlockValue]: 0.4,
			[Stat.StatResilienceRating]: 0.3,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 0.5,
		},
	),
);

// Default talents. Uses the wowhead calculator format, make the talents on
// https://wowhead.com/tbc/talent-calc and copy the numbers in the url.

export const DefaultTalents = {
	name: 'Default',
	data: SavedTalents.create({
		talentsString: '-0530513050000142521051-052050003003',
	}),
};

export const DefaultOptions = ProtectionPaladinOptions.create({
	classOptions: {},
});

export const DefaultConsumables = ConsumesSpec.create({
	flaskId: 22861, // Flask of Blinding Light
	foodId: 27657, // Blackened Basilisk
	potId: 22849, // Ironshield Potion
	conjuredId: 12662, // Dark Rune
	mhImbueId: 28017,
	explosiveId: 30217,
	superSapper: true,
	goblinSapper: true,
	nightmareSeed: true,
	scrollStr: true,
	scrollAgi: true,
	scrollArm: true,
});

export const DefaultRaidBuffs = RaidBuffs.create({
	bloodlust: true,
	divineSpirit: TristateEffect.TristateEffectImproved,
	arcaneBrilliance: true,
	giftOfTheWild: TristateEffect.TristateEffectImproved,
	powerWordFortitude: TristateEffect.TristateEffectImproved,
	shadowProtection: true,
	thorns: TristateEffect.TristateEffectImproved,
});

export const DefaultPartyBuffs = PartyBuffs.create({
	manaSpringTotem: TristateEffect.TristateEffectRegular,
	wrathOfAirTotem: TristateEffect.TristateEffectRegular,
	graceOfAirTotem: TristateEffect.TristateEffectMissing,
	strengthOfEarthTotem: TristateEffect.TristateEffectImproved,
	windfuryTotem: TristateEffect.TristateEffectMissing,
	battleShout: TristateEffect.TristateEffectMissing,
	drums: Drums.LesserDrumsOfBattle,
	sanctityAura: TristateEffect.TristateEffectMissing,
});

export const DefaultIndividualBuffs = IndividualBuffs.create({
	blessingOfKings: true,
	blessingOfWisdom: TristateEffect.TristateEffectImproved,
	blessingOfMight: TristateEffect.TristateEffectImproved,
	blessingOfSanctuary: true,
});

export const DefaultDebuffs = Debuffs.create({
	misery: true,
	curseOfElements: TristateEffect.TristateEffectImproved,
	improvedSealOfTheCrusader: TristateEffect.TristateEffectImproved,
	judgementOfWisdom: true,
	judgementOfLight: true,
	bloodFrenzy: true,
	huntersMark: TristateEffect.TristateEffectImproved,
	curseOfRecklessness: true,
	sunderArmor: true,
	faerieFire: TristateEffect.TristateEffectImproved,
	exposeArmor: TristateEffect.TristateEffectImproved,
	insectSwarm: true,
	...defaultExposeWeaknessSettings(),
});

export const OtherDefaults = {
	profession1: Profession.Engineering,
	profession2: Profession.Enchanting,
	distanceFromTarget: 5,
	iterationCount: 25000,
	healingModel: HealingModel.create({
		hps: 2200,
		cadenceSeconds: 0.4,
		cadenceVariation: 1.2,
		absorbFrac: 0.02,
		burstWindow: 6,
		inspirationUptime: 0.25,
	}),
};
