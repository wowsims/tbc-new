// Configuration for spec-specific UI elements on the settings tab.
// These don't need to be in a separate file but it keeps things cleaner.
import * as InputHelpers from '../../core/components/input_helpers';
import { Spec } from '../../core/proto/common';

export const PaladinRotationConfig = {
	inputs: [
		InputHelpers.makeRotationBooleanInput<Spec.SpecRetributionPaladin>({
			fieldName: 'useExorcism',
			label: 'Use Exorcism',
			labelTooltip: 'If true, will use Excorism in rotation if target is undead or demon.',
			changeEmitter: player => player.rotationChangeEmitter,
			getValue: player => player.getSimpleRotation().useExorcism,
		}),
		InputHelpers.makeRotationBooleanInput<Spec.SpecRetributionPaladin>({
			fieldName: 'useConsecrate',
			label: 'Use Consecrate',
			labelTooltip: 'If true, will use Consecrate in rotation. Exorcism is priority',
			changeEmitter: player => player.rotationChangeEmitter,
			getValue: player => player.getSimpleRotation().useConsecrate,
		}),
	],
};
