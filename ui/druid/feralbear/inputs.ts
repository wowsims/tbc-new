import * as InputHelpers from '../../core/components/input_helpers.js';
import { Player } from '../../core/player.js';
import { Spec } from '../../core/proto/common.js';
import { FeralBearDruid_Rotation_SwipeUsage as SwipeUsage } from '../../core/proto/druid.js';
import i18n from '../../i18n/config.js';

export const StartingRage = InputHelpers.makeSpecOptionsNumberInput<Spec.SpecFeralBearDruid>({
	fieldName: 'startingRage',
	label: i18n.t('settings_tab.other.starting_rage.label'),
	labelTooltip: i18n.t('settings_tab.other.starting_rage.tooltip'),
});

export const FeralBearRotationConfig = {
	inputs: [
		InputHelpers.makeRotationNumberInput<Spec.SpecFeralBearDruid>({
			fieldName: 'maulRageThreshold',
			label: i18n.t('rotation_tab.options.druid.feral_bear.maul_rage_threshold.label'),
			labelTooltip: i18n.t('rotation_tab.options.druid.feral_bear.maul_rage_threshold.tooltip'),
		}),
		InputHelpers.makeRotationEnumInput<Spec.SpecFeralBearDruid, SwipeUsage>({
			fieldName: 'swipeUsage',
			label: i18n.t('rotation_tab.options.druid.feral_bear.swipe_usage.label'),
			labelTooltip: i18n.t('rotation_tab.options.druid.feral_bear.swipe_usage.tooltip'),
			values: [
				{ name: i18n.t('rotation_tab.options.druid.feral_bear.swipe_usage.values.never'), value: SwipeUsage.SwipeUsage_Never },
				{ name: i18n.t('rotation_tab.options.druid.feral_bear.swipe_usage.values.with_enough_ap'), value: SwipeUsage.SwipeUsage_WithEnoughAP },
				{ name: i18n.t('rotation_tab.options.druid.feral_bear.swipe_usage.values.spam'), value: SwipeUsage.SwipeUsage_Spam },
			],
		}),
		InputHelpers.makeRotationNumberInput<Spec.SpecFeralBearDruid>({
			fieldName: 'swipeApThreshold',
			label: i18n.t('rotation_tab.options.druid.feral_bear.swipe_ap_threshold.label'),
			labelTooltip: i18n.t('rotation_tab.options.druid.feral_bear.swipe_ap_threshold.tooltip'),
			showWhen: (player: Player<Spec.SpecFeralBearDruid>) => player.getSimpleRotation().swipeUsage === SwipeUsage.SwipeUsage_WithEnoughAP,
		}),
		InputHelpers.makeRotationBooleanInput<Spec.SpecFeralBearDruid>({
			fieldName: 'maintainDemoralizingRoar',
			label: i18n.t('rotation_tab.options.druid.feral_bear.maintain_demo_roar.label'),
			labelTooltip: i18n.t('rotation_tab.options.druid.feral_bear.maintain_demo_roar.tooltip'),
		}),
		InputHelpers.makeRotationBooleanInput<Spec.SpecFeralBearDruid>({
			fieldName: 'maintainFaerieFire',
			label: i18n.t('rotation_tab.options.druid.feral_bear.maintain_faerie_fire.label'),
			labelTooltip: i18n.t('rotation_tab.options.druid.feral_bear.maintain_faerie_fire.tooltip'),
		}),
	],
};
