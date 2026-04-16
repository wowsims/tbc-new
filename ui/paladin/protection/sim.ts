import * as OtherInputs from '../../core/components/inputs/other_inputs.js';
import { IndividualSimUI, registerSpecConfig } from '../../core/individual_sim_ui.js';
import { Player } from '../../core/player.js';
import { PlayerClasses } from '../../core/player_classes';
import { APLRotation, APLRotation_Type, APLValueVariable, SimpleRotation } from '../../core/proto/apl.js';
import { Cooldowns, PseudoStat, Spec, Stat } from '../../core/proto/common.js';
import { PaladinAura, PaladinJudgement } from '../../core/proto/paladin.js';
import { Stats, UnitStat } from '../../core/proto_utils/stats.js';
import * as Presets from './presets.js';
import * as ProtPaladinInputs from './inputs.js';
import * as Mechanics from '../../core/constants/mechanics';
import { ReforgeOptimizer } from '../../core/components/suggest_reforges_action';

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
const PREPULL_SEAL_INDEX = 2; // Seal of Righteousness at -3s
const PRIORITY_JUDGE_ON_SEAL_INDEX = 0; // Const-false-gated: when maintenance seal is up, judge it
const PRIORITY_SWAP_SEAL_INDEX = 2; // Const-false-gated: when maintenance seal is down, JoX is down, and Judgement is ready, swap to maintenance seal
const PRIORITY_CONSECRATION_INDEX = 5; // Consecration rank 6

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

type JudgementSpec = { sealSpellId: number; sealRank: number; judgementAuraSpellId: number; judgementAuraRank: number };
const JUDGEMENT_CONFIG: Record<PaladinJudgement, JudgementSpec | null> = {
	[PaladinJudgement.JudgementNone]: null,
	[PaladinJudgement.JudgementOfLight]: { sealSpellId: 27160, sealRank: 5, judgementAuraSpellId: 27163, judgementAuraRank: 0 },
	[PaladinJudgement.JudgementOfWisdom]: { sealSpellId: 27166, sealRank: 4, judgementAuraSpellId: 27164, judgementAuraRank: 0 },
};

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

	simpleRotation: (player, simple): APLRotation => {
		const rotation = APLRotation.clone(Presets.APL_PRESET.rotation.rotation!);

		const {
			prioritizeHolyShield = true,
			consecrationRank = 6,
			useExorcism = false,
			useAvengersShield = true,
			maintainJudgement = PaladinJudgement.JudgementNone,
			aura: rawAura = PaladinAura.DevotionAura,
		} = simple;

		// Sanctity Aura requires the talent. If the user picked it without the
		// talent (e.g. dropped the point after selecting), fall back to None.
		const aura = rawAura === PaladinAura.SanctityAura && !player.getTalents().sanctityAura ? PaladinAura.AuraNone : rawAura;

		rotation.valueVariables = [
			APLValueVariable.fromJson({ name: 'Prioritize Holy Shield', value: { const: { val: String(prioritizeHolyShield) } } }),
			APLValueVariable.fromJson({ name: 'Use Exorcism', value: { const: { val: String(useExorcism) } } }),
			APLValueVariable.fromJson({ name: "Use Avenger's Shield", value: { const: { val: String(useAvengersShield) } } }),
		];

		const judgementConfig = JUDGEMENT_CONFIG[maintainJudgement];

		// For Light/Wisdom we activate the two maintenance actions that are
		// dormant in the default APL:
		//   - Judge-on-seal (action 0): [Const:false AND SoW active] -> Cast
		//     Judgement. Drop the Const:false so the action fires whenever the
		//     maintenance seal is up, consuming it to apply the Judgement
		//     debuff ASAP (before Holy Shield/Consecration).
		//   - Swap-seal (action 2): [Const:false AND SoW inactive AND JoW
		//     missing AND Judgement ready] -> Cast SoW. Drop the Const:false
		//     so the action prepares the maintenance seal whenever it's time
		//     to refresh the debuff.
		// Both actions also need the Seal of Wisdom / Judgement of Wisdom
		// references rewritten when the user picked Judgement of Light. The
		// prepull seal (SoR by default) is also swapped so combat starts with
		// the maintenance seal up and a free Judgement applies the debuff.
		//
		// For JudgementNone the Const:false keeps both actions dormant; no
		// mutation is needed.
		if (judgementConfig) {
			const prepullSealCast = (rotation.prepullActions[PREPULL_SEAL_INDEX].action!.action as any).castSpell;
			prepullSealCast.spellId.rawId = { oneofKind: 'spellId', spellId: judgementConfig.sealSpellId };
			prepullSealCast.spellId.rank = judgementConfig.sealRank;

			// Judge-on-seal: unwrap the AND keeping only the auraIsActive(SoW) clause, then swap its aura to the chosen seal.
			const judgeEntry = rotation.priorityList[PRIORITY_JUDGE_ON_SEAL_INDEX];
			const judgeCondition = (judgeEntry.action!.condition!.value as any).and;
			const sealActiveCheck = judgeCondition.vals[1];
			const sealActiveAuraId = (sealActiveCheck.value as any).auraIsActive.auraId;
			sealActiveAuraId.rawId = { oneofKind: 'spellId', spellId: judgementConfig.sealSpellId };
			sealActiveAuraId.rank = judgementConfig.sealRank;
			judgeCondition.vals = judgeCondition.vals.slice(1);

			// Swap-seal: unwrap the AND keeping [SoW inactive, JoX missing, Judgement ready]. Swap the seal and debuff auras.
			const swapEntry = rotation.priorityList[PRIORITY_SWAP_SEAL_INDEX];
			const swapAndVals = (swapEntry.action!.condition!.value as any).and.vals;
			const sealInactiveAuraId = (swapAndVals[1].value as any).auraIsInactive.auraId;
			sealInactiveAuraId.rawId = { oneofKind: 'spellId', spellId: judgementConfig.sealSpellId };
			sealInactiveAuraId.rank = judgementConfig.sealRank;
			const judgementInactiveAuraId = (swapAndVals[2].value as any).auraIsInactive.auraId;
			judgementInactiveAuraId.rawId = { oneofKind: 'spellId', spellId: judgementConfig.judgementAuraSpellId };
			judgementInactiveAuraId.rank = judgementConfig.judgementAuraRank;
			(swapEntry.action!.condition!.value as any).and.vals = swapAndVals.slice(1);
			const swapSealCast = (swapEntry.action!.action as any).castSpell;
			swapSealCast.spellId.rawId = { oneofKind: 'spellId', spellId: judgementConfig.sealSpellId };
			swapSealCast.spellId.rank = judgementConfig.sealRank;
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
			simple: SimpleRotation.create({
				cooldowns: Cooldowns.create(),
			}),
			prepullActions: prepullActions,
			priorityList: priorityList,
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
