import * as PresetUtils from '../../core/preset_utils';
import { ConsumesSpec, Debuffs, IndividualBuffs, PartyBuffs, Profession, RaidBuffs, Stat, TristateEffect, PseudoStat, Drums } from '../../core/proto/common';
import { Priest_Options as Options } from '../../core/proto/priest';
import { SavedTalents } from '../../core/proto/ui';
import { Stats } from '../../core/proto_utils/stats';
import { defaultImprovedShadowBoltSettings, defaultRaidBuffMajorDamageCooldowns } from '../../core/proto_utils/utils';
import DefaultApl from './apls/default.apl.json';
import P1Gear from './gear_sets/p1.gear.json';
import P2Gear from './gear_sets/p2.gear.json';
import PreRaidGear from './gear_sets/pre_raid.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.
export const PRE_RAID_PRESET = PresetUtils.makePresetGear('Pre Raid Preset', PreRaidGear);
export const P1_PRESET = PresetUtils.makePresetGear('P1 Preset', P1Gear);
export const P2_PRESET = PresetUtils.makePresetGear('P2 Preset', P2Gear);

export const ROTATION_PRESET_DEFAULT = PresetUtils.makePresetAPLRotation('Default', DefaultApl);

// Preset options for EP weights
export const P1_EP_PRESET = PresetUtils.makePresetEpWeights(
	'P1',
	Stats.fromMap(
		{
			[Stat.StatIntellect]: 0.06,
			[Stat.StatSpirit]: 0.12,
			[Stat.StatSpellDamage]: 1.0,
			[Stat.StatShadowDamage]: 1.0,
			[Stat.StatSpellHitRating]: 1.18,
			[Stat.StatSpellCritRating]: 0.18,
			[Stat.StatSpellHasteRating]: 0.69,
			[Stat.StatMP5]: 0.05,
		},
		{
			[PseudoStat.PseudoStatSchoolHitPercentShadow]: 1.15,
		},
	),
);

// Default talents. Uses the wowhead calculator format, make the talents on
// https://www.wowhead.com/tbc/talent-calc/priest and copy the numbers in the url.
export const StandardTalents = {
	name: 'Shadow',
	data: SavedTalents.create({
		talentsString: '500230013--503250510240103051451',
	}),
};

export const DefaultOptions = Options.create({
	classOptions: {
		preShadowform: true,
	},
});

export const DefaultConsumables = ConsumesSpec.create({
	flaskId: 22866, // Flask of Pure Death
	foodId: 27657, // Blackened Basilisk
	conjuredId: 12662, // Demonic Rune
	mhImbueId: 25122, // Brilliant Wizard Oil
	potId: 22839, // Destruction Potion
	explosiveId: 30217,
});

export const DefaultRaidBuffs = RaidBuffs.create({
	...defaultRaidBuffMajorDamageCooldowns(),
	arcaneBrilliance: true,
	giftOfTheWild: TristateEffect.TristateEffectImproved,
	powerWordFortitude: TristateEffect.TristateEffectImproved,
	divineSpirit: TristateEffect.TristateEffectImproved,
});

export const DefaultPartyBuffs = PartyBuffs.create({
	manaSpringTotem: TristateEffect.TristateEffectRegular,
	wrathOfAirTotem: TristateEffect.TristateEffectImproved,
	eyeOfTheNight: true,
	chainOfTheTwilightOwl: true,
	drums: Drums.LesserDrumsOfBattle,
});

export const DefaultIndividualBuffs = IndividualBuffs.create({
	blessingOfKings: true,
	blessingOfWisdom: TristateEffect.TristateEffectImproved,
	shadowPriestDps: 0,
});

export const DefaultDebuffs = Debuffs.create({
	improvedSealOfTheCrusader: true,
	judgementOfWisdom: true,
	misery: false,
	shadowWeaving: false,
	faerieFire: TristateEffect.TristateEffectImproved,
	shadowEmbrace: true,
	curseOfElements: TristateEffect.TristateEffectImproved,
	exposeArmor: TristateEffect.TristateEffectImproved,
	...defaultImprovedShadowBoltSettings(),
});

export const OtherDefaults = {
	channelClipDelay: 100,
	distanceFromTarget: 28,
	profession1: Profession.Enchanting,
	profession2: Profession.Tailoring,
};
