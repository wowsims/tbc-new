import * as PresetUtils from '../../core/preset_utils';
import { ConsumesSpec, Debuffs, Drums, IndividualBuffs, PartyBuffs, Profession, RaidBuffs, Stat, TristateEffect } from '../../core/proto/common';
import { Warlock_Options as WarlockOptions, WarlockOptions_Armor, WarlockOptions_CurseOptions, WarlockOptions_Summon } from '../../core/proto/warlock';
import { SavedTalents } from '../../core/proto/ui';
import { Stats } from '../../core/proto_utils/stats';
import BlankAPL from './apls/blank.apl.json';
import BlankGear from './gear_sets/blank.gear.json';
import PreRaid from './gear_sets/preraid.gear.json';
import PreRaidFire from './gear_sets/destro_fire_preraid.gear.json';
import T4Set from './gear_sets/t4.gear.json';
import T4Fire from './gear_sets/destro_fire_t4.gear.json';
import T5Set from './gear_sets/t5.gear.json';
import T6Set from './gear_sets/t6.gear.json';
import ZASet from './gear_sets/za.gear.json';
import SWPSet from './gear_sets/swp.gear.json';
import AfflictionRot from './apls/affliction.apl.json';
import DemoRot from './apls/demonology.apl.json';
import DestroRot from './apls/destruction.apl.json';
import DestroFireRot from './apls/destro_fire.apl.json';
import { defaultExposeWeaknessSettings, defaultImprovedShadowBoltSettings, defaultRaidBuffMajorDamageCooldowns } from '../../core/proto_utils/utils';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

export const BLANK_APL = PresetUtils.makePresetAPLRotation('Blank', BlankAPL);

export const BLANK_GEARSET = PresetUtils.makePresetGear('Blank', BlankGear);

export const PRE_RAID = PresetUtils.makePresetGear('Pre-Raid', PreRaid);
export const PRE_RAID_FIRE = PresetUtils.makePresetGear('Pre-Raid (Fire)', PreRaidFire);

export const T4 = PresetUtils.makePresetGear('T4', T4Set);
export const T4_FIRE = PresetUtils.makePresetGear('T4 (Fire)', T4Fire);

export const T5 = PresetUtils.makePresetGear('T5', T5Set);
export const T6 = PresetUtils.makePresetGear('T6', T6Set);
export const ZA = PresetUtils.makePresetGear("Zul'Aman", ZASet);
export const SWP = PresetUtils.makePresetGear('Sunwell Plateau', SWPSet);

// Preset options for EP weights
export const P1_AFFLI_DEMO_DESTRO_EP = PresetUtils.makePresetEpWeights(
	'P1 - Affli / Demo / Destro',
	Stats.fromMap({
		[Stat.StatIntellect]: 0.38,
		[Stat.StatSpellDamage]: 1,
		[Stat.StatFireDamage]: 0.07,
		[Stat.StatShadowDamage]: 0.92,
		[Stat.StatSpellHitRating]: 1.73,
		[Stat.StatSpellCritRating]: 0.82,
		[Stat.StatSpellHasteRating]: 1.21,
		[Stat.StatMP5]: 0.29,
	}),
);

export const P1_DESTRUCTION_FIRE_EP = PresetUtils.makePresetEpWeights(
	'P1 - Destro (Fire)',
	P1_AFFLI_DEMO_DESTRO_EP.epWeights.withStat(Stat.StatFireDamage, 0.92).withStat(Stat.StatShadowDamage, 0.07),
);

// Rotations
export const AfflictionAPL = PresetUtils.makePresetAPLRotation('Affliction', AfflictionRot);
export const DemoAPL = PresetUtils.makePresetAPLRotation('Demonology', DemoRot);
export const DestroAPL = PresetUtils.makePresetAPLRotation('Destruction', DestroRot);
export const DestroFireAPL = PresetUtils.makePresetAPLRotation('Destruction (Fire)', DestroFireRot);

// Default talents. Uses the wowhead calculator format, make the talents on
// https://wowhead.com/wotlk/talent-calc and copy the numbers in the url.
export const TalentsAffliction = {
	name: 'Affliction',
	data: SavedTalents.create({
		talentsString: '05022221112351055003--50500051220001',
	}),
};

export const TalentsDemoRuin = {
	name: 'Demo/Ruin',
	data: SavedTalents.create({
		talentsString: '01-205003213305010150134-50500251020001',
	}),
};

export const TalentsDemoFelguard = {
	name: 'Demonology Felguard',
	data: SavedTalents.create({
		talentsString: '01-2050030133250101501351-5050005112',
	}),
};

export const TalentsDestroNightfall = {
	name: 'Destro/Nightfall',
	data: SavedTalents.create({
		talentsString: '150222201023--505020510200510531051',
	}),
};

export const TalentsDestruction = {
	name: 'Destruction',
	data: SavedTalents.create({
		talentsString: '-20500301332101-50500051220051053105',
	}),
};

// Defaults
export const DefaultOptions = WarlockOptions.create({
	classOptions: {
		armor: WarlockOptions_Armor.FelArmor,
		curseOptions: WarlockOptions_CurseOptions.Recklessness,
		sacrificeSummon: true,
		summon: WarlockOptions_Summon.Succubus,
	},
});

export const DefaultConsumables = ConsumesSpec.create({
	flaskId: 22866, // Flask of Pure Death
	foodId: 27657, // Blackened Basilisk
	conjuredId: 12662, // Demonic Rune
	mhImbueId: 25122, // Brilliant Wizard Oil
	potId: 22839, // Destruction Potion
	explosiveId: 30217,
	petScrollAgi: true,
	petScrollStr: true,
});

export const OtherDefaults = {
	distanceFromTarget: 20,
	profession1: Profession.Engineering,
	profession2: Profession.Tailoring,
};

export const DefaultRaidBuffs = RaidBuffs.create({
	...defaultRaidBuffMajorDamageCooldowns(),
	arcaneBrilliance: true,
	giftOfTheWild: TristateEffect.TristateEffectImproved,
	powerWordFortitude: TristateEffect.TristateEffectImproved,
	divineSpirit: TristateEffect.TristateEffectImproved,
});

export const DefaultPartyBuffs = PartyBuffs.create({
	manaSpringTotem: TristateEffect.TristateEffectRegular,
	moonkinAura: TristateEffect.TristateEffectRegular,
	totemOfWrath: 1,
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
	...defaultExposeWeaknessSettings(),
	...defaultImprovedShadowBoltSettings(),
	improvedSealOfTheCrusader: true,
	judgementOfWisdom: true,
	misery: true,
	shadowWeaving: true,
	sunderArmor: true,
	screech: true,
	faerieFire: TristateEffect.TristateEffectImproved,
	curseOfRecklessness: true,
	shadowEmbrace: true,
	curseOfElements: TristateEffect.TristateEffectImproved,
	bloodFrenzy: true,
	giftOfArthas: true,
	mangle: true,
	exposeArmor: TristateEffect.TristateEffectImproved,
	huntersMark: TristateEffect.TristateEffectImproved,
});

export const P1_DEFAULT_SETTINGS: PresetUtils.PresetSettings = {
	name: 'Default',
	specOptions: DefaultOptions,
	consumables: DefaultConsumables,
	buffs: DefaultIndividualBuffs,
	partyBuffs: DefaultPartyBuffs,
	raidBuffs: DefaultRaidBuffs,
	debuffs: DefaultDebuffs,
};

export const P1_AFFLICTION_DEFAULT_SETTINGS: PresetUtils.PresetSettings = {
	...P1_DEFAULT_SETTINGS,
	name: 'Affliction',
	specOptions: WarlockOptions.create({
		...DefaultOptions,
		classOptions: {
			...DefaultOptions.classOptions,
			curseOptions: WarlockOptions_CurseOptions.Elements,
			summon: WarlockOptions_Summon.Imp,
			sacrificeSummon: false,
		},
	}),
	debuffs: DefaultDebuffs,
};

export const P1_DEMONOLOGY_DEFAULT_SETTINGS: PresetUtils.PresetSettings = {
	...P1_DEFAULT_SETTINGS,
	name: 'Demonology',
	specOptions: WarlockOptions.create({
		...DefaultOptions,
		classOptions: {
			...DefaultOptions.classOptions,
			curseOptions: WarlockOptions_CurseOptions.Recklessness,
			summon: WarlockOptions_Summon.Succubus,
			sacrificeSummon: false,
		},
	}),
	debuffs: DefaultDebuffs,
};

export const P1_FIRE_DEFAULT_SETTINGS: PresetUtils.PresetSettings = {
	...P1_DEFAULT_SETTINGS,
	name: 'Fire',
	specOptions: WarlockOptions.create({
		...DefaultOptions,
		classOptions: {
			...DefaultOptions.classOptions,
			summon: WarlockOptions_Summon.Imp,
			sacrificeSummon: true,
		},
	}),
	consumables: ConsumesSpec.create({
		...DefaultConsumables,
		conjuredId: 22788,
	}),
	debuffs: Debuffs.create({
		...DefaultDebuffs,
		improvedScorch: true,
	}),
};

// Builds
export const AFFLICTION_BUILD = PresetUtils.makePresetBuild('Affliction', {
	talents: TalentsAffliction,
	epWeights: P1_AFFLI_DEMO_DESTRO_EP,
	rotation: AfflictionAPL,
	settings: P1_AFFLICTION_DEFAULT_SETTINGS,
});

export const DEMONOLOGY_BUILD = PresetUtils.makePresetBuild('Demonology', {
	talents: TalentsDemoRuin,
	epWeights: P1_AFFLI_DEMO_DESTRO_EP,
	rotation: DemoAPL,
	settings: P1_DEMONOLOGY_DEFAULT_SETTINGS,
});

export const DESTRUCTION_BUILD = PresetUtils.makePresetBuild('Destruction', {
	talents: TalentsDestruction,
	epWeights: P1_AFFLI_DEMO_DESTRO_EP,
	rotation: DestroAPL,
	settings: P1_DEFAULT_SETTINGS,
});

export const DESTRUCTION_FIRE_BUILD = PresetUtils.makePresetBuild('Destruction (Fire)', {
	talents: TalentsDestruction,
	epWeights: P1_DESTRUCTION_FIRE_EP,
	rotation: DestroFireAPL,
	settings: P1_FIRE_DEFAULT_SETTINGS,
});
