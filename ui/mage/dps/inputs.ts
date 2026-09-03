import * as InputHelpers from '../../core/components/input_helpers';
import { MageArmor } from '../../core/proto/mage';
import { ActionId } from '../../core/proto_utils/action_id';
import { MageSpecs } from '../../core/proto_utils/utils';

// Configuration for class-specific UI elements on the settings tab.
// These don't need to be in a separate file but it keeps things cleaner.
export const MageArmorInputs = <SpecType extends MageSpecs>() =>
	InputHelpers.makeClassOptionsEnumIconInput<SpecType, MageArmor>({
		fieldName: 'defaultMageArmor',
		values: [
			{ value: MageArmor.MageArmorNone, tooltip: 'No Armor' },
			{ actionId: ActionId.fromSpellId(7302), value: MageArmor.MageArmorFrostArmor },
			{ actionId: ActionId.fromSpellId(6117), value: MageArmor.MageArmorMageArmor },
			{ actionId: ActionId.fromSpellId(30482), value: MageArmor.MageArmorMoltenArmor },
		],
	});

export const ArcaneMageRotationConfig = {
	inputs: [
		InputHelpers.makeRotationNumberInput<MageSpecs>({
			fieldName: 'conserveStart',
			label: 'Start Conserve Rotation %',
			labelTooltip: 'Starts the conserve mana rotation at %',
			changeEmitter: player => player.rotationChangeEmitter,
			getValue: player => player.getSimpleRotation().conserveStart,
			positive: true,
		}),
		InputHelpers.makeRotationNumberInput<MageSpecs>({
			fieldName: 'conserveEnd',
			label: 'End Conserve Rotation %',
			labelTooltip:
				'Ends the conserve mana rotation once mana reaches this threshold %, Conserve Rotation stops if its possible to spam AB till the end of the fight.',
			changeEmitter: player => player.rotationChangeEmitter,
			getValue: player => player.getSimpleRotation().conserveEnd,
			positive: true,
		}),
		InputHelpers.makeRotationNumberInput<MageSpecs>({
			fieldName: 'delayMajorCDs',
			label: 'Delay Major CDs',
			labelTooltip: 'Delays the first automatic use of major cooldowns (e.g. Bloodlust, Drums) by the specified number of seconds.',
			changeEmitter: player => player.rotationChangeEmitter,
			getValue: player => player.getSimpleRotation().delayMajorCDs,
			positive: true,
		}),
	],
};
