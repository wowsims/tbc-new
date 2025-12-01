import * as InputHelpers from '../../core/components/input_helpers.js';
import { Player } from '../../core/player.js';
import { APLRotation_Type } from '../../core/proto/apl.js';
import { Spec } from '../../core/proto/common.js';
import { FeralCatDruid_Rotation_AplType as AplType, FeralCatDruid_Rotation_HotwStrategy as HotwType } from '../../core/proto/druid.js';
import { TypedEvent } from '../../core/typed_event.js';
import i18n from '../../i18n/config.js';

// Configuration for spec-specific UI elements on the settings tab.
// These don't need to be in a separate file but it keeps things cleaner.

export const AssumeBleedActive = InputHelpers.makeSpecOptionsBooleanInput<Spec.SpecFeralCatDruid>({
	fieldName: 'assumeBleedActive',
	label: i18n.t('settings_tab.other.assume_bleed_active.label'),
	labelTooltip: i18n.t('settings_tab.other.assume_bleed_active.tooltip'),
	extraCssClasses: ['within-raid-sim-hide'],
});

export const CannotShredTarget = InputHelpers.makeSpecOptionsBooleanInput<Spec.SpecFeralCatDruid>({
	fieldName: 'cannotShredTarget',
	label: i18n.t('settings_tab.other.cannot_shred_target.label'),
	labelTooltip: i18n.t('settings_tab.other.cannot_shred_target.tooltip'),
});

function ShouldShowAdvParamST(player: Player<Spec.SpecFeralCatDruid>): boolean {
	const rot = player.getSimpleRotation();
	return rot.manualParams && rot.rotationType == AplType.SingleTarget;
}

function ShouldShowAdvParamAoe(player: Player<Spec.SpecFeralCatDruid>): boolean {
	const rot = player.getSimpleRotation();
	return rot.manualParams && rot.rotationType == AplType.Aoe;
}

export const FeralDruidRotationConfig = {
	inputs: [
		InputHelpers.makeRotationEnumInput<Spec.SpecFeralCatDruid, AplType>({
			fieldName: 'rotationType',
			label: i18n.t('rotation_tab.options.druid.feral.target_type.label'),
			values: [
				{ name: i18n.t('rotation_tab.options.druid.feral.target_type.single_target'), value: AplType.SingleTarget },
				{ name: i18n.t('rotation_tab.options.druid.feral.target_type.aoe'), value: AplType.Aoe },
			],
		}),
		InputHelpers.makeRotationBooleanInput<Spec.SpecFeralCatDruid>({
			fieldName: 'bearWeave',
			label: i18n.t('rotation_tab.options.druid.feral.bear_weave.label'),
			labelTooltip: i18n.t('rotation_tab.options.druid.feral.bear_weave.tooltip'),
		}),
		InputHelpers.makeRotationBooleanInput<Spec.SpecFeralCatDruid>({
			fieldName: 'snekWeave',
			label: i18n.t('rotation_tab.options.druid.feral.snek_weave.label'),
			labelTooltip: i18n.t('rotation_tab.options.druid.feral.snek_weave.tooltip'),
			showWhen: (player: Player<Spec.SpecFeralCatDruid>) => player.getSimpleRotation().bearWeave,
		}),
		// InputHelpers.makeRotationBooleanInput<Spec.SpecFeralCatDruid>({
		// 	fieldName: 'useNs',
		// 	label: i18n.t('rotation_tab.options.druid.feral.use_ns.label'),
		// 	labelTooltip: i18n.t('rotation_tab.options.druid.feral.use_ns.tooltip'),
		// 	showWhen: (player: Player<Spec.SpecFeralCatDruid>) => player.getTalents().dreamOfCenarius,
		// 	changeEmitter: (player: Player<Spec.SpecFeralCatDruid>) => TypedEvent.onAny([player.rotationChangeEmitter, player.talentsChangeEmitter]),
		// }),
		// InputHelpers.makeRotationEnumInput<Spec.SpecFeralCatDruid, HotwType>({
		// 	fieldName: 'hotwStrategy',
		// 	label: i18n.t('rotation_tab.options.druid.feral.hotw_strategy.label'),
		// 	labelTooltip: i18n.t('rotation_tab.options.druid.feral.hotw_strategy.tooltip'),
		// 	values: [
		// 		{ name: i18n.t('rotation_tab.options.druid.feral.hotw_strategy.values.passives_only'), value: HotwType.PassivesOnly },
		// 		{ name: i18n.t('rotation_tab.options.druid.feral.hotw_strategy.values.enhanced_bear_weaving'), value: HotwType.Bear },
		// 		{ name: i18n.t('rotation_tab.options.druid.feral.hotw_strategy.values.wrath_weaving'), value: HotwType.Wrath },
		// 	],
		// 	showWhen: (player: Player<Spec.SpecFeralCatDruid>) => player.getTalents().heartOfTheWild,
		// 	changeEmitter: (player: Player<Spec.SpecFeralCatDruid>) => TypedEvent.onAny([player.rotationChangeEmitter, player.talentsChangeEmitter]),
		// }),
		InputHelpers.makeRotationBooleanInput<Spec.SpecFeralCatDruid>({
			fieldName: 'allowAoeBerserk',
			label: i18n.t('rotation_tab.options.druid.feral.allow_aoe_berserk.label'),
			labelTooltip: i18n.t('rotation_tab.options.druid.feral.allow_aoe_berserk.tooltip'),
			showWhen: (player: Player<Spec.SpecFeralCatDruid>) => player.getSimpleRotation().rotationType == AplType.Aoe,
		}),
		InputHelpers.makeRotationBooleanInput<Spec.SpecFeralCatDruid>({
			fieldName: 'manualParams',
			label: i18n.t('rotation_tab.options.druid.feral.manual_params.label'),
			labelTooltip: i18n.t('rotation_tab.options.druid.feral.manual_params.tooltip'),
			showWhen: (player: Player<Spec.SpecFeralCatDruid>) => player.getSimpleRotation().rotationType == AplType.SingleTarget,
		}),
		InputHelpers.makeRotationNumberInput<Spec.SpecFeralCatDruid>({
			fieldName: 'minRoarOffset',
			label: i18n.t('rotation_tab.options.druid.feral.roar_offset.label'),
			labelTooltip: i18n.t('rotation_tab.options.druid.feral.roar_offset.tooltip'),
			showWhen: ShouldShowAdvParamST,
		}),
		InputHelpers.makeRotationNumberInput<Spec.SpecFeralCatDruid>({
			fieldName: 'ripLeeway',
			label: i18n.t('rotation_tab.options.druid.feral.rip_leeway.label'),
			labelTooltip: i18n.t('rotation_tab.options.druid.feral.rip_leeway.tooltip'),
			showWhen: ShouldShowAdvParamST,
		}),
		InputHelpers.makeRotationBooleanInput<Spec.SpecFeralCatDruid>({
			fieldName: 'useBite',
			label: i18n.t('rotation_tab.options.druid.feral.bite_during_rotation.label'),
			labelTooltip: i18n.t('rotation_tab.options.druid.feral.bite_during_rotation.tooltip'),
			showWhen: ShouldShowAdvParamST,
		}),
		InputHelpers.makeRotationNumberInput<Spec.SpecFeralCatDruid>({
			fieldName: 'biteTime',
			label: i18n.t('rotation_tab.options.druid.feral.bite_time.label'),
			labelTooltip: i18n.t('rotation_tab.options.druid.feral.bite_time.tooltip'),
			showWhen: (player: Player<Spec.SpecFeralCatDruid>) =>
				ShouldShowAdvParamST(player) && player.getSimpleRotation().useBite,
		}),
		InputHelpers.makeRotationNumberInput<Spec.SpecFeralCatDruid>({
			fieldName: 'berserkBiteTime',
			label: i18n.t('rotation_tab.options.druid.feral.berserk_bite_time.label'),
			labelTooltip: i18n.t('rotation_tab.options.druid.feral.berserk_bite_time.tooltip'),
			showWhen: (player: Player<Spec.SpecFeralCatDruid>) =>
				ShouldShowAdvParamST(player) && player.getSimpleRotation().useBite,
		}),
	],
};
