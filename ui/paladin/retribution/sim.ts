import { stat } from 'node:fs';
import * as OtherInputs from '../../core/components/inputs/other_inputs.js';
import { IndividualSimUI, registerSpecConfig } from '../../core/individual_sim_ui.js';
import { Player } from '../../core/player.js';
import { PlayerClasses } from '../../core/player_classes';
import { APLRotation, APLRotation_Type, APLValueVariable, SimpleRotation } from '../../core/proto/apl.js';
import { Cooldowns, Debuffs, Faction, IndividualBuffs, PartyBuffs, PseudoStat, Race, RaidBuffs, Spec, Stat, UnitStats } from '../../core/proto/common.js';
import { Stats, UnitStat } from '../../core/proto_utils/stats.js';
import { DefaultDebuffs, DefaultRaidBuffs, DefaultPartyBuffs, DefaultIndividualBuffs, DefaultConsumables, DefaultSimpleRotation } from './presets';
import * as Presets from './presets.js';
import * as Inputs from './inputs.js';
import * as Mechanics from '../../core/constants/mechanics';
import { ReforgeOptimizer } from '../../core/components/suggest_reforges_action';

const SPEC_CONFIG = registerSpecConfig(Spec.SpecRetributionPaladin, {
	cssClass: 'retribution-paladin-sim-ui',
	cssScheme: PlayerClasses.getCssClass(PlayerClasses.Paladin),
	// List any known bugs / issues here and they'll be shown on the site.
	knownIssues: [],
	consumableStats: [Stat.StatMana],

	// All stats for which EP should be calculated.
	epStats: [
		Stat.StatStrength,
		Stat.StatSpellDamage,
		Stat.StatAgility,
		Stat.StatAttackPower,
		Stat.StatArmorPenetration,
		Stat.StatMeleeHitRating,
		Stat.StatMeleeHasteRating,
		Stat.StatMeleeCritRating,
		Stat.StatExpertiseRating,
		Stat.StatMana,
	],
	epPseudoStats: [PseudoStat.PseudoStatMainHandDps, PseudoStat.PseudoStatOffHandDps],
	// Reference stat against which to calculate EP. I think all classes use either spell power or attack power.
	epReferenceStat: Stat.StatStrength,
	// Which stats to display in the Character Stats section, at the bottom of the left-hand sidebar.
	displayStats: UnitStat.createDisplayStatArray(
		[
			Stat.StatStrength,
			Stat.StatAgility,
			Stat.StatIntellect,
			Stat.StatAttackPower,
			Stat.StatSpellDamage,
			Stat.StatMana,
			Stat.StatHealth,
			Stat.StatStamina,
			Stat.StatExpertiseRating,
			Stat.StatHolyDamage,
		],
		[
			PseudoStat.PseudoStatMeleeHitPercent,
			PseudoStat.PseudoStatMeleeCritPercent,
			PseudoStat.PseudoStatMeleeHastePercent,
			PseudoStat.PseudoStatSpellHastePercent,
			PseudoStat.PseudoStatSpellCritPercent,
			PseudoStat.PseudoStatSpellHitPercent,
		],
	),

	defaults: {
		// Default equipped gear.
		gear: Presets.P1_GEAR_PRESET.gear,
		// Default EP weights for sorting gear in the gear picker.
		epWeights: Presets.P1_EP_PRESET.epWeights,
		statCaps: (() => {
			const hitCap = new Stats().withPseudoStat(PseudoStat.PseudoStatMeleeHitPercent, 9);
			const expCap = new Stats().withStat(Stat.StatExpertiseRating, 6.5 * 4 * Mechanics.EXPERTISE_PER_QUARTER_PERCENT_REDUCTION);

			return hitCap.add(expCap);
		})(),
		// Default consumes settings.
		consumables: DefaultConsumables,
		// Default talents.
		talents: Presets.DefaultTalents.data,
		// Default spec-specific settings.
		specOptions: Presets.DefaultOptions,
		other: Presets.OtherDefaults,
		// Default raid/party buffs settings.
		raidBuffs: DefaultRaidBuffs,
		partyBuffs: DefaultPartyBuffs,
		individualBuffs: DefaultIndividualBuffs,
		debuffs: DefaultDebuffs,

		rotationType: APLRotation_Type.TypeSimple,
		simpleRotation: DefaultSimpleRotation,
	},

	// IconInputs to include in the 'Player' section on the settings tab.
	playerIconInputs: [],
	rotationInputs: Inputs.PaladinRotationConfig,
	// Buff and Debuff inputs to include/exclude, overriding the EP-based defaults.
	includeBuffDebuffInputs: [Stat.StatMP5],
	excludeBuffDebuffInputs: [],
	// Inputs to include in the 'Other' section on the settings tab.
	otherInputs: {
		inputs: [OtherInputs.InputDelay, OtherInputs.TankAssignment, OtherInputs.InFrontOfTarget],
	},
	encounterPicker: {
		// Whether to include 'Execute Duration (%)' in the 'Encounter' section of the settings tab.
		showExecuteProportion: false,
	},

	presets: {
		epWeights: [Presets.P1_EP_PRESET],
		rotations: [Presets.APL_PRESET, Presets.APL_SIMPLE],
		// Preset talents that the user can quickly select.
		talents: [Presets.DefaultTalents, Presets.NoKingsTalents, Presets.ImpMightTalents],
		// Preset gear configurations that the user can quickly select.
		gear: [Presets.PRERAID_GEAR_PRESET, Presets.P1_GEAR_PRESET],
		builds: [],
	},

	autoRotation: (_: Player<Spec.SpecRetributionPaladin>): APLRotation => {
		return Presets.APL_PRESET.rotation.rotation!;
	},

	simpleRotation: (player, simple): APLRotation => {
		const rotation = APLRotation.clone(Presets.APL_PRESET.rotation.rotation!);

		const { useExorcism = false, useConsecrate = false, delayMajorCDs = 11, prepullSotC = true } = simple;

		const useExorcismBool = APLValueVariable.fromJson({
			name: 'Use Exorcism',
			value: { const: { val: String(useExorcism) } },
		});

		const useConsecrateBool = APLValueVariable.fromJson({
			name: 'Use Consecrate',
			value: { const: { val: String(useConsecrate) } },
		});

		const delayMajorCDsString = APLValueVariable.fromJson({
			name: 'Delay Major CDs',
			value: { const: { val: String(delayMajorCDs) + 's' } },
		});

		const prepullSotCBool = APLValueVariable.fromJson({
			name: 'Prepull Seal Of the Crusader',
			value: { const: { val: String(prepullSotC) } },
		});

		rotation.valueVariables[2] = useExorcismBool;
		rotation.valueVariables[3] = useConsecrateBool;
		rotation.valueVariables[4] = delayMajorCDsString;
		rotation.valueVariables[5] = prepullSotCBool;

		return APLRotation.create({
			simple: SimpleRotation.create({
				cooldowns: Cooldowns.create(),
			}),
			prepullActions: rotation.prepullActions,
			priorityList: rotation.priorityList,
			groups: rotation.groups,
			valueVariables: rotation.valueVariables,
		});
	},

	//Handled by APL for major cds
	hiddenMCDs: [2825, 28730, 31884, 351355, 22838, 23827, 12662, 29383, 22788, 22105, 23334, 23381, 35476, 23737, 10646],

	raidSimPresets: [
		{
			spec: Spec.SpecRetributionPaladin,
			talents: Presets.DefaultTalents.data,
			specOptions: Presets.DefaultOptions,
			consumables: Presets.DefaultConsumables,
			defaultFactionRaces: {
				[Faction.Unknown]: Race.RaceUnknown,
				[Faction.Alliance]: Race.RaceHuman,
				[Faction.Horde]: Race.RaceBloodElf,
			},
			defaultGear: {
				[Faction.Unknown]: {},
				[Faction.Alliance]: {
					1: Presets.PRERAID_GEAR_PRESET.gear,
				},
				[Faction.Horde]: {
					1: Presets.PRERAID_GEAR_PRESET.gear,
				},
			},
		},
	],
});

export class RetributionPaladinSimUI extends IndividualSimUI<Spec.SpecRetributionPaladin> {
	constructor(parentElem: HTMLElement, player: Player<Spec.SpecRetributionPaladin>) {
		super(parentElem, player, SPEC_CONFIG);
		this.reforger = new ReforgeOptimizer(this);
	}
}
