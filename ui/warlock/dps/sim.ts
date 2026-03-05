import * as BuffDebuffInputs from '../../core/components/inputs/buffs_debuffs';
import * as OtherInputs from '../../core/components/inputs/other_inputs';
import { ReforgeOptimizer } from '../../core/components/suggest_reforges_action';
import { IndividualSimUI, registerSpecConfig } from '../../core/individual_sim_ui';
import { Player } from '../../core/player';
import { PlayerClasses } from '../../core/player_classes';
import { APLRotation } from '../../core/proto/apl';
import { Debuffs, Faction, IndividualBuffs, ItemSlot, PartyBuffs, PseudoStat, Race, RaidBuffs, Spec, Stat, TristateEffect } from '../../core/proto/common';
import { DEFAULT_CASTER_GEM_STATS, Stats, UnitStat } from '../../core/proto_utils/stats';
import { defaultRaidBuffMajorDamageCooldowns } from '../../core/proto_utils/utils';
import * as WarlockInputs from './inputs';
import * as Presets from './presets';

const SPEC_CONFIG = registerSpecConfig(Spec.SpecWarlock, {
	cssClass: 'warlock-sim-ui',
	cssScheme: PlayerClasses.getCssClass(PlayerClasses.Warlock),
	// List any known bugs / issues here and they'll be shown on the site.
	knownIssues: [],

	// All stats for which EP should be calculated.
	epStats: [
		Stat.StatStamina,
		Stat.StatIntellect,
		Stat.StatSpellDamage,
		Stat.StatShadowDamage,
		Stat.StatFireDamage,
		Stat.StatSpellHitRating,
		Stat.StatSpellCritRating,
		Stat.StatSpellHasteRating,
		Stat.StatMP5,
	],
	// Reference stat against which to calculate EP. DPS classes use either spell power or attack power.
	epReferenceStat: Stat.StatSpellDamage,
	// Which stats to display in the Character Stats section, at the bottom of the left-hand sidebar.
	displayStats: UnitStat.createDisplayStatArray(
		[Stat.StatHealth, Stat.StatMana, Stat.StatStamina, Stat.StatIntellect, Stat.StatSpellDamage, Stat.StatShadowDamage, Stat.StatFireDamage, Stat.StatMP5],
		[PseudoStat.PseudoStatSpellHitPercent, PseudoStat.PseudoStatSpellCritPercent, PseudoStat.PseudoStatSpellHastePercent],
	),
	gemStats: DEFAULT_CASTER_GEM_STATS,

	defaults: {
		// Default equipped gear.
		gear: Presets.T4.gear,

		// Default EP weights for sorting gear in the gear picker.
		epWeights: Presets.P1_AFFLI_DEMO_DESTRO_EP.epWeights,
		statCaps: (() => {
			return new Stats().withPseudoStat(PseudoStat.PseudoStatSpellHitPercent, 16);
		})(),

		// Default consumes settings.
		consumables: Presets.DefaultConsumables,

		// Default talents.
		talents: Presets.TalentsDestruction.data,
		// Default spec-specific settings.
		specOptions: Presets.DefaultOptions,

		// Default buffs and debuffs settings.
		raidBuffs: RaidBuffs.create({
			...defaultRaidBuffMajorDamageCooldowns(),
			arcaneBrilliance: true,
			giftOfTheWild: TristateEffect.TristateEffectImproved,
			powerWordFortitude: TristateEffect.TristateEffectImproved,
			divineSpirit: TristateEffect.TristateEffectImproved,
		}),
		partyBuffs: PartyBuffs.create({
			manaSpringTotem: TristateEffect.TristateEffectImproved,
			moonkinAura: TristateEffect.TristateEffectImproved,
			totemOfWrath: 1,
			wrathOfAirTotem: TristateEffect.TristateEffectImproved,
		}),
		individualBuffs: IndividualBuffs.create({
			blessingOfKings: true,
			blessingOfWisdom: TristateEffect.TristateEffectImproved,
			shadowPriestDps: 800,
		}),
		debuffs: Debuffs.create({
			judgementOfWisdom: true,
			shadowWeaving: true,
			misery: true,
			curseOfElements: TristateEffect.TristateEffectRegular,
			sunderArmor: true,
			screech: true,
			faerieFire: TristateEffect.TristateEffectImproved,
			curseOfRecklessness: true,
		}),

		other: Presets.OtherDefaults,
	},

	consumableStats: [
		Stat.StatIntellect,
		Stat.StatSpirit,
		Stat.StatMP5,
		Stat.StatSpellDamage,
		Stat.StatSpellCritRating,
		Stat.StatSpellHitRating,
		Stat.StatSpellHasteRating,
		Stat.StatShadowDamage,
		Stat.StatFireDamage,
	],
	// IconInputs to include in the 'Player' section on the settings tab.
	playerIconInputs: [WarlockInputs.PetInput(), WarlockInputs.ArmorInput(), WarlockInputs.DemonicSacrificeInput()],

	// Buff and Debuff inputs to include/exclude, overriding the EP-based defaults.
	includeBuffDebuffInputs: [BuffDebuffInputs.DivineSpirit, BuffDebuffInputs.SanctityAura, BuffDebuffInputs.ManaSpringTotem, BuffDebuffInputs.ManaTideTotem],
	excludeBuffDebuffInputs: [],
	petConsumeInputs: [],
	// Inputs to include in the 'Other' section on the settings tab.
	otherInputs: {
		inputs: [OtherInputs.IsbUptime],
	},
	itemSwapSlots: [ItemSlot.ItemSlotMainHand, ItemSlot.ItemSlotOffHand, ItemSlot.ItemSlotTrinket1, ItemSlot.ItemSlotTrinket2],
	encounterPicker: {
		// Whether to include 'Execute Duration (%)' in the 'Encounter' section of the settings tab.
		showExecuteProportion: false,
	},

	presets: {
		epWeights: [Presets.P1_AFFLI_DEMO_DESTRO_EP, Presets.P1_DESTRUCTION_FIRE_EP],
		// Preset talents that the user can quickly select.
		talents: [Presets.TalentsAffliction, Presets.TalentsDemoFelguard, Presets.TalentsDemoRuin, Presets.TalentsDestroNightfall, Presets.TalentsDestruction],
		// Preset rotations that the user can quickly select.
		rotations: [Presets.AfflictionAPL, Presets.DemoAPL, Presets.DestroAPL, Presets.DestroFireAPL],

		// Preset gear configurations that the user can quickly select.
		gear: [Presets.PRE_RAID, Presets.PRE_RAID_FIRE, Presets.T4, Presets.T4_FIRE, Presets.T5, Presets.T6, Presets.ZA, Presets.SWP],
		itemSwaps: [],
		builds: [Presets.AFFLICTION_BUILD, Presets.DEMONOLOGY_BUILD, Presets.DESTRUCTION_BUILD, Presets.DESTRUCTION_FIRE_BUILD],
	},

	autoRotation: (_player: Player<Spec.SpecWarlock>): APLRotation => {
		return Presets.DestroAPL.rotation.rotation!;
	},
	customSections: [WarlockInputs.CursesSection],

	raidSimPresets: [
		{
			spec: Spec.SpecWarlock,
			talents: Presets.TalentsDestruction.data,
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
					1: Presets.PRE_RAID.gear,
				},
				[Faction.Horde]: {
					1: Presets.PRE_RAID.gear,
				},
			},
			otherDefaults: Presets.OtherDefaults,
		},
	],
});

export class WarlockSimUI extends IndividualSimUI<Spec.SpecWarlock> {
	constructor(parentElem: HTMLElement, player: Player<Spec.SpecWarlock>) {
		super(parentElem, player, SPEC_CONFIG);

		this.reforger = new ReforgeOptimizer(this);
	}
}
