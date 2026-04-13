import * as InputHelpers from '../../core/components/input_helpers.js';
import i18n from '../../i18n/config.js';
import { Player } from '../../core/player.js';
import { Spec } from '../../core/proto/common.js';
import {
	FeralCatDruid_Rotation_FinishingMove as FinishingMove,
} from '../../core/proto/druid.js';

// Configuration for spec-specific UI elements on the settings tab.
// These don't need to be in a separate file but it keeps things cleaner.

export const CannotShredTarget = InputHelpers.makeSpecOptionsBooleanInput<Spec.SpecFeralCatDruid>({
	fieldName: 'cannotShredTarget',
	label: i18n.t('settings_tab.other.cannot_shred_target.label'),
	labelTooltip: i18n.t('settings_tab.other.cannot_shred_target.tooltip'),
});

export const FeralDruidRotationConfig = {
	inputs: [
		InputHelpers.makeRotationEnumInput<Spec.SpecFeralCatDruid, FinishingMove>({
			fieldName: 'finishingMove',
			label: i18n.t('rotation_tab.options.druid.feral_cat.finishing_move.label'),
			labelTooltip: i18n.t('rotation_tab.options.druid.feral_cat.finishing_move.tooltip'),
			values: [
				{ name: i18n.t('rotation_tab.options.druid.feral_cat.finishing_move.values.rip'), value: FinishingMove.Rip },
				{ name: i18n.t('rotation_tab.options.druid.feral_cat.finishing_move.values.bite'), value: FinishingMove.Bite },
				{ name: i18n.t('rotation_tab.options.druid.feral_cat.finishing_move.values.none'), value: FinishingMove.None },
			],
		}),
		InputHelpers.makeRotationBooleanInput<Spec.SpecFeralCatDruid>({
			fieldName: 'biteweave',
			label: i18n.t('rotation_tab.options.druid.feral_cat.biteweave.label'),
			labelTooltip: i18n.t('rotation_tab.options.druid.feral_cat.biteweave.tooltip'),
			showWhen: (player: Player<Spec.SpecFeralCatDruid>) => player.getSimpleRotation().finishingMove === FinishingMove.Rip,
		}),
		InputHelpers.makeRotationBooleanInput<Spec.SpecFeralCatDruid>({
			fieldName: 'ripweave',
			label: i18n.t('rotation_tab.options.druid.feral_cat.ripweave.label'),
			labelTooltip: i18n.t('rotation_tab.options.druid.feral_cat.ripweave.tooltip'),
			showWhen: (player: Player<Spec.SpecFeralCatDruid>) => player.getSimpleRotation().finishingMove === FinishingMove.Bite,
		}),
		InputHelpers.makeRotationEnumInput<Spec.SpecFeralCatDruid, number>({
			fieldName: 'ripMinComboPoints',
			label: i18n.t('rotation_tab.options.druid.feral_cat.rip_min_combo_points.label'),
			labelTooltip: i18n.t('rotation_tab.options.druid.feral_cat.rip_min_combo_points.tooltip'),
			values: [
				{ name: '4', value: 4 },
				{ name: '5', value: 5 },
			],
			showWhen: (player: Player<Spec.SpecFeralCatDruid>) => {
				const rot = player.getSimpleRotation();
				return rot.finishingMove === FinishingMove.Rip || (rot.ripweave && rot.finishingMove !== FinishingMove.None);
			},
		}),
		InputHelpers.makeRotationEnumInput<Spec.SpecFeralCatDruid, number>({
			fieldName: 'biteMinComboPoints',
			label: i18n.t('rotation_tab.options.druid.feral_cat.bite_min_combo_points.label'),
			labelTooltip: i18n.t('rotation_tab.options.druid.feral_cat.bite_min_combo_points.tooltip'),
			values: [
				{ name: '4', value: 4 },
				{ name: '5', value: 5 },
			],
			showWhen: (player: Player<Spec.SpecFeralCatDruid>) => {
				const rot = player.getSimpleRotation();
				return rot.finishingMove === FinishingMove.Bite || (rot.biteweave && rot.finishingMove !== FinishingMove.None);
			},
		}),
		InputHelpers.makeRotationBooleanInput<Spec.SpecFeralCatDruid>({
			fieldName: 'mangleTrick',
			label: i18n.t('rotation_tab.options.druid.feral_cat.mangle_trick.label'),
			labelTooltip: i18n.t('rotation_tab.options.druid.feral_cat.mangle_trick.tooltip'),
		}),
		InputHelpers.makeRotationBooleanInput<Spec.SpecFeralCatDruid>({
			fieldName: 'rakeTrick',
			label: i18n.t('rotation_tab.options.druid.feral_cat.rake_trick.label'),
			labelTooltip: i18n.t('rotation_tab.options.druid.feral_cat.rake_trick.tooltip'),
		}),
		InputHelpers.makeRotationBooleanInput<Spec.SpecFeralCatDruid>({
			fieldName: 'maintainFaerieFire',
			label: i18n.t('rotation_tab.options.druid.feral_cat.maintain_faerie_fire.label'),
			labelTooltip: i18n.t('rotation_tab.options.druid.feral_cat.maintain_faerie_fire.tooltip'),
		}),
	],
};
