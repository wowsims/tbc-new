import * as InputHelpers from '../core/components/input_helpers';
import { WarriorSpecs } from '../core/proto_utils/utils';
import i18n from '../i18n/config.js';

// Configuration for class-specific UI elements on the settings tab.
// These don't need to be in a separate file but it keeps things cleaner.

export const StartingRage = <SpecType extends WarriorSpecs>() =>
	InputHelpers.makeSpecOptionsNumberInput<SpecType>({
		fieldName: 'startingRage',
		label: i18n.t('settings_tab.other.starting_rage.label'),
		labelTooltip: i18n.t('settings_tab.other.starting_rage.tooltip'),
	});

export const StanceSnapshot = <SpecType extends WarriorSpecs>() =>
	InputHelpers.makeSpecOptionsBooleanInput<SpecType>({
		fieldName: 'stanceSnapshot',
		label: i18n.t('settings_tab.other.stance_snapshot.label'),
		labelTooltip: i18n.t('settings_tab.other.stance_snapshot.tooltip'),
	});

export const QueueDelay = <SpecType extends WarriorSpecs>() =>
	InputHelpers.makeSpecOptionsNumberInput<SpecType>({
		fieldName: 'queueDelay',
		label: 'HS/Cleave Queue Delay (ms)',
		labelTooltip: 'How long (in milliseconds) to delay re-queueing Heroic Strike/Cleave in order to simulate real reaction time and game delay.',
	});
