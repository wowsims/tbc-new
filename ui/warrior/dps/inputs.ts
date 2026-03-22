// Configuration for spec-specific UI elements on the settings tab.
import * as InputHelpers from '../../core/components/input_helpers';
import { Spec } from '../../core/proto/common';
import { DpsWarriorSpec, WarriorSunder } from '../../core/proto/warrior';
import { TypedEvent } from '../../core/typed_event';
import i18n from '../../i18n/config';
import { isArmsKebabSpec, isFurySpec } from './presets';

// These don't need to be in a separate file but it keeps things cleaner.
export const RotationInputs = {
	inputs: [
		InputHelpers.makeRotationEnumInput<Spec.SpecDpsWarrior, DpsWarriorSpec>({
			fieldName: 'spec',
			label: i18n.t('rotation_tab.options.warrior.dps.spec.label'),
			values: [
				{ name: i18n.t('rotation_tab.options.warrior.dps.spec.fury'), value: DpsWarriorSpec.DpsWarriorSpecFury },
				{ name: i18n.t('rotation_tab.options.warrior.dps.spec.arms'), value: DpsWarriorSpec.DpsWarriorSpecArms },
			],
		}),
		InputHelpers.makeRotationNumberInput<Spec.SpecDpsWarrior>({
			fieldName: 'bloodlustTiming',
			label: i18n.t('rotation_tab.options.warrior.bloodlust_timing.label'),
			labelTooltip: i18n.t('rotation_tab.options.warrior.bloodlust_timing.tooltip'),
		}),
		InputHelpers.makeRotationEnumInput<Spec.SpecDpsWarrior, WarriorSunder>({
			fieldName: 'sunderArmor',
			label: i18n.t('rotation_tab.options.warrior.sunder_armor.label'),
			values: [
				{ name: i18n.t('rotation_tab.options.warrior.sunder_armor.none'), value: WarriorSunder.WarriorSunderNone },
				{ name: i18n.t('rotation_tab.options.warrior.sunder_armor.help'), value: WarriorSunder.WarriorSunderHelp },
				{ name: i18n.t('rotation_tab.options.warrior.sunder_armor.maintain'), value: WarriorSunder.WarriorSunderMaintain },
			],
		}),
		InputHelpers.makeRotationBooleanInput<Spec.SpecDpsWarrior>({
			fieldName: 'useRecklessness',
			label: i18n.t('rotation_tab.options.warrior.dps.use_recklessness.label'),
		}),
		InputHelpers.makeRotationBooleanInput<Spec.SpecDpsWarrior>({
			fieldName: 'useOverpower',
			label: i18n.t('rotation_tab.options.warrior.dps.use_overpower.label'),
			showWhen: player => isFurySpec(player) || isArmsKebabSpec(player),
			changeEmitter: player => TypedEvent.onAny([player.rotationChangeEmitter, player.talentsChangeEmitter, player.gearChangeEmitter]),
		}),
	],
};
