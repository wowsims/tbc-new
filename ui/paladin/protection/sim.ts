import * as OtherInputs from '../../core/components/inputs/other_inputs';
import { ReforgeOptimizer } from '../../core/components/suggest_reforges_action';
import { IndividualSimUI, registerSpecConfig } from '../../core/individual_sim_ui';
import { Player } from '../../core/player';
import { PlayerClasses } from '../../core/player_classes';
import { APLListItem, APLRotation, APLRotation_Type, APLValueVariable } from '../../core/proto/apl';
import { Cooldowns, PseudoStat, Spec, Stat } from '../../core/proto/common';
import { PaladinAura, PaladinJudgement } from '../../core/proto/paladin';
import { StatCapType } from '../../core/proto/ui';
import * as AplUtils from '../../core/proto_utils/apl_utils';
import { StatCap, UnitStat } from '../../core/proto_utils/stats';
import { SpecRotation } from '../../core/proto_utils/utils';
import * as Presets from './presets';
import * as ProtPaladinInputs from './inputs';

// Spell IDs for each rank of Consecration.
const CONSECRATION_RANK_SPELL_IDS: Record<number, number> = {
	1: 26573,
	2: 20116,
	3: 20922,
	4: 20923,
	5: 20924,
	6: 27173,
};

// Fixed indices into the default APL (apls/default.apl.json). simpleRotation
// relies on these — if you reorder the APL, update these too.
const PREPULL_AURA_INDEX = 1; // Devotion Aura at -18.5s
const PREPULL_SEAL_INDEX = 2; // Seal of Righteousness at -3s (shifted to -4s when precastAvengersShield is on)
const PREPULL_HOLY_SHIELD_INDEX = 3; // Holy Shield at -1.5s (shifted to -2.5s when precastAvengersShield is on)
const PREPULL_AVENGERS_SHIELD_INDEX = 4; // Avenger's Shield at -0.99s (hidden by default)
const PRIORITY_JUDGE_ON_SEAL_INDEX = 1; // First-global judge, and maintenance-seal judge once SoR is consumed
const PRIORITY_SWAP_SEAL_INDEX = 4; // When maintenance seal is down, JoX is down, and Judgement is ready, swap to maintenance seal
const PRIORITY_RIGHTEOUSNESS_JUDGE_INDEX = 7; // Judge -> Re-seal Righteousness
const PRIORITY_CONSECRATION_INDEX = 6; // Consecration rank 6

// SpellIDs for each paladin aura option.
const AURA_SPELL_IDS: Record<PaladinAura, number | null> = {
	[PaladinAura.AuraNone]: null,
	[PaladinAura.DevotionAura]: 27149,
	[PaladinAura.RetributionAura]: 27150,
	[PaladinAura.ConcentrationAura]: 19746,
	[PaladinAura.FireResistanceAura]: 27153,
	[PaladinAura.FrostResistanceAura]: 27152,
	[PaladinAura.ShadowResistanceAura]: 27151,
	[PaladinAura.SanctityAura]: 20218,
};

// Tags for each paladin aura option.
const AURA_TAGS: Record<PaladinAura, number | null> = {
	[PaladinAura.AuraNone]: 0,
	[PaladinAura.DevotionAura]: 0,
	[PaladinAura.RetributionAura]: 0,
	[PaladinAura.ConcentrationAura]: 0,
	[PaladinAura.FireResistanceAura]: 1,
	[PaladinAura.FrostResistanceAura]: 1,
	[PaladinAura.ShadowResistanceAura]: 1,
	[PaladinAura.SanctityAura]: 0,
};

type JudgementSpec = { sealSpellId: number; sealRank: number; judgementAuraSpellId: number; judgementAuraRank: number };
const JUDGEMENT_CONFIG: Record<PaladinJudgement, JudgementSpec | null> = {
	[PaladinJudgement.JudgementNone]: null,
	[PaladinJudgement.JudgementOfLight]: { sealSpellId: 27160, sealRank: 5, judgementAuraSpellId: 27162, judgementAuraRank: 5 },
	[PaladinJudgement.JudgementOfWisdom]: { sealSpellId: 27166, sealRank: 4, judgementAuraSpellId: 27164, judgementAuraRank: 4 },
};

const SPEC_CONFIG = registerSpecConfig(Spec.SpecProtectionPaladin, {
	cssClass: 'protection-paladin-sim-ui',
	cssScheme: PlayerClasses.getCssClass(PlayerClasses.Paladin),
	// List any known bugs / issues here and they'll be shown on the site.
	knownIssues: [],
	consumableStats: [Stat.StatStamina, Stat.StatHealth, Stat.StatMana, Stat.StatSpellDamage, Stat.StatHolyDamage],

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
			Stat.StatHolyDamage,
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
		gear: Presets.P2_GEAR_PRESET.gear,
		softCapBreakpoints: [
			StatCap.fromPseudoStat(PseudoStat.PseudoStatReducedCritTakenPercent, {
				breakpoints: [5.6],
				capType: StatCapType.TypeSoftCap,
				postCapEPs: [0],
			}),
		],
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
			OtherInputs.TotemTwisting,
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

	defaultBuild: Presets.MAGTHERIDON_PRESET_BUILD,

	presets: {
		epWeights: [Presets.P4_EP_PRESET],
		// Preset talents that the user can quickly select.
		talents: [Presets.DefaultTalents],
		// Preset rotations that the user can quickly select.
		rotations: [Presets.APL_SIMPLE, Presets.APL_PRESET],
		// Preset gear configurations that the user can quickly select.
		gear: [Presets.P1_GEAR_PRESET, Presets.P2_GEAR_PRESET, Presets.P3_GEAR_PRESET, Presets.P4_GEAR_PRESET, Presets.P5_GEAR_PRESET],
		builds: [
			Presets.DEFAULT_PRESET_BUILD,
			Presets.KARAZHAN_PRESET_BUILD,
			Presets.MAGTHERIDON_PRESET_BUILD,
			Presets.MOROGRIM_PRESET_BUILD,
			Presets.HYDROSS_PRESET_BUILD,
		],
	},

	autoRotation: (_player: Player<Spec.SpecProtectionPaladin>): APLRotation => {
		return Presets.APL_PRESET.rotation.rotation!;
	},

	simpleRotation: (player: Player<Spec.SpecProtectionPaladin>, simple: SpecRotation<Spec.SpecProtectionPaladin>, cooldowns: Cooldowns): APLRotation => {
		const actions = AplUtils.simpleCooldownActions(cooldowns);
		const rotation = APLRotation.clone(Presets.APL_PRESET.rotation.rotation!);

		let {
			prioritizeHolyShield = true,
			consecrationRank = 6,
			useExorcism = true,
			useAvengersShield = false,
			useHammerOfWrath = false,
			precastAvengersShield = true,
			maintainJudgement = PaladinJudgement.JudgementNone,
			aura: rawAura = PaladinAura.DevotionAura,
		} = simple;

		if (!player.getTalents().avengersShield) {
			useAvengersShield = false;
			precastAvengersShield = false;
		}

		if (!player.getTalents().holyShield) {
			prioritizeHolyShield = false;
		}

		// Sanctity Aura requires the talent. If the user picked it without the
		// talent (e.g. dropped the point after selecting), fall back to None.
		const aura = rawAura === PaladinAura.SanctityAura && !player.getTalents().sanctityAura ? PaladinAura.AuraNone : rawAura;

		const judgementConfig = JUDGEMENT_CONFIG[maintainJudgement];

		rotation.valueVariables = [
			APLValueVariable.fromJson({ name: 'Prioritize Holy Shield', value: { const: { val: String(prioritizeHolyShield) } } }),
			APLValueVariable.fromJson({ name: 'Use Exorcism', value: { const: { val: String(useExorcism) } } }),
			APLValueVariable.fromJson({ name: "Use Avenger's Shield", value: { const: { val: String(useAvengersShield) } } }),
			APLValueVariable.fromJson({ name: 'Use Hammer of Wrath', value: { const: { val: String(useHammerOfWrath) } } }),
			APLValueVariable.fromJson({ name: 'Maintain Judgement', value: { const: { val: String(!!judgementConfig) } } }),
		];

		// Avenger's Shield prepull is disabled in the default APL; flip it on
		// when the user enabled precast, and slide the seal and Holy Shield
		// casts earlier so AS can land at the pull without clipping them.
		if (precastAvengersShield) {
			rotation.prepullActions[PREPULL_AVENGERS_SHIELD_INDEX].hide = false;
			(rotation.prepullActions[PREPULL_SEAL_INDEX].doAtValue!.value as any).const.val = '-4s';
			(rotation.prepullActions[PREPULL_HOLY_SHIELD_INDEX].doAtValue!.value as any).const.val = '-2.5s';
		}

		// For Light/Wisdom we activate the two maintenance actions that are
		// dormant in the default APL (their "Maintain Judgement" variableRef
		// evaluates false, short-circuiting the AND):
		//   - Judge-on-seal: also doubles as the first-global judge via an OR
		//     with currentTime <= 0.5s, so we only need to swap the
		//     maintenance-seal aura inside the AND branch.
		//   - Swap-seal: fires once the maintenance seal is down, JoX is
		//     missing, and Judgement is ready, preparing the seal for the
		//     next refresh of the debuff.
		// Both also need the Seal of Wisdom / Judgement of Wisdom references
		// rewritten when the user picked Judgement of Light.
		//
		// The prepull seal stays SoR — starting combat in SoR gives a stronger
		// opening threat block (Judge SoR via the first-global branch) and
		// then the maintenance cycle takes over for the rest of the fight.
		if (judgementConfig) {
			// Judge-on-seal: swap the auraIsActive(SoW) check inside the AND branch of the OR.
			const judgeEntry = rotation.priorityList[PRIORITY_JUDGE_ON_SEAL_INDEX];
			const judgeAndCondition = ((judgeEntry.action!.condition!.value as any).or.vals[0].value as any).and;
			const sealActiveCheck = judgeAndCondition.vals[1];
			const sealActiveAuraId = (sealActiveCheck.value as any).auraIsActive.auraId;
			sealActiveAuraId.rawId = { oneofKind: 'spellId', spellId: judgementConfig.sealSpellId };
			sealActiveAuraId.rank = judgementConfig.sealRank;

			// Swap-seal: unwrap the AND keeping [SoW inactive, JoX missing, Judgement ready]. Swap the seal and debuff auras.
			const swapEntry = rotation.priorityList[PRIORITY_SWAP_SEAL_INDEX];
			const swapAndVals = (swapEntry.action!.condition!.value as any).and.vals;
			const sealInactiveAuraId = (swapAndVals[1].value as any).auraIsInactive.auraId;
			sealInactiveAuraId.rawId = { oneofKind: 'spellId', spellId: judgementConfig.sealSpellId };
			sealInactiveAuraId.rank = judgementConfig.sealRank;
			const judgementInactiveAuraId = (swapAndVals[2].value as any).auraIsInactive.auraId;
			judgementInactiveAuraId.rawId = { oneofKind: 'spellId', spellId: judgementConfig.judgementAuraSpellId };
			judgementInactiveAuraId.rank = judgementConfig.judgementAuraRank;
			// (swapEntry.action!.condition!.value as any).and.vals = swapAndVals.slice(1);
			const swapSealCast = (swapEntry.action!.action as any).castSpell;
			swapSealCast.spellId.rawId = { oneofKind: 'spellId', spellId: judgementConfig.sealSpellId };
			swapSealCast.spellId.rank = judgementConfig.sealRank;

			// Righteousness judge: replace target aura spell check with judgement aura
			const righteousnessJudgeEntry = rotation.priorityList[PRIORITY_RIGHTEOUSNESS_JUDGE_INDEX];
			const judgeAndVals = (righteousnessJudgeEntry.action!.condition!.value as any).and.vals;
			const orVals = (judgeAndVals[2].value as any).or.vals;
			const auraRemainingTime = (orVals[1].value as any).cmp.lhs.value.auraRemainingTime.auraId;
			auraRemainingTime.rawId = { oneofKind: 'spellId', spellId: judgementConfig.judgementAuraSpellId };
			auraRemainingTime.rank = judgementConfig.judgementAuraRank;
		}

		// Consecration rank swap (removal handled by the filter below).
		if (consecrationRank !== 0) {
			const consecrationCast = (rotation.priorityList[PRIORITY_CONSECRATION_INDEX].action!.action as any).castSpell;
			consecrationCast.spellId.rawId = { oneofKind: 'spellId', spellId: CONSECRATION_RANK_SPELL_IDS[consecrationRank] };
			consecrationCast.spellId.rank = consecrationRank;
		}

		// Aura swap: replace the SpellID of the prepull aura cast. If None is
		// picked the action is filtered out entirely.
		const auraSpellId = AURA_SPELL_IDS[aura];
		if (auraSpellId !== null) {
			const auraCast = (rotation.prepullActions[PREPULL_AURA_INDEX].action!.action as any).castSpell;
			auraCast.spellId.rawId = { oneofKind: 'spellId', spellId: auraSpellId };
			auraCast.spellId.rank = 0;
			auraCast.spellId.tag = AURA_TAGS[aura];
		}

		// Drop Consecration when disabled, and the prepull aura when the user
		// picked None. (The maintenance actions stay in place for
		// JudgementNone — their Const:false keeps them dormant.)
		const priorityList = rotation.priorityList.filter((_, i) => {
			if (i === PRIORITY_CONSECRATION_INDEX && consecrationRank === 0) return false;
			return true;
		});
		const prepullActions = rotation.prepullActions.filter((_, i) => {
			if (i === PREPULL_AURA_INDEX && auraSpellId === null) return false;
			return true;
		});

		return APLRotation.create({
			prepullActions: prepullActions,
			priorityList: [
				...actions.map(action =>
					APLListItem.create({
						action: action,
					}),
				),
				...priorityList,
			],
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
