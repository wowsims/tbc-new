import * as OtherInputs from '../../core/components/inputs/other_inputs';
import * as BuffDebuffInputs from '../../core/components/inputs/buffs_debuffs';
import * as Mechanics from '../../core/constants/mechanics';
import { IndividualSimUI, registerSpecConfig } from '../../core/individual_sim_ui';
import { Player } from '../../core/player';
import { PlayerClasses } from '../../core/player_classes';
import { APLAction, APLListItem, APLPrepullAction, APLRotation, APLRotation_Type as APLRotationType } from '../../core/proto/apl';
import { Cooldowns, Debuffs, EquipmentSpec, Faction, IndividualBuffs, ItemSlot, PartyBuffs, PseudoStat, Race, RaidBuffs, Spec, Stat } from '../../core/proto/common';
import {
	FeralCatDruid_Rotation as DruidRotation,
	FeralCatDruid_Rotation_AplType as FeralRotationType,
} from '../../core/proto/druid';
import * as AplUtils from '../../core/proto_utils/apl_utils';
import { Stats, UnitStat } from '../../core/proto_utils/stats';
import { defaultRaidBuffMajorDamageCooldowns } from '../../core/proto_utils/utils';
import { TypedEvent } from '../../core/typed_event';
import * as FeralInputs from './inputs';
import * as Presets from './presets';

const SPEC_CONFIG = registerSpecConfig(Spec.SpecFeralCatDruid, {
	cssClass: 'feral-druid-sim-ui',
	cssScheme: PlayerClasses.getCssClass(PlayerClasses.Druid),
	// Override required talent rows - Feral only requires rows 3 and 5 instead of all rows
	requiredTalentRows: [0, 3, 5],
	// List any known bugs / issues here and they'll be shown on the site.
	knownIssues: [],
	warnings: [],

	// All stats for which EP should be calculated.
	epStats: [
		Stat.StatStrength,
		Stat.StatAgility,
		Stat.StatAttackPower,
	],
	epPseudoStats: [PseudoStat.PseudoStatMainHandDps],
	// Reference stat against which to calculate EP. I think all classes use either spell power or attack power.
	epReferenceStat: Stat.StatAgility,
	// Which stats to display in the Character Stats section, at the bottom of the left-hand sidebar.
	displayStats: UnitStat.createDisplayStatArray(
		[Stat.StatHealth, Stat.StatStrength, Stat.StatAgility, Stat.StatAttackPower, Stat.StatMana],
		[PseudoStat.PseudoStatMeleeHitPercent, PseudoStat.PseudoStatMeleeCritPercent, PseudoStat.PseudoStatMeleeHastePercent],
	),

	defaults: {
		// Default equipped gear.
		gear: EquipmentSpec.create(),
		// Default EP weights for sorting gear in the gear picker.
		epWeights: Stats.fromMap({}),
		other: Presets.OtherDefaults,
		// Default consumes settings.
		consumables: Presets.DefaultConsumables,
		// Default rotation settings.
		rotationType: APLRotationType.TypeAPL,
		// Default talents.
		talents: Presets.StandardTalents.data,
		// Default spec-specific settings.
		specOptions: Presets.DefaultOptions,
		// Default raid/party buffs settings.
		raidBuffs: RaidBuffs.create({
			...defaultRaidBuffMajorDamageCooldowns(),
		}),
		partyBuffs: PartyBuffs.create({

		}),
		individualBuffs: IndividualBuffs.create({

		}),
		debuffs: Debuffs.create({

		}),
	},

	// IconInputs to include in the 'Player' section on the settings tab.
	playerIconInputs: [],
	// Inputs to include in the 'Rotation' section on the settings tab.
	rotationInputs: FeralInputs.FeralDruidRotationConfig,
	// Buff and Debuff inputs to include/exclude, overriding the EP-based defaults.
	includeBuffDebuffInputs: [, , ],
	excludeBuffDebuffInputs: [],
	// Inputs to include in the 'Other' section on the settings tab.
	otherInputs: {
		inputs: [
			FeralInputs.AssumeBleedActive,
			OtherInputs.InputDelay,
			OtherInputs.DistanceFromTarget,
			OtherInputs.TankAssignment,
			OtherInputs.InFrontOfTarget,
			FeralInputs.CannotShredTarget,
		],
	},
	itemSwapSlots: [ItemSlot.ItemSlotMainHand, ItemSlot.ItemSlotOffHand, ItemSlot.ItemSlotTrinket1, ItemSlot.ItemSlotTrinket2],
	encounterPicker: {
		// Whether to include 'Execute Duration (%)' in the 'Encounter' section of the settings tab.
		showExecuteProportion: true,
	},

	presets: {
		epWeights: [],
		// Preset talents that the user can quickly select.
		talents: [Presets.StandardTalents],
		rotations: [],
		// Preset gear configurations that the user can quickly select.
		gear: [],
		itemSwaps: [],
		builds: [],
	},

	autoRotation: (_player: Player<Spec.SpecFeralCatDruid>): APLRotation => {
		return APLRotation.create();
	},

	simpleRotation: (player: Player<Spec.SpecFeralCatDruid>, simple: DruidRotation, cooldowns: Cooldowns): APLRotation => {
		// TODO: Implement TBC-specific rotation logic
		// This is a placeholder for clean state - build TBC rotations from scratch
		return APLRotation.create({
			prepullActions: [],
			priorityList: [],
		});
	},

	hiddenMCDs: [],

	raidSimPresets: [
		{
			spec: Spec.SpecFeralCatDruid,
			talents: Presets.StandardTalents.data,
			specOptions: Presets.DefaultOptions,
			consumables: Presets.DefaultConsumables,
			defaultFactionRaces: {
				[Faction.Unknown]: Race.RaceUnknown,
				[Faction.Alliance]: Race.RaceNightElf,
				[Faction.Horde]: Race.RaceTauren,
			},
			defaultGear: {
				[Faction.Unknown]: {},
				[Faction.Alliance]: {},
				[Faction.Horde]: {},
			},
			otherDefaults: Presets.OtherDefaults,
		},
	],
});

export class FeralCatDruidSimUI extends IndividualSimUI<Spec.SpecFeralCatDruid> {
	constructor(parentElem: HTMLElement, player: Player<Spec.SpecFeralCatDruid>) {
		super(parentElem, player, SPEC_CONFIG);
	}
}
