// Configuration for spec-specific UI elements on the settings tab.
// These don't need to be in a separate file but it keeps things cleaner.
import * as InputHelpers from '../../core/components/input_helpers';
import { Spec } from '../../core/proto/common';
import { PaladinAura, PaladinJudgement } from '../../core/proto/paladin';
import { ActionId } from '../../core/proto_utils/action_id';

export const PaladinRotationConfig = {
	inputs: [
		InputHelpers.makeRotationBooleanInput<Spec.SpecProtectionPaladin>({
			fieldName: 'prioritizeHolyShield',
			label: 'Prioritize Holy Shield',
			labelTooltip: 'If <b>true</b>, Holy Shield is cast at highest priority. If <b>false</b>, Holy Shield is cast after Consecration and Judgement.',
			changeEmitter: player => player.rotationChangeEmitter,
			getValue: player => player.getSimpleRotation().prioritizeHolyShield,
		}),
		InputHelpers.makeRotationEnumInput<Spec.SpecProtectionPaladin, number>({
			fieldName: 'consecrationRank',
			label: 'Consecration Rank',
			labelTooltip: 'Which rank of Consecration to use in the rotation. Select <b>Do not use</b> to disable.',
			values: [
				{ name: 'Do not use', value: 0 },
				{ name: 'Rank 1', value: 1 },
				{ name: 'Rank 2', value: 2 },
				{ name: 'Rank 3', value: 3 },
				{ name: 'Rank 4', value: 4 },
				{ name: 'Rank 5', value: 5 },
				{ name: 'Rank 6', value: 6 },
			],
			changeEmitter: player => player.rotationChangeEmitter,
			getValue: player => player.getSimpleRotation().consecrationRank,
		}),
		InputHelpers.makeRotationBooleanInput<Spec.SpecProtectionPaladin>({
			fieldName: 'useExorcism',
			label: 'Use Exorcism',
			labelTooltip: 'If <b>true</b>, will use Exorcism in the rotation (only effective against Undead and Demons).',
			changeEmitter: player => player.rotationChangeEmitter,
			getValue: player => player.getSimpleRotation().useExorcism,
		}),
		InputHelpers.makeRotationBooleanInput<Spec.SpecProtectionPaladin>({
			fieldName: 'useAvengersShield',
			label: "Use Avenger's Shield",
			labelTooltip: "If <b>true</b>, will use Avenger's Shield in the rotation.",
			changeEmitter: player => player.rotationChangeEmitter,
			getValue: player => player.getSimpleRotation().useAvengersShield,
		}),
		InputHelpers.makeRotationEnumIconInput<Spec.SpecProtectionPaladin, PaladinJudgement>({
			fieldName: 'maintainJudgement',
			label: 'Maintain Judgement',
			labelTooltip:
				'Which Judgement debuff to keep active on the target. The matching Seal will be used before each Judgement. Pick <b>None</b> to keep Seal of Righteousness up and skip Judgement maintenance.',
			values: [
				{ color: 'grey', value: PaladinJudgement.JudgementNone, tooltip: 'None' },
				{ actionId: ActionId.fromSpellId(27162), value: PaladinJudgement.JudgementOfLight, tooltip: 'Judgement of Light' },
				{ actionId: ActionId.fromSpellId(27164), value: PaladinJudgement.JudgementOfWisdom, tooltip: 'Judgement of Wisdom' },
			],
			changeEmitter: player => player.rotationChangeEmitter,
			getValue: player => player.getSimpleRotation().maintainJudgement,
		}),
		InputHelpers.makeRotationEnumIconInput<Spec.SpecProtectionPaladin, PaladinAura>({
			fieldName: 'aura',
			label: 'Aura',
			labelTooltip: 'Which paladin aura to activate in the prepull. Pick <b>None</b> to skip casting an aura.',
			values: [
				{ color: 'grey', value: PaladinAura.AuraNone, tooltip: 'None' },
				{ actionId: ActionId.fromSpellId(27149), value: PaladinAura.DevotionAura, tooltip: 'Devotion Aura' },
				{ actionId: ActionId.fromSpellId(27150), value: PaladinAura.RetributionAura, tooltip: 'Retribution Aura' },
				{ actionId: ActionId.fromSpellId(19746), value: PaladinAura.ConcentrationAura, tooltip: 'Concentration Aura' },
				{ actionId: ActionId.fromSpellId(27153), value: PaladinAura.FireResistanceAura, tooltip: 'Fire Resistance Aura' },
				{ actionId: ActionId.fromSpellId(27152), value: PaladinAura.FrostResistanceAura, tooltip: 'Frost Resistance Aura' },
				{ actionId: ActionId.fromSpellId(27151), value: PaladinAura.ShadowResistanceAura, tooltip: 'Shadow Resistance Aura' },
			],
			changeEmitter: player => TypedEvent.onAny([player.rotationChangeEmitter, player.talentsChangeEmitter]),
			getValue: player => player.getSimpleRotation().aura,
		}),
	],
};
