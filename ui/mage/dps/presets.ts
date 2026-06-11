import * as PresetUtils from '../../core/preset_utils';
import {
	Debuffs,
	PseudoStat,
	RaidBuffs,
	Stat,
	ConsumesSpec,
	PartyBuffs,
	IndividualBuffs,
	TristateEffect,
	Profession,
	Drums,
	Spec,
} from '../../core/proto/common';
import { defaultImprovedShadowBoltSettings } from '../../core/proto_utils/utils';
import { Stats } from '../../core/proto_utils/stats';
import { SavedTalents } from '../../core/proto/ui';
import { Mage_Rotation, MageArmor, Mage_Options as MageOptions } from '../../core/proto/mage';
import BlankAPL from './apls/blank.apl.json';
import BlankGear from './gear_sets/blank.gear.json';
import ArcaneApl from './apls/arcane.apl.json';
import PreBISArcaneGear from './gear_sets/preBisArcane.gear.json';
import P1BISArcaneGear from './gear_sets/p1Arcane.gear.json';
import P2BISArcaneGear from './gear_sets/p2Arcane.gear.json';
import { Phase } from '../../core/constants/other';
import { APLRotation_Type } from '../../core/proto/apl';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

export const BLANK_APL = PresetUtils.makePresetAPLRotation('Blank', BlankAPL);
export const PREBIS_ARCANE = PresetUtils.makePresetGear('Arcane PreRaid - BIS', PreBISArcaneGear, { phase: Phase.Phase1 });
export const P1_BIS_ARCANE = PresetUtils.makePresetGear('Arcane - BIS', P1BISArcaneGear, { phase: Phase.Phase1 });
export const P2_BIS_ARCANE = PresetUtils.makePresetGear('Arcane - BIS', P2BISArcaneGear, { phase: Phase.Phase2 });
//export const P3_BIS_ARCANE = PresetUtils.makePresetGear('Arcane P3 - BIS', P3BISArcaneGear);

export const ARCANE_TALENTS = PresetUtils.makePresetTalents('Arcane', SavedTalents.create({ talentsString: '2500052300030150330125--053500031003001' }));
export const ROTATION_PRESET_ARCANE = PresetUtils.makePresetAPLRotation('Arcane', ArcaneApl);
export const BLANK_GEARSET = PresetUtils.makePresetGear('Blank', BlankGear);

export const ArcaneMageSimpleRotation = Mage_Rotation.create({
	conserveStart: 20,
	conserveEnd: 30,
	delayMajorCDs: 10,
});

export const APL_ARCANE_SIMPLE = PresetUtils.makePresetSimpleRotation('Arcane Simple', Spec.SpecMage, ArcaneMageSimpleRotation);

// Preset options for EP weights
export const P1_EP_PRESET = PresetUtils.makePresetEpWeights(
	'P1 - Arcane',
	Stats.fromMap(
		{
			[Stat.StatMana]: 0.03,
			[Stat.StatIntellect]: 1.52,
			[Stat.StatSpirit]: 1,
			[Stat.StatSpellDamage]: 1,
			[Stat.StatArcaneDamage]: 0.92,
			[Stat.StatFrostDamage]: 0.08,
			[Stat.StatSpellHitRating]: 2.36,
			[Stat.StatSpellCritRating]: 0.83,
			[Stat.StatSpellHasteRating]: 0.53,
			[Stat.StatSpellPenetration]: 0,
			[Stat.StatMP5]: 0.56,
		},
		{
			[PseudoStat.PseudoStatSchoolHitPercentArcane]: 2.14,
			[PseudoStat.PseudoStatSchoolHitPercentFrost]: 0.15,
		},
	),
);

export const P2_EP_PRESET = PresetUtils.makePresetEpWeights(
	'P2 - Arcane',
	Stats.fromMap(
		{
			[Stat.StatMana]: 0.03,
			[Stat.StatIntellect]: 1.31,
			[Stat.StatSpirit]: 0.9,
			[Stat.StatSpellDamage]: 1,
			[Stat.StatArcaneDamage]: 0.9,
			[Stat.StatFrostDamage]: 0.1,
			[Stat.StatSpellHitRating]: 2.3,
			[Stat.StatSpellCritRating]: 0.77,
			[Stat.StatSpellHasteRating]: 0.55,
			[Stat.StatSpellPenetration]: 0,
			[Stat.StatMP5]: 0.48,
		},
		{
			[PseudoStat.PseudoStatSchoolHitPercentArcane]: 2.09,
			[PseudoStat.PseudoStatSchoolHitPercentFrost]: 0.2,
		},
	),
);

export const Talents = {
	name: 'Blank',
	data: SavedTalents.create({
		talentsString: '',
	}),
};

export const DefaultOptions = MageOptions.create({
	classOptions: {
		defaultMageArmor: MageArmor.MageArmorMageArmor,
	},
});

export const OtherDefaults = {
	distanceFromTarget: 20,
	profession1: Profession.Engineering,
	profession2: Profession.Tailoring,
};

export const DefaultConsumables = ConsumesSpec.create({
	guardianElixirId: 32067, // Elixir of Draenic Wisdom
	battleElixirId: 28103, // Adept's Elixir
	foodId: 27657, // Blackened Basilisk
	mhImbueId: 25122, // Brilliant Wizard Oil
	potId: 22839, // Destruction Potion
});

export const DefaultRaidBuffs = RaidBuffs.create({
	bloodlust: true,
	divineSpirit: 2,
	arcaneBrilliance: true,
	giftOfTheWild: 2,
	powerWordFortitude: 2,
	shadowProtection: true,
});

export const DefaultPartyBuffs = PartyBuffs.create({
	manaSpringTotem: 2,
	manaTideTotems: 1,
	wrathOfAirTotem: 1,
	drums: Drums.LesserDrumsOfBattle,
});

export const DefaultIndividualBuffs = IndividualBuffs.create({
	blessingOfKings: true,
	blessingOfWisdom: 2,
	innervates: 1,
	powerInfusions: 1,
	shadowPriestDps: 1400,
});

export const DefaultDebuffs = Debuffs.create({
	misery: true,
	curseOfElements: 2,
	improvedSealOfTheCrusader: TristateEffect.TristateEffectImproved,
	judgementOfWisdom: true,
	...defaultImprovedShadowBoltSettings(),
});

export const P1_PLAYER_SETTINGS: PresetUtils.PresetSettings = {
	name: 'P1',
	playerOptions: OtherDefaults,
	debuffs: Debuffs.create({
		...DefaultDebuffs,
	}),
	reforgeSettings: {
		maxGemPhase: Phase.Phase1,
	},
};

export const P2_PLAYER_SETTINGS: PresetUtils.PresetSettings = {
	name: 'P2',
	playerOptions: OtherDefaults,
	partyBuffs: PartyBuffs.create({
		...DefaultPartyBuffs,
	}),
	debuffs: Debuffs.create({
		...DefaultDebuffs,
	}),
	reforgeSettings: {
		maxGemPhase: Phase.Phase2,
	},
};

export const P1_PRESET_BUILD_ARC = PresetUtils.makePresetBuild('P1', {
	group: 'Arcane',
	phase: Phase.Phase1,
	gear: P1_BIS_ARCANE,
	talents: ARCANE_TALENTS,
	epWeights: P1_EP_PRESET,
	rotationType: APLRotation_Type.TypeSimple,
	rotation: APL_ARCANE_SIMPLE,
	settings: P1_PLAYER_SETTINGS,
});

export const P2_PRESET_BUILD_ARC = PresetUtils.makePresetBuild('P2', {
	group: 'Arcane',
	phase: Phase.Phase2,
	gear: P2_BIS_ARCANE,
	talents: ARCANE_TALENTS,
	epWeights: P2_EP_PRESET,
	rotationType: APLRotation_Type.TypeSimple,
	rotation: APL_ARCANE_SIMPLE,
	settings: P2_PLAYER_SETTINGS,
});
