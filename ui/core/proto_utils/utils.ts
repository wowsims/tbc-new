import { CURRENT_API_VERSION, REPO_NAME } from '../constants/other.js';
import { PlayerClass } from '../player_class.js';
import { PlayerClasses } from '../player_classes';
import { PlayerSpec } from '../player_spec.js';
import { PlayerSpecs } from '../player_specs';
import { Player } from '../proto/api.js';
import {
	Class,
	EnchantType,
	Faction,
	HandType,
	ItemSlot,
	ItemType,
	Profession,
	Race,
	RaidBuffs,
	RangedWeaponType,
	Spec,
	UnitReference,
	UnitReference_Type,
	WeaponType,
} from '../proto/common.js';
import {
	BalanceDruid,
	BalanceDruid_Options,
	BalanceDruid_Rotation,
	DruidOptions,
	DruidTalents,
	FeralCatDruid,
	FeralCatDruid_Options,
	FeralCatDruid_Rotation,
	FeralBearDruid,
	FeralBearDruid_Options,
	FeralBearDruid_Rotation,
	RestorationDruid,
	RestorationDruid_Options,
	RestorationDruid_Rotation,
} from '../proto/druid.js';
import {
	Hunter,
	Hunter_Rotation,
	Hunter_Options,
	HunterOptions,
	HunterTalents,
} from '../proto/hunter.js';
import {
	Mage,
	Mage_Options,
	Mage_Rotation,
	MageOptions,
	MageTalents,
} from '../proto/mage.js';
import {
	Blessings,
	HolyPaladin,
	HolyPaladin_Options,
	HolyPaladin_Rotation,
	PaladinOptions,
	PaladinTalents,
	ProtectionPaladin,
	ProtectionPaladin_Options,
	ProtectionPaladin_Rotation,
	RetributionPaladin,
	RetributionPaladin_Options,
	RetributionPaladin_Rotation,
} from '../proto/paladin.js';
import {
	DisciplinePriest,
	DisciplinePriest_Options,
	DisciplinePriest_Rotation,
	HolyPriest,
	HolyPriest_Options,
	HolyPriest_Rotation,
	PriestOptions,
	PriestTalents,
	ShadowPriest,
	ShadowPriest_Options,
	ShadowPriest_Rotation,
} from '../proto/priest.js';
import {
	Rogue,
	Rogue_Options,
	Rogue_Rotation,
	RogueOptions,
	RogueTalents,
} from '../proto/rogue.js';
import {
	ElementalShaman,
	ElementalShaman_Options,
	ElementalShaman_Rotation,
	EnhancementShaman,
	EnhancementShaman_Options,
	EnhancementShaman_Rotation,
	RestorationShaman,
	RestorationShaman_Options,
	RestorationShaman_Rotation,
	ShamanOptions,
	ShamanTalents,
} from '../proto/shaman.js';
import { ResourceType } from '../proto/spell';
import { BlessingsAssignment, BlessingsAssignments, UIEnchant as Enchant, UIGem as Gem, UIItem as Item } from '../proto/ui.js';
import {
	Warlock,
	Warlock_Options,
	Warlock_Rotation,
	WarlockOptions,
	WarlockTalents,
} from '../proto/warlock.js';
import {
	DPSWarrior,
	DPSWarrior_Options,
	DPSWarrior_Rotation,
	ProtectionWarrior,
	ProtectionWarrior_Options,
	ProtectionWarrior_Rotation,
	WarriorOptions,
	WarriorTalents,
} from '../proto/warrior.js';
import { getEnumValues, intersection, sum } from '../utils.js';
import { Stats } from './stats.js';

export const NUM_SPECS = getEnumValues(Spec).length;

export const raidSimIcon = '/tbc/assets/img/raid_icon.png';
export const raidSimLabel = 'Full Raid Sim';

// Converts '1231321-12313123-0' to [40, 21, 0].
export function getTalentTreePoints(talentsString: string): Array<number> {
	const trees = talentsString.split('-');
	if (trees.length == 2) {
		trees.push('0');
	}
	return trees.map(tree => sum([...tree].map(char => parseInt(char) || 0)));
}

export function getTalentPoints(talentsString: string): number {
	return getTalentTreePoints(talentsString).filter(Boolean).length;
}

// Gets the URL for the individual sim corresponding to the given spec.
export function getSpecSiteUrl(classString: string, specString: string): string {
	const specSiteUrlTemplate = new URL(`${window.location.protocol}//${window.location.host}/${REPO_NAME}/CLASS/SPEC/`).toString();
	return specSiteUrlTemplate.replace('CLASS', classString).replace('SPEC', specString);
}
export const raidSimSiteUrl = new URL(`${window.location.protocol}//${window.location.host}/${REPO_NAME}/raid/`).toString();

export function textCssClassForClass<ClassType extends Class>(playerClass: PlayerClass<ClassType>): string {
	return `text-${PlayerClasses.getCssClass(playerClass)}`;
}
export function textCssClassForSpec<SpecType extends Spec>(playerSpec: PlayerSpec<SpecType>): string {
	return textCssClassForClass(PlayerSpecs.getPlayerClass(playerSpec));
}

// Placeholder classes to fill the Unknown Spec Type Functions entry below
type UnknownSpecs = Spec.SpecUnknown;
class UnknownRotation {
	// eslint-disable-next-line @typescript-eslint/no-empty-function
	constructor() {}
}
class UnknownTalents {
	// eslint-disable-next-line @typescript-eslint/no-empty-function
	constructor() {}
}
class UnknownClassOptions {
	// eslint-disable-next-line @typescript-eslint/no-empty-function
	constructor() {}
}
class UnknownSpecOptions {
	classOptions: UnknownClassOptions;
	// eslint-disable-next-line @typescript-eslint/no-empty-function
	constructor() {
		this.classOptions = new UnknownClassOptions();
	}
}

export type DruidSpecs = Spec.SpecBalanceDruid | Spec.SpecFeralCatDruid | Spec.SpecFeralBearDruid | Spec.SpecRestorationDruid;
export type HunterSpecs = Spec.SpecHunter;
export type MageSpecs = Spec.SpecMage;
export type PaladinSpecs = Spec.SpecHolyPaladin | Spec.SpecRetributionPaladin | Spec.SpecProtectionPaladin;
export type PriestSpecs = Spec.SpecDisciplinePriest | Spec.SpecHolyPriest | Spec.SpecShadowPriest;
export type RogueSpecs = Spec.SpecRogue;
export type ShamanSpecs = Spec.SpecElementalShaman | Spec.SpecEnhancementShaman | Spec.SpecRestorationShaman;
export type WarlockSpecs = Spec.SpecWarlock;
export type WarriorSpecs = Spec.SpecDPSWarrior | Spec.SpecProtectionWarrior;

export type ClassSpecs<T extends Class> = T extends Class.ClassDruid
	? DruidSpecs
	: T extends Class.ClassHunter
		? HunterSpecs
		: T extends Class.ClassMage
			? MageSpecs
			: T extends Class.ClassPaladin
				? PaladinSpecs
				: T extends Class.ClassPriest
					? PriestSpecs
					: T extends Class.ClassRogue
						? RogueSpecs
						: T extends Class.ClassShaman
							? ShamanSpecs
							: T extends Class.ClassWarlock
								? WarlockSpecs
								: T extends Class.ClassWarrior
									? WarriorSpecs
									: // Should never reach this case
										UnknownSpecs;

export type SpecClasses<T extends Spec> =
	// Druid
	T extends DruidSpecs
		? Class.ClassDruid
		: // Hunter
			T extends HunterSpecs
			? Class.ClassHunter
			: // Mage
				T extends MageSpecs
				? Class.ClassMage
				: // Paladin
					T extends PaladinSpecs
					? Class.ClassPaladin
					: // Priest
						T extends PriestSpecs
						? Class.ClassPriest
						: // Rogue
							T extends RogueSpecs
							? Class.ClassRogue
							: // Shaman
								T extends ShamanSpecs
								? Class.ClassShaman
								: // Warlock
									T extends WarlockSpecs
									? Class.ClassWarlock
									: // Warrior
										T extends WarriorSpecs
										? Class.ClassWarrior
										: // Should never reach this case
											Class.ClassUnknown;

export type SpecRotation<T extends Spec> =
	// Druid
	T extends Spec.SpecBalanceDruid
	? BalanceDruid_Rotation
	: T extends Spec.SpecFeralCatDruid
		? FeralCatDruid_Rotation
		: T extends Spec.SpecFeralBearDruid
			? FeralBearDruid_Rotation
			: T extends Spec.SpecRestorationDruid
				? RestorationDruid_Rotation
				: // Hunter
					T extends Spec.SpecHunter
					? Hunter_Rotation
							: // Mage
								T extends Spec.SpecMage
								? Mage_Rotation
										: // Paladin
											T extends Spec.SpecHolyPaladin
											? HolyPaladin_Rotation
											: T extends Spec.SpecProtectionPaladin
												? ProtectionPaladin_Rotation
												: T extends Spec.SpecRetributionPaladin
													? RetributionPaladin_Rotation
													: // Priest
														T extends Spec.SpecDisciplinePriest
														? DisciplinePriest_Rotation
														: T extends Spec.SpecHolyPriest
															? HolyPriest_Rotation
															: T extends Spec.SpecShadowPriest
																? ShadowPriest_Rotation
																: // Rogue
																	T extends Spec.SpecRogue
																	? Rogue_Rotation
																			: // Shaman
																				T extends Spec.SpecElementalShaman
																				? ElementalShaman_Rotation
																				: T extends Spec.SpecEnhancementShaman
																					? EnhancementShaman_Rotation
																					: T extends Spec.SpecRestorationShaman
																						? RestorationShaman_Rotation
																						: // Warlock
																							T extends Spec.SpecWarlock
																							? Warlock_Rotation
																									: // Warrior
																										T extends Spec.SpecDPSWarrior
																										? DPSWarrior_Rotation
																											: T extends Spec.SpecProtectionWarrior
																												? ProtectionWarrior_Rotation
																												: // Should never reach this case
																													UnknownRotation;

export type SpecTalents<T extends Spec> =
	// Druid
	T extends DruidSpecs
	? DruidTalents
	: // Hunter
		T extends HunterSpecs
		? HunterTalents
		: // Mage
			T extends MageSpecs
			? MageTalents
			: // Paladin
				T extends PaladinSpecs
				? PaladinTalents
				: // Priest
					T extends PriestSpecs
					? PriestTalents
					: // Rogue
						T extends RogueSpecs
						? RogueTalents
						: // Shaman
							T extends ShamanSpecs
							? ShamanTalents
							: // Warlock
								T extends WarlockSpecs
								? WarlockTalents
								: // Warrior
									T extends WarriorSpecs
									? WarriorTalents
									: // Should never reach this case
										UnknownTalents;

export type ClassOptions<T extends Spec> =
	// Druid
	T extends DruidSpecs
	? DruidOptions
	: // Hunter
		T extends HunterSpecs
		? HunterOptions
		: // Mage
			T extends MageSpecs
			? MageOptions
				: // Paladin
					T extends PaladinSpecs
					? PaladinOptions
					: // Priest
						T extends PriestSpecs
						? PriestOptions
						: // Rogue
							T extends RogueSpecs
							? RogueOptions
							: // Shaman
								T extends ShamanSpecs
								? ShamanOptions
								: // Warlock
									T extends WarlockSpecs
									? WarlockOptions
									: // Warrior
										T extends WarriorSpecs
										? WarriorOptions
										: // Should never reach this case
											UnknownClassOptions;

export type SpecOptions<T extends Spec> =
	// Druid
	T extends Spec.SpecBalanceDruid
	? BalanceDruid_Options
	: T extends Spec.SpecFeralCatDruid
		? FeralCatDruid_Options
		: T extends Spec.SpecFeralBearDruid
			? FeralBearDruid_Options
			: T extends Spec.SpecRestorationDruid
				? RestorationDruid_Options
				: // Hunter
					T extends Spec.SpecHunter
					? Hunter_Options
							: // Mage
								T extends Spec.SpecMage
								? Mage_Options
										: // Paladin
											T extends Spec.SpecHolyPaladin
											? HolyPaladin_Options
											: T extends Spec.SpecProtectionPaladin
												? ProtectionPaladin_Options
												: T extends Spec.SpecRetributionPaladin
													? RetributionPaladin_Options
													: // Priest
														T extends Spec.SpecDisciplinePriest
														? DisciplinePriest_Options
														: T extends Spec.SpecHolyPriest
															? HolyPriest_Options
															: T extends Spec.SpecShadowPriest
																? ShadowPriest_Options
																: // Rogue
																	T extends Spec.SpecRogue
																	? Rogue_Options
																			: // Shaman
																				T extends Spec.SpecElementalShaman
																				? ElementalShaman_Options
																				: T extends Spec.SpecEnhancementShaman
																					? EnhancementShaman_Options
																					: T extends Spec.SpecRestorationShaman
																						? RestorationShaman_Options
																						: // Warlock
																							T extends Spec.SpecWarlock
																							? Warlock_Options
																									: // Warrior
																										T extends Spec.SpecDPSWarrior
																										? DPSWarrior_Options
																											: T extends Spec.SpecProtectionWarrior
																												? ProtectionWarrior_Options
																												: // Should never reach this case
																													UnknownSpecOptions;

export type SpecType<T extends Spec> =
	// Druid
	T extends Spec.SpecBalanceDruid
	? BalanceDruid
	: T extends Spec.SpecFeralCatDruid
		? FeralCatDruid
		: T extends Spec.SpecFeralBearDruid
			? FeralBearDruid
			: T extends Spec.SpecRestorationDruid
				? RestorationDruid
				: // Hunter
					T extends Spec.SpecHunter
					? Hunter
							: // Mage
								T extends Spec.SpecMage
								? Mage
										: // Paladin
											T extends Spec.SpecHolyPaladin
											? HolyPaladin
											: T extends Spec.SpecProtectionPaladin
												? ProtectionPaladin
												: T extends Spec.SpecRetributionPaladin
													? RetributionPaladin
													: // Priest
														T extends Spec.SpecDisciplinePriest
														? DisciplinePriest
														: T extends Spec.SpecHolyPriest
															? HolyPriest
															: T extends Spec.SpecShadowPriest
																? ShadowPriest
																: // Rogue
																	T extends Spec.SpecRogue
																	? Rogue
																			: // Shaman
																				T extends Spec.SpecElementalShaman
																				? ElementalShaman
																				: T extends Spec.SpecEnhancementShaman
																					? EnhancementShaman
																					: T extends Spec.SpecRestorationShaman
																						? RestorationShaman
																						: // Warlock
																							T extends Spec.SpecWarlock
																							? Warlock
																									: // Warrior
																										T extends Spec.SpecDPSWarrior
																										? DPSWarrior
																											: T extends Spec.SpecProtectionWarrior
																												? ProtectionWarrior
																												: // Should never reach this case
																													Spec.SpecUnknown;

export type SpecTypeFunctions<SpecType extends Spec> = {
	rotationCreate: () => SpecRotation<SpecType>;
	rotationEquals: (a: SpecRotation<SpecType>, b: SpecRotation<SpecType>) => boolean;
	rotationCopy: (a: SpecRotation<SpecType>) => SpecRotation<SpecType>;
	rotationToJson: (a: SpecRotation<SpecType>) => any;
	rotationFromJson: (obj: any) => SpecRotation<SpecType>;

	talentsCreate: () => SpecTalents<SpecType>;
	talentsEquals: (a: SpecTalents<SpecType>, b: SpecTalents<SpecType>) => boolean;
	talentsCopy: (a: SpecTalents<SpecType>) => SpecTalents<SpecType>;
	talentsToJson: (a: SpecTalents<SpecType>) => any;
	talentsFromJson: (obj: any) => SpecTalents<SpecType>;

	optionsCreate: () => SpecOptions<SpecType>;
	optionsEquals: (a: SpecOptions<SpecType>, b: SpecOptions<SpecType>) => boolean;
	optionsCopy: (a: SpecOptions<SpecType>) => SpecOptions<SpecType>;
	optionsToJson: (a: SpecOptions<SpecType>) => any;
	optionsFromJson: (obj: any) => SpecOptions<SpecType>;
	optionsFromPlayer: (player: Player) => SpecOptions<SpecType>;
};

export const specTypeFunctions: Record<Spec, SpecTypeFunctions<any>> = {
	[Spec.SpecUnknown]: {
		rotationCreate: () => new UnknownRotation(),
		rotationEquals: (_a, _b) => true,
		rotationCopy: _a => new UnknownRotation(),
		rotationToJson: _a => undefined,
		rotationFromJson: _obj => new UnknownRotation(),

		talentsCreate: () => new UnknownTalents(),
		talentsEquals: (_a, _b) => true,
		talentsCopy: _a => new UnknownTalents(),
		talentsToJson: _a => undefined,
		talentsFromJson: _obj => new UnknownTalents(),

		optionsCreate: () => new UnknownSpecOptions(),
		optionsEquals: (_a, _b) => true,
		optionsCopy: _a => new UnknownSpecOptions(),
		optionsToJson: _a => undefined,
		optionsFromJson: _obj => new UnknownSpecOptions(),
		optionsFromPlayer: _player => new UnknownSpecOptions(),
	},

	// Druid
	[Spec.SpecBalanceDruid]: {
		rotationCreate: () => BalanceDruid_Rotation.create(),
		rotationEquals: (a, b) => BalanceDruid_Rotation.equals(a as BalanceDruid_Rotation, b as BalanceDruid_Rotation),
		rotationCopy: a => BalanceDruid_Rotation.clone(a as BalanceDruid_Rotation),
		rotationToJson: a => BalanceDruid_Rotation.toJson(a as BalanceDruid_Rotation),
		rotationFromJson: obj => BalanceDruid_Rotation.fromJson(obj),

		talentsCreate: () => DruidTalents.create(),
		talentsEquals: (a, b) => DruidTalents.equals(a as DruidTalents, b as DruidTalents),
		talentsCopy: a => DruidTalents.clone(a as DruidTalents),
		talentsToJson: a => DruidTalents.toJson(a as DruidTalents),
		talentsFromJson: obj => DruidTalents.fromJson(obj),

		optionsCreate: () => BalanceDruid_Options.create({ classOptions: {} }),
		optionsEquals: (a, b) => BalanceDruid_Options.equals(a as BalanceDruid_Options, b as BalanceDruid_Options),
		optionsCopy: a => BalanceDruid_Options.clone(a as BalanceDruid_Options),
		optionsToJson: a => BalanceDruid_Options.toJson(a as BalanceDruid_Options),
		optionsFromJson: obj => BalanceDruid_Options.fromJson(obj),
		optionsFromPlayer: player =>
			player.spec.oneofKind == 'balanceDruid'
				? player.spec.balanceDruid.options || BalanceDruid_Options.create()
				: BalanceDruid_Options.create({ classOptions: {} }),
	},
	[Spec.SpecFeralCatDruid]: {
		rotationCreate: () => FeralCatDruid_Rotation.create(),
		rotationEquals: (a, b) => FeralCatDruid_Rotation.equals(a as FeralCatDruid_Rotation, b as FeralCatDruid_Rotation),
		rotationCopy: a => FeralCatDruid_Rotation.clone(a as FeralCatDruid_Rotation),
		rotationToJson: a => FeralCatDruid_Rotation.toJson(a as FeralCatDruid_Rotation),
		rotationFromJson: obj => FeralCatDruid_Rotation.fromJson(obj),

		talentsCreate: () => DruidTalents.create(),
		talentsEquals: (a, b) => DruidTalents.equals(a as DruidTalents, b as DruidTalents),
		talentsCopy: a => DruidTalents.clone(a as DruidTalents),
		talentsToJson: a => DruidTalents.toJson(a as DruidTalents),
		talentsFromJson: obj => DruidTalents.fromJson(obj),

		optionsCreate: () => FeralCatDruid_Options.create({ classOptions: {} }),
		optionsEquals: (a, b) => FeralCatDruid_Options.equals(a as FeralCatDruid_Options, b as FeralCatDruid_Options),
		optionsCopy: a => FeralCatDruid_Options.clone(a as FeralCatDruid_Options),
		optionsToJson: a => FeralCatDruid_Options.toJson(a as FeralCatDruid_Options),
		optionsFromJson: obj => FeralCatDruid_Options.fromJson(obj),
		optionsFromPlayer: player =>
			player.spec.oneofKind == 'feralDruid'
				? player.spec.feralDruid.options || FeralCatDruid_Options.create()
				: FeralCatDruid_Options.create({ classOptions: {} }),
	},
	[Spec.SpecFeralBearDruid]: {
		rotationCreate: () => FeralBearDruid_Rotation.create(),
		rotationEquals: (a, b) => FeralBearDruid_Rotation.equals(a as FeralBearDruid_Rotation, b as FeralBearDruid_Rotation),
		rotationCopy: a => FeralBearDruid_Rotation.clone(a as FeralBearDruid_Rotation),
		rotationToJson: a => FeralBearDruid_Rotation.toJson(a as FeralBearDruid_Rotation),
		rotationFromJson: obj => FeralBearDruid_Rotation.fromJson(obj),

		talentsCreate: () => DruidTalents.create(),
		talentsEquals: (a, b) => DruidTalents.equals(a as DruidTalents, b as DruidTalents),
		talentsCopy: a => DruidTalents.clone(a as DruidTalents),
		talentsToJson: a => DruidTalents.toJson(a as DruidTalents),
		talentsFromJson: obj => DruidTalents.fromJson(obj),

		optionsCreate: () => FeralBearDruid_Options.create({ classOptions: {} }),
		optionsEquals: (a, b) => FeralBearDruid_Options.equals(a as FeralBearDruid_Options, b as FeralBearDruid_Options),
		optionsCopy: a => FeralBearDruid_Options.clone(a as FeralBearDruid_Options),
		optionsToJson: a => FeralBearDruid_Options.toJson(a as FeralBearDruid_Options),
		optionsFromJson: obj => FeralBearDruid_Options.fromJson(obj),
		optionsFromPlayer: player =>
			player.spec.oneofKind == 'guardianDruid'
				? player.spec.guardianDruid.options || FeralBearDruid_Options.create()
				: FeralBearDruid_Options.create({ classOptions: {} }),
	},
	[Spec.SpecRestorationDruid]: {
		rotationCreate: () => RestorationDruid_Rotation.create(),
		rotationEquals: (a, b) => RestorationDruid_Rotation.equals(a as RestorationDruid_Rotation, b as RestorationDruid_Rotation),
		rotationCopy: a => RestorationDruid_Rotation.clone(a as RestorationDruid_Rotation),
		rotationToJson: a => RestorationDruid_Rotation.toJson(a as RestorationDruid_Rotation),
		rotationFromJson: obj => RestorationDruid_Rotation.fromJson(obj),

		talentsCreate: () => DruidTalents.create(),
		talentsEquals: (a, b) => DruidTalents.equals(a as DruidTalents, b as DruidTalents),
		talentsCopy: a => DruidTalents.clone(a as DruidTalents),
		talentsToJson: a => DruidTalents.toJson(a as DruidTalents),
		talentsFromJson: obj => DruidTalents.fromJson(obj),

		optionsCreate: () => RestorationDruid_Options.create({ classOptions: {} }),
		optionsEquals: (a, b) => RestorationDruid_Options.equals(a as RestorationDruid_Options, b as RestorationDruid_Options),
		optionsCopy: a => RestorationDruid_Options.clone(a as RestorationDruid_Options),
		optionsToJson: a => RestorationDruid_Options.toJson(a as RestorationDruid_Options),
		optionsFromJson: obj => RestorationDruid_Options.fromJson(obj),
		optionsFromPlayer: player =>
			player.spec.oneofKind == 'restorationDruid'
				? player.spec.restorationDruid.options || RestorationDruid_Options.create()
				: RestorationDruid_Options.create({ classOptions: {} }),
	},
	// Hunter
	[Spec.SpecHunter]: {
		rotationCreate: () => Hunter_Rotation.create(),
		rotationEquals: (a, b) => Hunter_Rotation.equals(a as Hunter_Rotation, b as Hunter_Rotation),
		rotationCopy: a => Hunter_Rotation.clone(a as Hunter_Rotation),
		rotationToJson: a => Hunter_Rotation.toJson(a as Hunter_Rotation),
		rotationFromJson: obj => Hunter_Rotation.fromJson(obj),

		talentsCreate: () => HunterTalents.create(),
		talentsEquals: (a, b) => HunterTalents.equals(a as HunterTalents, b as HunterTalents),
		talentsCopy: a => HunterTalents.clone(a as HunterTalents),
		talentsToJson: a => HunterTalents.toJson(a as HunterTalents),
		talentsFromJson: obj => HunterTalents.fromJson(obj),

		optionsCreate: () => Hunter_Options.create({ classOptions: {} }),
		optionsEquals: (a, b) => Hunter_Options.equals(a as Hunter_Options, b as Hunter_Options),
		optionsCopy: a => Hunter_Options.clone(a as Hunter_Options),
		optionsToJson: a => Hunter_Options.toJson(a as Hunter_Options),
		optionsFromJson: obj => Hunter_Options.fromJson(obj),
		optionsFromPlayer: player =>
			player.spec.oneofKind == 'hunter'
				? player.spec.hunter.options || Hunter_Options.create()
				: Hunter_Options.create({ classOptions: {} }),
	},
	// Mage
	[Spec.SpecMage]: {
		rotationCreate: () => Mage_Rotation.create(),
		rotationEquals: (a, b) => Mage_Rotation.equals(a as Mage_Rotation, b as Mage_Rotation),
		rotationCopy: a => Mage_Rotation.clone(a as Mage_Rotation),
		rotationToJson: a => Mage_Rotation.toJson(a as Mage_Rotation),
		rotationFromJson: obj => Mage_Rotation.fromJson(obj),

		talentsCreate: () => MageTalents.create(),
		talentsEquals: (a, b) => MageTalents.equals(a as MageTalents, b as MageTalents),
		talentsCopy: a => MageTalents.clone(a as MageTalents),
		talentsToJson: a => MageTalents.toJson(a as MageTalents),
		talentsFromJson: obj => MageTalents.fromJson(obj),

		optionsCreate: () => Mage_Options.create({ classOptions: {} }),
		optionsEquals: (a, b) => Mage_Options.equals(a as Mage_Options, b as Mage_Options),
		optionsCopy: a => Mage_Options.clone(a as Mage_Options),
		optionsToJson: a => Mage_Options.toJson(a as Mage_Options),
		optionsFromJson: obj => Mage_Options.fromJson(obj),
		optionsFromPlayer: player =>
			player.spec.oneofKind == 'mage'
				? player.spec.mage.options || Mage_Options.create()
				: Mage_Options.create({ classOptions: {} }),
	},
	// Paladin
	[Spec.SpecHolyPaladin]: {
		rotationCreate: () => HolyPaladin_Rotation.create(),
		rotationEquals: (a, b) => HolyPaladin_Rotation.equals(a as HolyPaladin_Rotation, b as HolyPaladin_Rotation),
		rotationCopy: a => HolyPaladin_Rotation.clone(a as HolyPaladin_Rotation),
		rotationToJson: a => HolyPaladin_Rotation.toJson(a as HolyPaladin_Rotation),
		rotationFromJson: obj => HolyPaladin_Rotation.fromJson(obj),

		talentsCreate: () => PaladinTalents.create(),
		talentsEquals: (a, b) => PaladinTalents.equals(a as PaladinTalents, b as PaladinTalents),
		talentsCopy: a => PaladinTalents.clone(a as PaladinTalents),
		talentsToJson: a => PaladinTalents.toJson(a as PaladinTalents),
		talentsFromJson: obj => PaladinTalents.fromJson(obj),

		optionsCreate: () => HolyPaladin_Options.create({ classOptions: {} }),
		optionsEquals: (a, b) => HolyPaladin_Options.equals(a as HolyPaladin_Options, b as HolyPaladin_Options),
		optionsCopy: a => HolyPaladin_Options.clone(a as HolyPaladin_Options),
		optionsToJson: a => HolyPaladin_Options.toJson(a as HolyPaladin_Options),
		optionsFromJson: obj => HolyPaladin_Options.fromJson(obj),
		optionsFromPlayer: player =>
			player.spec.oneofKind == 'holyPaladin'
				? player.spec.holyPaladin.options || HolyPaladin_Options.create()
				: HolyPaladin_Options.create({ classOptions: {} }),
	},
	[Spec.SpecProtectionPaladin]: {
		rotationCreate: () => ProtectionPaladin_Rotation.create(),
		rotationEquals: (a, b) => ProtectionPaladin_Rotation.equals(a as ProtectionPaladin_Rotation, b as ProtectionPaladin_Rotation),
		rotationCopy: a => ProtectionPaladin_Rotation.clone(a as ProtectionPaladin_Rotation),
		rotationToJson: a => ProtectionPaladin_Rotation.toJson(a as ProtectionPaladin_Rotation),
		rotationFromJson: obj => ProtectionPaladin_Rotation.fromJson(obj),

		talentsCreate: () => PaladinTalents.create(),
		talentsEquals: (a, b) => PaladinTalents.equals(a as PaladinTalents, b as PaladinTalents),
		talentsCopy: a => PaladinTalents.clone(a as PaladinTalents),
		talentsToJson: a => PaladinTalents.toJson(a as PaladinTalents),
		talentsFromJson: obj => PaladinTalents.fromJson(obj),

		optionsCreate: () => ProtectionPaladin_Options.create({ classOptions: {} }),
		optionsEquals: (a, b) => ProtectionPaladin_Options.equals(a as ProtectionPaladin_Options, b as ProtectionPaladin_Options),
		optionsCopy: a => ProtectionPaladin_Options.clone(a as ProtectionPaladin_Options),
		optionsToJson: a => ProtectionPaladin_Options.toJson(a as ProtectionPaladin_Options),
		optionsFromJson: obj => ProtectionPaladin_Options.fromJson(obj),
		optionsFromPlayer: player =>
			player.spec.oneofKind == 'protectionPaladin'
				? player.spec.protectionPaladin.options || ProtectionPaladin_Options.create()
				: ProtectionPaladin_Options.create({ classOptions: {} }),
	},
	[Spec.SpecRetributionPaladin]: {
		rotationCreate: () => RetributionPaladin_Rotation.create(),
		rotationEquals: (a, b) => RetributionPaladin_Rotation.equals(a as RetributionPaladin_Rotation, b as RetributionPaladin_Rotation),
		rotationCopy: a => RetributionPaladin_Rotation.clone(a as RetributionPaladin_Rotation),
		rotationToJson: a => RetributionPaladin_Rotation.toJson(a as RetributionPaladin_Rotation),
		rotationFromJson: obj => RetributionPaladin_Rotation.fromJson(obj),

		talentsCreate: () => PaladinTalents.create(),
		talentsEquals: (a, b) => PaladinTalents.equals(a as PaladinTalents, b as PaladinTalents),
		talentsCopy: a => PaladinTalents.clone(a as PaladinTalents),
		talentsToJson: a => PaladinTalents.toJson(a as PaladinTalents),
		talentsFromJson: obj => PaladinTalents.fromJson(obj),

		optionsCreate: () => RetributionPaladin_Options.create({ classOptions: {} }),
		optionsEquals: (a, b) => RetributionPaladin_Options.equals(a as RetributionPaladin_Options, b as RetributionPaladin_Options),
		optionsCopy: a => RetributionPaladin_Options.clone(a as RetributionPaladin_Options),
		optionsToJson: a => RetributionPaladin_Options.toJson(a as RetributionPaladin_Options),
		optionsFromJson: obj => RetributionPaladin_Options.fromJson(obj),
		optionsFromPlayer: player =>
			player.spec.oneofKind == 'retributionPaladin'
				? player.spec.retributionPaladin.options || RetributionPaladin_Options.create()
				: RetributionPaladin_Options.create({ classOptions: {} }),
	},
	// Priest
	[Spec.SpecDisciplinePriest]: {
		rotationCreate: () => DisciplinePriest_Rotation.create(),
		rotationEquals: (a, b) => DisciplinePriest_Rotation.equals(a as DisciplinePriest_Rotation, b as DisciplinePriest_Rotation),
		rotationCopy: a => DisciplinePriest_Rotation.clone(a as DisciplinePriest_Rotation),
		rotationToJson: a => DisciplinePriest_Rotation.toJson(a as DisciplinePriest_Rotation),
		rotationFromJson: obj => DisciplinePriest_Rotation.fromJson(obj),

		talentsCreate: () => PriestTalents.create(),
		talentsEquals: (a, b) => PriestTalents.equals(a as PriestTalents, b as PriestTalents),
		talentsCopy: a => PriestTalents.clone(a as PriestTalents),
		talentsToJson: a => PriestTalents.toJson(a as PriestTalents),
		talentsFromJson: obj => PriestTalents.fromJson(obj),

		optionsCreate: () => DisciplinePriest_Options.create({ classOptions: {} }),
		optionsEquals: (a, b) => DisciplinePriest_Options.equals(a as DisciplinePriest_Options, b as DisciplinePriest_Options),
		optionsCopy: a => DisciplinePriest_Options.clone(a as DisciplinePriest_Options),
		optionsToJson: a => DisciplinePriest_Options.toJson(a as DisciplinePriest_Options),
		optionsFromJson: obj => DisciplinePriest_Options.fromJson(obj),
		optionsFromPlayer: player =>
			player.spec.oneofKind == 'disciplinePriest'
				? player.spec.disciplinePriest.options || DisciplinePriest_Options.create()
				: DisciplinePriest_Options.create({ classOptions: {} }),
	},
	[Spec.SpecHolyPriest]: {
		rotationCreate: () => HolyPriest_Rotation.create(),
		rotationEquals: (a, b) => HolyPriest_Rotation.equals(a as HolyPriest_Rotation, b as HolyPriest_Rotation),
		rotationCopy: a => HolyPriest_Rotation.clone(a as HolyPriest_Rotation),
		rotationToJson: a => HolyPriest_Rotation.toJson(a as HolyPriest_Rotation),
		rotationFromJson: obj => HolyPriest_Rotation.fromJson(obj),

		talentsCreate: () => PriestTalents.create(),
		talentsEquals: (a, b) => PriestTalents.equals(a as PriestTalents, b as PriestTalents),
		talentsCopy: a => PriestTalents.clone(a as PriestTalents),
		talentsToJson: a => PriestTalents.toJson(a as PriestTalents),
		talentsFromJson: obj => PriestTalents.fromJson(obj),

		optionsCreate: () => HolyPriest_Options.create({ classOptions: {} }),
		optionsEquals: (a, b) => HolyPriest_Options.equals(a as HolyPriest_Options, b as HolyPriest_Options),
		optionsCopy: a => HolyPriest_Options.clone(a as HolyPriest_Options),
		optionsToJson: a => HolyPriest_Options.toJson(a as HolyPriest_Options),
		optionsFromJson: obj => HolyPriest_Options.fromJson(obj),
		optionsFromPlayer: player =>
			player.spec.oneofKind == 'holyPriest'
				? player.spec.holyPriest.options || HolyPriest_Options.create()
				: HolyPriest_Options.create({ classOptions: {} }),
	},
	[Spec.SpecShadowPriest]: {
		rotationCreate: () => ShadowPriest_Rotation.create(),
		rotationEquals: (a, b) => ShadowPriest_Rotation.equals(a as ShadowPriest_Rotation, b as ShadowPriest_Rotation),
		rotationCopy: a => ShadowPriest_Rotation.clone(a as ShadowPriest_Rotation),
		rotationToJson: a => ShadowPriest_Rotation.toJson(a as ShadowPriest_Rotation),
		rotationFromJson: obj => ShadowPriest_Rotation.fromJson(obj),

		talentsCreate: () => PriestTalents.create(),
		talentsEquals: (a, b) => PriestTalents.equals(a as PriestTalents, b as PriestTalents),
		talentsCopy: a => PriestTalents.clone(a as PriestTalents),
		talentsToJson: a => PriestTalents.toJson(a as PriestTalents),
		talentsFromJson: obj => PriestTalents.fromJson(obj),

		optionsCreate: () => ShadowPriest_Options.create({ classOptions: {} }),
		optionsEquals: (a, b) => ShadowPriest_Options.equals(a as ShadowPriest_Options, b as ShadowPriest_Options),
		optionsCopy: a => ShadowPriest_Options.clone(a as ShadowPriest_Options),
		optionsToJson: a => ShadowPriest_Options.toJson(a as ShadowPriest_Options),
		optionsFromJson: obj => ShadowPriest_Options.fromJson(obj),
		optionsFromPlayer: player =>
			player.spec.oneofKind == 'shadowPriest'
				? player.spec.shadowPriest.options || ShadowPriest_Options.create()
				: ShadowPriest_Options.create({ classOptions: {} }),
	},
	// Rogue
	[Spec.SpecRogue]: {
		rotationCreate: () => Rogue_Rotation.create(),
		rotationEquals: (a, b) => Rogue_Rotation.equals(a as Rogue_Rotation, b as Rogue_Rotation),
		rotationCopy: a => Rogue_Rotation.clone(a as Rogue_Rotation),
		rotationToJson: a => Rogue_Rotation.toJson(a as Rogue_Rotation),
		rotationFromJson: obj => Rogue_Rotation.fromJson(obj),

		talentsCreate: () => RogueTalents.create(),
		talentsEquals: (a, b) => RogueTalents.equals(a as RogueTalents, b as RogueTalents),
		talentsCopy: a => RogueTalents.clone(a as RogueTalents),
		talentsToJson: a => RogueTalents.toJson(a as RogueTalents),
		talentsFromJson: obj => RogueTalents.fromJson(obj),

		optionsCreate: () => Rogue_Options.create({ classOptions: {} }),
		optionsEquals: (a, b) => Rogue_Options.equals(a as Rogue_Options, b as Rogue_Options),
		optionsCopy: a => Rogue_Options.clone(a as Rogue_Options),
		optionsToJson: a => Rogue_Options.toJson(a as Rogue_Options),
		optionsFromJson: obj => Rogue_Options.fromJson(obj),
		optionsFromPlayer: player =>
			player.spec.oneofKind == 'rogue'
				? player.spec.rogue.options || Rogue_Options.create()
				: Rogue_Options.create({ classOptions: {} }),
	},
	// Shaman
	[Spec.SpecElementalShaman]: {
		rotationCreate: () => ElementalShaman_Rotation.create(),
		rotationEquals: (a, b) => ElementalShaman_Rotation.equals(a as ElementalShaman_Rotation, b as ElementalShaman_Rotation),
		rotationCopy: a => ElementalShaman_Rotation.clone(a as ElementalShaman_Rotation),
		rotationToJson: a => ElementalShaman_Rotation.toJson(a as ElementalShaman_Rotation),
		rotationFromJson: obj => ElementalShaman_Rotation.fromJson(obj),

		talentsCreate: () => ShamanTalents.create(),
		talentsEquals: (a, b) => ShamanTalents.equals(a as ShamanTalents, b as ShamanTalents),
		talentsCopy: a => ShamanTalents.clone(a as ShamanTalents),
		talentsToJson: a => ShamanTalents.toJson(a as ShamanTalents),
		talentsFromJson: obj => ShamanTalents.fromJson(obj),

		optionsCreate: () => ElementalShaman_Options.create({ classOptions: {} }),
		optionsEquals: (a, b) => ElementalShaman_Options.equals(a as ElementalShaman_Options, b as ElementalShaman_Options),
		optionsCopy: a => ElementalShaman_Options.clone(a as ElementalShaman_Options),
		optionsToJson: a => ElementalShaman_Options.toJson(a as ElementalShaman_Options),
		optionsFromJson: obj => ElementalShaman_Options.fromJson(obj),
		optionsFromPlayer: player =>
			player.spec.oneofKind == 'elementalShaman'
				? player.spec.elementalShaman.options || ElementalShaman_Options.create()
				: ElementalShaman_Options.create({ classOptions: {} }),
	},
	[Spec.SpecEnhancementShaman]: {
		rotationCreate: () => EnhancementShaman_Rotation.create(),
		rotationEquals: (a, b) => EnhancementShaman_Rotation.equals(a as EnhancementShaman_Rotation, b as EnhancementShaman_Rotation),
		rotationCopy: a => EnhancementShaman_Rotation.clone(a as EnhancementShaman_Rotation),
		rotationToJson: a => EnhancementShaman_Rotation.toJson(a as EnhancementShaman_Rotation),
		rotationFromJson: obj => EnhancementShaman_Rotation.fromJson(obj),

		talentsCreate: () => ShamanTalents.create(),
		talentsEquals: (a, b) => ShamanTalents.equals(a as ShamanTalents, b as ShamanTalents),
		talentsCopy: a => ShamanTalents.clone(a as ShamanTalents),
		talentsToJson: a => ShamanTalents.toJson(a as ShamanTalents),
		talentsFromJson: obj => ShamanTalents.fromJson(obj),

		optionsCreate: () => EnhancementShaman_Options.create({ classOptions: {} }),
		optionsEquals: (a, b) => EnhancementShaman_Options.equals(a as EnhancementShaman_Options, b as EnhancementShaman_Options),
		optionsCopy: a => EnhancementShaman_Options.clone(a as EnhancementShaman_Options),
		optionsToJson: a => EnhancementShaman_Options.toJson(a as EnhancementShaman_Options),
		optionsFromJson: obj => EnhancementShaman_Options.fromJson(obj),
		optionsFromPlayer: player =>
			player.spec.oneofKind == 'enhancementShaman'
				? player.spec.enhancementShaman.options || EnhancementShaman_Options.create()
				: EnhancementShaman_Options.create({ classOptions: {} }),
	},
	[Spec.SpecRestorationShaman]: {
		rotationCreate: () => RestorationShaman_Rotation.create(),
		rotationEquals: (a, b) => RestorationShaman_Rotation.equals(a as RestorationShaman_Rotation, b as RestorationShaman_Rotation),
		rotationCopy: a => RestorationShaman_Rotation.clone(a as RestorationShaman_Rotation),
		rotationToJson: a => RestorationShaman_Rotation.toJson(a as RestorationShaman_Rotation),
		rotationFromJson: obj => RestorationShaman_Rotation.fromJson(obj),

		talentsCreate: () => ShamanTalents.create(),
		talentsEquals: (a, b) => ShamanTalents.equals(a as ShamanTalents, b as ShamanTalents),
		talentsCopy: a => ShamanTalents.clone(a as ShamanTalents),
		talentsToJson: a => ShamanTalents.toJson(a as ShamanTalents),
		talentsFromJson: obj => ShamanTalents.fromJson(obj),

		optionsCreate: () => RestorationShaman_Options.create({ classOptions: {} }),
		optionsEquals: (a, b) => RestorationShaman_Options.equals(a as RestorationShaman_Options, b as RestorationShaman_Options),
		optionsCopy: a => RestorationShaman_Options.clone(a as RestorationShaman_Options),
		optionsToJson: a => RestorationShaman_Options.toJson(a as RestorationShaman_Options),
		optionsFromJson: obj => RestorationShaman_Options.fromJson(obj),
		optionsFromPlayer: player =>
			player.spec.oneofKind == 'restorationShaman'
				? player.spec.restorationShaman.options || RestorationShaman_Options.create()
				: RestorationShaman_Options.create({ classOptions: {} }),
	},
	// Warlock
	[Spec.SpecWarlock]: {
		rotationCreate: () => Warlock_Rotation.create(),
		rotationEquals: (a, b) => Warlock_Rotation.equals(a as Warlock_Rotation, b as Warlock_Rotation),
		rotationCopy: a => Warlock_Rotation.clone(a as Warlock_Rotation),
		rotationToJson: a => Warlock_Rotation.toJson(a as Warlock_Rotation),
		rotationFromJson: obj => Warlock_Rotation.fromJson(obj),

		talentsCreate: () => WarlockTalents.create(),
		talentsEquals: (a, b) => WarlockTalents.equals(a as WarlockTalents, b as WarlockTalents),
		talentsCopy: a => WarlockTalents.clone(a as WarlockTalents),
		talentsToJson: a => WarlockTalents.toJson(a as WarlockTalents),
		talentsFromJson: obj => WarlockTalents.fromJson(obj),

		optionsCreate: () => Warlock_Options.create({ classOptions: {} }),
		optionsEquals: (a, b) => Warlock_Options.equals(a as Warlock_Options, b as Warlock_Options),
		optionsCopy: a => Warlock_Options.clone(a as Warlock_Options),
		optionsToJson: a => Warlock_Options.toJson(a as Warlock_Options),
		optionsFromJson: obj => Warlock_Options.fromJson(obj),
		optionsFromPlayer: player =>
			player.spec.oneofKind == 'warlock'
				? player.spec.warlock.options || Warlock_Options.create()
				: Warlock_Options.create({ classOptions: {} }),
	},
	// Warrior
	[Spec.SpecDPSWarrior]: {
		rotationCreate: () => DPSWarrior_Rotation.create(),
		rotationEquals: (a, b) => DPSWarrior_Rotation.equals(a as DPSWarrior_Rotation, b as DPSWarrior_Rotation),
		rotationCopy: a => DPSWarrior_Rotation.clone(a as DPSWarrior_Rotation),
		rotationToJson: a => DPSWarrior_Rotation.toJson(a as DPSWarrior_Rotation),
		rotationFromJson: obj => DPSWarrior_Rotation.fromJson(obj),

		talentsCreate: () => WarriorTalents.create(),
		talentsEquals: (a, b) => WarriorTalents.equals(a as WarriorTalents, b as WarriorTalents),
		talentsCopy: a => WarriorTalents.clone(a as WarriorTalents),
		talentsToJson: a => WarriorTalents.toJson(a as WarriorTalents),
		talentsFromJson: obj => WarriorTalents.fromJson(obj),

		optionsCreate: () => DPSWarrior_Options.create({ classOptions: {} }),
		optionsEquals: (a, b) => DPSWarrior_Options.equals(a as DPSWarrior_Options, b as DPSWarrior_Options),
		optionsCopy: a => DPSWarrior_Options.clone(a as DPSWarrior_Options),
		optionsToJson: a => DPSWarrior_Options.toJson(a as DPSWarrior_Options),
		optionsFromJson: obj => DPSWarrior_Options.fromJson(obj),
		optionsFromPlayer: player =>
			player.spec.oneofKind == 'dpsWarrior'
				? player.spec.dpsWarrior.options || DPSWarrior_Options.create()
				: DPSWarrior_Options.create({ classOptions: {} }),
	},
	[Spec.SpecProtectionWarrior]: {
		rotationCreate: () => ProtectionWarrior_Rotation.create(),
		rotationEquals: (a, b) => ProtectionWarrior_Rotation.equals(a as ProtectionWarrior_Rotation, b as ProtectionWarrior_Rotation),
		rotationCopy: a => ProtectionWarrior_Rotation.clone(a as ProtectionWarrior_Rotation),
		rotationToJson: a => ProtectionWarrior_Rotation.toJson(a as ProtectionWarrior_Rotation),
		rotationFromJson: obj => ProtectionWarrior_Rotation.fromJson(obj),

		talentsCreate: () => WarriorTalents.create(),
		talentsEquals: (a, b) => WarriorTalents.equals(a as WarriorTalents, b as WarriorTalents),
		talentsCopy: a => WarriorTalents.clone(a as WarriorTalents),
		talentsToJson: a => WarriorTalents.toJson(a as WarriorTalents),
		talentsFromJson: obj => WarriorTalents.fromJson(obj),

		optionsCreate: () => ProtectionWarrior_Options.create({ classOptions: {} }),
		optionsEquals: (a, b) => ProtectionWarrior_Options.equals(a as ProtectionWarrior_Options, b as ProtectionWarrior_Options),
		optionsCopy: a => ProtectionWarrior_Options.clone(a as ProtectionWarrior_Options),
		optionsToJson: a => ProtectionWarrior_Options.toJson(a as ProtectionWarrior_Options),
		optionsFromJson: obj => ProtectionWarrior_Options.fromJson(obj),
		optionsFromPlayer: player =>
			player.spec.oneofKind == 'protectionWarrior'
				? player.spec.protectionWarrior.options || ProtectionWarrior_Options.create()
				: ProtectionWarrior_Options.create(),
	},
};

export const raceToFaction: Record<Race, Faction> = {
	[Race.RaceUnknown]: Faction.Unknown,

	[Race.RaceDraenei]: Faction.Alliance,
	[Race.RaceDwarf]: Faction.Alliance,
	[Race.RaceGnome]: Faction.Alliance,
	[Race.RaceHuman]: Faction.Alliance,
	[Race.RaceNightElf]: Faction.Alliance,

	[Race.RaceBloodElf]: Faction.Horde,
	[Race.RaceOrc]: Faction.Horde,
	[Race.RaceTauren]: Faction.Horde,
	[Race.RaceTroll]: Faction.Horde,
	[Race.RaceUndead]: Faction.Horde,
};

// Returns a copy of playerOptions, with the class field set.
export function withSpec<SpecType extends Spec>(spec: Spec, player: Player, specOptions: SpecOptions<SpecType>): Player {
	const copy = Player.clone(player);

	switch (spec) {
		// Druid
		case Spec.SpecBalanceDruid:
			copy.spec = {
				oneofKind: 'balanceDruid',
				balanceDruid: BalanceDruid.create({
					options: specOptions as BalanceDruid_Options,
				}),
			};
			return copy;
		case Spec.SpecFeralCatDruid:
			copy.spec = {
				oneofKind: 'feralDruid',
				feralDruid: FeralCatDruid.create({
					options: specOptions as FeralCatDruid_Options,
				}),
			};
			return copy;
		case Spec.SpecFeralBearDruid:
			copy.spec = {
				oneofKind: 'guardianDruid',
				guardianDruid: FeralBearDruid.create({
					options: specOptions as FeralBearDruid_Options,
				}),
			};
			return copy;
		case Spec.SpecRestorationDruid:
			copy.spec = {
				oneofKind: 'restorationDruid',
				restorationDruid: RestorationDruid.create({
					options: specOptions as RestorationDruid_Options,
				}),
			};
			return copy;
		// Hunter
		case Spec.SpecHunter:
			copy.spec = {
				oneofKind: 'hunter',
				hunter: Hunter.create({
					options: specOptions as Hunter_Options,
				}),
			};
			return copy;
		// Mage
		case Spec.SpecMage:
			copy.spec = {
				oneofKind: 'mage',
				mage: Mage.create({
					options: specOptions as Mage_Options,
				}),
			};
			return copy;
		// Paladin
		case Spec.SpecHolyPaladin:
			copy.spec = {
				oneofKind: 'holyPaladin',
				holyPaladin: HolyPaladin.create({
					options: specOptions as HolyPaladin_Options,
				}),
			};
			return copy;
		case Spec.SpecProtectionPaladin:
			copy.spec = {
				oneofKind: 'protectionPaladin',
				protectionPaladin: ProtectionPaladin.create({
					options: specOptions as ProtectionPaladin_Options,
				}),
			};
			return copy;
		case Spec.SpecRetributionPaladin:
			copy.spec = {
				oneofKind: 'retributionPaladin',
				retributionPaladin: RetributionPaladin.create({
					options: specOptions as RetributionPaladin_Options,
				}),
			};
			return copy;
		// Priest
		case Spec.SpecDisciplinePriest:
			copy.spec = {
				oneofKind: 'disciplinePriest',
				disciplinePriest: DisciplinePriest.create({
					options: specOptions as DisciplinePriest_Options,
				}),
			};
			return copy;
		case Spec.SpecHolyPriest:
			copy.spec = {
				oneofKind: 'holyPriest',
				holyPriest: HolyPriest.create({
					options: specOptions as HolyPriest_Options,
				}),
			};
			return copy;
		case Spec.SpecShadowPriest:
			copy.spec = {
				oneofKind: 'shadowPriest',
				shadowPriest: ShadowPriest.create({
					options: specOptions as ShadowPriest_Options,
				}),
			};
			return copy;
		// Rogue
		case Spec.SpecRogue:
			copy.spec = {
				oneofKind: 'rogue',
				rogue: Rogue.create({
					options: specOptions as Rogue_Options,
				}),
			};
			return copy;
		// Shaman
		case Spec.SpecElementalShaman:
			copy.spec = {
				oneofKind: 'elementalShaman',
				elementalShaman: ElementalShaman.create({
					options: specOptions as ElementalShaman_Options,
				}),
			};
			return copy;
		case Spec.SpecEnhancementShaman:
			copy.spec = {
				oneofKind: 'enhancementShaman',
				enhancementShaman: EnhancementShaman.create({
					options: specOptions as EnhancementShaman_Options,
				}),
			};
			return copy;
		case Spec.SpecRestorationShaman:
			copy.spec = {
				oneofKind: 'restorationShaman',
				restorationShaman: RestorationShaman.create({
					options: specOptions as RestorationShaman_Options,
				}),
			};
			return copy;
		// Warlock
		case Spec.SpecWarlock:
			copy.spec = {
				oneofKind: 'warlock',
				warlock: Warlock.create({
					options: specOptions as Warlock_Options,
				}),
			};
			return copy;
		// Warrior
		case Spec.SpecDPSWarrior:
			copy.spec = {
				oneofKind: 'dpsWarrior',
				dpsWarrior: DPSWarrior.create({
					options: specOptions as DPSWarrior_Options,
				}),
			};
			return copy;
		case Spec.SpecProtectionWarrior:
			copy.spec = {
				oneofKind: 'protectionWarrior',
				protectionWarrior: ProtectionWarrior.create({
					options: specOptions as ProtectionWarrior_Options,
				}),
			};
			return copy;
		default:
			return copy;
	}
}

export function getPlayerSpecFromPlayer<SpecType extends Spec>(player: Player): PlayerSpec<SpecType> {
	const specValues = getEnumValues(Spec);
	for (let i = 0; i < specValues.length; i++) {
		const spec = specValues[i] as SpecType;
		let specString = Spec[spec]; // Returns 'SpecBalanceDruid' for BalanceDruid.
		specString = specString.substring('Spec'.length); // 'BalanceDruid'
		specString = specString.charAt(0).toLowerCase() + specString.slice(1); // 'balanceDruid'

		if (player.spec.oneofKind == specString) {
			return PlayerSpecs.fromProto(spec);
		}
	}

	throw new Error('Unable to parse spec from player proto: ' + JSON.stringify(Player.toJson(player), null, 2));
}

export function isSharpWeaponType(weaponType: WeaponType): boolean {
	return [WeaponType.WeaponTypeAxe, WeaponType.WeaponTypeDagger, WeaponType.WeaponTypePolearm, WeaponType.WeaponTypeSword].includes(weaponType);
}

export function isBluntWeaponType(weaponType: WeaponType): boolean {
	return [WeaponType.WeaponTypeFist, WeaponType.WeaponTypeMace, WeaponType.WeaponTypeStaff].includes(weaponType);
}

// Custom functions for determining the EP value of meta gem effects.
// Default meta effect EP value is 0, so just handle the ones relevant to your spec.
const metaGemEffectEPs: Partial<Record<Spec, (gem: Gem, playerStats: Stats) => number>> = {};

export function getMetaGemEffectEP<SpecType extends Spec>(playerSpec: PlayerSpec<SpecType>, gem: Gem, playerStats: Stats) {
	if (metaGemEffectEPs[playerSpec.specID]) {
		return metaGemEffectEPs[playerSpec.specID]!(gem, playerStats);
	} else {
		return 0;
	}
}

// Returns true if this item may be equipped in at least 1 slot for the given Spec.
export function canEquipItem<SpecType extends Spec>(item: Item, playerSpec: PlayerSpec<SpecType>, slot: ItemSlot | undefined): boolean {
	const playerClass = PlayerSpecs.getPlayerClass(playerSpec);
	if (item.classAllowlist.length > 0 && !item.classAllowlist.includes(playerClass.classID)) {
		return false;
	}

	if ([ItemType.ItemTypeFinger, ItemType.ItemTypeTrinket].includes(item.type)) {
		return true;
	}

	if (item.type == ItemType.ItemTypeWeapon) {
		const eligibleWeaponType = playerClass.weaponTypes.find(wt => wt.weaponType == item.weaponType);
		if (!eligibleWeaponType) {
			return false;
		}

		if (
			(item.handType == HandType.HandTypeOffHand || (item.handType == HandType.HandTypeOneHand && slot == ItemSlot.ItemSlotOffHand)) &&
			![WeaponType.WeaponTypeShield, WeaponType.WeaponTypeOffHand].includes(item.weaponType) &&
			!playerSpec.canDualWield
		) {
			return false;
		}

		if (item.handType == HandType.HandTypeTwoHand && !eligibleWeaponType.canUseTwoHand) {
			return false;
		}
		if (item.handType == HandType.HandTypeTwoHand && slot == ItemSlot.ItemSlotOffHand && playerSpec.specID != Spec.SpecDPSWarrior) {
			return false;
		}

		return true;
	}

	if (item.type == ItemType.ItemTypeRanged) {
		return playerClass.rangedWeaponTypes.includes(item.rangedWeaponType);
	}

	// At this point, we know the item is an armor piece (feet, chest, legs, etc).
	return playerClass.armorTypes[0] >= item.armorType;
}

const pvpSeasonFromName: Record<string, string> = {
	Wrathful: 'Season 8',
	Bloodthirsty: 'Season 8.5',
	Vicious: 'Season 9',
	Ruthless: 'Season 10',
	Cataclysmic: 'Season 11',
};

export const isPVPItem = (item: Item) => item?.name?.includes('Gladiator') || false;

export const getPVPSeasonFromItem = (item: Item) => {
	const seasonName = item.name.substring(0, item.name.indexOf(' '));
	return pvpSeasonFromName[seasonName] || undefined;
};

const itemTypeToSlotsMap: Partial<Record<ItemType, Array<ItemSlot>>> = {
	[ItemType.ItemTypeUnknown]: [],
	[ItemType.ItemTypeHead]: [ItemSlot.ItemSlotHead],
	[ItemType.ItemTypeNeck]: [ItemSlot.ItemSlotNeck],
	[ItemType.ItemTypeShoulder]: [ItemSlot.ItemSlotShoulder],
	[ItemType.ItemTypeBack]: [ItemSlot.ItemSlotBack],
	[ItemType.ItemTypeChest]: [ItemSlot.ItemSlotChest],
	[ItemType.ItemTypeWrist]: [ItemSlot.ItemSlotWrist],
	[ItemType.ItemTypeHands]: [ItemSlot.ItemSlotHands],
	[ItemType.ItemTypeWaist]: [ItemSlot.ItemSlotWaist],
	[ItemType.ItemTypeLegs]: [ItemSlot.ItemSlotLegs],
	[ItemType.ItemTypeFeet]: [ItemSlot.ItemSlotFeet],
	[ItemType.ItemTypeFinger]: [ItemSlot.ItemSlotFinger1, ItemSlot.ItemSlotFinger2],
	[ItemType.ItemTypeTrinket]: [ItemSlot.ItemSlotTrinket1, ItemSlot.ItemSlotTrinket2],
	[ItemType.ItemTypeRanged]: [ItemSlot.ItemSlotRanged],
};

export function getEligibleItemSlots(item: Item, isFuryWarrior?: boolean): Array<ItemSlot> {
	if (itemTypeToSlotsMap[item.type]) {
		return itemTypeToSlotsMap[item.type]!;
	}

	if (item.type == ItemType.ItemTypeWeapon) {
		if (isFuryWarrior) {
			return [ItemSlot.ItemSlotMainHand, ItemSlot.ItemSlotOffHand];
		}

		if (item.handType == HandType.HandTypeMainHand) {
			return [ItemSlot.ItemSlotMainHand];
		} else if (item.handType == HandType.HandTypeOffHand) {
			return [ItemSlot.ItemSlotOffHand];
		} else {
			return [ItemSlot.ItemSlotMainHand, ItemSlot.ItemSlotOffHand];
		}
	}

	// Should never reach here
	throw new Error('Could not find item slots for item: ' + Item.toJsonString(item));
}

export const isSecondaryItemSlot = (slot: ItemSlot) => slot === ItemSlot.ItemSlotFinger2 || slot === ItemSlot.ItemSlotTrinket2;

// Returns whether the given main-hand and off-hand items can be worn at the
// same time.
export function validWeaponCombo(mainHand: Item | null | undefined, offHand: Item | null | undefined, canDW2h: boolean): boolean {
	if (mainHand?.handType == HandType.HandTypeTwoHand && !canDW2h) {
		return false;
	} else if (mainHand?.handType == HandType.HandTypeTwoHand && mainHand?.weaponType == WeaponType.WeaponTypeStaff) {
		return false;
	}
	if (offHand?.handType == HandType.HandTypeTwoHand && !canDW2h) {
		return false;
	} else if (offHand?.handType == HandType.HandTypeTwoHand && offHand?.weaponType == WeaponType.WeaponTypeStaff) {
		return false;
	}

	return true;
}

// Returns all item slots to which the enchant might be applied.
//
// Note that this alone is not enough; some items have further restrictions,
// e.g. some weapon enchants may only be applied to 2H weapons.
export function getEligibleEnchantSlots(enchant: Enchant): Array<ItemSlot> {
	return [enchant.type]
		.concat(enchant.extraTypes || [])
		.map(type => {
			if (itemTypeToSlotsMap[type]) {
				return itemTypeToSlotsMap[type]!;
			}

			if (type == ItemType.ItemTypeWeapon) {
				return [ItemSlot.ItemSlotMainHand, ItemSlot.ItemSlotOffHand];
			}

			// Should never reach here
			throw new Error('Could not find item slots for enchant: ' + Enchant.toJsonString(enchant));
		})
		.flat();
}

export function enchantAppliesToItem(enchant: Enchant, item: Item): boolean {
	const sharedSlots = intersection(getEligibleEnchantSlots(enchant), getEligibleItemSlots(item));
	if (!sharedSlots.length) return false;

	if (enchant.enchantType === EnchantType.EnchantTypeTwoHand && item.handType !== HandType.HandTypeTwoHand) return false;

	if (enchant.enchantType === EnchantType.EnchantTypeStaff && item.weaponType !== WeaponType.WeaponTypeStaff) return false;

	if (enchant.enchantType === EnchantType.EnchantTypeShield && item.weaponType !== WeaponType.WeaponTypeShield) return false;

	if (
		(enchant.enchantType === EnchantType.EnchantTypeOffHand) !==
		(item.weaponType === WeaponType.WeaponTypeOffHand ||
			// All off-hand enchants can be applied to shields as well
			(item.weaponType === WeaponType.WeaponTypeShield && enchant.enchantType !== EnchantType.EnchantTypeShield))
	)
		return false;

	if (enchant.type == ItemType.ItemTypeRanged) {
		if (
			![RangedWeaponType.RangedWeaponTypeBow, RangedWeaponType.RangedWeaponTypeCrossbow, RangedWeaponType.RangedWeaponTypeGun].includes(
				item.rangedWeaponType,
			)
		)
			return false;
	}

	if (item.rangedWeaponType != RangedWeaponType.RangedWeaponTypeWand && item.rangedWeaponType > 0 && enchant.type != ItemType.ItemTypeRanged) {
		return false;
	}

	return true;
}

export function canEquipEnchant<SpecType extends Spec>(enchant: Enchant, playerSpec: PlayerSpec<SpecType>): boolean {
	if (enchant.classAllowlist.length > 0 && !enchant.classAllowlist.includes(playerSpec.classID)) {
		return false;
	}

	// This is a Tinker and we handle them differently
	if (enchant.requiredProfession == Profession.Engineering) {
		return false;
	}

	return true;
}

export function newUnitReference(raidIndex: number): UnitReference {
	return UnitReference.create({
		type: UnitReference_Type.Player,
		index: raidIndex,
	});
}

export function emptyUnitReference(): UnitReference {
	return UnitReference.create();
}

// Makes a new set of assignments with everything 0'd out.
export function makeBlankBlessingsAssignments(numPaladins: number): BlessingsAssignments {
	const assignments = BlessingsAssignments.create();
	for (let i = 0; i < numPaladins; i++) {
		assignments.paladins.push(
			BlessingsAssignment.create({
				blessings: new Array(NUM_SPECS).fill(Blessings.BlessingUnknown),
			}),
		);
	}
	return assignments;
}

export function makeBlessingsAssignments(numPaladins: number): BlessingsAssignments {
	const assignments = makeBlankBlessingsAssignments(numPaladins);
	for (let i = 1; i < Object.keys(Spec).length; i++) {
		const spec = i;
		const blessings = [Blessings.BlessingOfKings, Blessings.BlessingOfMight];
		for (let j = 0; j < blessings.length; j++) {
			if (j >= assignments.paladins.length) {
				// Can't assign more blessings since we ran out of paladins
				break;
			}
			assignments.paladins[j].blessings[spec] = blessings[j];
		}
	}
	return assignments;
}

// Default blessings settings in the raid sim UI.
export function makeDefaultBlessings(numPaladins: number): BlessingsAssignments {
	return makeBlessingsAssignments(numPaladins);
}

export const orderedResourceTypes: Array<ResourceType> = [
	ResourceType.ResourceTypeHealth,
	ResourceType.ResourceTypeMana,
	ResourceType.ResourceTypeEnergy,
	ResourceType.ResourceTypeRage,
	ResourceType.ResourceTypeComboPoints,
	ResourceType.ResourceTypeFocus,
	ResourceType.ResourceTypeGenericResource,
];

export const AL_CATEGORY_HARD_MODE = 'Hard Mode';
export const AL_CATEGORY_TITAN_RUNE = 'Titan Rune';

export const defaultRaidBuffMajorDamageCooldowns = (classID?: Class): Partial<RaidBuffs> => {
	return RaidBuffs.create({
		//skullBannerCount: classID == Class.ClassWarrior ? 1 : 2,
		//stormlashTotemCount: classID == Class.ClassShaman ? 3 : 4,
	});
};

// Utilities for migrating protos between versions

// Each key is an API version, each value is a function that up-converts a proto
// to that version from the previous one. If there are missing keys between
// successive entries, then it is assumed that no intermediate conversions are
// required (i.e. the intermediate version changes did not affect this
// particular proto).
export type ProtoConversionMap<Type> = Map<number, (arg: Type) => Type>;

export function migrateOldProto<Type>(oldProto: Type, oldApiVersion: number, conversionMap: ProtoConversionMap<Type>, targetApiVersion?: number): Type {
	let migratedProto = oldProto;
	const finalVersion = targetApiVersion || CURRENT_API_VERSION;
	for (let nextVersion = oldApiVersion + 1; nextVersion <= finalVersion; nextVersion++) {
		if (conversionMap.has(nextVersion)) {
			migratedProto = conversionMap.get(nextVersion)!(migratedProto);
		}
	}

	return migratedProto;
}
