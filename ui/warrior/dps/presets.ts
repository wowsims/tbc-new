import { Player } from '../../core/player';
import * as PresetUtils from '../../core/preset_utils';
import { ConsumesSpec, Debuffs, HandType, ItemSlot, PartyBuffs, Profession, PseudoStat, Race, Spec, Stat, TristateEffect } from '../../core/proto/common';
import { SavedTalents } from '../../core/proto/ui';
import {
	DpsWarriorSpec,
	DpsWarrior_Rotation,
	DpsWarrior_Options as WarriorOptions,
	WarriorShout,
	WarriorStance,
	WarriorSunder,
} from '../../core/proto/warrior';
import { Stats } from '../../core/proto_utils/stats';
import * as WarriorPresets from '../presets';
import DefaultArmsApl from './apls/arms.apl.json';
import DefaultFuryApl from './apls/fury.apl.json';
import PreraidArmsGear from './gear_sets/preraid_arms.gear.json';
import P1ArmsGear from './gear_sets/p1_arms.gear.json';
import P2ArmsGear from './gear_sets/p2_arms.gear.json';
import P3ArmsGear from './gear_sets/p3_arms.gear.json';
import P35ArmsGear from './gear_sets/p3.5_arms.gear.json';
import P4ArmsGear from './gear_sets/p4_arms.gear.json';
import PreraidFuryGear from './gear_sets/preraid_fury.gear.json';
import P1FuryGear from './gear_sets/p1_fury.gear.json';
import P2FuryGear from './gear_sets/p2_fury.gear.json';
import P3FuryGear from './gear_sets/p3_fury.gear.json';
import P35FuryGear from './gear_sets/p3.5_fury.gear.json';
import P4FuryGear from './gear_sets/p4_fury.gear.json';
import { Phase } from '../../core/constants/other';
import { defaultExposeWeaknessSettings } from '../../core/proto_utils/utils';
import { APLRotation_Type } from '../../core/proto/apl';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

export const isArmsSpec = (player: Player<Spec.SpecDpsWarrior>) =>
	player.getEquippedItem(ItemSlot.ItemSlotMainHand)?.item.handType === HandType.HandTypeTwoHand;

export const isArmsKebabSpec = (player: Player<Spec.SpecDpsWarrior>) => player.getTalents().mortalStrike && isFurySpec(player);

export const isFurySpec = (player: Player<Spec.SpecDpsWarrior>) =>
	player.getTalents().bloodthirst ||
	player.getEquippedItem(ItemSlot.ItemSlotMainHand)?.item.handType === HandType.HandTypeMainHand ||
	player.getEquippedItem(ItemSlot.ItemSlotMainHand)?.item.handType === HandType.HandTypeOneHand;

// Handlers for spec specific load checks
const FURY_PRESET_OPTIONS = {
	onLoad: (player: Player<Spec.SpecDpsWarrior>) => {
		PresetUtils.makeSpecChangeWarningToast(
			[
				{
					condition: isArmsSpec,
					message: 'Check your gear: You have a two-handed weapon equipped, but the selected option is for dual wield.',
				},
				{
					condition: (player: Player<Spec.SpecDpsWarrior>) => !player.getTalents().dualWieldSpecialization,
					message: "Check your talents: You have selected a dual-wield spec but don't have [Dual Wield Specialization] talented.",
				},
			],
			player,
		);
	},
};
const ARMS_PRESET_OPTIONS = {
	onLoad: (player: Player<any>) => {
		PresetUtils.makeSpecChangeWarningToast(
			[
				{
					condition: isFurySpec,
					message: 'Check your gear: You have a one-handed weapon equipped, but the selected option is for two-handed weapons.',
				},
			],
			player,
		);
	},
};

export const P1_PRERAID_FURY_PRESET = PresetUtils.makePresetGear('Preraid - Fury', PreraidFuryGear, FURY_PRESET_OPTIONS);
export const P1_BIS_FURY_PRESET = PresetUtils.makePresetGear('P1 - Fury', P1FuryGear, FURY_PRESET_OPTIONS);
export const P2_BIS_FURY_PRESET = PresetUtils.makePresetGear('P2 - Fury', P2FuryGear, FURY_PRESET_OPTIONS);
export const P3_BIS_FURY_PRESET = PresetUtils.makePresetGear('P3 - Fury', P3FuryGear, FURY_PRESET_OPTIONS);
export const P35_BIS_FURY_PRESET = PresetUtils.makePresetGear('P3.5 - Fury', P35FuryGear, FURY_PRESET_OPTIONS);
export const P4_BIS_FURY_PRESET = PresetUtils.makePresetGear('P4 - Fury', P4FuryGear, FURY_PRESET_OPTIONS);

export const P1_PRERAID_ARMS_PRESET = PresetUtils.makePresetGear('Preraid - Arms', PreraidArmsGear, ARMS_PRESET_OPTIONS);
export const P1_BIS_ARMS_PRESET = PresetUtils.makePresetGear('P1 - Arms', P1ArmsGear, ARMS_PRESET_OPTIONS);
export const P2_BIS_ARMS_PRESET = PresetUtils.makePresetGear('P2 - Arms', P2ArmsGear, ARMS_PRESET_OPTIONS);
export const P3_BIS_ARMS_PRESET = PresetUtils.makePresetGear('P3 - Arms', P3ArmsGear, ARMS_PRESET_OPTIONS);
export const P35_BIS_ARMS_PRESET = PresetUtils.makePresetGear('P3.5 - Arms', P35ArmsGear, ARMS_PRESET_OPTIONS);
export const P4_BIS_ARMS_PRESET = PresetUtils.makePresetGear('P4 - Arms', P4ArmsGear, ARMS_PRESET_OPTIONS);

export const FURY_DEFAULT_ROTATION = PresetUtils.makePresetAPLRotation('Fury', DefaultFuryApl);
export const ARMS_DEFAULT_ROTATION = PresetUtils.makePresetAPLRotation('Arms', DefaultArmsApl);

export const SIMPLE_ROTATION = DpsWarrior_Rotation.create({
	spec: DpsWarriorSpec.DpsWarriorSpecFury,
	sunderArmor: WarriorSunder.WarriorSunderHelp,
	useOverpower: true,
	useRecklessness: false,
	bloodlustTiming: 5,
});
export const SIMPLE_DEFAULT_ROTATION = PresetUtils.makePresetSimpleRotation('Simple', Spec.SpecDpsWarrior, SIMPLE_ROTATION);
export const SIMPLE_ARMS_DEFAULT_ROTATION = PresetUtils.makePresetSimpleRotation('Simple', Spec.SpecDpsWarrior, {
	...SIMPLE_ROTATION,
	spec: DpsWarriorSpec.DpsWarriorSpecArms,
});

// Preset options for EP weights
export const P1_FURY_EP_PRESET = PresetUtils.makePresetEpWeights(
	'P1 - Fury',
	Stats.fromMap(
		{
			[Stat.StatStrength]: 1.0,
			[Stat.StatAgility]: 0.68,
			[Stat.StatAttackPower]: 0.45,
			[Stat.StatMeleeHitRating]: 0.48,
			[Stat.StatMeleeCritRating]: 0.92,
			[Stat.StatMeleeHasteRating]: 0.81,
			[Stat.StatArmorPenetration]: 0.15,
			[Stat.StatExpertiseRating]: 1.03,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 2.79,
			[PseudoStat.PseudoStatOffHandDps]: 1.47,
		},
	),
	FURY_PRESET_OPTIONS,
);

export const P2_FURY_EP_PRESET = PresetUtils.makePresetEpWeights(
	'P2, P3 & P4 - Fury',
	Stats.fromMap(
		{
			[Stat.StatStrength]: 1.0,
			[Stat.StatAgility]: 0.75,
			[Stat.StatAttackPower]: 0.45,
			[Stat.StatMeleeHitRating]: 0.56,
			[Stat.StatMeleeCritRating]: 0.9,
			[Stat.StatMeleeHasteRating]: 0.86,
			[Stat.StatArmorPenetration]: 0.2,
			[Stat.StatExpertiseRating]: 1.31,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 2.8,
			[PseudoStat.PseudoStatOffHandDps]: 1.5,
		},
	),
	FURY_PRESET_OPTIONS,
);

export const P1_ARMS_EP_PRESET = PresetUtils.makePresetEpWeights(
	'P1, P2 - Arms',
	Stats.fromMap(
		{
			[Stat.StatStrength]: 1.0,
			[Stat.StatAgility]: 0.7,
			[Stat.StatAttackPower]: 0.46,
			[Stat.StatMeleeHitRating]: 0.5,
			[Stat.StatMeleeCritRating]: 0.95,
			[Stat.StatMeleeHasteRating]: 0.8,
			[Stat.StatArmorPenetration]: 0.19,
			[Stat.StatExpertiseRating]: 1.46,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 5.85,
		},
	),
	ARMS_PRESET_OPTIONS,
);

export const P3_ARMS_EP_PRESET = PresetUtils.makePresetEpWeights(
	'P3 & P4 - Arms',
	Stats.fromMap(
		{
			[Stat.StatStrength]: 1.0,
			[Stat.StatAgility]: 0.8,
			[Stat.StatAttackPower]: 0.45,
			[Stat.StatMeleeHitRating]: 1.01,
			[Stat.StatMeleeCritRating]: 1.05,
			[Stat.StatMeleeHasteRating]: 0.85,
			[Stat.StatArmorPenetration]: 0.23,
			[Stat.StatExpertiseRating]: 1.66,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 6.0,
		},
	),
	ARMS_PRESET_OPTIONS,
);

// Default talents. Uses the wowhead calculator format, make the talents on
// https://wowhead.com/tbc/talent-calc and copy the numbers in the url.
export const FuryTalents = {
	name: 'Fury',
	data: SavedTalents.create({
		talentsString: '3400502130201-05050005505012050115',
	}),
	...FURY_PRESET_OPTIONS,
};

export const ArmsTalents = {
	name: 'Arms',
	data: SavedTalents.create({
		talentsString: '32005020352010500221-0550000500521203',
	}),
	...ARMS_PRESET_OPTIONS,
};

export const ArmsKebabTalents = {
	name: 'Arms - Kebab',
	data: SavedTalents.create({
		talentsString: '34005021302010510321-0550000520501203',
	}),
	...FURY_PRESET_OPTIONS,
};

export const DefaultOptions = WarriorOptions.create({
	classOptions: {
		queueDelay: 250,
		startingRage: 50,
		defaultShout: WarriorShout.WarriorShoutBattle,
		defaultStance: WarriorStance.WarriorStanceBerserker,
		hasBsT2: true,
		stanceSnapshot: true,
	},
});

export const DefaultConsumables = ConsumesSpec.create({
	...WarriorPresets.DefaultConsumables,
});

export const OtherDefaults = {
	race: Race.RaceOrc,
	profession1: Profession.Engineering,
	profession2: Profession.Blacksmithing,
	distanceFromTarget: 25,
};

export const PRESET_BUILD_FURY = PresetUtils.makePresetBuild('Fury', {
	talents: FuryTalents,
	rotationType: APLRotation_Type.TypeSimple,
	rotation: SIMPLE_DEFAULT_ROTATION,
});

export const PRESET_BUILD_ARMS = PresetUtils.makePresetBuild('Arms', {
	talents: ArmsTalents,
	rotationType: APLRotation_Type.TypeSimple,
	rotation: SIMPLE_ARMS_DEFAULT_ROTATION,
});

export const PRESET_BUILD_ARMS_KEBAB = PresetUtils.makePresetBuild('Arms - Kebab', {
	talents: ArmsKebabTalents,
	rotationType: APLRotation_Type.TypeSimple,
	rotation: SIMPLE_ARMS_DEFAULT_ROTATION,
});

export const P1_PLAYER_SETTINGS: PresetUtils.PresetSettings = {
	name: 'P1',
	playerOptions: OtherDefaults,
	debuffs: WarriorPresets.DefaultDebuffs,
	reforgeSettings: {
		maxGemPhase: Phase.Phase1,
	},
};

export const P2_PLAYER_SETTINGS: PresetUtils.PresetSettings = {
	name: 'P2',
	playerOptions: OtherDefaults,
	partyBuffs: PartyBuffs.create({
		...WarriorPresets.DefaultPartyBuffs,
		leaderOfThePack: TristateEffect.TristateEffectImproved,
	}),
	debuffs: Debuffs.create({
		...WarriorPresets.DefaultDebuffs,
		...defaultExposeWeaknessSettings(Phase.Phase2),
	}),
	reforgeSettings: {
		maxGemPhase: Phase.Phase2,
	},
};

export const P3_PLAYER_SETTINGS: PresetUtils.PresetSettings = {
	name: 'P3',
	playerOptions: {
		...OtherDefaults,
		profession2: Profession.Jewelcrafting,
	},
	partyBuffs: P2_PLAYER_SETTINGS.partyBuffs,
	debuffs: Debuffs.create({
		...WarriorPresets.DefaultDebuffs,
		...defaultExposeWeaknessSettings(Phase.Phase3),
	}),
	reforgeSettings: {
		maxGemPhase: Phase.Phase3,
	},
};

export const P35_PLAYER_SETTINGS: PresetUtils.PresetSettings = {
	name: 'P3.5',
	playerOptions: {
		...OtherDefaults,
		profession2: Profession.Jewelcrafting,
	},
	partyBuffs: P2_PLAYER_SETTINGS.partyBuffs,
	debuffs: Debuffs.create({
		...WarriorPresets.DefaultDebuffs,
		...defaultExposeWeaknessSettings(Phase.Phase4),
	}),
	reforgeSettings: {
		maxGemPhase: Phase.Phase4,
	},
};

export const P4_PLAYER_SETTINGS: PresetUtils.PresetSettings = {
	name: 'P4',
	playerOptions: {
		...OtherDefaults,
		profession2: Profession.Jewelcrafting,
	},
	partyBuffs: P2_PLAYER_SETTINGS.partyBuffs,
	debuffs: Debuffs.create({
		...WarriorPresets.DefaultDebuffs,
		...defaultExposeWeaknessSettings(Phase.Phase5),
	}),
	reforgeSettings: {
		maxGemPhase: Phase.Phase5,
	},
};

export const P1_PRESET_BUILD_FURY = PresetUtils.makePresetBuild('P1 - Fury', {
	gear: P1_BIS_FURY_PRESET,
	talents: FuryTalents,
	epWeights: P1_FURY_EP_PRESET,
	rotationType: APLRotation_Type.TypeSimple,
	rotation: SIMPLE_DEFAULT_ROTATION,
	settings: P1_PLAYER_SETTINGS,
});

export const P2_PRESET_BUILD_FURY = PresetUtils.makePresetBuild('P2 - Fury', {
	gear: P2_BIS_FURY_PRESET,
	talents: FuryTalents,
	epWeights: P2_FURY_EP_PRESET,
	rotationType: APLRotation_Type.TypeSimple,
	rotation: SIMPLE_DEFAULT_ROTATION,
	settings: P2_PLAYER_SETTINGS,
});

export const P3_PRESET_BUILD_FURY = PresetUtils.makePresetBuild('P3 - Fury', {
	gear: P3_BIS_FURY_PRESET,
	talents: FuryTalents,
	epWeights: P2_FURY_EP_PRESET,
	rotationType: APLRotation_Type.TypeSimple,
	rotation: SIMPLE_DEFAULT_ROTATION,
	settings: P3_PLAYER_SETTINGS,
});

export const P35_PRESET_BUILD_FURY = PresetUtils.makePresetBuild('P3.5 - Fury', {
	gear: P35_BIS_FURY_PRESET,
	talents: FuryTalents,
	epWeights: P2_FURY_EP_PRESET,
	rotationType: APLRotation_Type.TypeSimple,
	rotation: SIMPLE_DEFAULT_ROTATION,
	settings: P35_PLAYER_SETTINGS,
});

export const P4_PRESET_BUILD_FURY = PresetUtils.makePresetBuild('P4 - Fury', {
	gear: P4_BIS_FURY_PRESET,
	talents: FuryTalents,
	epWeights: P2_FURY_EP_PRESET,
	rotationType: APLRotation_Type.TypeSimple,
	rotation: SIMPLE_DEFAULT_ROTATION,
	settings: P4_PLAYER_SETTINGS,
});

export const P1_PRESET_BUILD_ARMS = PresetUtils.makePresetBuild('P1 - Arms', {
	gear: P1_BIS_ARMS_PRESET,
	talents: ArmsTalents,
	epWeights: P1_ARMS_EP_PRESET,
	rotationType: APLRotation_Type.TypeSimple,
	rotation: SIMPLE_ARMS_DEFAULT_ROTATION,
	settings: P1_PLAYER_SETTINGS,
});

export const P2_PRESET_BUILD_ARMS = PresetUtils.makePresetBuild('P2 - Arms', {
	gear: P2_BIS_ARMS_PRESET,
	talents: ArmsTalents,
	epWeights: P1_ARMS_EP_PRESET,
	rotationType: APLRotation_Type.TypeSimple,
	rotation: SIMPLE_ARMS_DEFAULT_ROTATION,
	settings: P2_PLAYER_SETTINGS,
});

export const P3_PRESET_BUILD_ARMS = PresetUtils.makePresetBuild('P3 - Arms', {
	gear: P3_BIS_ARMS_PRESET,
	talents: ArmsTalents,
	epWeights: P3_ARMS_EP_PRESET,
	rotationType: APLRotation_Type.TypeSimple,
	rotation: SIMPLE_ARMS_DEFAULT_ROTATION,
	settings: P3_PLAYER_SETTINGS,
});

export const P35_PRESET_BUILD_ARMS = PresetUtils.makePresetBuild('P3.5 - Arms', {
	gear: P35_BIS_ARMS_PRESET,
	talents: ArmsTalents,
	epWeights: P3_ARMS_EP_PRESET,
	rotationType: APLRotation_Type.TypeSimple,
	rotation: SIMPLE_ARMS_DEFAULT_ROTATION,
	settings: P35_PLAYER_SETTINGS,
});

export const P4_PRESET_BUILD_ARMS = PresetUtils.makePresetBuild('P4 - Arms', {
	gear: P4_BIS_ARMS_PRESET,
	talents: ArmsTalents,
	epWeights: P3_ARMS_EP_PRESET,
	rotationType: APLRotation_Type.TypeSimple,
	rotation: SIMPLE_ARMS_DEFAULT_ROTATION,
	settings: P4_PLAYER_SETTINGS,
});
