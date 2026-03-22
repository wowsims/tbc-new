import * as OtherInputs from '../../core/components/inputs/other_inputs';
import { IndividualSimUI, registerSpecConfig } from '../../core/individual_sim_ui';
import { Player } from '../../core/player';
import { PlayerClasses } from '../../core/player_classes';
import { Mage } from '../../core/player_classes/mage';
import { APLRotation } from '../../core/proto/apl';
import { Faction, ItemSlot, PseudoStat, Race, Spec, Stat } from '../../core/proto/common';
import { DEFAULT_CASTER_GEM_STATS, Stats, UnitStat } from '../../core/proto_utils/stats';
import { DefaultDebuffs, DefaultRaidBuffs, DefaultPartyBuffs, DefaultIndividualBuffs, DefaultConsumables } from './presets';
import * as Presets from './presets';
import * as MageInputs from './inputs';
import { Mage_Rotation } from '../../core/proto/mage';
import { ReforgeOptimizer } from '../../core/components/suggest_reforges_action';

const SPEC_CONFIG = registerSpecConfig(Spec.SpecMage, {
	requiredTalentRows: [],
	cssClass: 'mage-sim-ui',
	cssScheme: PlayerClasses.getCssClass(Mage),
	// List any known bugs / issues here and they'll be shown on the site.
	knownIssues: [],

	// All stats for which EP should be calculated.
	epStats: [
		Stat.StatIntellect,
		Stat.StatSpirit,
		Stat.StatSpellDamage,
		Stat.StatArcaneDamage,
		Stat.StatFrostDamage,
		Stat.StatFireDamage,
		Stat.StatSpellPenetration,
		Stat.StatSpellHitRating,
		Stat.StatSpellCritRating,
		Stat.StatSpellHasteRating,
		Stat.StatMana,
		Stat.StatMP5,
	],
	epPseudoStats: [PseudoStat.PseudoStatSchoolHitPercentArcane, PseudoStat.PseudoStatSchoolHitPercentFire, PseudoStat.PseudoStatSchoolHitPercentFrost],
	// Reference stat against which to calculate EP. I think all classes use either spell power or attack power.
	epReferenceStat: Stat.StatSpellDamage,
	// Which stats to display in the Character Stats section, at the bottom of the left-hand sidebar.
	displayStats: UnitStat.createDisplayStatArray(
		[
			Stat.StatHealth,
			Stat.StatMana,
			Stat.StatStamina,
			Stat.StatIntellect,
			Stat.StatSpirit,
			Stat.StatSpellDamage,
			Stat.StatFrostDamage,
			Stat.StatFireDamage,
			Stat.StatArcaneDamage,
		],
		[
			PseudoStat.PseudoStatSpellHitPercent,
			PseudoStat.PseudoStatSchoolHitPercentArcane,
			PseudoStat.PseudoStatSchoolHitPercentFire,
			PseudoStat.PseudoStatSchoolHitPercentFrost,
			PseudoStat.PseudoStatSpellCritPercent,
			PseudoStat.PseudoStatSpellHastePercent,
		],
	),

	modifyDisplayStats: (player: Player<Spec.SpecMage>) => {
		return {
			talents: new Stats().addPseudoStat(PseudoStat.PseudoStatSpellCritPercent, player.getTalents().arcaneInstability),
		};
	},

	gemStats: DEFAULT_CASTER_GEM_STATS,

	consumableStats: [
		Stat.StatIntellect,
		Stat.StatSpirit,
		Stat.StatMP5,
		Stat.StatMana,
		Stat.StatSpellDamage,
		Stat.StatFrostDamage,
		Stat.StatFireDamage,
		Stat.StatArcaneDamage,
		Stat.StatSpellCritRating,
		Stat.StatSpellHitRating,
		Stat.StatSpellHasteRating,
	],

	defaults: {
		// Default equipped gear.
		gear: Presets.P1_BIS_ARCANE.gear,
		// Default EP weights for sorting gear in the gear picker.
		epWeights: Presets.P1_EP_PRESET.epWeights,
		statCaps: (() => {
			return new Stats().withPseudoStat(PseudoStat.PseudoStatSchoolHitPercentArcane, 16);
		})(),
		// Default consumes settings.
		consumables: Presets.DefaultConsumables,
		// Default talents.
		talents: Presets.ARCANE_TALENTS.data,
		// Default spec-specific settings.
		specOptions: Presets.DefaultOptions,
		other: Presets.OtherDefaults,
		// Default raid/party buffs settings.
		raidBuffs: DefaultRaidBuffs,

		partyBuffs: DefaultPartyBuffs,
		individualBuffs: DefaultIndividualBuffs,
		debuffs: DefaultDebuffs,
	},

	// IconInputs to include in the 'Player' section on the settings tab.
	playerIconInputs: [MageInputs.MageArmorInputs()],
	// Buff and Debuff inputs to include/exclude, overriding the EP-based defaults.
	includeBuffDebuffInputs: [Stat.StatMP5],
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
		epWeights: [Presets.P1_EP_PRESET],
		// Preset rotations that the user can quickly select.
		rotations: [Presets.ROTATION_PRESET_ARCANE],
		// Preset talents that the user can quickly select.
		talents: [Presets.ARCANE_TALENTS],
		// Preset gear configurations that the user can quickly select.
		gear: [Presets.PREBIS_ARCANE, Presets.P1_BIS_ARCANE],

		builds: [],
	},

	autoRotation: (player: Player<Spec.SpecMage>): APLRotation => {
		// const numTargets = player.sim.encounter.targets.length;
		// if (numTargets >= 2) {
		// 	return Presets.ROTATION_PRESET_CLEAVE.rotation.rotation!;
		// } else {
		return Presets.ROTATION_PRESET_ARCANE.rotation.rotation!;
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

		this.reforger = new ReforgeOptimizer(this);
	}
}
