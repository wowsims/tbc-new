import * as OtherInputs from '../../core/components/inputs/other_inputs';
import * as Mechanics from '../../core/constants/mechanics';
import { IndividualSimUI, registerSpecConfig } from '../../core/individual_sim_ui';
import { Player } from '../../core/player';
import { PlayerClasses } from '../../core/player_classes';

import { APLRotation, APLRotation_Type } from '../../core/proto/apl';
import { Faction, ItemSlot, PseudoStat, Race, Spec, Stat } from '../../core/proto/common';
import { StatCapType } from '../../core/proto/ui';
import { DEFAULT_HYBRID_CASTER_GEM_STATS, StatCap, Stats, UnitStat, UnitStatPresets } from '../../core/proto_utils/stats';
import { formatToNumber } from '../../core/utils';
import * as DruidInputs from '../inputs';
import * as BalanceInputs from './inputs';
import * as Presets from './presets';

const SPEC_CONFIG = registerSpecConfig(Spec.SpecBalanceDruid, {
	cssClass: 'balance-druid-sim-ui',
	cssScheme: PlayerClasses.getCssClass(PlayerClasses.Druid),
	// List any known bugs / issues here, and they'll be shown on the site.
	knownIssues: [],

	// All stats for which EP should be calculated.
	epStats: [Stat.StatIntellect, Stat.StatSpirit, Stat.StatSpellDamage],
	// Reference stat against which to calculate EP. I think all classes use either spell power or attack power.
	epReferenceStat: Stat.StatIntellect,
	// Which stats to display in the Character Stats section, at the bottom of the left-hand sidebar.
	displayStats: UnitStat.createDisplayStatArray(
		[
			Stat.StatHealth,
			Stat.StatMana,
			Stat.StatStamina,
			Stat.StatIntellect,
			Stat.StatSpirit,
			Stat.StatSpellDamage,
		],
		[PseudoStat.PseudoStatSpellHitPercent, PseudoStat.PseudoStatSpellCritPercent, PseudoStat.PseudoStatSpellHastePercent],
	),
	gemStats: DEFAULT_HYBRID_CASTER_GEM_STATS,

	modifyDisplayStats: (player: Player<Spec.SpecBalanceDruid>) => {
		const playerStats = player.getCurrentStats();
		const gearStats = Stats.fromProto(playerStats.gearStats);
		const talentsStats = Stats.fromProto(playerStats.talentsStats);
		const talentsDelta = talentsStats.subtract(gearStats);
		const talentsMod = new Stats().withStat(
			Stat.StatSpellHitRating,
			talentsDelta.getPseudoStat(PseudoStat.PseudoStatSpellHitPercent) * Mechanics.SPELL_HIT_RATING_PER_HIT_PERCENT,
		);

		return {
			talents: talentsMod,
		};
	},

	defaults: {
		// Default equipped gear.
		gear: Presets.T14PresetGear.gear,
		// Default EP weights for sorting gear in the gear picker.
		epWeights: Presets.StandardEPWeights.epWeights,
		// Default consumes settings.
		consumables: Presets.DefaultConsumables,
		// Default talents.
		talents: Presets.StandardTalents.data,
		// Default spec-specific settings.
		specOptions: Presets.DefaultOptions,
		// Default raid/party buffs settings.
		raidBuffs: Presets.DefaultRaidBuffs,
		partyBuffs: Presets.DefaultPartyBuffs,
		individualBuffs: Presets.DefaultIndividualBuffs,
		debuffs: Presets.DefaultDebuffs,
		other: Presets.OtherDefaults,
		rotationType: APLRotation_Type.TypeAuto,
	},

	// IconInputs to include in the 'Player' section on the settings tab.
	playerIconInputs: [DruidInputs.SelfInnervate()],
	// Buff and Debuff inputs to include/exclude, overriding the EP-based defaults.
	includeBuffDebuffInputs: [],
	excludeBuffDebuffInputs: [],
	// Inputs to include in the 'Other' section on the settings tab.
	otherInputs: {
		inputs: [BalanceInputs.OkfUptime, OtherInputs.TankAssignment, OtherInputs.InputDelay, OtherInputs.DistanceFromTarget],
	},
	itemSwapSlots: [ItemSlot.ItemSlotMainHand, ItemSlot.ItemSlotOffHand, ItemSlot.ItemSlotTrinket1, ItemSlot.ItemSlotTrinket2],
	encounterPicker: {
		// Whether to include 'Execute Duration (%)' in the 'Encounter' section of the settings tab.
		showExecuteProportion: false,
	},

	presets: {
		epWeights: [Presets.StandardEPWeights],
		// Preset talents that the user can quickly select.
		talents: [Presets.StandardTalents],
		rotations: [Presets.StandardRotation],
		// Preset gear configurations that the user can quickly select.
		gear: [Presets.PreraidPresetGear, Presets.T14PresetGear, Presets.T14UpgradedPresetGear, Presets.T15PresetGear /*, Presets.T16PresetGear*/],
		builds: [Presets.PresetPreraidBuild, Presets.T14PresetBuild,Presets.T15PresetBuild /*, Presets.T16PresetBuild*/],
	},

	autoRotation: (_player: Player<Spec.SpecBalanceDruid>): APLRotation => {
		return Presets.StandardRotation.rotation.rotation!;
	},

	raidSimPresets: [
		{
			spec: Spec.SpecBalanceDruid,
			talents: Presets.StandardTalents.data,
			specOptions: Presets.DefaultOptions,
			consumables: Presets.DefaultConsumables,
			otherDefaults: Presets.OtherDefaults,
			defaultFactionRaces: {
				[Faction.Unknown]: Race.RaceUnknown,
				[Faction.Alliance]: Race.RaceTauren,
				[Faction.Horde]: Race.RaceTroll,
			},
			defaultGear: {
				[Faction.Unknown]: {},
				[Faction.Alliance]: {
					1: Presets.PreraidPresetGear.gear,
				},
				[Faction.Horde]: {
					1: Presets.PreraidPresetGear.gear,
				},
			},
		},
	],
});

export class BalanceDruidSimUI extends IndividualSimUI<Spec.SpecBalanceDruid> {
	constructor(parentElem: HTMLElement, player: Player<Spec.SpecBalanceDruid>) {
		super(parentElem, player, SPEC_CONFIG);
		const statSelectionHastePreset = {
			unitStat: UnitStat.fromPseudoStat(PseudoStat.PseudoStatSpellHastePercent),
			presets: new Map<string, number>([]),
		};

		const modifyHaste = (oldHastePercent: number, modifier: number) =>
			Number(formatToNumber(((oldHastePercent / 100 + 1) / modifier - 1) * 100, { maximumFractionDigits: 5 }));

		const createHasteBreakpointVariants = (name: string, breakpoint: number, prefix?: string) => {
			const breakpoints = new Map<string, number>();
			breakpoints.set(`${prefix ? `${prefix} - ` : ''}${name}`, breakpoint);

			const blBreakpoint = modifyHaste(breakpoint, 1.3);
			if (blBreakpoint > 0) {
				breakpoints.set(`${prefix ? `${prefix} - ` : ''}BL - ${name}`, blBreakpoint);
			}

			const berserkingBreakpoint = modifyHaste(breakpoint, 1.2);
			if (berserkingBreakpoint > 0) {
				breakpoints.set(`${prefix ? `${prefix} - ` : ''}Zerk - ${name}`, berserkingBreakpoint);
			}

			const blZerkingBreakpoint = modifyHaste(blBreakpoint, 1.2);
			if (blZerkingBreakpoint > 0) {
				breakpoints.set(`${prefix ? `${prefix} - ` : ''}BL+Zerk - ${name}`, blZerkingBreakpoint);
			}

			return breakpoints;
		};

		for (const [name, breakpoint] of Presets.BALANCE_T14_4P_BREAKPOINTS!.presets) {
			const variants = createHasteBreakpointVariants(name, breakpoint, 'T14 4P');
			for (const [variantName, variantValue] of variants) {
				statSelectionHastePreset.presets.set(variantName, variantValue);
			}
		}

		for (const [name, breakpoint] of Presets.BALANCE_BREAKPOINTS!.presets) {
			const variants = createHasteBreakpointVariants(name, breakpoint);
			for (const [variantName, variantValue] of variants) {
				statSelectionHastePreset.presets.set(variantName, variantValue);
			}
		}
	}
}
