import { ReforgeOptimizer } from '../../core/components/suggest_reforges_action';
import * as other_inputs from '../../core/components/inputs/other_inputs';
import { IndividualSimUI, registerSpecConfig } from '../../core/individual_sim_ui';
import { Player } from '../../core/player';
import { PlayerClasses } from '../../core/player_classes';
import { APLRotation } from '../../core/proto/apl';
import { Faction, HandType, ItemSlot, PseudoStat, Race, Spec, Stat } from '../../core/proto/common';
import { StatCapType } from '../../core/proto/ui';
import { StatCap, UnitStat } from '../../core/proto_utils/stats';
import * as HunterInputs from './inputs';
import * as Presets from './presets';

const SPEC_CONFIG = registerSpecConfig(Spec.SpecHunter, {
	cssClass: 'hunter-sim-ui',
	cssScheme: PlayerClasses.getCssClass(PlayerClasses.Hunter),
	// List any known bugs / issues here and they'll be shown on the site.
	knownIssues: [],
	warnings: [],
	// All stats for which EP should be calculated.
	epStats: [
		Stat.StatAgility,
		Stat.StatStrength,
		Stat.StatIntellect,
		Stat.StatMP5,
		Stat.StatAttackPower,
		Stat.StatRangedAttackPower,
		Stat.StatArmorPenetration,
		Stat.StatMeleeHitRating,
		Stat.StatMeleeHasteRating,
		Stat.StatMeleeCritRating,
		Stat.StatArmorPenetration,
		Stat.StatExpertiseRating,
		Stat.StatPhysicalDamage,
	],
	gemStats: [Stat.StatStamina, Stat.StatAgility],
	epPseudoStats: [PseudoStat.PseudoStatRangedHitPercent, PseudoStat.PseudoStatRangedCritPercent, PseudoStat.PseudoStatRangedDps],
	consumableStats: [Stat.StatStamina, Stat.StatHealth, Stat.StatMana],
	// Reference stat against which to calculate EP.
	epReferenceStat: Stat.StatAgility,
	// Which stats to display in the Character Stats section, at the bottom of the left-hand sidebar.
	displayStats: UnitStat.createDisplayStatArray(
		[
			Stat.StatHealth,
			Stat.StatStamina,
			Stat.StatStrength,
			Stat.StatAgility,
			Stat.StatAttackPower,
			Stat.StatRangedAttackPower,
			Stat.StatExpertiseRating,
			Stat.StatArmorPenetration,
		],
		[
			PseudoStat.PseudoStatMeleeHitPercent,
			PseudoStat.PseudoStatMeleeCritPercent,
			PseudoStat.PseudoStatRangedHitPercent,
			PseudoStat.PseudoStatRangedCritPercent,
			PseudoStat.PseudoStatRangedHastePercent,
		],
	),
	itemSwapSlots: [ItemSlot.ItemSlotMainHand, ItemSlot.ItemSlotOffHand, ItemSlot.ItemSlotRanged, ItemSlot.ItemSlotTrinket1, ItemSlot.ItemSlotTrinket2],
	defaults: {
		// Default equipped gear.
		gear: Presets.P1_BM_2H_6P_GEARSET.gear,
		// Default EP weights for sorting gear in the gear picker.
		epWeights: Presets.P1_BM_EP_PRESET.epWeights,
		softCapBreakpoints: [
			StatCap.fromPseudoStat(PseudoStat.PseudoStatRangedHitPercent, {
				breakpoints: [9],
				capType: StatCapType.TypeSoftCap,
				postCapEPs: [0],
			}),
		],
		other: Presets.OtherDefaults,
		// Default consumes settings.
		consumables: Presets.DefaultConsumables,
		// Default talents.
		talents: Presets.BMTalents.data,
		// Default spec-specific settings.
		specOptions: Presets.DefaultOptions,
		// Default raid/party buffs settings.
		raidBuffs: Presets.DefaultRaidBuffs,
		partyBuffs: Presets.DefaultPartyBuffs,
		individualBuffs: Presets.DefaultIndividualBuffs,
		debuffs: Presets.DefaultDebuffs,
	},

	// IconInputs to include in the 'Player' section on the settings tab.
	playerIconInputs: [HunterInputs.PetTypeInput(), HunterInputs.QuiverInput(), HunterInputs.AmmoInput()],
	// Buff and Debuff inputs to include/exclude, overriding the EP-based defaults.
	includeBuffDebuffInputs: [],
	excludeBuffDebuffInputs: [],
	// Inputs to include in the 'Other' section on the settings tab.
	otherInputs: {
		inputs: [
			other_inputs.TotemTwisting,
			HunterInputs.PetUptime(),
			HunterInputs.PetSingleAbility(),
			other_inputs.InputDelay,
			other_inputs.DistanceFromTarget,
			other_inputs.TankAssignment,
		],
	},
	encounterPicker: {
		// Whether to include 'Execute Duration (%)' in the 'Encounter' section of the settings tab.
		showExecuteProportion: false,
	},

	presets: {
		epWeights: [Presets.P1_BM_EP_PRESET, Presets.P1_SV_EP_PRESET],
		// Preset talents that the user can quickly select.
		talents: [Presets.BMTalents, Presets.SVTalents],
		// Preset rotations that the user can quickly select.
		rotations: [Presets.WEAVE_APL, Presets.TURRET_APL],
		// Preset gear configurations that the user can quickly select.
		builds: [
			Presets.P1_PRESET_BUILD_PRE_RAID,
			Presets.P1_PRESET_BUILD_BM_2H,
			Presets.P1_PRESET_BUILD_BM_DW,
			Presets.P1_PRESET_BUILD_SV_2H,
			Presets.P1_PRESET_BUILD_SV_DW,
		],
		gear: [
			Presets.P1_PreRaid_GEARSET,
			Presets.P1_BM_2H_6P_GEARSET,
			Presets.P1_BM_2H_9P_GEARSET,
			Presets.P1_BM_DW_6P_GEARSET,
			Presets.P1_BM_DW_9P_GEARSET,
			Presets.P1_SV_2H_3P_GEARSET,
			Presets.P1_SV_2H_6P_GEARSET,
			Presets.P1_SV_DW_3P_GEARSET,
			Presets.P1_SV_DW_6P_GEARSET,
		],
	},

	autoRotation: (player: Player<Spec.SpecHunter>): APLRotation => {
		const gear = player.getGear();
		const mainHandType = gear.getEquippedItem(ItemSlot.ItemSlotMainHand)?.item.handType;
		return mainHandType == HandType.HandTypeTwoHand ? Presets.WEAVE_APL.rotation.rotation! : Presets.TURRET_APL.rotation.rotation!;
	},

	raidSimPresets: [
		{
			spec: Spec.SpecHunter,
			talents: Presets.BMTalents.data,
			specOptions: Presets.DefaultOptions,
			consumables: Presets.DefaultConsumables,
			defaultFactionRaces: {
				[Faction.Unknown]: Race.RaceUnknown,
				[Faction.Alliance]: Race.RaceNightElf,
				[Faction.Horde]: Race.RaceOrc,
			},
			defaultGear: {
				[Faction.Unknown]: {},
				[Faction.Alliance]: {
					1: Presets.P1_BM_2H_6P_GEARSET.gear,
				},
				[Faction.Horde]: {
					1: Presets.P1_BM_2H_6P_GEARSET.gear,
				},
			},
			otherDefaults: Presets.OtherDefaults,
		},
	],
});

export class HunterSimUI extends IndividualSimUI<Spec.SpecHunter> {
	constructor(parentElem: HTMLElement, player: Player<Spec.SpecHunter>) {
		super(parentElem, player, SPEC_CONFIG);
		this.reforger = new ReforgeOptimizer(this);
	}
}
