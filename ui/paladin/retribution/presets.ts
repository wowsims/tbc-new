import * as PresetUtils from '../../core/preset_utils';
import {
	ConsumesSpec,
	Debuffs,
	RaidBuffs,
	Profession,
	PseudoStat,
	PartyBuffs,
	IndividualBuffs,
	TristateEffect,
	Race,
	Stat,
	Spec,
	Drums,
} from '../../core/proto/common';
import {
	PaladinAura,
	RetributionPaladin_Options as RetributionPaladinOptions,
	RetributionPaladin_Rotation as PaladinRotation,
} from '../../core/proto/paladin';
import { SavedTalents } from '../../core/proto/ui';
import { Stats } from '../../core/proto_utils/stats';
import DefaultApl from './apls/default.apl.json';
import P1_Gear from './gear_sets/p1.gear.json';
import Preraid_Gear from './gear_sets/preraid.gear.json';
import { defaultExposeWeaknessSettings } from '../../core/proto_utils/utils';

export const P1_GEAR_PRESET = PresetUtils.makePresetGear('P1', P1_Gear);
export const PRERAID_GEAR_PRESET = PresetUtils.makePresetGear('Pre-raid', Preraid_Gear);

export const DefaultSimpleRotation = PaladinRotation.create({
	useExorcism: false,
	consecrationRank: 0,
	delayMajorCDs: 11,
	prepullSotC: true,
	aura: PaladinAura.SanctityAura,
});

export const APL_PRESET = PresetUtils.makePresetAPLRotation('Default', DefaultApl);
export const APL_SIMPLE = PresetUtils.makePresetSimpleRotation('Simple', Spec.SpecRetributionPaladin, DefaultSimpleRotation);

export const P1_EP_PRESET = PresetUtils.makePresetEpWeights(
	'P1',
	Stats.fromMap(
		{
			[Stat.StatStrength]: 1.0,
			[Stat.StatAgility]: 0.75,
			[Stat.StatAttackPower]: 0.41,
			[Stat.StatMeleeHitRating]: 2.19,
			[Stat.StatMeleeCritRating]: 0.77,
			[Stat.StatMeleeHasteRating]: 1.3,
			[Stat.StatArmorPenetration]: 0.1,
			[Stat.StatExpertiseRating]: 2.18,
			[Stat.StatSpellDamage]: 0.14,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 5.88,
		},
	),
);

export const DefaultTalents = {
	name: 'Default',
	data: SavedTalents.create({
		talentsString: '5-053201-0523005120033125331051',
	}),
};

export const NoKingsTalents = {
	name: 'No Kings',
	data: SavedTalents.create({
		talentsString: '5-0532-0523005130033125331051',
	}),
};

export const ImpMightTalents = {
	name: 'Imp Might',
	data: SavedTalents.create({
		talentsString: '5-053201-5023005120033125331051',
	}),
};

export const DefaultOptions = RetributionPaladinOptions.create({
	classOptions: {},
});

export const DefaultConsumables = ConsumesSpec.create({
	potId: 22838,
	flaskId: 22854,
	foodId: 27658,
	conjuredId: 12662,
	superSapper: true,
	goblinSapper: true,
	scrollAgi: true,
	scrollStr: true,
	explosiveId: 30217,
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
	leaderOfThePack: TristateEffect.TristateEffectImproved,
	battleShout: TristateEffect.TristateEffectImproved,
	strengthOfEarthTotem: TristateEffect.TristateEffectImproved,
	totemTwisting: true,
	windfuryTotem: TristateEffect.TristateEffectImproved,
	graceOfAirTotem: TristateEffect.TristateEffectImproved,
	drums: Drums.LesserDrumsOfBattle,
	sanctityAura: TristateEffect.TristateEffectImproved,
});

export const DefaultIndividualBuffs = IndividualBuffs.create({
	blessingOfKings: true,
	blessingOfWisdom: TristateEffect.TristateEffectImproved,
	blessingOfMight: TristateEffect.TristateEffectImproved,
	unleashedRage: true,
});

export const DefaultDebuffs = Debuffs.create({
	misery: true,
	curseOfElements: TristateEffect.TristateEffectImproved,
	improvedSealOfTheCrusader: TristateEffect.TristateEffectImproved,
	jocRetribution2Pt4: true,
	judgementOfWisdom: true,
	bloodFrenzy: true,
	huntersMark: TristateEffect.TristateEffectImproved,
	curseOfRecklessness: true,
	sunderArmor: true,
	faerieFire: TristateEffect.TristateEffectImproved,
	exposeArmor: TristateEffect.TristateEffectImproved,
	...defaultExposeWeaknessSettings(),
});

export const OtherDefaults = {
	profession1: Profession.Engineering,
	profession2: Profession.Blacksmithing,
	distanceFromTarget: 5,
	iterationCount: 25000,
	race: Race.RaceBloodElf,
};
