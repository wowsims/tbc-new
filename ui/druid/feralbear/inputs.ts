import * as InputHelpers from '../../core/components/input_helpers.js';
import { Player } from '../../core/player.js';
import { Class, Spec } from '../../core/proto/common.js';
import i18n from '../../i18n/config.js';

// Configuration for spec-specific UI elements on the settings tab.
// These don't need to be in a separate file but it keeps things cleaner.
export const SymbiosisSelection = InputHelpers.makeSpecOptionsEnumInput<Spec.SpecFeralBearDruid, Player<Spec.SpecFeralBearDruid>>({
	fieldName: 'symbiosisTarget',
	label: 'Symbiosis Target',
	labelTooltip: 'Class from which to receive a Symbiosis spell',
	values: [
	],
});

export const GuardianDruidRotationConfig = {
	inputs: [
		InputHelpers.makeRotationBooleanInput<Spec.SpecFeralBearDruid>({
			fieldName: 'maintainFaerieFire',
			label: i18n.t('rotation_tab.options.druid.guardian.maintain_faerie_fire.label'),
			labelTooltip: i18n.t('rotation_tab.options.druid.guardian.maintain_faerie_fire.tooltip'),
		}),
		InputHelpers.makeRotationBooleanInput<Spec.SpecFeralBearDruid>({
			fieldName: 'maintainDemoralizingRoar',
			label: i18n.t('rotation_tab.options.druid.guardian.maintain_demo_roar.label'),
			labelTooltip: i18n.t('rotation_tab.options.druid.guardian.maintain_demo_roar.tooltip'),
		}),
		InputHelpers.makeRotationNumberInput<Spec.SpecFeralBearDruid>({
			fieldName: 'demoTime',
			label: i18n.t('rotation_tab.options.druid.guardian.demo_roar_refresh_leeway.label'),
			labelTooltip: i18n.t('rotation_tab.options.druid.guardian.demo_roar_refresh_leeway.tooltip'),
			showWhen: (player: Player<Spec.SpecFeralBearDruid>) => player.getSimpleRotation().maintainDemoralizingRoar,
		}),
		InputHelpers.makeRotationNumberInput<Spec.SpecFeralBearDruid>({
			fieldName: 'pulverizeTime',
			label: i18n.t('rotation_tab.options.druid.guardian.pulverize_refresh_leeway.label'),
			labelTooltip: i18n.t('rotation_tab.options.druid.guardian.pulverize_refresh_leeway.tooltip'),
		}),
		// InputHelpers.makeRotationBooleanInput<Spec.SpecGuardianDruid>({
		// 	fieldName: 'prepullStampede',
		// 	label: 'Assume pre-pull Stampede',
		// 	labelTooltip: 'Activate Stampede Haste buff at the start of each pull. Models the effects of initiating the pull with Feral Charge.',
		// 	showWhen: (player: Player<Spec.SpecGuardianDruid>) =>
		// 		player.getTalents().stampede > 0,
		// 	changeEmitter: (player: Player<Spec.SpecGuardianDruid>) => TypedEvent.onAny([player.rotationChangeEmitter, player.talentsChangeEmitter]),
		// }),
	],
};
