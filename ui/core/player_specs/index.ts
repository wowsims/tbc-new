import { translatePlayerClass, translatePlayerSpec } from '../../i18n/localization';
import { LOCAL_STORAGE_PREFIX } from '../constants/other';
import { PlayerClass } from '../player_class';
import { PlayerClasses } from '../player_classes';
import { PlayerSpec } from '../player_spec';
import { Spec } from '../proto/common';
import { SpecClasses } from '../proto_utils/utils';
import * as DruidSpecs from './druid';
import * as HunterSpecs from './hunter';
import * as MageSpecs from './mage';
import * as PaladinSpecs from './paladin';
import * as PriestSpecs from './priest';
import * as RogueSpecs from './rogue';
import * as ShamanSpecs from './shaman';
import * as WarlockSpecs from './warlock';
import * as WarriorSpecs from './warrior';

const specToPlayerSpec: Record<Spec, PlayerSpec<any> | undefined> = {
	[Spec.SpecUnknown]: undefined,
	// Druid
	[Spec.SpecBalanceDruid]: DruidSpecs.BalanceDruid,
	[Spec.SpecFeralCatDruid]: DruidSpecs.FeralCatDruid,
	[Spec.SpecFeralBearDruid]: DruidSpecs.FeralBearDruid,
	[Spec.SpecRestorationDruid]: DruidSpecs.RestorationDruid,
	// Hunter
	[Spec.SpecHunter]: HunterSpecs.Hunter,
	// Mage
	[Spec.SpecMage]: MageSpecs.Mage,
	// Paladin
	[Spec.SpecHolyPaladin]: PaladinSpecs.HolyPaladin,
	[Spec.SpecProtectionPaladin]: PaladinSpecs.ProtectionPaladin,
	[Spec.SpecRetributionPaladin]: PaladinSpecs.RetributionPaladin,
	// Priest
	[Spec.SpecDisciplinePriest]: PriestSpecs.DisciplinePriest,
	[Spec.SpecHolyPriest]: PriestSpecs.HolyPriest,
	[Spec.SpecShadowPriest]: PriestSpecs.ShadowPriest,
	// Rogue
	[Spec.SpecRogue]: RogueSpecs.Rogue,
	// Shaman
	[Spec.SpecElementalShaman]: ShamanSpecs.ElementalShaman,
	[Spec.SpecEnhancementShaman]: ShamanSpecs.EnhancementShaman,
	[Spec.SpecRestorationShaman]: ShamanSpecs.RestorationShaman,
	// Warlock
	[Spec.SpecWarlock]: WarlockSpecs.Warlock,
	// Warrior
	[Spec.SpecDPSWarrior]: WarriorSpecs.DPSWarrior,
	[Spec.SpecProtectionWarrior]: WarriorSpecs.ProtectionWarrior,
};

const getPlayerClass = <SpecType extends Spec>(playerSpec: PlayerSpec<SpecType>): PlayerClass<SpecClasses<SpecType>> => {
	if (playerSpec.specID == Spec.SpecUnknown) {
		throw new Error('Invalid Spec');
	}

	return PlayerClasses.fromProto(playerSpec.classID);
};

export const PlayerSpecs = {
	...DruidSpecs,
	...HunterSpecs,
	...MageSpecs,
	...PaladinSpecs,
	...PriestSpecs,
	...RogueSpecs,
	...ShamanSpecs,
	...WarlockSpecs,
	...WarriorSpecs,
	getPlayerClass,
	getFullSpecName: <SpecType extends Spec>(playerSpec: PlayerSpec<SpecType>): string => {
		const translatedSpec = translatePlayerSpec(playerSpec);
		const translatedClass = translatePlayerClass(getPlayerClass(playerSpec));
		return `${translatedSpec} ${translatedClass}`;
	},
	getSpecNumber: <SpecType extends Spec>(playerSpec: PlayerSpec<SpecType>): number => {
		return Object.values(getPlayerClass(playerSpec).specs).findIndex(spec => spec == playerSpec) ?? 0;
	},
	// Prefixes used for storing browser data for each site. Even if a Spec is
	// renamed, DO NOT change these values or people will lose their saved data.
	getLocalStorageKey: <SpecType extends Spec>(playerSpec: PlayerSpec<SpecType>): string => {
		return `${LOCAL_STORAGE_PREFIX}_${playerSpec.friendlyName.toLowerCase().replace(/\s/, '_')}_${getPlayerClass(playerSpec)
			.friendlyName.toLowerCase()
			.replace(/\s/, '_')}`;
	},
	fromProto: <SpecType extends Spec>(spec: SpecType): PlayerSpec<SpecType> => {
		if (spec == Spec.SpecUnknown) {
			throw new Error('Invalid Spec');
		}

		return specToPlayerSpec[spec] as PlayerSpec<SpecType>;
	},
};
