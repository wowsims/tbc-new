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
	FeralCatDruid_Rotation_HotwStrategy as HotwStrategy,
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
		const [prepullActions, actions] = AplUtils.standardCooldownDefaults(cooldowns);

		// Rotation entries
		const agiTrinkets = APLAction.fromJsonString(
			`{"condition":{"or":{"vals":[{"auraIsActive":{"auraId":{"spellId":5217}}},{"cmp":{"op":"OpLt","lhs":{"remainingTime":{}},"rhs":{"const":{"val":"16s"}}}}]}},"castAllStatBuffCooldowns":{"statType1":1,"statType2":-1,"statType3":-1}}`,
		);
		const synapseSprings = APLAction.fromJsonString(
			`{"condition":{"or":{"vals":[{"auraIsActive":{"auraId":{"spellId":5217}}},{"cmp":{"op":"OpLt","lhs":{"remainingTime":{}},"rhs":{"const":{"val":"11s"}}}}]}},"castSpell":{"spellId":{"spellId":126734}}}`,
		);
		const hasteTrinkets = APLAction.fromJsonString(
			`{"condition":{"or":{"vals":[{"auraIsActive":{"auraId":{"spellId":5217}}},{"cmp":{"op":"OpLt","lhs":{"remainingTime":{}},"rhs":{"const":{"val":"16s"}}}}]}},"castAllStatBuffCooldowns":{"statType1":7,"statType2":-1,"statType3":-1}}`,
		);
		const trees = APLAction.fromJsonString(
			`{"condition":{"or":{"vals":[{"anyStatBuffCooldownsActive":{"statType1":1,"statType2":-1,"statType3":-1}},{"cmp":{"op":"OpLt","lhs":{"remainingTime":{}},"rhs":{"const":{"val":"16s"}}}}]}},"castSpell":{"spellId":{"spellId":106737}}}`,
		);
		const potion = APLAction.fromJsonString(
			`{"condition":{"or":{"vals":[{"and":{"vals":[{"auraIsActive":{"auraId":{"spellId":5217}}},{"cmp":{"op":"OpLt","lhs":{"remainingTime":{}},"rhs":{"math":{"op":"OpAdd","lhs":{"spellTimeToReady":{"spellId":{"spellId":106952}}},"rhs":{"const":{"val":"26s"}}}}}}]}},{"cmp":{"op":"OpLt","lhs":{"remainingTime":{}},"rhs":{"const":{"val":"26s"}}}},{"auraIsActive":{"auraId":{"spellId":106951}}}]}},"castSpell":{"spellId":{"itemId":76089}}}`,
		);
		const trollRacial = APLAction.fromJsonString(`{"condition":{"auraIsActive":{"auraId":{"spellId":106951}}},"castSpell":{"spellId":{"spellId":26297}}}`);
		const blockZerk = APLAction.fromJsonString(`{"condition":{"const":{"val":"false"}},"castSpell":{"spellId":{"spellId":106952}}}`);
		const blockNS = APLAction.fromJsonString(`{"condition":{"const":{"val":"false"}},"castSpell":{"spellId":{"spellId":132158}}}`);
		const blockHotw = APLAction.fromJsonString(`{"condition":{"const":{"val":"false"}},"castSpell":{"spellId":{"spellId":108292}}}`);
		const shouldUseHotw = player.getTalents().heartOfTheWild && simple.hotwStrategy != HotwStrategy.PassivesOnly;
		const shouldWrathWeave = shouldUseHotw && simple.hotwStrategy == HotwStrategy.Wrath;
		const doRotation = APLAction.fromJsonString(
			`{"catOptimalRotationAction":{"rotationType":${simple.rotationType},"manualParams":${simple.manualParams},"allowAoeBerserk":${simple.allowAoeBerserk},"bearWeave":${simple.bearWeave},"snekWeave":${simple.snekWeave},"useNs":${simple.useNs},"wrathWeave":${shouldWrathWeave},"minRoarOffset":${simple.minRoarOffset.toFixed(2)},"ripLeeway":${simple.ripLeeway.toFixed(2)},"useBite":${simple.useBite},"biteTime":${simple.biteTime.toFixed(2)},"berserkBiteTime":${simple.berserkBiteTime.toFixed(2)}}}`,
		);

		const singleTarget = simple.rotationType == FeralRotationType.SingleTarget;
		actions.push(
			...([
				singleTarget ? agiTrinkets : null,
				singleTarget ? synapseSprings : null,
				singleTarget ? hasteTrinkets : null,
				singleTarget ? trees : null,
				singleTarget ? potion : null,
				singleTarget ? trollRacial : null,
				blockZerk,
				blockNS,
				!shouldUseHotw ? blockHotw : null,
				doRotation,
			].filter(a => a) as Array<APLAction>),
		);

		// Pre-pull entries
		const healingTouch = APLPrepullAction.fromJsonString(`{"action":{"castSpell":{"spellId":{"spellId":5185}}},"doAtValue":{"const":{"val":"-5.2s"}}}`);
		const shiftCat = APLPrepullAction.fromJsonString(`{"action":{"castSpell":{"spellId":{"spellId":768}}},"doAtValue":{"const":{"val":"-2.6s"}}}`);
		const preRoar = APLPrepullAction.fromJsonString(`{"action":{"castSpell":{"spellId":{"spellId":52610}}},"doAtValue":{"const":{"val":"-1s"}}}`);

		prepullActions.push(
			...([
				// player.getTalents().dreamOfCenarius ? healingTouch : null,
				// player.getTalents().dreamOfCenarius ? shiftCat : null,
			].filter(a => a) as Array<APLPrepullAction>),
		);

		return APLRotation.create({
			prepullActions: prepullActions,
			priorityList: actions.map(action =>
				APLListItem.create({
					action: action,
				}),
			),
		});
	},

	hiddenMCDs: [126734, 106737, 76089, 26297, 106952, 132158, 108292, 55004],

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
