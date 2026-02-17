import * as InputHelpers from '../../core/components/input_helpers.js';
import { Player } from '../../core/player.js';
import { RotationType, Spec } from '../../core/proto/common.js';
import { WarlockOptions_Summon as Summon, WarlockOptions_Armor, WarlockOptions_CurseOptions } from '../../core/proto/warlock.js';
import { ActionId } from '../../core/proto_utils/action_id.js';
import { WarlockSpecs } from '../../core/proto_utils/utils.js';
import { ContentBlock } from '../../core/components/content_block';
import i18n from '../../i18n/config.js';
import { IndividualSimUI } from '../../core/individual_sim_ui';
import { Input } from '../../core/components/input';
import { buildIconInput } from '../../core/components/icon_inputs';
import { EventID } from '../../core/typed_event';

// Configuration for spec-specific UI elements on the settings tab.
// These don't need to be in a separate file but it keeps things cleaner.

export const PetInput = <SpecType extends WarlockSpecs>() =>
	InputHelpers.makeClassOptionsEnumIconInput<SpecType, Summon>({
		fieldName: 'summon',
		values: [
			{ value: Summon.NoSummon, tooltip: 'No Pet' },
			{ actionId: ActionId.fromSpellId(691), value: Summon.Felhunter },
			{ actionId: ActionId.fromSpellId(688), value: Summon.Imp },
			{ actionId: ActionId.fromSpellId(712), value: Summon.Succubus },
			{ actionId: ActionId.fromSpellId(697), value: Summon.Voidwalker },
		],
		changeEmitter: (player: Player<SpecType>) => player.changeEmitter,
	});
export const PetAttackOptions = <SpecType extends WarlockSpecs>() =>
	InputHelpers.makeClassOptionsEnumInput<SpecType, Summon>({
		fieldName: 'summon',
		label: 'peinput2' ,
		values: [
			{ value: Summon.NoSummon, name: 'No Pet', },
			{ value: Summon.Felhunter, name: 'FelHunter'},
			{ value: Summon.Imp, name: 'Imp'},
			{ value: Summon.Succubus, name: 'Succubus'},
			{ value: Summon.Voidwalker, name: 'VoidWalker'},
		],
		changeEmitter: (player: Player<SpecType>) => player.changeEmitter,
	});

export const ArmorInput = <SpecType extends WarlockSpecs>() =>
	InputHelpers.makeClassOptionsEnumIconInput<SpecType, WarlockOptions_Armor>({
		fieldName: 'armor',
		values: [
			{ value: WarlockOptions_Armor.NoArmor, tooltip: 'No Armor'},
			{ actionId: ActionId.fromSpellId(28176), value: WarlockOptions_Armor.FelArmor},
			{ actionId: ActionId.fromSpellId(706), value: WarlockOptions_Armor.DemonArmor}
		]
	})

export const DemonicSacrificeInput = <SpecType extends WarlockSpecs>() =>
	InputHelpers.makeClassOptionsBooleanIconInput<SpecType>({
		fieldName: 'sacrificeSummon',
		actionId: () => ActionId.fromSpellId(18788),
		getValue: (player: Player<SpecType>) => player.getClassOptions().sacrificeSummon && player.getTalents().demonicSacrifice && player.getClassOptions().summon != Summon.NoSummon,
		setValue: (eventID: number, player: Player<SpecType>, newValue: boolean) => {
			const newOptions = player.getClassOptions();
			newOptions.sacrificeSummon = newValue;
			player.setClassOptions(eventID, newOptions);
		},
		changeEmitter: (player: Player<SpecType>) => player.specOptionsChangeEmitter,
	})

export function CursesSection(parentElem: HTMLElement, simUI: IndividualSimUI<any>): ContentBlock {
	const contentBlock = new ContentBlock(parentElem, 'curses-settings', {
		header: { title: 'Curses' },
	});

	const curses = Input.newGroupContainer();
	curses.classList.add('curses-toggle-container', 'icon-group');

	contentBlock.bodyElement.appendChild(curses);

	buildIconInput(curses, simUI.player,
		makeCursePicker(WarlockOptions_CurseOptions.Agony, 980));

	buildIconInput(curses, simUI.player,
		makeCursePicker(WarlockOptions_CurseOptions.Doom, 603));

	buildIconInput(curses, simUI.player,
		makeCursePicker(WarlockOptions_CurseOptions.Elements, 1490));

	buildIconInput(curses, simUI.player,
		makeCursePicker(WarlockOptions_CurseOptions.Recklessness, 704));

	contentBlock.bodyElement.querySelectorAll('.input-root').forEach(elem => {
		elem.classList.add('input-inline');
	});

	return contentBlock;
}

const makeCursePicker = <SpecType extends WarlockSpecs>(
	curse: WarlockOptions_CurseOptions,
	spellId: number,
) =>
	InputHelpers.makeClassOptionsBooleanIconInput<SpecType>({
		fieldName: 'curseOptions',
		actionId: () => ActionId.fromSpellId(spellId),

		getValue: (player: Player<SpecType>) =>
			player.getClassOptions().curseOptions === curse,

		setValue: (eventID: EventID, player: Player<SpecType>, newValue: boolean) => {
			if (!newValue) return;

			const newOptions = player.getClassOptions();
			newOptions.curseOptions = curse;
			player.setClassOptions(eventID, newOptions);
		},

		changeEmitter: (player: Player<SpecType>) =>
			player.specOptionsChangeEmitter,
	});
