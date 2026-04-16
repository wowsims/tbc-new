// Configuration for spec-specific UI elements on the settings tab.
// These don't need to be in a separate file but it keeps things cleaner.
import * as InputHelpers from '../../core/components/input_helpers';
import { Spec } from '../../core/proto/common';
import { PaladinJudgement } from '../../core/proto/paladin';
import { ActionId } from '../../core/proto_utils/action_id';
import * as SharedPaladinInputs from '../inputs';

export const PaladinRotationConfig = {
	inputs: [
		InputHelpers.makeRotationBooleanInput<Spec.SpecProtectionPaladin>({
			fieldName: 'prioritizeHolyShield',
			label: 'Prioritize Holy Shield',
			labelTooltip: 'If <b>true</b>, Holy Shield is cast at highest priority. If <b>false</b>, Holy Shield is cast after Consecration and Judgement.',
			changeEmitter: player => player.rotationChangeEmitter,
			getValue: player => player.getSimpleRotation().prioritizeHolyShield,
		}),
		SharedPaladinInputs.ConsecrationRankInput<Spec.SpecProtectionPaladin>(
			'Which rank of Consecration to use in the rotation. Select <b>Do not use</b> to disable.',
		),
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
		SharedPaladinInputs.AuraInput<Spec.SpecProtectionPaladin>(
			'Which paladin aura to activate in the prepull. <b>Sanctity Aura</b> requires the talent. Pick <b>None</b> to skip casting an aura.',
		),
	],
};
