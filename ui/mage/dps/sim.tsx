import * as OtherInputs from '../../core/components/inputs/other_inputs';
import { IndividualSimUI, registerSpecConfig } from '../../core/individual_sim_ui';
import { Player } from '../../core/player';
import { PlayerClasses } from '../../core/player_classes';
import { Mage } from '../../core/player_classes/mage';
import { APLListItem, APLRotation, APLRotation_Type, APLValueVariable } from '../../core/proto/apl';
import { Cooldowns, Faction, ItemSlot, PseudoStat, Race, Spec, Stat } from '../../core/proto/common';
import { DEFAULT_CASTER_GEM_STATS, Stats, UnitStat } from '../../core/proto_utils/stats';
import { DefaultDebuffs, DefaultRaidBuffs, DefaultPartyBuffs, DefaultIndividualBuffs, DefaultConsumables } from './presets';
import { SpecRotation } from '../../core/proto_utils/utils';
import * as AplUtils from '../../core/proto_utils/apl_utils';
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
		consumables: DefaultConsumables,
		// Default talents.
		talents: Presets.ARCANE_TALENTS.data,
		// Default spec-specific settings.
		specOptions: Presets.DefaultOptions,
		other: Presets.OtherDefaults,
		// Default raid/party buffs settings.
		raidBuffs: DefaultRaidBuffs,

		partyBuffs: DefaultPartyBuffs,
		individualBuffs: DefaultIndividualBuffs,

		rotationType: APLRotation_Type.TypeSimple,
		simpleRotation: Presets.ArcaneMageSimpleRotation,
		debuffs: DefaultDebuffs,
	},

	// IconInputs to include in the 'Player' section on the settings tab.
	playerIconInputs: [MageInputs.MageArmorInputs()],
	rotationInputs: MageInputs.ArcaneMageRotationConfig,
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
		epWeights: [Presets.P1_EP_PRESET, Presets.P2_EP_PRESET],
		// Preset rotations that the user can quickly select.
		rotations: [Presets.ROTATION_PRESET_ARCANE, Presets.APL_ARCANE_SIMPLE, Presets.ROTATION_PRESET_ARCANEBRAID],
		// Preset talents that the user can quickly select.
		talents: [Presets.ARCANE_TALENTS],
		// Preset gear configurations that the user can quickly select.
		gear: [Presets.PREBIS_ARCANE, Presets.P1_BIS_ARCANE, Presets.P2_BIS_ARCANE],

		builds: [Presets.P1_PRESET_BUILD_ARC, Presets.P2_PRESET_BUILD_ARC],
	},

	autoRotation: (player: Player<Spec.SpecMage>): APLRotation => {
		// const numTargets = player.sim.encounter.targets.length;
		// if (numTargets >= 2) {
		// 	return Presets.ROTATION_PRESET_CLEAVE.rotation.rotation!;
		// } else {
		return Presets.ROTATION_PRESET_ARCANE.rotation.rotation!;
		// }
	},

	simpleRotation: (player: Player<Spec.SpecMage>, simple: SpecRotation<Spec.SpecMage>, cooldowns: Cooldowns): APLRotation => {
		const actions = AplUtils.simpleCooldownActions(cooldowns);
		const rotation = APLRotation.clone(Presets.ROTATION_PRESET_ARCANE.rotation.rotation!);

		const { conserveStart = 20, conserveEnd = 30, delayMajorCDs = 10 } = simple;

		const conserveStartString = APLValueVariable.fromJson({
			name: 'Conserve Start',
			value: { const: { val: String(conserveStart) + '%' } },
		});

		const conserveEndString = APLValueVariable.fromJson({
			name: 'Conserve End',
			value: { const: { val: String(conserveEnd) + '%' } },
		});

		const delayMajorCDsString = APLValueVariable.fromJson({
			name: 'Delay Major CDs',
			value: { const: { val: String(delayMajorCDs) + 's' } },
		});

		rotation.valueVariables[0] = conserveStartString;
		rotation.valueVariables[1] = conserveEndString;
		rotation.valueVariables[2] = delayMajorCDsString;

		return APLRotation.create({
			prepullActions: rotation.prepullActions,
			priorityList: [
				...actions.map(action =>
					APLListItem.create({
						action: action,
					}),
				),
				...rotation.priorityList,
			],
			groups: rotation.groups,
			valueVariables: rotation.valueVariables,
		});
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
