import { EligibleWeaponType, IconSize, PlayerClass } from '../player_class';
import { PlayerSpec } from '../player_spec';
import { Hunter as HunterSpec } from '../player_specs/hunter';
import { ArmorType, Class, Race, RangedWeaponType, WeaponType } from '../proto/common';
import { HunterSpecs } from '../proto_utils/utils';

export class Hunter extends PlayerClass<Class.ClassHunter> {
	static classID = Class.ClassHunter as Class.ClassHunter;
	static friendlyName = 'Hunter';
	static hexColor = '#abd473';
	static specs: Record<string, PlayerSpec<HunterSpecs>> = {
		[HunterSpec.friendlyName]: HunterSpec,
	};
	static races: Race[] = [
		// [A]
		Race.RaceDwarf,
		Race.RaceNightElf,
		Race.RaceDraenei,
		// [H]
		Race.RaceOrc,
		Race.RaceTauren,
		Race.RaceTroll,
		Race.RaceBloodElf,
	];
	static armorTypes: ArmorType[] = [ArmorType.ArmorTypeMail, ArmorType.ArmorTypeLeather, ArmorType.ArmorTypeCloth];
	static weaponTypes: EligibleWeaponType[] = [
		{ weaponType: WeaponType.WeaponTypeDagger },
		{ weaponType: WeaponType.WeaponTypeFist },
		{ weaponType: WeaponType.WeaponTypeAxe, canUseTwoHand: true },
		{ weaponType: WeaponType.WeaponTypeOffHand },
		{ weaponType: WeaponType.WeaponTypeSword, canUseTwoHand: true },
		{ weaponType: WeaponType.WeaponTypeStaff, canUseTwoHand: true },
		{ weaponType: WeaponType.WeaponTypePolearm, canUseTwoHand: true },
	];
	static rangedWeaponTypes: RangedWeaponType[] = [
		RangedWeaponType.RangedWeaponTypeBow,
		RangedWeaponType.RangedWeaponTypeCrossbow,
		RangedWeaponType.RangedWeaponTypeGun,
	];

	readonly classID = Hunter.classID;
	readonly friendlyName = Hunter.name;
	readonly hexColor = Hunter.hexColor;
	readonly specs = Hunter.specs;
	readonly races = Hunter.races;
	readonly armorTypes = Hunter.armorTypes;

	readonly weaponTypes = Hunter.weaponTypes;
	readonly rangedWeaponTypes = Hunter.rangedWeaponTypes;

	static getIcon = (size: IconSize): string => {
		return `https://wow.zamimg.com/images/wow/icons/${size}/class_hunter.jpg`;
	};

	getIcon = (size: IconSize): string => {
		return Hunter.getIcon(size);
	};
}
