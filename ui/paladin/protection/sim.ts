import * as OtherInputs from '../../core/components/inputs/other_inputs.js';
import { IndividualSimUI, registerSpecConfig } from '../../core/individual_sim_ui.js';
import { Player } from '../../core/player.js';
import { PlayerClasses } from '../../core/player_classes';
import { APLRotation, APLRotation_Type, APLValueVariable, SimpleRotation } from '../../core/proto/apl.js';
import { Cooldowns, Faction, PseudoStat, Race, Spec, Stat } from '../../core/proto/common.js';
import { Stats, UnitStat } from '../../core/proto_utils/stats.js';
import * as Presets from './presets.js';
import * as ProtPaladinInputs from './inputs.js';
import * as Mechanics from '../../core/constants/mechanics';
import { ReforgeOptimizer } from '../../core/components/suggest_reforges_action';

const SPEC_CONFIG = registerSpecConfig(Spec.SpecProtectionPaladin, {
	cssClass: 'protection-paladin-sim-ui',
	cssScheme: PlayerClasses.getCssClass(PlayerClasses.Paladin),
	// List any known bugs / issues here and they'll be shown on the site.
	knownIssues: [],
	consumableStats: [Stat.StatStamina, Stat.StatHealth, Stat.StatMana],

	// All stats for which EP should be calculated.
	epStats: [
		Stat.StatStamina,
		Stat.StatStrength,
		Stat.StatSpellDamage,
		Stat.StatAgility,
		Stat.StatAttackPower,
		Stat.StatMeleeHitRating,
		Stat.StatMeleeHasteRating,
		Stat.StatMeleeCritRating,
		Stat.StatArmorPenetration,
		Stat.StatExpertiseRating,
		Stat.StatResilienceRating,
		Stat.StatDefenseRating,
		Stat.StatDodgeRating,
		Stat.StatParryRating,
		Stat.StatArmor,
		Stat.StatBonusArmor,
		Stat.StatExpertiseRating,
	],
	epPseudoStats: [PseudoStat.PseudoStatMainHandDps],
	// Reference stat against which to calculate EP. I think all classes use either spell power or attack power.
	epReferenceStat: Stat.StatStrength,
	// Which stats to display in the Character Stats section, at the bottom of the left-hand sidebar.
	displayStats: UnitStat.createDisplayStatArray(
		[
			Stat.StatHealth,
			Stat.StatMana,
			Stat.StatArmor,
			Stat.StatBonusArmor,
			Stat.StatStamina,
			Stat.StatStrength,
			Stat.StatSpellDamage,
			Stat.StatAgility,
			Stat.StatAttackPower,
			Stat.StatBlockValue,
			Stat.StatDefenseRating,
			Stat.StatResilienceRating,
			Stat.StatArcaneResistance,
			Stat.StatFireResistance,
			Stat.StatFrostResistance,
			Stat.StatNatureResistance,
			Stat.StatShadowResistance,
			Stat.StatExpertiseRating,
		],
		[
			PseudoStat.PseudoStatMeleeHitPercent,
			PseudoStat.PseudoStatMeleeCritPercent,
			PseudoStat.PseudoStatMeleeHastePercent,
			PseudoStat.PseudoStatSpellHitPercent,
			PseudoStat.PseudoStatSpellCritPercent,
			PseudoStat.PseudoStatBlockPercent,
			PseudoStat.PseudoStatDodgePercent,
			PseudoStat.PseudoStatParryPercent,
		],
	),

	defaults: {
		// Default equipped gear.
		gear: Presets.P1_GEAR_PRESET.gear,
		statCaps: (() => {
			const hitCap = new Stats().withPseudoStat(PseudoStat.PseudoStatMeleeHitPercent, 9);
			const expCap = new Stats().withStat(Stat.StatExpertiseRating, 6.5 * 4 * Mechanics.EXPERTISE_PER_QUARTER_PERCENT_REDUCTION);
			const critImmunityCap = new Stats().withPseudoStat(PseudoStat.PseudoStatReducedCritTakenPercent, 5.6);

			return hitCap.add(expCap).add(critImmunityCap);
		})(),
		// Default EP weights for sorting gear in the gear picker.
		// Values for now are pre-Cata initial WAG
		epWeights: Presets.P4_EP_PRESET.epWeights,
		// Default consumes settings.
		consumables: Presets.DefaultConsumables,
		// Default talents.
		talents: Presets.DefaultTalents.data,
		// Default spec-specific settings.
		specOptions: Presets.DefaultOptions,
		other: Presets.OtherDefaults,
		// Default raid/party buffs settings.
		raidBuffs: Presets.DefaultRaidBuffs,
		partyBuffs: Presets.DefaultPartyBuffs,
		individualBuffs: Presets.DefaultIndividualBuffs,
		debuffs: Presets.DefaultDebuffs,
		simpleRotation: Presets.DefaultSimpleRotation,
		rotationType: APLRotation_Type.TypeSimple,
		encounter: "Magtheridon's Lair/Magtheridon 25",
	},

	rotationInputs: ProtPaladinInputs.PaladinRotationConfig,
	// IconInputs to include in the 'Player' section on the settings tab.
	playerIconInputs: [],
	// Buff and Debuff inputs to include/exclude, overriding the EP-based defaults.
	includeBuffDebuffInputs: [Stat.StatMP5, Stat.StatIntellect],
	excludeBuffDebuffInputs: [],
	// Inputs to include in the 'Other' section on the settings tab.
	otherInputs: {
		inputs: [
			OtherInputs.InputDelay,
			OtherInputs.TankAssignment,
			OtherInputs.InspirationUptime,
			OtherInputs.IncomingHps,
			OtherInputs.HealingCadence,
			OtherInputs.HealingCadenceVariation,
			OtherInputs.AbsorbFrac,
			OtherInputs.BurstWindow,
			OtherInputs.HpPercentForDefensives,
			OtherInputs.InFrontOfTarget,
		],
	},
	encounterPicker: {
		// Whether to include 'Execute Duration (%)' in the 'Encounter' section of the settings tab.
		showExecuteProportion: false,
	},

	presets: {
		epWeights: [Presets.P4_EP_PRESET],
		// Preset talents that the user can quickly select.
		talents: [Presets.DefaultTalents],
		// Preset rotations that the user can quickly select.
		rotations: [Presets.APL_SIMPLE, Presets.APL_PRESET],
		// Preset gear configurations that the user can quickly select.
		gear: [Presets.P1_GEAR_PRESET, Presets.P2_GEAR_PRESET, Presets.P3_GEAR_PRESET, Presets.P4_GEAR_PRESET, Presets.P5_GEAR_PRESET],
		builds: [],
	},

	autoRotation: (_player: Player<Spec.SpecProtectionPaladin>): APLRotation => {
		return Presets.APL_PRESET.rotation.rotation!;
	},

	simpleRotation: (_player, simple): APLRotation => {
		const rotation = APLRotation.clone(Presets.APL_PRESET.rotation.rotation!);

		const {
			prioritizeHolyShield = true,
			useConsecrate = true,
			useExorcism = false,
			useAvengersShield = true,
			maintainJudgementOfWisdom = true
		} = simple;

		rotation.valueVariables = [
			APLValueVariable.fromJson({ name: 'Prioritize Holy Shield', value: { const: { val: String(prioritizeHolyShield) } } }),
			APLValueVariable.fromJson({ name: 'Use Consecrate', value: { const: { val: String(useConsecrate) } } }),
			APLValueVariable.fromJson({ name: 'Use Exorcism', value: { const: { val: String(useExorcism) } } }),
			APLValueVariable.fromJson({ name: "Use Avenger's Shield", value: { const: { val: String(useAvengersShield) } } }),
			APLValueVariable.fromJson({ name: 'Maintain Judgement of Wisdom', value: { const: { val: String(maintainJudgementOfWisdom) } } }),
		];

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

	raidSimPresets: [],
});

export class ProtectionPaladinSimUI extends IndividualSimUI<Spec.SpecProtectionPaladin> {
	constructor(parentElem: HTMLElement, player: Player<Spec.SpecProtectionPaladin>) {
		super(parentElem, player, SPEC_CONFIG);
		this.reforger = new ReforgeOptimizer(this);
	}
}
