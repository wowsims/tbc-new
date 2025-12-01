// Configuration for spec-specific UI elements on the settings tab.
// These don't need to be in a separate file but it keeps things cleaner.

import * as InputHelpers from '../../core/components/input_helpers.js';
import { Profession, Spec, Stat } from '../../core/proto/common.js';
import { WarriorSyncType } from '../../core/proto/warrior';
import { Stats } from '../../core/proto_utils/stats';
import i18n from '../../i18n/config.js';

export const SyncTypeInput = InputHelpers.makeSpecOptionsEnumInput<Spec.SpecDPSWarrior, WarriorSyncType>({
	fieldName: 'syncType',
	label: i18n.t('settings_tab.other.sync_type.label'),
	labelTooltip: i18n.t('settings_tab.other.sync_type.tooltip'),
	values: [
		{ name: i18n.t('settings_tab.other.sync_type.values.none'), value: WarriorSyncType.WarriorNoSync },
		{ name: i18n.t('settings_tab.other.sync_type.values.perfect_sync'), value: WarriorSyncType.WarriorSyncMainhandOffhandSwings },
	],
});
