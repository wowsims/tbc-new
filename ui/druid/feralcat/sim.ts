import * as OtherInputs from '../../core/components/inputs/other_inputs';
import { ReforgeOptimizer } from '../../core/components/suggest_reforges_action';
import { Phase } from '../../core/constants/other';
import * as Mechanics from '../../core/constants/mechanics';
import { IndividualSimUI, registerSpecConfig } from '../../core/individual_sim_ui';
import { Player } from '../../core/player';
import { PlayerClasses } from '../../core/player_classes';
import { APLAction, APLListItem, APLRotation, APLRotation_Type as APLRotationType } from '../../core/proto/apl';
import { Cooldowns, Debuffs, Drums, Faction, IndividualBuffs, ItemSlot, PartyBuffs, PseudoStat, Race, RaidBuffs, Spec, Stat, TristateEffect } from '../../core/proto/common';
import { FeralCatDruid_Rotation as DruidRotation } from '../../core/proto/druid';
import { StatCap, Stats, UnitStat } from '../../core/proto_utils/stats';
import { StatCapType } from '../../core/proto/ui';
import { defaultExposeWeaknessSettings, defaultRaidBuffMajorDamageCooldowns } from '../../core/proto_utils/utils';
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

	// Extra stats shown in consumables picker (beyond epStats).
	consumableStats: [Stat.StatMeleeHasteRating, Stat.StatMana, Stat.StatSpirit, Stat.StatMP5],

	// All stats for which EP should be calculated.
	epStats: [
		Stat.StatStrength,
		Stat.StatAgility,
		Stat.StatAttackPower,
		Stat.StatFeralAttackPower,
		Stat.StatMeleeHitRating,
		Stat.StatExpertiseRating,
		Stat.StatMeleeCritRating,
		Stat.StatMeleeHasteRating,
		Stat.StatArmorPenetration,
	],
	gemStats: [Stat.StatAgility, Stat.StatStrength],
	epPseudoStats: [],
	// Reference stat against which to calculate EP. I think all classes use either spell power or attack power.
	epReferenceStat: Stat.StatAgility,
	// Which stats to display in the Character Stats section, at the bottom of the left-hand sidebar.
	displayStats: UnitStat.createDisplayStatArray(
		[Stat.StatHealth, Stat.StatStrength, Stat.StatAgility, Stat.StatStamina, Stat.StatIntellect, Stat.StatSpirit, Stat.StatAttackPower, Stat.StatMana, Stat.StatExpertiseRating, Stat.StatArmorPenetration],
		[PseudoStat.PseudoStatMeleeHitPercent, PseudoStat.PseudoStatMeleeCritPercent, PseudoStat.PseudoStatMeleeHastePercent],
	),

	defaults: {
		// Default equipped gear.
		gear: Presets.P1_GEARSET.gear,
		// Default EP weights for sorting gear in the gear picker.
		epWeights: Presets.P1_EP_PRESET.epWeights,
		statCaps: (() => {
			return new Stats()
				.withPseudoStat(PseudoStat.PseudoStatMeleeHitPercent, 9)
				.withStat(Stat.StatExpertiseRating, 6.5 * 4 * Mechanics.EXPERTISE_PER_QUARTER_PERCENT_REDUCTION);
		})(),
		other: Presets.OtherDefaults,
		// Default consumes settings.
		consumables: Presets.DefaultConsumables,
		// Default rotation settings.
		rotationType: APLRotationType.TypeSimple,
		simpleRotation: Presets.DefaultRotation,
		// Default talents.
		talents: Presets.StandardTalents.data,
		// Default spec-specific settings.
		specOptions: Presets.DefaultOptions,
		// Default raid/party buffs settings.
		raidBuffs: RaidBuffs.create({
			...defaultRaidBuffMajorDamageCooldowns(),
			arcaneBrilliance: true,
			giftOfTheWild: TristateEffect.TristateEffectImproved,
			powerWordFortitude: TristateEffect.TristateEffectImproved,
		}),
		partyBuffs: PartyBuffs.create({
			drums: Drums.LesserDrumsOfBattle,
			ferociousInspiration: 2,
			battleShout: TristateEffect.TristateEffectImproved,
			graceOfAirTotem: TristateEffect.TristateEffectImproved,
			windfuryTotem: TristateEffect.TristateEffectImproved,
			manaSpringTotem: TristateEffect.TristateEffectRegular,
			strengthOfEarthTotem: TristateEffect.TristateEffectImproved,
			totemTwisting: true,
		}),
		individualBuffs: IndividualBuffs.create({
			blessingOfKings: true,
			blessingOfMight: TristateEffect.TristateEffectImproved,
			unleashedRage: true,
		}),
		debuffs: Debuffs.create({
			...defaultExposeWeaknessSettings(Phase.Phase1),
			bloodFrenzy: true,
			exposeArmor: TristateEffect.TristateEffectImproved,
			huntersMark: TristateEffect.TristateEffectImproved,
			improvedSealOfTheCrusader: TristateEffect.TristateEffectImproved,
			judgementOfWisdom: true,
			misery: true,
			curseOfRecklessness: true,
			faerieFire: TristateEffect.TristateEffectImproved,
			sunderArmor: true,
			giftOfArthas: true,
		}),
	},

	// IconInputs to include in the 'Player' section on the settings tab.
	playerIconInputs: [],
	// Inputs to include in the 'Rotation' section on the settings tab.
	rotationInputs: FeralInputs.FeralDruidRotationConfig,
	// Buff and Debuff inputs to include/exclude, overriding the EP-based defaults.
	includeBuffDebuffInputs: [Stat.StatMP5, Stat.StatIntellect, Stat.StatStamina],
	excludeBuffDebuffInputs: [Stat.StatParryRating],
	// Inputs to include in the 'Other' section on the settings tab.
	otherInputs: {
		inputs: [
			OtherInputs.TotemTwisting,
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
		epWeights: [Presets.P1_EP_PRESET],
		// Preset talents that the user can quickly select.
		talents: [Presets.StandardTalents, Presets.MonocatTalents],
		rotations: [Presets.SIMPLE, Presets.APL],
		// Preset gear configurations that the user can quickly select.
		gear: [Presets.PRE_RAID_GEARSET, Presets.P1_GEARSET, Presets.P2_GEARSET, Presets.P3_GEARSET, Presets.P4_GEARSET, Presets.P5_GEARSET],
		itemSwaps: [],
		builds: [],
	},

	autoRotation: (_player: Player<Spec.SpecFeralCatDruid>): APLRotation => {
		return APLRotation.create();
	},

	simpleRotation: (_player: Player<Spec.SpecFeralCatDruid>, simple: DruidRotation, _cooldowns: Cooldowns): APLRotation => {
		// All cooldowns (potions, sappers, runes) are fired by the Go rotation
		// during power shifts (ClearForm → fire MCDs → CatForm), not through the APL.
		const doRotation = APLAction.fromJsonString(
			`{"catOptimalRotationAction":{"finishingMove":${simple.finishingMove},"biteweave":${simple.biteweave},"ripweave":${simple.ripweave},"ripMinComboPoints":${simple.ripMinComboPoints},"biteMinComboPoints":${simple.biteMinComboPoints},"mangleTrick":${simple.mangleTrick},"rakeTrick":${simple.rakeTrick},"maintainFaerieFire":${simple.maintainFaerieFire}}}`,
		);

		return APLRotation.create({
			priorityList: [APLListItem.create({ action: doRotation })],
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
		this.reforger = new ReforgeOptimizer(this);
	}
}
