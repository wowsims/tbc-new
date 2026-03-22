import * as InputHelpers from '../../core/components/input_helpers';
import { PriestOptions_Armor } from '../../core/proto/priest';
import { ActionId } from '../../core/proto_utils/action_id';
import { PriestSpecs } from '../../core/proto_utils/utils';

// Configuration for class-specific UI elements on the settings tab.
// These don't need to be in a separate file but it keeps things cleaner.

export const ShadowformInput = <SpecType extends PriestSpecs>() =>
	InputHelpers.makeClassOptionsBooleanIconInput<SpecType>({
		fieldName: 'preShadowform',
		actionId: () => ActionId.fromSpellId(15473),
		label: 'Shadowform',
		showWhen: player => player.getTalents().shadowform,
	});
