import * as InputHelpers from '../core/components/input_helpers';
import { WarriorShout, WarriorStance } from '../core/proto/warrior';
import { ActionId } from '../core/proto_utils/action_id';
import { WarriorSpecs } from '../core/proto_utils/utils';
import i18n from '../i18n/config.js';

// Configuration for class-specific UI elements on the settings tab.
// These don't need to be in a separate file but it keeps things cleaner.
export const ShoutPicker = <SpecType extends WarriorSpecs>() =>
	InputHelpers.makeClassOptionsEnumIconInput<SpecType, WarriorShout>({
		fieldName: 'defaultShout',
		label: i18n.t('settings_tab.other.default_shout.label'),
		labelTooltip: i18n.t('settings_tab.other.default_shout.label'),
		values: [
			{ actionId: ActionId.fromSpellId(6673), value: WarriorShout.WarriorShoutBattle },
			{ actionId: ActionId.fromSpellId(469), value: WarriorShout.WarriorShoutCommanding },
		],
	});
export const StancePicker = <SpecType extends WarriorSpecs>() =>
	InputHelpers.makeClassOptionsEnumIconInput<SpecType, WarriorStance>({
		fieldName: 'defaultStance',
		label: i18n.t('settings_tab.other.default_stance.label'),
		labelTooltip: i18n.t('settings_tab.other.default_stance.label'),
		values: [
			{ actionId: ActionId.fromSpellId(2457), value: WarriorStance.WarriorStanceBattle },
			{ actionId: ActionId.fromSpellId(2458), value: WarriorStance.WarriorStanceBerserker },
			{ actionId: ActionId.fromSpellId(71), value: WarriorStance.WarriorStanceDefensive },
		],
	});

export const StartingRage = <SpecType extends WarriorSpecs>() =>
	InputHelpers.makeClassOptionsNumberInput<SpecType>({
		fieldName: 'startingRage',
		label: i18n.t('settings_tab.other.starting_rage.label'),
		labelTooltip: i18n.t('settings_tab.other.starting_rage.tooltip'),
	});

export const StanceSnapshot = <SpecType extends WarriorSpecs>() =>
	InputHelpers.makeClassOptionsBooleanInput<SpecType>({
		fieldName: 'stanceSnapshot',
		label: i18n.t('settings_tab.other.stance_snapshot.label'),
		labelTooltip: i18n.t('settings_tab.other.stance_snapshot.tooltip'),
	});

export const QueueDelay = <SpecType extends WarriorSpecs>() =>
	InputHelpers.makeClassOptionsNumberInput<SpecType>({
		fieldName: 'queueDelay',
		label: i18n.t('settings_tab.other.queue_delay.label'),
		labelTooltip: i18n.t('settings_tab.other.queue_delay.tooltip'),
	});

export const BattleShoutSolarianSapphire = <SpecType extends WarriorSpecs>() =>
	InputHelpers.makeClassOptionsBooleanIconInput<SpecType>({
		fieldName: 'hasBsSolarianSapphire',
		label: i18n.t('settings_tab.other.has_bs_solarian_sapphire.label'),
		labelTooltip: i18n.t('settings_tab.other.has_bs_solarian_sapphire.tooltip'),
		actionId: () => ActionId.fromItemId(30446),
	});

export const BattleShoutT2 = <SpecType extends WarriorSpecs>() =>
	InputHelpers.makeClassOptionsBooleanIconInput<SpecType>({
		fieldName: 'hasBsT2',
		label: i18n.t('settings_tab.other.has_bs_tier_2.label'),
		labelTooltip: i18n.t('settings_tab.other.has_bs_tier_2.tooltip'),
		actionId: () => ActionId.fromSpellId(23563),
	});
