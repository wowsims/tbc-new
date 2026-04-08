import { Phase } from './constants/other';
import { Player } from './player';
import { Spec } from './proto/common';

// This file is for anything related to launching a new sim. DO NOT touch this
// file until your sim is ready to launch!

export enum LaunchStatus {
	Unlaunched,
	Alpha,
	Beta,
	Launched,
}

export type SimStatus = {
	phase: Phase;
	status: LaunchStatus;
	oldSimLink?: string;
};

export const raidSimStatus: SimStatus = {
	phase: Phase.Phase1,
	status: LaunchStatus.Unlaunched,
	oldSimLink: 'https://wowsims.github.io/tbc/raid/',
};

// This list controls which links are shown in the top-left dropdown menu.
export const simLaunchStatuses: Record<Spec, SimStatus> = {
	[Spec.SpecUnknown]: {
		phase: Phase.Phase1,
		status: LaunchStatus.Unlaunched,
	},
	// Druid
	[Spec.SpecBalanceDruid]: {
		phase: Phase.Phase1,
		status: LaunchStatus.Alpha,
		oldSimLink: 'https://wowsims.github.io/tbc/balance_druid/',
	},
	[Spec.SpecFeralCatDruid]: {
		phase: Phase.Phase1,
		status: LaunchStatus.Alpha,
		oldSimLink: 'https://wowsims.github.io/tbc/feral_druid/',
	},
	[Spec.SpecFeralBearDruid]: {
		phase: Phase.Phase1,
		status: LaunchStatus.Alpha,
		oldSimLink: 'https://wowsims.github.io/tbc/feral_tank_druid/',
	},
	[Spec.SpecRestorationDruid]: {
		phase: Phase.Phase1,
		status: LaunchStatus.Unlaunched,
	},
	// Hunter
	[Spec.SpecHunter]: {
		phase: Phase.Phase1,
		status: LaunchStatus.Alpha,
	},
	// Mage
	[Spec.SpecMage]: {
		phase: Phase.Phase1,
		status: LaunchStatus.Alpha,
	},
	// Paladin
	[Spec.SpecHolyPaladin]: {
		phase: Phase.Phase1,
		status: LaunchStatus.Unlaunched,
	},
	[Spec.SpecProtectionPaladin]: {
		phase: Phase.Phase1,
		status: LaunchStatus.Unlaunched,
		oldSimLink: 'https://wowsims.github.io/tbc/protection_paladin/',
	},
	[Spec.SpecRetributionPaladin]: {
		phase: Phase.Phase1,
		status: LaunchStatus.Alpha,
	},
	// Priest
	[Spec.SpecPriest]: {
		phase: Phase.Phase1,
		status: LaunchStatus.Alpha,
	},
	// Rogue
	[Spec.SpecRogue]: {
		phase: Phase.Phase1,
		status: LaunchStatus.Alpha,
	},
	// Shaman
	[Spec.SpecElementalShaman]: {
		phase: Phase.Phase1,
		status: LaunchStatus.Alpha,
	},
	[Spec.SpecEnhancementShaman]: {
		phase: Phase.Phase1,
		status: LaunchStatus.Alpha,
	},
	[Spec.SpecRestorationShaman]: {
		phase: Phase.Phase1,
		status: LaunchStatus.Unlaunched,
	},
	// Warlock
	[Spec.SpecWarlock]: {
		phase: Phase.Phase1,
		status: LaunchStatus.Alpha,
	},
	// Warrior
	[Spec.SpecDpsWarrior]: {
		phase: Phase.Phase1,
		status: LaunchStatus.Alpha,
	},
	[Spec.SpecProtectionWarrior]: {
		phase: Phase.Phase1,
		status: LaunchStatus.Alpha,
	},
};

export const getSpecLaunchStatus = (player: Player<any>) => simLaunchStatuses[player.getSpec() as Spec].status;
