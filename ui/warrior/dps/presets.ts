import { Player } from '../../core/player';
import * as PresetUtils from '../../core/preset_utils';
import { ConsumesSpec, HandType, ItemSlot, Profession, PseudoStat, Race, Spec, Stat } from '../../core/proto/common';
import { SavedTalents } from '../../core/proto/ui';
import { DpsWarrior_Options as WarriorOptions, WarriorShout, WarriorStance } from '../../core/proto/warrior';
import { Stats } from '../../core/proto_utils/stats';
import * as WarriorPresets from '../presets';
import DefaultArmsApl from './apls/arms.apl.json';
import DefaultFuryApl from './apls/fury.apl.json';
import PreraidArmsGear from './gear_sets/preraid_arms.gear.json';
import PreraidFuryGear from './gear_sets/preraid_fury.gear.json';
import P1ArmsGear from './gear_sets/p1_arms.gear.json';
import P1FuryGear from './gear_sets/p1_fury.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

// Handlers for spec specific load checks
const FURY_PRESET_OPTIONS = {
	onLoad: (player: Player<Spec.SpecDpsWarrior>) => {
		PresetUtils.makeSpecChangeWarningToast(
			[
				{
					condition: (player: Player<Spec.SpecDpsWarrior>) =>
						player.getEquippedItem(ItemSlot.ItemSlotMainHand)?.item.handType === HandType.HandTypeTwoHand,
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
					condition: (player: Player<Spec.SpecDpsWarrior>) =>
						player.getEquippedItem(ItemSlot.ItemSlotMainHand)?.item.handType === HandType.HandTypeOneHand,
					message: 'Check your gear: You have a one-handed weapon equipped, but the selected option is for two-handed weapons.',
				},
			],
			player,
		);
	},
};

export const P1_PRERAID_FURY_PRESET = PresetUtils.makePresetGear('Preraid - Fury', PreraidFuryGear, FURY_PRESET_OPTIONS);
export const P1_BIS_FURY_PRESET = PresetUtils.makePresetGear('P1 - Fury', P1FuryGear, FURY_PRESET_OPTIONS);

export const P1_PRERAID_ARMS_PRESET = PresetUtils.makePresetGear('Preraid - Arms', PreraidArmsGear, ARMS_PRESET_OPTIONS);
export const P1_BIS_ARMS_PRESET = PresetUtils.makePresetGear('P1 - Arms', P1ArmsGear, ARMS_PRESET_OPTIONS);

export const FURY_DEFAULT_ROTATION = PresetUtils.makePresetAPLRotation('Fury', DefaultFuryApl);
export const ARMS_DEFAULT_ROTATION = PresetUtils.makePresetAPLRotation('Arms', DefaultArmsApl);

// Preset options for EP weights
export const P1_FURY_EP_PRESET = PresetUtils.makePresetEpWeights(
	'P1 - Fury',
	Stats.fromMap(
		{
			[Stat.StatStrength]: 1.0,
			[Stat.StatAgility]: 0.65,
			[Stat.StatAttackPower]: 0.45,
			[Stat.StatMeleeHitRating]: 0.55,
			[Stat.StatMeleeCritRating]: 0.88,
			[Stat.StatMeleeHasteRating]: 0.85,
			[Stat.StatArmorPenetration]: 0.15,
			[Stat.StatExpertiseRating]: 0.89,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 2.72,
			[PseudoStat.PseudoStatOffHandDps]: 1.55,
		},
	),
	FURY_PRESET_OPTIONS,
);

export const P1_ARMS_EP_PRESET = PresetUtils.makePresetEpWeights(
	'P1 - Arms',
	Stats.fromMap(
		{
			[Stat.StatStrength]: 1.0,
			[Stat.StatAgility]: 0.65,
			[Stat.StatAttackPower]: 0.45,
			[Stat.StatMeleeHitRating]: 0.55,
			[Stat.StatMeleeCritRating]: 0.88,
			[Stat.StatMeleeHasteRating]: 0.85,
			[Stat.StatArmorPenetration]: 0.15,
			[Stat.StatExpertiseRating]: 0.89,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 2.72,
		},
	),
	ARMS_PRESET_OPTIONS,
);

// Default talents. Uses the wowhead calculator format, make the talents on
// https://wowhead.com/tbc/talent-calc and copy the numbers in the url.

export const FuryTalents = {
	name: 'Fury',
	data: SavedTalents.create({
		talentsString: '3500501130201-05050005505012050115',
	}),
	...FURY_PRESET_OPTIONS,
};

export const ArmsTalents = {
	name: 'Arms',
	data: SavedTalents.create({
		talentsString: '33005001352010500221-0550000500521203',
	}),
	...ARMS_PRESET_OPTIONS,
};

export const DefaultOptions = WarriorOptions.create({
	classOptions: {
		queueDelay: 250,
		startingRage: 0,
		defaultShout: WarriorShout.WarriorShoutBattle,
		defaultStance: WarriorStance.WarriorStanceBerserker,
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

export const P1_PRESET_BUILD_FURY = PresetUtils.makePresetBuild('P1 - Fury', {
	gear: P1_BIS_FURY_PRESET,
	talents: FuryTalents,
	epWeights: P1_FURY_EP_PRESET,
});

export const P1_PRESET_BUILD_ARMS = PresetUtils.makePresetBuild('P1 - Arms', {
	gear: P1_BIS_ARMS_PRESET,
	talents: ArmsTalents,
	epWeights: P1_ARMS_EP_PRESET,
});
