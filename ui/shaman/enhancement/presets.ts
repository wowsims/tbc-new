import * as Mechanics from '../../core/constants/mechanics';
import * as PresetUtils from '../../core/preset_utils.js';
import {
	Class,
	ConsumesSpec,
	Debuffs,
	IndividualBuffs,
	PartyBuffs,
	Profession,
	PseudoStat,
	Race,
	RaidBuffs,
	Stat,
	TristateEffect,
} from '../../core/proto/common.js';
import { EnhancementShaman_Options as EnhancementShamanOptions, ShamanImbue, ShamanSyncType } from '../../core/proto/shaman.js';
import { SavedTalents } from '../../core/proto/ui.js';
import { Stats } from '../../core/proto_utils/stats';
import { defaultExposeWeaknessSettings, defaultRaidBuffMajorDamageCooldowns } from '../../core/proto_utils/utils';
import DefaultApl from './apls/default.apl.json';
import P1Gear from './gear_sets/p1.gear.json';
import P2Gear from './gear_sets/p2.gear.json';
import P3Gear from './gear_sets/p3.gear.json';
import P4Gear from './gear_sets/p4.gear.json';
import P5Gear from './gear_sets/p5.gear.json';
import P1ItemSwap from './gear_sets/p1.itemswap.json';
import PreraidGear from './gear_sets/preraid.gear.json';
import { Phase } from '../../core/constants/other';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

export const PRERAID_PRESET = PresetUtils.makePresetGear('Pre-raid', PreraidGear);

export const P1_PRESET = PresetUtils.makePresetGear('P1 Preset', P1Gear);
export const P2_PRESET = PresetUtils.makePresetGear('P2 Preset', P2Gear);
export const P3_PRESET = PresetUtils.makePresetGear('P3 Preset', P3Gear);
export const P4_PRESET = PresetUtils.makePresetGear('P4 Preset', P4Gear);
export const P5_PRESET = PresetUtils.makePresetGear('P5 Preset', P5Gear);

export const P1_ITEMSWAP_PRESET = PresetUtils.makePresetItemSwapGear('P1 ItemSwap Preset', P1ItemSwap);

export const ROTATION_PRESET_DEFAULT = PresetUtils.makePresetAPLRotation('Default', DefaultApl);

// Preset options for EP weights
export const P1_EP_PRESET = PresetUtils.makePresetEpWeights(
	'Default',
	Stats.fromMap(
		{
			// calculated in p1 bis after building out new default APL
			[Stat.StatStrength]: 2.2,
			[Stat.StatAgility]: 1.62,
			[Stat.StatIntellect]: 0.08,
			[Stat.StatSpellDamage]: 0.56,
			[Stat.StatNatureDamage]: 0.4, // As simulated using Fire Ele Totem Only
			[Stat.StatSpellHitRating]: 0.55,
			[Stat.StatSpellCritRating]: 0.13,
			[Stat.StatAttackPower]: 1.0,
			[Stat.StatMeleeHitRating]: 1.9,
			[Stat.StatMeleeCritRating]: 1.73,
			[Stat.StatMeleeHasteRating]: 1.37,
			[Stat.StatArmorPenetration]: 0.3,
			[Stat.StatExpertiseRating]: 2.49,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 8.19,
			[PseudoStat.PseudoStatOffHandDps]: 3.59,
		},
	),
);

export const P3_EP_PRESET = PresetUtils.makePresetEpWeights(
	'P3 (WiP)',
	Stats.fromMap(
		{
			// calculated in p3 bis after building out new default APL
			[Stat.StatIntellect]: 0.1,
			[Stat.StatAgility]: 1.69,
			[Stat.StatStrength]: 2.2,
			[Stat.StatAttackPower]: 1.0,
			[Stat.StatSpellDamage]: 0.48,
			[Stat.StatNatureDamage]: 0.35, // As simulated using Fire Ele Totem Only
			[Stat.StatMeleeHitRating]: 1.91,
			[Stat.StatMeleeCritRating]: 1.74,
			[Stat.StatMeleeHasteRating]: 1.94,
			[Stat.StatArmorPenetration]: 0.33,
			[Stat.StatExpertiseRating]: 2.73,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 8.25,
			[PseudoStat.PseudoStatOffHandDps]: 3.61,
		},
	),
);

// Default talents. Uses the wowhead calculator format, make the talents on
// https://wowhead.com/tbc/talent-calc and copy the numbers in the url.
export const SubRestoIWT = {
	name: 'Sub-Restoration IWT',
	data: SavedTalents.create({
		talentsString: '03-500502210501133531151-50005301',
	}),
};

export const SubRestoILS = {
	name: 'Sub-Restoration ILS',
	data: SavedTalents.create({
		talentsString: '03-500503210500133531151-50005301',
	}),
};

export const SubEle = {
	name: 'Sub-Elemental',
	data: SavedTalents.create({
		talentsString: '250031501-500503210500133531151',
	}),
};

export const DefaultIndividualBuffs = IndividualBuffs.create({
	blessingOfKings: true,
	blessingOfMight: TristateEffect.TristateEffectImproved,
});

export const DefaultOptions = EnhancementShamanOptions.create({
	classOptions: {
		shieldProcrate: 0,
		imbueMh: ShamanImbue.WindfuryWeapon,
		imbueMhSwap: ShamanImbue.WindfuryWeapon,
	},
	imbueOh: ShamanImbue.WindfuryWeapon,
	syncType: ShamanSyncType.Auto,
});

export const OtherDefaults = {
	distanceFromTarget: 5,
	profession1: Profession.Engineering,
	profession2: Profession.Leatherworking,
	race: Race.RaceOrc,
};

export const DefaultConsumables = ConsumesSpec.create({
	potId: 22838, // Haste Potion
	flaskId: 22854, // Flask of Relentless Assault
	foodId: 27658, // Roasted Clefthoof
	drumsId: 351355,
	conjuredId: 22788,
	explosiveId: 30217,
	superSapper: true,
	goblinSapper: true,
	scrollAgi: true,
	scrollStr: true,
});

export const DefaultPartyBuffs = PartyBuffs.create({
	ferociousInspiration: 2,
	braidedEterniumChain: true,
	leaderOfThePack: TristateEffect.TristateEffectRegular,
	battleShout: TristateEffect.TristateEffectImproved,
});

export const DefaultRaidBuffs = RaidBuffs.create({
	...defaultRaidBuffMajorDamageCooldowns(Class.ClassShaman),
	powerWordFortitude: TristateEffect.TristateEffectImproved,
	giftOfTheWild: TristateEffect.TristateEffectImproved,
	arcaneBrilliance: true,
});

export const DefaultDebuffs = Debuffs.create({
	...defaultExposeWeaknessSettings(Phase.Phase1),
	improvedSealOfTheCrusader: true,
	judgementOfWisdom: true,
	screech: true,
	misery: true,
	bloodFrenzy: true,
	giftOfArthas: true,
	mangle: true,
	exposeArmor: TristateEffect.TristateEffectImproved,
	faerieFire: TristateEffect.TristateEffectImproved,
	sunderArmor: true,
	curseOfElements: TristateEffect.TristateEffectImproved,
	curseOfRecklessness: true,
	huntersMark: TristateEffect.TristateEffectImproved,
});
