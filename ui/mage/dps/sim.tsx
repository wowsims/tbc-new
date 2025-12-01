import * as OtherInputs from '../../core/components/inputs/other_inputs';
import * as Mechanics from '../../core/constants/mechanics';
import { IndividualSimUI, registerSpecConfig } from '../../core/individual_sim_ui';
import { Player } from '../../core/player';
import { PlayerClasses } from '../../core/player_classes';
import { Mage } from '../../core/player_classes/mage';
import { APLRotation } from '../../core/proto/apl';
import { Faction, IndividualBuffs, ItemSlot, PartyBuffs, PseudoStat, Race, Spec, Stat } from '../../core/proto/common';
import { StatCapType } from '../../core/proto/ui';
import { DEFAULT_CASTER_GEM_STATS, StatCap, Stats, UnitStat } from '../../core/proto_utils/stats';
import { formatToNumber } from '../../core/utils';
import { DefaultDebuffs, DefaultRaidBuffs } from './presets';
import * as Inputs from './inputs';
import * as Presets from './presets';
import * as MageInputs from './inputs';

const SPEC_CONFIG = registerSpecConfig(Spec.SpecMage, {
	cssClass: 'mage-sim-ui',
	cssScheme: PlayerClasses.getCssClass(Mage),
	// List any known bugs / issues here and they'll be shown on the site.
	knownIssues: [],

	// All stats for which EP should be calculated.
	epStats: [Stat.StatIntellect, Stat.StatSpellPower], // Reference stat against which to calculate EP. I think all classes use either spell power or attack power.
	epReferenceStat: Stat.StatSpellPower,
	// Which stats to display in the Character Stats section, at the bottom of the left-hand sidebar.
	displayStats: UnitStat.createDisplayStatArray(
		[
			Stat.StatHealth,
			Stat.StatMana,
			Stat.StatStamina,
			Stat.StatIntellect,
			Stat.StatSpirit,
			Stat.StatSpellPower,
		],
		[PseudoStat.PseudoStatSpellHitPercent, PseudoStat.PseudoStatSpellCritPercent, PseudoStat.PseudoStatSpellHastePercent],
	),
	gemStats: DEFAULT_CASTER_GEM_STATS,

	defaults: {
		// Default equipped gear.
		gear: Presets.BLANK_GEARSET.gear,
		// Default EP weights for sorting gear in the gear picker.
		epWeights: Presets.P1_EP_PRESET.epWeights,
		// Default consumes settings.
		consumables: Presets.DefaultConsumables,
		// Default talents.
		talents: Presets.Talents.data,
		// Default spec-specific settings.
		specOptions: Presets.DefaultOptions,
		other: Presets.OtherDefaults,
		// Default raid/party buffs settings.
		raidBuffs: DefaultRaidBuffs,

		partyBuffs: PartyBuffs.create({}),
		individualBuffs: IndividualBuffs.create({}),
		debuffs: DefaultDebuffs,
	},

	// IconInputs to include in the 'Player' section on the settings tab.
	playerIconInputs: [MageInputs.MageArmorInputs()],
	// Buff and Debuff inputs to include/exclude, overriding the EP-based defaults.
	includeBuffDebuffInputs: [],
	excludeBuffDebuffInputs: [],
	// Inputs to include in the 'Other' section on the settings tab.
	otherInputs: {
		inputs: [OtherInputs.InputDelay, OtherInputs.DistanceFromTarget, OtherInputs.TankAssignment],
	},
	itemSwapSlots: [ItemSlot.ItemSlotMainHand, ItemSlot.ItemSlotOffHand, ItemSlot.ItemSlotTrinket1, ItemSlot.ItemSlotTrinket2],
	encounterPicker: {
		// Whether to include 'Execute Duration (%)' in the 'Encounter' section of the settings tab.
		showExecuteProportion: true,
	},

	presets: {
		epWeights: [],
		// Preset rotations that the user can quickly select.
		rotations: [],
		// Preset talents that the user can quickly select.
		talents: [],
		// Preset gear configurations that the user can quickly select.
		gear: [],

		builds: [],
	},

	autoRotation: (player: Player<Spec.SpecMage>): APLRotation => {
		// const numTargets = player.sim.encounter.targets.length;
		// if (numTargets >= 2) {
		// 	return Presets.ROTATION_PRESET_CLEAVE.rotation.rotation!;
		// } else {
		return Presets.BLANK_APL.rotation.rotation!;
		// }
	},

	raidSimPresets: [
		{
			spec: Spec.SpecMage,
			talents: Presets.Talents.data,
			specOptions: Presets.DefaultOptions,
			consumables: Presets.DefaultConsumables,
			otherDefaults: Presets.OtherDefaults,
			defaultFactionRaces: {
				[Faction.Unknown]: Race.RaceUnknown,
				[Faction.Alliance]: Race.RaceGnome,
				[Faction.Horde]: Race.RaceTroll,
			},
			defaultGear: {
				[Faction.Unknown]: {},
				[Faction.Alliance]: {
					1: Presets.BLANK_GEARSET.gear,
				},
				[Faction.Horde]: {
					1: Presets.BLANK_GEARSET.gear,
				},
			},
		},
	],
});

export class MageSimUI extends IndividualSimUI<Spec.SpecMage> {
	constructor(parentElem: HTMLElement, player: Player<Spec.SpecMage>) {
		super(parentElem, player, SPEC_CONFIG);
	}
}
