// Configuration for spec-specific UI elements on the settings tab.
// These don't need to be in a separate file but it keeps things cleaner.
import * as InputHelpers from '../../core/components/input_helpers';
import { Spec } from '../../core/proto/common';
import { PaladinAura } from '../../core/proto/paladin';
import { ActionId } from '../../core/proto_utils/action_id';
import { TypedEvent } from '../../core/typed_event';

export const PaladinRotationConfig = {
	inputs: [
		InputHelpers.makeRotationBooleanInput<Spec.SpecRetributionPaladin>({
			fieldName: 'useExorcism',
			label: 'Use Exorcism',
			labelTooltip: 'If <b>true</b>, will use Excorism in rotation if target is undead or demon.',
			changeEmitter: player => player.rotationChangeEmitter,
			getValue: player => player.getSimpleRotation().useExorcism,
		}),
		InputHelpers.makeRotationEnumInput<Spec.SpecRetributionPaladin, number>({
			fieldName: 'consecrationRank',
			label: 'Consecration Rank',
			labelTooltip: 'Which rank of Consecration to use in the rotation. Exorcism takes priority. Select <b>Do not use</b> to disable.',
			values: [
				{ name: 'Do not use', value: 0 },
				{ name: 'Rank 1', value: 1 },
				{ name: 'Rank 2', value: 2 },
				{ name: 'Rank 3', value: 3 },
				{ name: 'Rank 4', value: 4 },
				{ name: 'Rank 5', value: 5 },
				{ name: 'Rank 6', value: 6 },
			],
			changeEmitter: player => player.rotationChangeEmitter,
			getValue: player => player.getSimpleRotation().consecrationRank,
		}),
		InputHelpers.makeRotationNumberInput<Spec.SpecRetributionPaladin>({
			fieldName: 'delayMajorCDs',
			label: 'Delay Major CDs',
			labelTooltip: 'Delays the first automatic use of major cooldowns (e.g. Bloodlust, Drums) by the specified number of seconds.',
			changeEmitter: player => player.rotationChangeEmitter,
			getValue: player => player.getSimpleRotation().delayMajorCDs,
			positive: true,
		}),
		InputHelpers.makeRotationBooleanInput<Spec.SpecRetributionPaladin>({
			fieldName: 'prepullSotC',
			label: 'Prepull Seal of the Crusader',
			labelTooltip:
				'If <b>true</b>, will use Seal of the Crusader on prepull for the target Debuff. Set this to true if you are the only paladin applying SotC. <br/><br/> If <b>false</b>, make sure to enable SotC in settings under debuffs.',
			changeEmitter: player => player.rotationChangeEmitter,
			getValue: player => player.getSimpleRotation().prepullSotC,
		}),
		InputHelpers.makeRotationEnumIconInput<Spec.SpecRetributionPaladin, PaladinAura>({
			fieldName: 'aura',
			label: 'Aura',
			labelTooltip:
				'Which paladin aura to activate in the prepull. <b>Sanctity Aura</b> requires the talent. Pick <b>None</b> to skip casting an aura.',
			values: [
				{ color: 'grey', value: PaladinAura.AuraNone, tooltip: 'None' },
				{ actionId: ActionId.fromSpellId(27149), value: PaladinAura.DevotionAura, tooltip: 'Devotion Aura' },
				{ actionId: ActionId.fromSpellId(27150), value: PaladinAura.RetributionAura, tooltip: 'Retribution Aura' },
				{ actionId: ActionId.fromSpellId(19746), value: PaladinAura.ConcentrationAura, tooltip: 'Concentration Aura' },
				{ actionId: ActionId.fromSpellId(27153), value: PaladinAura.FireResistanceAura, tooltip: 'Fire Resistance Aura' },
				{ actionId: ActionId.fromSpellId(27152), value: PaladinAura.FrostResistanceAura, tooltip: 'Frost Resistance Aura' },
				{ actionId: ActionId.fromSpellId(27151), value: PaladinAura.ShadowResistanceAura, tooltip: 'Shadow Resistance Aura' },
				{
					actionId: ActionId.fromSpellId(20218),
					value: PaladinAura.SanctityAura,
					tooltip: 'Sanctity Aura',
					showWhen: player => player.getTalents().sanctityAura,
				},
			],
			changeEmitter: player => TypedEvent.onAny([player.rotationChangeEmitter, player.talentsChangeEmitter]),
			getValue: player => player.getSimpleRotation().aura,
		}),
	],
};
