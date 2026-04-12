// Configuration for spec-specific UI elements on the settings tab.
// These don't need to be in a separate file but it keeps things cleaner.
import * as InputHelpers from '../../core/components/input_helpers';
import { Spec } from '../../core/proto/common';

export const PaladinRotationConfig = {
	inputs: [
		InputHelpers.makeRotationBooleanInput<Spec.SpecProtectionPaladin>({
			fieldName: 'prioritizeHolyShield',
			label: 'Prioritize Holy Shield',
			labelTooltip: 'If <b>true</b>, Holy Shield is cast at highest priority. If <b>false</b>, Holy Shield is cast after Consecration and Judgement.',
			changeEmitter: player => player.rotationChangeEmitter,
			getValue: player => player.getSimpleRotation().prioritizeHolyShield,
		}),
		InputHelpers.makeRotationBooleanInput<Spec.SpecProtectionPaladin>({
			fieldName: 'useConsecrate',
			label: 'Use Consecrate',
			labelTooltip: 'If <b>true</b>, will use Consecration (Rank 6) in the rotation.',
			changeEmitter: player => player.rotationChangeEmitter,
			getValue: player => player.getSimpleRotation().useConsecrate,
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
		InputHelpers.makeRotationBooleanInput<Spec.SpecProtectionPaladin>({
			fieldName: 'maintainJudgementOfWisdom',
			label: 'Maintain Judgement of Wisdom',
			labelTooltip: 'If <b>true</b>, will prioritize judging to keep Judgement of Wisdom active on the target for mana sustain. Disable this if JoW is already provided by raid settings.',
			changeEmitter: player => player.rotationChangeEmitter,
			getValue: player => player.getSimpleRotation().maintainJudgementOfWisdom,
		}),
	],
};
