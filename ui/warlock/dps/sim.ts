import * as BuffDebuffInputs from '../../core/components/inputs/buffs_debuffs';
import * as OtherInputs from '../../core/components/inputs/other_inputs';
import { IndividualSimUI, registerSpecConfig } from '../../core/individual_sim_ui';
import { Player } from '../../core/player';
import { PlayerClasses } from '../../core/player_classes';
import { APLRotation } from '../../core/proto/apl';
import { Debuffs, Drums, Faction, IndividualBuffs, ItemSlot, PartyBuffs, PseudoStat, Race, RaidBuffs, Spec, Stat, TristateEffect } from '../../core/proto/common';
import { DEFAULT_CASTER_GEM_STATS, Stats, UnitStat } from '../../core/proto_utils/stats';
import { defaultRaidBuffMajorDamageCooldowns } from '../../core/proto_utils/utils';
import { TypedEvent } from '../../core/typed_event';
import * as WarlockInputs from './inputs';
import * as Presets from './presets';

const modifyDisplayStats = (player: Player<Spec.SpecWarlock>) => {
	let stats = new Stats();

	TypedEvent.freezeAllAndDo(() => {
		const currentStats = player.getCurrentStats().finalStats?.stats;
		if (currentStats === undefined) {
			return {};
		}

		// stats = stats.addStat(Stat.StatMP5, (currentStats[Stat.StatMP5] * currentStats[Stat.StatSpellHasteRating]) / HASTE_RATING_PER_HASTE_PERCENT / 100);
	});

	return {
		talents: stats,
	};
};

const SPEC_CONFIG = registerSpecConfig(Spec.SpecWarlock, {
	cssClass: 'warlock-sim-ui',
	cssScheme: PlayerClasses.getCssClass(PlayerClasses.Warlock),
	// List any known bugs / issues here and they'll be shown on the site.
	knownIssues: [],

	// All stats for which EP should be calculated.
	epStats: [
		Stat.StatIntellect,
		Stat.StatSpellDamage,
		Stat.StatSpellCritRating,
		Stat.StatSpellHasteRating
	],
	// Reference stat against which to calculate EP. DPS classes use either spell power or attack power.
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
			Stat.StatShadowDamage,
			Stat.StatFireDamage,
			Stat.StatMP5,
		],
		[
			PseudoStat.PseudoStatSpellHitPercent,
			PseudoStat.PseudoStatSpellCritPercent,
			PseudoStat.PseudoStatSpellHastePercent
		],
	),
	gemStats: DEFAULT_CASTER_GEM_STATS,

	modifyDisplayStats,
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

		// Default buffs and debuffs settings.
		raidBuffs: RaidBuffs.create({
			...defaultRaidBuffMajorDamageCooldowns(),
			arcaneBrilliance: true,
			powerWordFortitude: TristateEffect.TristateEffectImproved,
			divineSpirit: TristateEffect.TristateEffectImproved,
			giftOfTheWild: TristateEffect.TristateEffectImproved,
		}),
		partyBuffs: PartyBuffs.create({
			bloodPact: TristateEffect.TristateEffectMissing,
			moonkinAura: TristateEffect.TristateEffectRegular,
			totemOfWrath: 1,
			wrathOfAirTotem: TristateEffect.TristateEffectRegular,
			manaSpringTotem: TristateEffect.TristateEffectRegular,
			draeneiRacialCaster: false,
			ferociousInspiration: 0,
			sanctityAura: TristateEffect.TristateEffectMissing,
			drums: Drums.DrumsOfBattle
		}),
		individualBuffs: IndividualBuffs.create({
			blessingOfKings: true,
			blessingOfSalvation: false,
			blessingOfWisdom: TristateEffect.TristateEffectRegular,
			innervates: 0,
			powerInfusions: 0,
			shadowPriestDps: 0,
		}),
		debuffs: Debuffs.create({
			bloodFrenzy: true,
			curseOfElements: TristateEffect.TristateEffectRegular,
			curseOfRecklessness: true,
			faerieFire: TristateEffect.TristateEffectRegular,
			huntersMark: TristateEffect.TristateEffectImproved,
			exposeArmor: TristateEffect.TristateEffectRegular,
			improvedScorch: false,
			improvedSealOfTheCrusader: true,
			hemorrhageUptime: 0,
			isbUptime: 0,
			misery: true,
			shadowWeaving: true,
			sunderArmor: true,
			wintersChill: true,
		}),

		other: Presets.OtherDefaults,
	},

	// IconInputs to include in the 'Player' section on the settings tab.
	playerIconInputs: [
		WarlockInputs.PetInput(),
		WarlockInputs.ArmorInput(),
		WarlockInputs.DemonicSacrificeInput()
	],

	// Buff and Debuff inputs to include/exclude, overriding the EP-based defaults.
	includeBuffDebuffInputs: [],
	excludeBuffDebuffInputs: [],
	petConsumeInputs: [],
	// Inputs to include in the 'Other' section on the settings tab.
	otherInputs: {
		inputs: [
			OtherInputs.IsbUptime,
			OtherInputs.HemoUptime,
			OtherInputs.DistanceFromTarget,
			OtherInputs.TankAssignment,
		],
	},
	itemSwapSlots: [ItemSlot.ItemSlotMainHand, ItemSlot.ItemSlotOffHand, ItemSlot.ItemSlotTrinket1, ItemSlot.ItemSlotTrinket2],
	encounterPicker: {
		// Whether to include 'Execute Duration (%)' in the 'Encounter' section of the settings tab.
		showExecuteProportion: false,
	},

	presets: {
		epWeights: [],
		// Preset talents that the user can quickly select.
		talents: [],
		// Preset rotations that the user can quickly select.
		rotations: [],

		// Preset gear configurations that the user can quickly select.
		gear: [],
		itemSwaps: [],
	},

	autoRotation: (_player: Player<Spec.SpecWarlock>): APLRotation => {
		return Presets.BLANK_APL.rotation.rotation!;
	},

	raidSimPresets: [
		{
			spec: Spec.SpecWarlock,
			talents: Presets.Talents.data,
			specOptions: Presets.DefaultOptions,
			consumables: Presets.DefaultConsumables,
			defaultFactionRaces: {
				[Faction.Unknown]: Race.RaceUnknown,
				[Faction.Alliance]: Race.RaceHuman,
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
			otherDefaults: Presets.OtherDefaults,
		},
	],
});

export class WarlockSimUI extends IndividualSimUI<Spec.SpecWarlock> {
	constructor(parentElem: HTMLElement, player: Player<Spec.SpecWarlock>) {
		super(parentElem, player, SPEC_CONFIG);
	}
}
