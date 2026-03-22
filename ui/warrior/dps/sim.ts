import * as OtherInputs from '../../core/components/inputs/other_inputs';
import { ReforgeOptimizer } from '../../core/components/suggest_reforges_action';
import { IndividualSimUI, registerSpecConfig } from '../../core/individual_sim_ui';
import { Player } from '../../core/player';
import { PlayerClasses } from '../../core/player_classes';
import { APLRotation, APLRotation_Type, SimpleRotation } from '../../core/proto/apl';
import { Cooldowns, HandType, ItemSlot, PseudoStat, Spec, Stat } from '../../core/proto/common';
import { StatCapType } from '../../core/proto/ui';
import { DEFAULT_MELEE_GEM_STATS, StatCap, Stats, UnitStat } from '../../core/proto_utils/stats';

import * as Mechanics from '../../core/constants/mechanics';
import * as WarriorInputs from '../inputs';
import * as DpsWarriorInputs from './inputs';
import * as WarriorPresets from '../presets';
import * as Presets from './presets';
import { SpecRotation } from '../../core/proto_utils/utils';
import { DpsWarriorSpec, WarriorSunder } from '../../core/proto/warrior';

const SPEC_CONFIG = registerSpecConfig(Spec.SpecDpsWarrior, {
	cssClass: 'dps-warrior-sim-ui',
	cssScheme: PlayerClasses.getCssClass(PlayerClasses.Warrior),
	// List any known bugs / issues here and they'll be shown on the site.
	knownIssues: [],

	// All stats for which EP should be calculated.
	epStats: [
		Stat.StatStrength,
		Stat.StatAgility,
		Stat.StatAttackPower,
		Stat.StatArmorPenetration,
		Stat.StatMeleeHitRating,
		Stat.StatMeleeHasteRating,
		Stat.StatMeleeCritRating,
		Stat.StatArmorPenetration,
		Stat.StatExpertiseRating,
	],
	epPseudoStats: [PseudoStat.PseudoStatMainHandDps, PseudoStat.PseudoStatOffHandDps],
	// Reference stat against which to calculate EP. I think all classes use either spell power or attack power.
	epReferenceStat: Stat.StatStrength,
	gemStats: DEFAULT_MELEE_GEM_STATS,
	// Which stats to display in the Character Stats section, at the bottom of the left-hand sidebar.
	displayStats: UnitStat.createDisplayStatArray(
		[Stat.StatHealth, Stat.StatStamina, Stat.StatStrength, Stat.StatAgility, Stat.StatAttackPower, Stat.StatExpertiseRating, Stat.StatArmorPenetration],
		[PseudoStat.PseudoStatMeleeHitPercent, PseudoStat.PseudoStatMeleeCritPercent, PseudoStat.PseudoStatMeleeHastePercent],
	),

	defaults: {
		// Default equipped gear.
		gear: Presets.P1_BIS_FURY_PRESET.gear,
		// Default EP weights for sorting gear in the gear picker.
		epWeights: Presets.P1_FURY_EP_PRESET.epWeights,
		statCaps: (() => {
			const expCap = new Stats().withStat(Stat.StatExpertiseRating, 6.5 * 4 * Mechanics.EXPERTISE_PER_QUARTER_PERCENT_REDUCTION);
			return expCap;
		})(),
		softCapBreakpoints: (() => {
			const meleeHitSoftCapConfig = StatCap.fromPseudoStat(PseudoStat.PseudoStatMeleeHitPercent, {
				breakpoints: [9, 28],
				capType: StatCapType.TypeSoftCap,
				postCapEPs: [0.57 * Mechanics.PHYSICAL_HIT_RATING_PER_HIT_PERCENT, 0],
			});

			return [meleeHitSoftCapConfig];
		})(),
		rotationType: APLRotation_Type.TypeSimple,
		simpleRotation: Presets.SIMPLE_ROTATION,
		other: Presets.OtherDefaults,
		// Default consumes settings.
		consumables: Presets.DefaultConsumables,
		// Default talents.
		talents: Presets.FuryTalents.data,
		// Default spec-specific settings.
		specOptions: Presets.DefaultOptions,
		// Default raid/party buffs settings.
		raidBuffs: WarriorPresets.DefaultRaidBuffs,
		partyBuffs: WarriorPresets.DefaultPartyBuffs,
		individualBuffs: WarriorPresets.DefaultIndividualBuffs,
		debuffs: WarriorPresets.DefaultDebuffs,
	},

	// IconInputs to include in the 'Player' section on the settings tab.
	playerIconInputs: [WarriorInputs.ShoutPicker(), WarriorInputs.StancePicker()],
	// Buff and Debuff inputs to include/exclude, overriding the EP-based defaults.
	includeBuffDebuffInputs: [],
	excludeBuffDebuffInputs: [],
	rotationInputs: DpsWarriorInputs.RotationInputs,
	// Inputs to include in the 'Other' section on the settings tab.
	otherInputs: {
		inputs: [
			OtherInputs.TotemTwisting,
			WarriorInputs.BattleShoutSolarianSapphire(),
			WarriorInputs.BattleShoutT2(),
			WarriorInputs.StartingRage(),
			WarriorInputs.StanceSnapshot(),
			OtherInputs.DistanceFromTarget,
			WarriorInputs.QueueDelay(),
			OtherInputs.InputDelay,
			OtherInputs.TankAssignment,
			OtherInputs.InFrontOfTarget,
		],
	},
	itemSwapSlots: [ItemSlot.ItemSlotTrinket1, ItemSlot.ItemSlotTrinket2, ItemSlot.ItemSlotMainHand, ItemSlot.ItemSlotOffHand],
	encounterPicker: {
		// Whether to include 'Execute Duration (%)' in the 'Encounter' section of the settings tab.
		showExecuteProportion: true,
	},

	presets: {
		epWeights: [Presets.P1_FURY_EP_PRESET, Presets.P2_FURY_EP_PRESET, Presets.P1_ARMS_EP_PRESET, Presets.P3_ARMS_EP_PRESET],
		// Preset talents that the user can quickly select.
		talents: [Presets.FuryTalents, Presets.ArmsTalents, Presets.ArmsKebabTalents],
		// Preset rotations that the user can quickly select.
		rotations: [Presets.SIMPLE_DEFAULT_ROTATION, Presets.FURY_DEFAULT_ROTATION, Presets.ARMS_DEFAULT_ROTATION],
		// Preset gear configurations that the user can quickly select.
		gear: [
			Presets.P1_PRERAID_FURY_PRESET,
			Presets.P1_BIS_FURY_PRESET,
			Presets.P2_BIS_FURY_PRESET,
			Presets.P3_BIS_FURY_PRESET,
			Presets.P35_BIS_FURY_PRESET,
			Presets.P4_BIS_FURY_PRESET,
			Presets.P1_PRERAID_ARMS_PRESET,
			Presets.P1_BIS_ARMS_PRESET,
			Presets.P2_BIS_ARMS_PRESET,
			Presets.P3_BIS_ARMS_PRESET,
			Presets.P35_BIS_ARMS_PRESET,
			Presets.P4_BIS_ARMS_PRESET,
		],
		builds: [
			Presets.PRESET_BUILD_FURY,
			Presets.PRESET_BUILD_ARMS,
			Presets.PRESET_BUILD_ARMS_KEBAB,
			Presets.P1_PRESET_BUILD_FURY,
			Presets.P2_PRESET_BUILD_FURY,
			Presets.P3_PRESET_BUILD_FURY,
			Presets.P35_PRESET_BUILD_FURY,
			Presets.P4_PRESET_BUILD_FURY,
			Presets.P1_PRESET_BUILD_ARMS,
			Presets.P2_PRESET_BUILD_ARMS,
			Presets.P3_PRESET_BUILD_ARMS,
			Presets.P35_PRESET_BUILD_ARMS,
			Presets.P4_PRESET_BUILD_ARMS,
		],
	},

	autoRotation: (player: Player<Spec.SpecDpsWarrior>): APLRotation => {
		if (Presets.isArmsSpec(player) || Presets.isArmsKebabSpec(player)) {
			return Presets.ARMS_DEFAULT_ROTATION.rotation.rotation!;
		}

		return Presets.FURY_DEFAULT_ROTATION.rotation.rotation!;
	},

	simpleRotation: (player: Player<Spec.SpecDpsWarrior>, simple: SpecRotation<Spec.SpecDpsWarrior>, _: Cooldowns): APLRotation => {
		let { spec, sunderArmor = WarriorSunder.WarriorSunderHelp, useOverpower = true, useRecklessness = false, bloodlustTiming = 5 } = simple;

		if (!spec) {
			if (Presets.isArmsSpec(player) || Presets.isArmsKebabSpec(player)) {
				spec = DpsWarriorSpec.DpsWarriorSpecArms;
			} else {
				spec = DpsWarriorSpec.DpsWarriorSpecFury;
			}
		}

		const rotation = APLRotation.clone(
			spec == DpsWarriorSpec.DpsWarriorSpecFury ? Presets.FURY_DEFAULT_ROTATION.rotation.rotation! : Presets.ARMS_DEFAULT_ROTATION.rotation.rotation!,
		);

		const bloodlustTimingVariable = rotation.valueVariables.find(variable => variable.name === 'Bloodlust time');
		if (bloodlustTimingVariable && bloodlustTimingVariable.value?.value.oneofKind === 'const')
			bloodlustTimingVariable.value.value.const.val = String(bloodlustTiming);

		const recklessnessAction = rotation.priorityList.find(
			action => action.action?.action.oneofKind === 'groupReference' && action.action.action.groupReference.groupName === 'Recklessness ON/OFF',
		);
		if (recklessnessAction) recklessnessAction.hide = !useRecklessness;

		const sunderArmorAction = rotation.priorityList.find(
			action => action.action?.action.oneofKind === 'groupReference' && action.action?.action.groupReference.groupName === 'Sunder Armor',
		);
		if (sunderArmorAction) sunderArmorAction.hide = sunderArmor == WarriorSunder.WarriorSunderNone;

		const opWeaveAction = rotation.priorityList.find(
			action => action.action?.action.oneofKind === 'groupReference' && action.action?.action.groupReference.groupName === 'Overpower Weaving',
		);
		if (opWeaveAction) opWeaveAction.hide = !useOverpower;

		return APLRotation.create({
			simple: SimpleRotation.create({}),
			...rotation,
		});
	},

	raidSimPresets: [],
});

export class DpsWarriorSimUI extends IndividualSimUI<Spec.SpecDpsWarrior> {
	constructor(parentElem: HTMLElement, player: Player<Spec.SpecDpsWarrior>) {
		super(parentElem, player, SPEC_CONFIG);

		this.reforger = new ReforgeOptimizer(this, {
			updateSoftCaps: softCaps => {
				const gear = player.getGear();
				const mainHandType = gear.getEquippedItem(ItemSlot.ItemSlotMainHand)?.item.handType;
				const offHandType = gear.getEquippedItem(ItemSlot.ItemSlotOffHand)?.item.handType;
				const isFury =
					mainHandType &&
					[HandType.HandTypeOneHand, HandType.HandTypeMainHand].includes(mainHandType) &&
					offHandType &&
					[HandType.HandTypeOneHand, HandType.HandTypeOffHand].includes(offHandType);

				const softCapToModify = softCaps.find(sc => sc.unitStat.equalsPseudoStat(PseudoStat.PseudoStatMeleeHitPercent));
				if (softCapToModify) {
					if (isFury) {
						softCapToModify.breakpoints = this.individualConfig.defaults.softCapBreakpoints?.[0].breakpoints || [];
						softCapToModify.postCapEPs = this.individualConfig.defaults.softCapBreakpoints?.[0].postCapEPs || [];
					} else {
						softCapToModify.breakpoints = [9];
						softCapToModify.postCapEPs = [0];
					}
				}

				return softCaps;
			},
		});
	}
}
