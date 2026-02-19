import * as OtherInputs from '../../core/components/inputs/other_inputs';
import { IndividualSimUI, registerSpecConfig } from '../../core/individual_sim_ui';
import { Player } from '../../core/player';
import { PlayerClasses } from '../../core/player_classes';
import { APLRotation } from '../../core/proto/apl';
import { Debuffs, Faction, IndividualBuffs, ItemSlot, PartyBuffs, PseudoStat, Race, RaidBuffs, Spec, Stat, TristateEffect } from '../../core/proto/common';
import { UnitStat } from '../../core/proto_utils/stats';
import { defaultRaidBuffMajorDamageCooldowns } from '../../core/proto_utils/utils';

import * as Presets from './presets';

const SPEC_CONFIG = registerSpecConfig(Spec.SpecRogue, {
	cssClass: 'rogue-sim-ui',
	cssScheme: PlayerClasses.getCssClass(PlayerClasses.Rogue),
	// List any known bugs / issues here and they'll be shown on the site.
	knownIssues: [],

	// All stats for which EP should be calculated.
	epStats: [
		Stat.StatStamina,
		Stat.StatAgility,
		Stat.StatStrength,
		Stat.StatMeleeCritRating,
		Stat.StatMeleeHasteRating,
		Stat.StatMeleeHitRating,
		Stat.StatArmorPenetration,
		Stat.StatExpertiseRating,
		Stat.StatAttackPower,
		Stat.StatPhysicalDamage,
	],
	epPseudoStats: [PseudoStat.PseudoStatMainHandDps, PseudoStat.PseudoStatOffHandDps],
	// Reference stat against which to calculate EP.
	epReferenceStat: Stat.StatAttackPower,
	// Which stats to display in the Character Stats section, at the bottom of the left-hand sidebar.
	displayStats: UnitStat.createDisplayStatArray(
		[Stat.StatHealth, Stat.StatStamina, Stat.StatAgility, Stat.StatStrength, Stat.StatAttackPower, Stat.StatArmorPenetration, Stat.StatExpertiseRating],
		[PseudoStat.PseudoStatMeleeHitPercent, PseudoStat.PseudoStatMeleeCritPercent, PseudoStat.PseudoStatMeleeHastePercent],
	),

	defaults: {
		// Default equipped gear.
		gear: Presets.P1_SWORDS_GEAR.gear,
		// Default EP weights for sorting gear in the gear picker.
		epWeights: Presets.P1_EP_PRESET.epWeights,
		other: Presets.OtherDefaults,
		// Default consumes settings.
		consumables: Presets.DefaultConsumables,
		// Default talents.
		talents: Presets.Talents.data,
		// Default spec-specific settings.
		specOptions: Presets.DefaultOptions,
		// Default raid/party buffs settings.
		raidBuffs: RaidBuffs.create({
			...defaultRaidBuffMajorDamageCooldowns(),
			giftOfTheWild: TristateEffect.TristateEffectImproved
		}),
		partyBuffs: PartyBuffs.create({
			battleShout: TristateEffect.TristateEffectImproved,
			ferociousInspiration: 1,
			strengthOfEarthTotem: TristateEffect.TristateEffectImproved,
			graceOfAirTotem: TristateEffect.TristateEffectImproved,
			windfuryTotem: TristateEffect.TristateEffectImproved,
			leaderOfThePack: TristateEffect.TristateEffectRegular
		}),
		individualBuffs: IndividualBuffs.create({
			blessingOfKings: true,
			blessingOfMight: TristateEffect.TristateEffectImproved,
			unleashedRage: true,
		}),
		debuffs: Debuffs.create({
			bloodFrenzy: true,
			huntersMark: TristateEffect.TristateEffectImproved,
			improvedSealOfTheCrusader: true,
			mangle: true,
			misery: true,
			curseOfRecklessness: true,
			faerieFire: TristateEffect.TristateEffectImproved,
			exposeWeaknessUptime: 0.9,
			exposeWeaknessHunterAgility: 1080,
			giftOfArthas: true,
			sunderArmor: true,
		}),
	},

	playerInputs: {
		inputs: [],
	},
	// IconInputs to include in the 'Player' section on the settings tab.
	playerIconInputs: [],
	// Buff and Debuff inputs to include/exclude, overriding the EP-based defaults.
	includeBuffDebuffInputs: [],
	excludeBuffDebuffInputs: [],
	// Inputs to include in the 'Other' section on the settings tab.
	otherInputs: {
		inputs: [OtherInputs.InFrontOfTarget, OtherInputs.InputDelay],
	},
	itemSwapSlots: [ItemSlot.ItemSlotTrinket1, ItemSlot.ItemSlotTrinket2, ItemSlot.ItemSlotMainHand, ItemSlot.ItemSlotOffHand],
	encounterPicker: {
		// Whether to include 'Execute Duration (%)' in the 'Encounter' section of the settings tab.
		showExecuteProportion: false,
	},

	presets: {
		epWeights: [Presets.P1_EP_PRESET],
		// Preset talents that the user can quickly select.
		talents: [Presets.Talents],
		// Preset rotations that the user can quickly select.
		rotations: [Presets.SINSITER_APL, Presets.SHIV_APL],
		// Preset gear configurations that the user can quickly select.
		gear: [Presets.PREARAID_SWORDS_GEAR, Presets.P1_SWORDS_GEAR],
	},

	autoRotation: (player: Player<Spec.SpecRogue>): APLRotation => {
		return Presets.SINSITER_APL.rotation.rotation!;
	},

	raidSimPresets: [
		{
			spec: Spec.SpecRogue,
			talents: Presets.Talents.data,
			specOptions: Presets.DefaultOptions,
			consumables: Presets.DefaultConsumables,
			defaultFactionRaces: {
				[Faction.Unknown]: Race.RaceUnknown,
				[Faction.Alliance]: Race.RaceHuman,
				[Faction.Horde]: Race.RaceOrc,
			},
			defaultGear: {
				[Faction.Unknown]: {},
				[Faction.Alliance]: {
					1: Presets.P1_SWORDS_GEAR.gear,
				},
				[Faction.Horde]: {
					1: Presets.P1_SWORDS_GEAR.gear,
				},
			},
			otherDefaults: Presets.OtherDefaults,
		},
	],
});

export class RogueSimUI extends IndividualSimUI<Spec.SpecRogue> {
	constructor(parentElem: HTMLElement, player: Player<Spec.SpecRogue>) {
		super(parentElem, player, SPEC_CONFIG);

		this.player.changeEmitter.on(c => {
			const options = this.player.getSpecOptions();
			this.player.setSpecOptions(c, options);
		});
		this.sim.encounter.changeEmitter.on(c => {
			const options = this.player.getSpecOptions();
			this.player.setSpecOptions(c, options);
		});
	}
}
