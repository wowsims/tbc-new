import * as InputHelpers from '../../core/components/input_helpers';
import { HunterSpecs } from '../../core/proto_utils/utils';
import { makePetTypeInputConfig } from '../../core/talents/hunter_pet';
import { RotationType, Spec } from '../../core/proto/common';
import i18n from '../../i18n/config.js';
// import { makePetTypeInputConfig } from '../core/talents/hunter_pet';

// // Configuration for class-specific UI elements on the settings tab.
// // These don't need to be in a separate file but it keeps things cleaner.

// export const PetTypeInput = <SpecType extends HunterSpecs>() => makePetTypeInputConfig<SpecType>();
export const PetTypeInput = <SpecType extends HunterSpecs>() => makePetTypeInputConfig<SpecType>();

export const PetUptime = <SpecType extends HunterSpecs>() =>
	InputHelpers.makeClassOptionsNumberInput<SpecType>({
		fieldName: 'petUptime',
		label: i18n.t('settings_tab.other.pet_uptime.label'),
		labelTooltip: i18n.t('settings_tab.other.pet_uptime.tooltip'),
		percent: true,
	});

export const GlaiveTossChance = <SpecType extends HunterSpecs>() =>
	InputHelpers.makeClassOptionsNumberInput<SpecType>({
		fieldName: 'glaiveTossSuccess',
		label: i18n.t('settings_tab.other.glaive_toss_chance.label'),
		labelTooltip: i18n.t('settings_tab.other.glaive_toss_chance.tooltip'),
		percent: true,
	});

export const MMRotationConfig = {
	inputs: [
		InputHelpers.makeRotationEnumInput<Spec.SpecHunter, RotationType>({
			fieldName: 'type',
			label: i18n.t('rotation_tab.common.rotation_type.label'),
			values: [
				{ name: i18n.t('rotation_tab.common.rotation_type.single_target'), value: RotationType.SingleTarget },
				{ name: i18n.t('rotation_tab.common.rotation_type.aoe'), value: RotationType.Aoe },
			],
		}),
	]};
