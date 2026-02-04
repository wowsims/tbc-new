import * as OtherInputs from '../../core/components/inputs/other_inputs';
import { IndividualSimUI, registerSpecConfig } from '../../core/individual_sim_ui';
import { Player } from '../../core/player';
import { PlayerClasses } from '../../core/player_classes';
import { APLRotation } from '../../core/proto/apl';
import { Faction, IndividualBuffs, ItemSlot, PseudoStat, Race, Spec, Stat } from '../../core/proto/common';
import { DEFAULT_MELEE_GEM_STATS, UnitStat } from '../../core/proto_utils/stats';

import * as WarriorInputs from '../inputs';
import * as WarriorPresets from '../presets';
import * as Presets from './presets';

const SPEC_CONFIG = registerSpecConfig(Spec.SpecDpsWarrior, {
	cssClass: 'dps-warrior-sim-ui',
	cssScheme: PlayerClasses.getCssClass(PlayerClasses.Warrior),
	// List any known bugs / issues here and they'll be shown on the site.
	knownIssues: [],

	// All stats for which EP should be calculated.
	epStats: [Stat.StatStrength, Stat.StatAgility, Stat.StatAttackPower, Stat.StatExpertiseRating, Stat.StatMeleeHasteRating, Stat.StatMeleeCritRating],
	epPseudoStats: [PseudoStat.PseudoStatMainHandDps, PseudoStat.PseudoStatOffHandDps],
	// Reference stat against which to calculate EP. I think all classes use either spell power or attack power.
	epReferenceStat: Stat.StatStrength,
	gemStats: DEFAULT_MELEE_GEM_STATS,
	// Which stats to display in the Character Stats section, at the bottom of the left-hand sidebar.
	displayStats: UnitStat.createDisplayStatArray(
		[Stat.StatHealth, Stat.StatStamina, Stat.StatStrength, Stat.StatAgility, Stat.StatAttackPower, Stat.StatExpertiseRating],
		[
			PseudoStat.PseudoStatMeleeHitPercent,
			PseudoStat.PseudoStatMeleeHitPercent,
			PseudoStat.PseudoStatMeleeCritPercent,
			PseudoStat.PseudoStatMeleeHastePercent,
		],
	),

	defaults: {
		// Default equipped gear.
		gear: Presets.P1_BIS_FURY_PRESET.gear,
		// Default EP weights for sorting gear in the gear picker.
		epWeights: Presets.P1_FURY_EP_PRESET.epWeights,
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
	// Inputs to include in the 'Other' section on the settings tab.
	otherInputs: {
		inputs: [
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
		epWeights: [Presets.P1_FURY_EP_PRESET, Presets.P1_ARMS_EP_PRESET],
		// Preset talents that the user can quickly select.
		talents: [Presets.FuryTalents, Presets.ArmsTalents],
		// Preset rotations that the user can quickly select.
		rotations: [Presets.FURY_DEFAULT_ROTATION, Presets.ARMS_DEFAULT_ROTATION],
		// Preset gear configurations that the user can quickly select.
		gear: [Presets.P1_PRERAID_FURY_PRESET, Presets.P1_PRERAID_ARMS_PRESET, Presets.P1_BIS_FURY_PRESET, Presets.P1_BIS_ARMS_PRESET],
		builds: [Presets.P1_PRESET_BUILD_FURY, Presets.P1_PRESET_BUILD_ARMS],
	},

	autoRotation: (_player: Player<Spec.SpecDpsWarrior>): APLRotation => {
		return Presets.FURY_DEFAULT_ROTATION.rotation.rotation!;
	},

	raidSimPresets: [
		{
			spec: Spec.SpecDpsWarrior,
			talents: Presets.FuryTalents.data,
			specOptions: Presets.DefaultOptions,
			consumables: Presets.DefaultConsumables,
			defaultFactionRaces: {
				[Faction.Unknown]: Race.RaceUnknown,
				[Faction.Alliance]: Race.RaceHuman,
				[Faction.Horde]: Race.RaceOrc,
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

export class DpsWarriorSimUI extends IndividualSimUI<Spec.SpecDpsWarrior> {
	constructor(parentElem: HTMLElement, player: Player<Spec.SpecDpsWarrior>) {
		super(parentElem, player, SPEC_CONFIG);
	}
}
