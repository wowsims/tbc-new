import { ContentBlock } from '../core/components/content_block';
import { buildIconInput } from '../core/components/icon_inputs';
import { Input } from '../core/components/input';
import * as InputHelpers from '../core/components/input_helpers';
import { BooleanPicker } from '../core/components/pickers/boolean_picker';
import { NumberPicker } from '../core/components/pickers/number_picker';
import { IndividualSimUI } from '../core/individual_sim_ui';
import { Player } from '../core/player';
import { Spec } from '../core/proto/common';
import { ShamanImbue, ShamanShield } from '../core/proto/shaman';
import { ActionId } from '../core/proto_utils/action_id';
import { ShamanSpecs } from '../core/proto_utils/utils';
import { EventID, TypedEvent } from '../core/typed_event';
import i18n from '../i18n/config';

// Configuration for class-specific UI elements on the settings tab.
// These don't need to be in a separate file but it keeps things cleaner.

export const ShamanShieldInput = <SpecType extends ShamanSpecs>() =>
	InputHelpers.makeClassOptionsEnumIconInput<SpecType, ShamanShield>({
		fieldName: 'shield',
		values: [
			{ value: ShamanShield.NoShield, tooltip: 'No Shield' },
			{ actionId: ActionId.fromSpellId(52127), value: ShamanShield.WaterShield },
			{ actionId: ActionId.fromSpellId(324), value: ShamanShield.LightningShield },
		],
	});

export const ShamanImbueMH = <SpecType extends ShamanSpecs>() =>
	InputHelpers.makeClassOptionsEnumIconInput<SpecType, ShamanImbue>({
		fieldName: 'imbueMh',
		values: [
			{ value: ShamanImbue.NoImbue, tooltip: 'No Main Hand Enchant' },
			{ actionId: ActionId.fromSpellId(8232), value: ShamanImbue.WindfuryWeapon },
			{ actionId: ActionId.fromSpellId(8024), value: ShamanImbue.FlametongueWeapon },
			{ actionId: ActionId.fromSpellId(8033), value: ShamanImbue.FrostbrandWeapon },
		],
	});

export const ShamanImbueMHSwap = <SpecType extends ShamanSpecs>() =>
	InputHelpers.makeClassOptionsEnumIconInput<SpecType, ShamanImbue>({
		fieldName: 'imbueMhSwap',
		values: [
			{ value: ShamanImbue.NoImbue, tooltip: 'No Main Hand Swap Enchant' },
			{ actionId: ActionId.fromSpellId(8232), value: ShamanImbue.WindfuryWeapon },
			{ actionId: ActionId.fromSpellId(8024), value: ShamanImbue.FlametongueWeapon },
		],
		showWhen: (player: Player<SpecType>) => player.itemSwapSettings.getEnableItemSwap(),
		changeEmitter: (player: Player<SpecType>) => TypedEvent.onAny([player.specOptionsChangeEmitter, player.itemSwapSettings.changeEmitter]),
	});

export function TotemsSection(parentElem: HTMLElement, simUI: IndividualSimUI<any>): ContentBlock {
	const contentBlock = new ContentBlock(parentElem, 'totems-settings', {
		header: { title: 'Totems' },
	});

	contentBlock.bodyElement.querySelectorAll('.input-root').forEach(elem => {
		elem.classList.add('input-inline');
	});

	return contentBlock;
}
