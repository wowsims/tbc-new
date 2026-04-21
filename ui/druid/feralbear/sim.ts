import * as OtherInputs from '../../core/components/inputs/other_inputs.js';
import * as Mechanics from '../../core/constants/mechanics.js';
import { IndividualSimUI, registerSpecConfig } from '../../core/individual_sim_ui.js';
import { Player } from '../../core/player.js';
import { PlayerClasses } from '../../core/player_classes';
import { APLAction, APLListItem, APLRotation, APLRotation_Type as APLRotationType } from '../../core/proto/apl.js';
import { Cooldowns, Debuffs, Drums, Faction, IndividualBuffs, ItemSlot, PartyBuffs, PseudoStat, Race, RaidBuffs, Spec, Stat, TristateEffect } from '../../core/proto/common.js';
import { defaultExposeWeaknessSettings } from '../../core/proto_utils/utils.js';
import { FeralBearDruid_Rotation as DruidRotation } from '../../core/proto/druid.js';
import { ReforgeOptimizer } from '../../core/components/suggest_reforges_action';
import { Stats, UnitStat } from '../../core/proto_utils/stats.js';
import * as FeralBearInputs from './inputs.js';
import * as Presets from './presets.js';

const SPEC_CONFIG = registerSpecConfig(Spec.SpecFeralBearDruid, {
	cssClass: 'feral-bear-druid-sim-ui',
	cssScheme: PlayerClasses.getCssClass(PlayerClasses.Druid),
    // List any known bugs / issues here and they'll be shown on the site.
	knownIssues: [],
	warnings: [],

	epRatios: [0, 0, 0.6, 0, 1.0, 0],
	// All stats for which EP should be calculated.
	epStats: [
		Stat.StatStamina,
		Stat.StatAgility,
		Stat.StatStrength,
		Stat.StatAttackPower,
		Stat.StatArmor,
		Stat.StatBonusArmor,
		Stat.StatDodgeRating,
		Stat.StatDefenseRating,
		Stat.StatMeleeHitRating,
		Stat.StatMeleeCritRating,
        Stat.StatMeleeHasteRating,
		Stat.StatExpertiseRating,
        Stat.StatResilienceRating,
		Stat.StatPhysicalDamage,
        Stat.StatArmorPenetration,
	],
	epPseudoStats: [PseudoStat.PseudoStatMainHandDps],
	epReferenceStat: Stat.StatAgility,
	tankRefStat: Stat.StatStamina,
	displayStats: UnitStat.createDisplayStatArray(
		[
			Stat.StatHealth,
			Stat.StatStamina,
			Stat.StatAgility,
			Stat.StatStrength,
			Stat.StatAttackPower,
			Stat.StatArmor,
			Stat.StatBonusArmor,
			Stat.StatDodgeRating,
			Stat.StatDefenseRating,
			Stat.StatExpertiseRating,
            Stat.StatResilienceRating,
			Stat.StatNatureResistance,
			Stat.StatFireResistance,
			Stat.StatFrostResistance,
			Stat.StatArcaneResistance,
			Stat.StatShadowResistance,
		],
		[
			PseudoStat.PseudoStatMeleeHitPercent,
			PseudoStat.PseudoStatMeleeCritPercent,
			PseudoStat.PseudoStatMeleeHastePercent,
			PseudoStat.PseudoStatDodgePercent,
		],
	),

	defaults: {
		gear: Presets.PRERAID_PRESET.gear,
		epWeights: Presets.BALANCED_EP_PRESET.epWeights,
		statCaps: (() => {
			const hitCap = new Stats().withPseudoStat(PseudoStat.PseudoStatMeleeHitPercent, 9);
			const expCap = new Stats().withStat(Stat.StatExpertiseRating, 6.5 * 4 * Mechanics.EXPERTISE_PER_QUARTER_PERCENT_REDUCTION);
			const critImmunityCap = new Stats().withPseudoStat(PseudoStat.PseudoStatReducedCritTakenPercent, 5.6);
			return hitCap.add(expCap).add(critImmunityCap);
		})(),
		other: Presets.OtherDefaults,
		consumables: Presets.DefaultConsumables,
		rotationType: APLRotationType.TypeAPL,
		aplRotation: Presets.ROTATION_DEFAULT.rotation.rotation!,
		talents: Presets.StandardTalents.data,
		specOptions: Presets.DefaultOptions,
		// Default encounter
		encounter: "Magtheridon's Lair/Magtheridon 25",
		raidBuffs: RaidBuffs.create({
			arcaneBrilliance: true,
			giftOfTheWild: TristateEffect.TristateEffectImproved,
			powerWordFortitude: TristateEffect.TristateEffectImproved,
            bloodlust: true,
			shadowProtection: true,
			thorns: TristateEffect.TristateEffectRegular,
			divineSpirit: TristateEffect.TristateEffectImproved,
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
			blessingOfSanctuary: true,
			unleashedRage: true,
		}),
		debuffs: Debuffs.create({
            ...defaultExposeWeaknessSettings(),
			bloodFrenzy: true,
			exposeArmor: TristateEffect.TristateEffectImproved,
			faerieFire: TristateEffect.TristateEffectImproved,
			giftOfArthas: false,
			huntersMark: TristateEffect.TristateEffectImproved,
			improvedSealOfTheCrusader: TristateEffect.TristateEffectImproved,
            curseOfRecklessness: true,
			insectSwarm: true,
			judgementOfWisdom: true,
			misery: true,
			screech: true,
			shadowEmbrace: true,
			sunderArmor: true,
		}),
	},

	playerIconInputs: [],
	rotationInputs: FeralBearInputs.FeralBearRotationConfig,
	includeBuffDebuffInputs: [Stat.StatStamina, Stat.StatArmor],
	excludeBuffDebuffInputs: [Stat.StatParryRating],
	otherInputs: {
		inputs: [
			OtherInputs.TotemTwisting,
			FeralBearInputs.StartingRage,
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
	itemSwapSlots: [ItemSlot.ItemSlotTrinket1, ItemSlot.ItemSlotTrinket2, ItemSlot.ItemSlotMainHand, ItemSlot.ItemSlotRanged],
	defaultBuild: Presets.MAGTHERIDON_PRESET_BUILD,

	encounterPicker: {
		showExecuteProportion: false,
	},

	presets: {
		epWeights: [Presets.SURVIVAL_EP_PRESET, Presets.BALANCED_EP_PRESET],
		talents: [Presets.StandardTalents, Presets.DemoRoarTalents],
		// ROTATION_SIMPLE is kept in presets.ts for reference but omitted here —
		// the APL rotation is more user-friendly and handles CDs, re-shifting, and
		// on-use items more easily.
		rotations: [Presets.ROTATION_DEFAULT],
		gear: [Presets.PRERAID_PRESET, Presets.P1_PRESET, Presets.P2_HYDROSS_FROST_PRESET, Presets.P2_HYDROSS_NATURE_PRESET],
		builds: [
				Presets.DEFAULT_PRESET_BUILD,
				Presets.KARAZHAN_PRESET_BUILD,
				Presets.MAGTHERIDON_PRESET_BUILD,
				Presets.MOROGRIM_PRESET_BUILD,
				Presets.HYDROSS_PRESET_BUILD,
			],
	},

	autoRotation: (_player: Player<Spec.SpecFeralBearDruid>): APLRotation => {
		return Presets.ROTATION_DEFAULT.rotation.rotation!;
	},

	simpleRotation: (_player: Player<Spec.SpecFeralBearDruid>, simple: DruidRotation, _cooldowns: Cooldowns): APLRotation => {
		const doRotation = APLAction.fromJsonString(
			`{"bearOptimalRotationAction":{"maintainFaerieFire":${simple.maintainFaerieFire},"maintainDemoralizingRoar":${simple.maintainDemoralizingRoar},"maulRageThreshold":${simple.maulRageThreshold},"swipeUsage":${simple.swipeUsage},"swipeApThreshold":${simple.swipeApThreshold}}}`,
		);
		return APLRotation.create({
			priorityList: [APLListItem.create({ action: doRotation })],
		});
	},

	raidSimPresets: [
		{
			spec: Spec.SpecFeralBearDruid,
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

export class FeralBearDruidSimUI extends IndividualSimUI<Spec.SpecFeralBearDruid> {
	constructor(parentElem: HTMLElement, player: Player<Spec.SpecFeralBearDruid>) {
		super(parentElem, player, SPEC_CONFIG);
		this.reforger = new ReforgeOptimizer(this);
	}
}
