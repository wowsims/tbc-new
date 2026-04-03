import * as InputHelpers from '../../core/components/input_helpers.js';
import { Player } from '../../core/player.js';
import { Spec } from '../../core/proto/common.js';
import {
	FeralCatDruid_Rotation_FinishingMove as FinishingMove,
} from '../../core/proto/druid.js';

// Configuration for spec-specific UI elements on the settings tab.
// These don't need to be in a separate file but it keeps things cleaner.

export const CannotShredTarget = InputHelpers.makeSpecOptionsBooleanInput<Spec.SpecFeralCatDruid>({
	fieldName: 'cannotShredTarget',
	label: 'Cannot Shred Target',
	labelTooltip: 'Prevent the sim from casting Shred (e.g. when you cannot get behind the target).',
});

export const FeralDruidRotationConfig = {
	inputs: [
		InputHelpers.makeRotationEnumInput<Spec.SpecFeralCatDruid, FinishingMove>({
			fieldName: 'finishingMove',
			label: 'Finishing Move',
			labelTooltip: 'Choose whether Rip or Ferocious Bite is the primary finisher in the rotation.',
			values: [
				{ name: 'Rip', value: FinishingMove.Rip },
				{ name: 'Ferocious Bite', value: FinishingMove.Bite },
				{ name: 'None', value: FinishingMove.None },
			],
		}),
		InputHelpers.makeRotationBooleanInput<Spec.SpecFeralCatDruid>({
			fieldName: 'biteweave',
			label: 'Enable Bite-weaving',
			labelTooltip: 'Spend Combo Points on Ferocious Bite when Rip is already active on the target.',
			showWhen: (player: Player<Spec.SpecFeralCatDruid>) => player.getSimpleRotation().finishingMove === FinishingMove.Rip,
		}),
		InputHelpers.makeRotationBooleanInput<Spec.SpecFeralCatDruid>({
			fieldName: 'ripweave',
			label: 'Enable Rip-weaving',
			labelTooltip: 'Spend Combo Points on Rip when at 52 Energy or above, even when Bite is the primary finisher.',
			showWhen: (player: Player<Spec.SpecFeralCatDruid>) => player.getSimpleRotation().finishingMove === FinishingMove.Bite,
		}),
		InputHelpers.makeRotationEnumInput<Spec.SpecFeralCatDruid, number>({
			fieldName: 'ripMinComboPoints',
			label: 'Rip CP Threshold',
			labelTooltip: 'Minimum Combo Points to accumulate before casting Rip.',
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
			label: 'Bite CP Threshold',
			labelTooltip: 'Minimum Combo Points to accumulate before casting Ferocious Bite.',
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
			label: 'Use Mangle trick',
			labelTooltip:
				'Cast Mangle rather than Shred when between 50–56 Energy with 2pT6, or 60–61 Energy without, with less than 1 second until the next Energy tick.',
		}),
		InputHelpers.makeRotationBooleanInput<Spec.SpecFeralCatDruid>({
			fieldName: 'rakeTrick',
			label: 'Use Rake/Bite tricks',
			labelTooltip:
				'Cast Rake or Ferocious Bite rather than powershifting when between 35–39 Energy without 2pT6, with more than 1 second until the next Energy tick.',
		}),
		InputHelpers.makeRotationBooleanInput<Spec.SpecFeralCatDruid>({
			fieldName: 'maintainFaerieFire',
			label: 'Maintain Faerie Fire',
			labelTooltip: 'Use Faerie Fire (Feral) whenever it is not active on the target.',
		}),
	],
};
