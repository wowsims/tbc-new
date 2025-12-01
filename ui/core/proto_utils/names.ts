import { ArmorType, Class, Profession, Race, Spec, Stat } from '../proto/common';
import { ResourceType } from '../proto/spell';
import { DungeonDifficulty, RepFaction, RepLevel, StatCapType } from '../proto/ui';

export const armorTypeNames: Map<ArmorType, string> = new Map([
	[ArmorType.ArmorTypeUnknown, 'Unknown'],
	[ArmorType.ArmorTypeCloth, 'Cloth'],
	[ArmorType.ArmorTypeLeather, 'Leather'],
	[ArmorType.ArmorTypeMail, 'Mail'],
	[ArmorType.ArmorTypePlate, 'Plate'],
]);

export const raceNames: Map<Race, string> = new Map([
	[Race.RaceUnknown, 'None'],
	[Race.RaceBloodElf, 'Blood Elf'],
	[Race.RaceDraenei, 'Draenei'],
	[Race.RaceDwarf, 'Dwarf'],
	[Race.RaceGnome, 'Gnome'],
	[Race.RaceHuman, 'Human'],
	[Race.RaceNightElf, 'Night Elf'],
	[Race.RaceOrc, 'Orc'],
	[Race.RaceTauren, 'Tauren'],
	[Race.RaceTroll, 'Troll'],
	[Race.RaceUndead, 'Undead'],
]);

export function nameToRace(name: string): Race {
	const normalized = name.toLowerCase().replaceAll(' ', '');
	for (const [key, value] of raceNames) {
		if (value.toLowerCase().replaceAll(' ', '') == normalized) {
			return key;
		}
	}
	return Race.RaceUnknown;
}

export const classNames: Map<Class, string> = new Map([
	[Class.ClassUnknown, 'None'],
	[Class.ClassDruid, 'Druid'],
	[Class.ClassHunter, 'Hunter'],
	[Class.ClassMage, 'Mage'],
	[Class.ClassPaladin, 'Paladin'],
	[Class.ClassPriest, 'Priest'],
	[Class.ClassRogue, 'Rogue'],
	[Class.ClassShaman, 'Shaman'],
	[Class.ClassWarlock, 'Warlock'],
	[Class.ClassWarrior, 'Warrior'],
]);

export function nameToClass(name: string): Class {
	const lower = name.toLowerCase();
	for (const [key, value] of classNames) {
		if (value.toLowerCase().replace(/\s+/g, '') == lower) {
			return key;
		}
	}
	return Class.ClassUnknown;
}

export const professionNames: Map<Profession, string> = new Map([
	[Profession.ProfessionUnknown, 'None'],
	[Profession.Alchemy, 'Alchemy'],
	[Profession.Blacksmithing, 'Blacksmithing'],
	[Profession.Enchanting, 'Enchanting'],
	[Profession.Engineering, 'Engineering'],
	[Profession.Herbalism, 'Herbalism'],
	[Profession.Inscription, 'Inscription'],
	[Profession.Jewelcrafting, 'Jewelcrafting'],
	[Profession.Leatherworking, 'Leatherworking'],
	[Profession.Mining, 'Mining'],
	[Profession.Skinning, 'Skinning'],
	[Profession.Tailoring, 'Tailoring'],
]);

export function nameToProfession(name: string): Profession {
	const lower = name.toLowerCase();
	for (const [key, value] of professionNames) {
		if (value.toLowerCase() == lower) {
			return key;
		}
	}
	return Profession.ProfessionUnknown;
}

export function getStatName(stat: Stat): string {
	if (stat == Stat.StatRangedAttackPower) {
		return 'Ranged AP';
	} else {
		return Stat[stat]
			.split(/(?<![A-Z])(?=[A-Z])/)
			.slice(1)
			.join(' ');
	}
}

// TODO: Make sure BE exports the spell schools properly
export enum SpellSchool {
	None = 0,
	Physical = 1 << 1,
	Arcane = 1 << 2,
	Fire = 1 << 3,
	Frost = 1 << 4,
	Holy = 1 << 5,
	Nature = 1 << 6,
	Shadow = 1 << 7,
}

export const spellSchoolNames: Map<number, string> = new Map([
	[SpellSchool.Physical, 'Physical'],
	[SpellSchool.Arcane, 'Arcane'],
	[SpellSchool.Fire, 'Fire'],
	[SpellSchool.Frost, 'Frost'],
	[SpellSchool.Holy, 'Holy'],
	[SpellSchool.Nature, 'Nature'],
	[SpellSchool.Shadow, 'Shadow'],
	[SpellSchool.Nature + SpellSchool.Arcane, 'Astral'],
	[SpellSchool.Shadow + SpellSchool.Fire, 'Shadowflame'],
	[SpellSchool.Fire + SpellSchool.Arcane, 'Spellfire'],
	[SpellSchool.Arcane + SpellSchool.Frost, 'Spellfrost'],
	[SpellSchool.Frost + SpellSchool.Fire, 'Frostfire'],
	[SpellSchool.Shadow + SpellSchool.Frost, 'Shadowfrost'],
	[SpellSchool.Nature + SpellSchool.Shadow, 'Plague'],
	[SpellSchool.Fire + SpellSchool.Nature, 'Firestorm'],
	[SpellSchool.Fire + SpellSchool.Frost + SpellSchool.Nature, 'Elemental'],
]);

export const resourceNames: Map<ResourceType, string> = new Map([
	[ResourceType.ResourceTypeNone, 'None'],
	[ResourceType.ResourceTypeHealth, 'Health'],
	[ResourceType.ResourceTypeMana, 'Mana'],
	[ResourceType.ResourceTypeEnergy, 'Energy'],
	[ResourceType.ResourceTypeRage, 'Rage'],
	[ResourceType.ResourceTypeComboPoints, 'Combo Points'],
	[ResourceType.ResourceTypeFocus, 'Focus'],
	[ResourceType.ResourceTypeGenericResource, 'Generic Resource'],
]);

export const resourceColors: Map<ResourceType, string> = new Map([
	[ResourceType.ResourceTypeNone, '#ffffff'],
	[ResourceType.ResourceTypeHealth, '#22ba00'],
	[ResourceType.ResourceTypeMana, '#2e93fa'],
	[ResourceType.ResourceTypeEnergy, '#ffd700'],
	[ResourceType.ResourceTypeRage, '#ff0000'],
	[ResourceType.ResourceTypeComboPoints, '#ffa07a'],
	[ResourceType.ResourceTypeFocus, '#cd853f'],
	[ResourceType.ResourceTypeGenericResource, '#ffffff'],
]);

export function stringToResourceType(str: string): [ResourceType] {
	for (const [key, val] of resourceNames) {
		if (val.toLowerCase() == str.toLowerCase()) {
			return [key];
		}
	}

	return [ResourceType.ResourceTypeNone];
}

export const difficultyNames: Map<DungeonDifficulty, string> = new Map([
	[DungeonDifficulty.DifficultyUnknown, 'Unknown'],
	[DungeonDifficulty.DifficultyNormal, 'N'],
	[DungeonDifficulty.DifficultyHeroic, 'H'],
	[DungeonDifficulty.DifficultyCelestial, 'CEL'],
	[DungeonDifficulty.DifficultyRaid10, '10N'],
	[DungeonDifficulty.DifficultyRaid10H, '10H'],
	[DungeonDifficulty.DifficultyRaid25RF, 'RF'],
	[DungeonDifficulty.DifficultyRaid25, 'RN'],
	[DungeonDifficulty.DifficultyRaid25H, 'RH'],
	[DungeonDifficulty.DifficultyRaidFlex, 'RFL'],
]);

export const REP_LEVEL_NAMES: Record<RepLevel, string> = {
	[RepLevel.RepLevelUnknown]: 'Unknown',
	[RepLevel.RepLevelHated]: 'Hated',
	[RepLevel.RepLevelHostile]: 'Hostile',
	[RepLevel.RepLevelUnfriendly]: 'Unfriendly',
	[RepLevel.RepLevelNeutral]: 'Neutral',
	[RepLevel.RepLevelFriendly]: 'Friendly',
	[RepLevel.RepLevelHonored]: 'Honored',
	[RepLevel.RepLevelRevered]: 'Revered',
	[RepLevel.RepLevelExalted]: 'Exalted',
};

export const REP_FACTION_NAMES: Record<RepFaction, string> = {
	[RepFaction.RepFactionUnknown]: 'Unknown',
	[RepFaction.RepFactionTheEarthenRing]: 'The Earthen Ring',
	[RepFaction.RepFactionGuardiansOfHyjal]: 'Guardians of Hyjal',
	[RepFaction.RepFactionTherazane]: 'Therazane',
	[RepFaction.RepFactionDragonmawClan]: 'Dragonmaw Clan',
	[RepFaction.RepFactionRamkahen]: 'Ramkahen',
	[RepFaction.RepFactionWildhammerClan]: 'Wildhammer Clan',
	[RepFaction.RepFactionBaradinsWardens]: "Baradin's Wardens",
	[RepFaction.RepFactionHellscreamsReach]: "Hellscream's Reach",
	[RepFaction.RepFactionAvengersOfHyjal]: 'Avengers of Hyjal',

	[RepFaction.RepFactionGoldenLotus]: 'Golden Lotus',
	[RepFaction.RepFactionTheTillers]: 'The Tillers',
	[RepFaction.RepFactionShadoPan]: 'Shado-Pan',
	[RepFaction.RepFactionShadoPanAssault]: 'Shado-Pan Assault',
	[RepFaction.RepFactionTheBrewmasters]: 'The Brewmasters',
	[RepFaction.RepFactionTheKlaxxi]: 'The Klaxxi',
	[RepFaction.RepFactionTheAugustCelestials]: 'The August Celestials',
	[RepFaction.RepFactionTheAnglers]: 'The Anglers',
	[RepFaction.RepFactionEmperorShaohao]: 'Emperor Shaohao',
	[RepFaction.RepFactionSunreaverOnslaught]: 'Sunreaver Onslaught',
	[RepFaction.RepFactionKirinTorOffensive]: 'Kirin Tor Offensive',
	[RepFaction.RepFactionDominanceOffensive]: 'Dominance Offensive',
	[RepFaction.RepFactionOrderOfTheCloudSerpent]: 'Order of the Cloud Serpent',
	[RepFaction.RepFactionShangXisAcademy]: 'Shang Xi’s Academy',
	[RepFaction.RepFactionTheLorewalkers]: 'The Lorewalkers',
	[RepFaction.RepFactionTheBlackPrince]: 'The Black Prince',
	[RepFaction.RepFactionForestHozen]: 'Forest Hozen',
	[RepFaction.RepFactionPearlfinJinyu]: 'Pearlfin Jinyu',
	[RepFaction.RepFactionHozen]: 'Hozen',
	[RepFaction.RepFactionOperationShieldwall]: 'Operation: Shieldwall',
};

export const REP_FACTION_QUARTERMASTERS: Record<RepFaction, number> = {
	[RepFaction.RepFactionUnknown]: 0,
	[RepFaction.RepFactionTheEarthenRing]: 50324,
	[RepFaction.RepFactionGuardiansOfHyjal]: 50314,
	[RepFaction.RepFactionTherazane]: 45408,
	[RepFaction.RepFactionDragonmawClan]: 49387,
	[RepFaction.RepFactionRamkahen]: 48617,
	[RepFaction.RepFactionWildhammerClan]: 49386,
	[RepFaction.RepFactionBaradinsWardens]: 47328,
	[RepFaction.RepFactionHellscreamsReach]: 48531,
	[RepFaction.RepFactionAvengersOfHyjal]: 54401,

	[RepFaction.RepFactionGoldenLotus]: 59908,
	[RepFaction.RepFactionTheTillers]: 58706,
	[RepFaction.RepFactionShadoPan]: 64595,
	[RepFaction.RepFactionShadoPanAssault]: 70346,
	[RepFaction.RepFactionTheBrewmasters]: 64596,
	[RepFaction.RepFactionTheKlaxxi]: 64599,
	[RepFaction.RepFactionTheAugustCelestials]: 64032,
	[RepFaction.RepFactionTheAnglers]: 63721,
	[RepFaction.RepFactionEmperorShaohao]: 0,
	[RepFaction.RepFactionSunreaverOnslaught]: 67672,
	[RepFaction.RepFactionKirinTorOffensive]: 68000,
	[RepFaction.RepFactionDominanceOffensive]: 69060,
	[RepFaction.RepFactionOrderOfTheCloudSerpent]: 58414,
	[RepFaction.RepFactionShangXisAcademy]: 0,
	[RepFaction.RepFactionTheLorewalkers]: 64605,
	[RepFaction.RepFactionTheBlackPrince]: 0,
	[RepFaction.RepFactionForestHozen]: 0,
	[RepFaction.RepFactionPearlfinJinyu]: 0,
	[RepFaction.RepFactionHozen]: 0,
	[RepFaction.RepFactionOperationShieldwall]: 69059,
};

export const statCapTypeNames = new Map<StatCapType, string>([
	[StatCapType.TypeHardCap, 'Hard cap'],
	[StatCapType.TypeSoftCap, 'Soft cap'],
	[StatCapType.TypeThreshold, 'Threshold'],
]);
