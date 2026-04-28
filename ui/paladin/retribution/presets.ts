import { DrumsBattle } from '../../core/components/inputs/consumables';
import { RetributionPaladin } from '../../core/player_specs/paladin';
import * as PresetUtils from '../../core/preset_utils.js';
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
} from '../../core/proto/common.js';
import { RetributionPaladin_Options as RetributionPaladinOptions, RetributionPaladin_Rotation as PaladinRotation } from '../../core/proto/paladin.js';
import { SavedTalents } from '../../core/proto/ui.js';
import { Stats } from '../../core/proto_utils/stats';
import DefaultApl from './apls/default.apl.json';
import P1_Gear from './gear_sets/p1.gear.json';
import P2_Gear from './gear_sets/p2.gear.json';
import Preraid_Gear from './gear_sets/preraid.gear.json';
import { defaultExposeWeaknessSettings } from '../../core/proto_utils/utils';
import { Phase } from '../../core/constants/other';

export const P1_GEAR_PRESET = PresetUtils.makePresetGear('P1', P1_Gear, { phase: Phase.Phase1 });
export const P2_GEAR_PRESET = PresetUtils.makePresetGear('P2', P2_Gear, { phase: Phase.Phase2 });
export const PRERAID_GEAR_PRESET = PresetUtils.makePresetGear('Pre-raid', Preraid_Gear, { phase: Phase.Phase1 });

export const DefaultSimpleRotation = PaladinRotation.create({
	useExorcism: false,
	useConsecrate: false,
	delayMajorCDs: 11,
	prepullSotC: true,
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

export const P2_EP_PRESET = PresetUtils.makePresetEpWeights(
	'P2',
	Stats.fromMap(
		{
			[Stat.StatStrength]: 1.0,
			[Stat.StatAgility]: 0.75,
			[Stat.StatAttackPower]: 0.41,
			[Stat.StatMeleeHitRating]: 2.15,
			[Stat.StatMeleeCritRating]: 0.77,
			[Stat.StatMeleeHasteRating]: 1.17,
			[Stat.StatArmorPenetration]: 0.1,
			[Stat.StatExpertiseRating]: 2.14,
			[Stat.StatSpellDamage]: 0.17,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 5.34,
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
	divineSpirit: 2,
	arcaneBrilliance: true,
	giftOfTheWild: 2,
	powerWordFortitude: 2,
	shadowProtection: true,
	thorns: 2,
});

export const DefaultPartyBuffs = PartyBuffs.create({
	manaSpringTotem: 1,
	leaderOfThePack: 2,
	battleShout: 2,
	strengthOfEarthTotem: 2,
	windfuryTotem: 2,
	graceOfAirTotem: 2,
	drums: Drums.LesserDrumsOfBattle,
	sanctityAura: 2,
	totemTwisting: true,
});

export const DefaultIndividualBuffs = IndividualBuffs.create({
	blessingOfKings: true,
	blessingOfWisdom: 2,
	blessingOfMight: 2,
	unleashedRage: true,
});

export const DefaultDebuffs = Debuffs.create({
	misery: true,
	curseOfElements: 2,
	improvedSealOfTheCrusader: TristateEffect.TristateEffectImproved,
	jocRetribution2Pt4: true,
	judgementOfWisdom: true,
	bloodFrenzy: true,
	huntersMark: 2,
	curseOfRecklessness: true,
	sunderArmor: true,
	faerieFire: 2,
	exposeArmor: 2,
	...defaultExposeWeaknessSettings(),
});

export const OtherDefaults = {
	profession1: Profession.Engineering,
	profession2: Profession.Blacksmithing,
	distanceFromTarget: 5,
	iterationCount: 25000,
	race: Race.RaceBloodElf,
};
