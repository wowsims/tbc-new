import { Player } from '../../player';
import { Class, ConsumesSpec, Drums, ItemSlot, Profession, Spec, Stat, TristateEffect } from '../../proto/common';
import { Consumable } from '../../proto/db';
import { ActionId } from '../../proto_utils/action_id';
import { EventID, TypedEvent } from '../../typed_event';
import * as InputHelpers from '../input_helpers';
import { IconEnumValueConfig } from '../pickers/icon_enum_picker';
import { ActionInputConfig, ItemStatOption } from './stat_options';
import i18n from '../../../i18n/config.js';
import { makeBooleanConsumeInput } from '../icon_inputs';
import { CURRENT_PHASE, Phase } from '../../constants/other';

export interface ConsumableInputConfig<T> extends ActionInputConfig<T> {
	value: T;
}

export interface ConsumableStatOption<T> extends ItemStatOption<T> {
	config: ConsumableInputConfig<T>;
}

export interface ConsumeInputFactoryArgs<T extends number> {
	consumesFieldName: keyof ConsumesSpec;
	// Additional callback if logic besides syncing consumes is required
	onSet?: (eventactionId: EventID, player: Player<any>, newValue: T) => void;
	showWhen?: (player: Player<any>) => boolean;
	enableWhen?: (player: Player<any>) => boolean;
	changedEvent?: (player: Player<any>) => void;
}

function makeConsumeInputFactory<T extends number, SpecType extends Spec>(
	args: ConsumeInputFactoryArgs<T>,
): (options: ConsumableStatOption<T>[], tooltip?: string) => InputHelpers.TypedIconEnumPickerConfig<Player<SpecType>, T> {
	return (options: ConsumableStatOption<T>[], tooltip?: string) => {
		const valueOptions = options.map(
			option =>
				({
					actionId: option.config.actionId,
					value: option.config.value,
					showWhen: (player: Player<SpecType>) =>
						(!option.config.showWhen || option.config.showWhen(player)) && (option.config.faction || player.getFaction()) == player.getFaction(),
				}) satisfies IconEnumValueConfig<Player<SpecType>, T>,
		);
		return {
			type: 'iconEnum',
			tooltip: tooltip,
			numColumns: options.length > 5 ? 2 : 1,
			values: [{ value: 0, iconUrl: '', tooltip: i18n.t('common.none') } as unknown as IconEnumValueConfig<Player<SpecType>, T>].concat(valueOptions),
			equals: (a: T, b: T) => a == b,
			zeroValue: 0 as T,
			changedEvent: (player: Player<any>) =>
				(args.changedEvent && args.changedEvent(player)) ||
				TypedEvent.onAny([player.consumesChangeEmitter, player.gearChangeEmitter, player.professionChangeEmitter]),
			showWhen: (player: Player<any>) => (!args.showWhen || args.showWhen(player)) && valueOptions.some(option => option.showWhen?.(player)),
			enableWhen: args.enableWhen,
			getValue: (player: Player<any>) => player.getConsumes()[args.consumesFieldName] as T,
			setValue: (eventID: EventID, player: Player<any>, newValue: number) => {
				const newConsumes = player.getConsumes();
				if (newConsumes[args.consumesFieldName] === newValue) {
					return;
				}

				(newConsumes[args.consumesFieldName] as number) = newValue;
				TypedEvent.freezeAllAndDo(() => {
					player.setConsumes(eventID, newConsumes);
					if (args.onSet) {
						args.onSet(eventID, player, newValue as T);
					}
				});
			},
		};
	};
}

///////////////////////////////////////////////////////////////////////////
//                                 CONJURED
///////////////////////////////////////////////////////////////////////////

export const ConjuredDarkRune = {
	actionId: ActionId.fromItemId(12662),
	value: 12662,
};
export const ConjuredHealthstone = {
	actionId: ActionId.fromItemId(22105),
	value: 22105,
};
export const ConjuredRogueThistleTea = {
	actionId: ActionId.fromItemId(7676),
	value: 7676,
	showWhen: <SpecType extends Spec>(player: Player<SpecType>) => player.getClass() == Class.ClassRogue,
};
export const ConjuredFlameCap = {
	actionId: ActionId.fromItemId(22788),
	value: 22788,
};
export const ConjuredCrackedPowerCore = {
	actionId: ActionId.fromItemId(23334),
	value: 23334,
};
export const ChippedCrackedPowerCore = {
	actionId: ActionId.fromItemId(23381),
	value: 23381,
};
export const ConjuredNightmareSeed = {
	actionId: ActionId.fromItemId(22797),
	value: 22797,
	showWhen: <SpecType extends Spec>(player: Player<SpecType>) => player.getPlayerSpec().isTankSpec,
};

export const CONJURED_CONFIG = [
	{ config: ConjuredRogueThistleTea, stats: [] },
	{ config: ConjuredHealthstone, stats: [Stat.StatStamina] },
	{ config: ConjuredNightmareSeed, stats: [Stat.StatStamina] },
	{ config: ConjuredDarkRune, stats: [Stat.StatIntellect] },
	{ config: ConjuredFlameCap, stats: [] },
	{ config: ConjuredCrackedPowerCore, stats: [Stat.StatSpellDamage] },
	{ config: ChippedCrackedPowerCore, stats: [Stat.StatSpellDamage] },
] as ConsumableStatOption<number>[];

export const makeConjuredInput = makeConsumeInputFactory({ consumesFieldName: 'conjuredId' });

///////////////////////////////////////////////////////////////////////////
//                               ENGINEERING
///////////////////////////////////////////////////////////////////////////

export const EzThroDynamiteTwo = {
	actionId: ActionId.fromItemId(18588),
	value: 18588,
};

export const CrystalCharge = {
	actionId: ActionId.fromItemId(11566),
	value: 15239,
};

export const AdamantiteGrenade = {
	actionId: ActionId.fromItemId(23737),
	value: 30217,
	showWhen: (player: Player<any>) => player.hasProfession(Profession.Engineering),
};

export const FelIronBomb = {
	actionId: ActionId.fromItemId(23736),
	value: 30216,
	showWhen: (player: Player<any>) => player.hasProfession(Profession.Engineering),
};

export const GnomishFlameTurrent = {
	actionId: ActionId.fromItemId(23841),
	value: 30526,
	showWhen: (player: Player<any>) => player.hasProfession(Profession.Engineering),
};

export const EXPLOSIVE_CONFIG = [
	{ config: AdamantiteGrenade, stats: [] },
	{ config: FelIronBomb, stats: [] },
	{ config: CrystalCharge, stats: [] },
	{ config: EzThroDynamiteTwo, stats: [] },
	// { config: GnomishFlameTurrent, stats: [] }, Excluding this thing for now because it's weird and I don't like it
] as ConsumableStatOption<number>[];
export const makeExplosivesInput = makeConsumeInputFactory({ consumesFieldName: 'explosiveId' });

export const GoblinSapper = makeBooleanConsumeInput({
	actionId: () => ActionId.fromItemId(10646),
	fieldName: 'goblinSapper',
	showWhen: (player: Player<any>) => player.hasProfession(Profession.Engineering),
});

export const SuperSapper = makeBooleanConsumeInput({
	actionId: () => ActionId.fromItemId(23827),
	fieldName: 'superSapper',
	showWhen: (player: Player<any>) => player.hasProfession(Profession.Engineering),
});

///////////////////////////////////////////////////////////////////////////
//                               WEAPON IMBUES
///////////////////////////////////////////////////////////////////////////

// Oils
export const ManaOil = {
	actionId: ActionId.fromItemId(20748),
	value: 25123,
};
export const BrilWizardOil = {
	actionId: ActionId.fromItemId(20749),
	value: 25122,
};
export const SupWizardOil = {
	actionId: ActionId.fromItemId(22522),
	value: 28017,
};
// Stones
export const AdamantiteSharpeningMH = {
	actionId: ActionId.fromItemId(23529),
	value: 29453,
	showWhen: (player: Player<any>) => player.getGear().hasSharpMHWeapon(),
};
export const AdamantiteWeightMH = {
	actionId: ActionId.fromItemId(28421),
	value: 34340,
	showWhen: (player: Player<any>) => player.getGear().hasBluntMHWeapon(),
};
export const ConsecratedSharpeningStoneMH = {
	actionId: ActionId.fromItemId(23122),
	value: 28891,
	showWhen: (player: Player<any>) => player.getGear().hasMHWeapon(),
};

export const AdamantiteSharpeningOH = {
	actionId: ActionId.fromItemId(23529),
	value: 29453,
	showWhen: (player: Player<any>) => player.getGear().hasSharpOHWeapon(),
};
export const AdamantiteWeightOH = {
	actionId: ActionId.fromItemId(28421),
	value: 34340,
	showWhen: (player: Player<any>) => player.getGear().hasBluntOHWeapon(),
};
export const ConsecratedSharpeningStoneOH = {
	actionId: ActionId.fromItemId(23122),
	value: 28891,
	showWhen: (player: Player<any>) => player.getGear().hasOHWeapon(),
};

// Rogue Poisons
export const RogueInstantPoison = {
	actionId: ActionId.fromItemId(21927),
	value: 26891,
	showWhen: (player: Player<any>) => player.getClass() == Class.ClassRogue,
};
export const RogueDeadlyPoison = {
	actionId: ActionId.fromItemId(22054),
	value: 27186,
	showWhen: (player: Player<any>) => player.getClass() == Class.ClassRogue,
};
export const RogueWoundPoison = {
	actionId: ActionId.fromItemId(22055),
	value: 27188,
	showWhen: (player: Player<any>) => player.getClass() == Class.ClassRogue,
};
// Shaman Imbues
export const ShamanImbueWindfury = {
	actionId: ActionId.fromSpellId(25505),
	value: 25505,
	showWhen: (player: Player<any>) => player.getClass() == Class.ClassShaman,
};
export const ShamanImbueFlametongue = {
	actionId: ActionId.fromSpellId(25489),
	value: 25489,
	showWhen: (player: Player<any>) => player.getClass() == Class.ClassShaman,
};

export const ShamanImbueFrostbrand = {
	actionId: ActionId.fromSpellId(25500),
	value: 25500,
	showWhen: (player: Player<any>) => player.getClass() == Class.ClassShaman,
};

export const ShamanImbueRockbiter = {
	actionId: ActionId.fromSpellId(25485),
	value: 25485,
	showWhen: (player: Player<any>) => player.getClass() == Class.ClassShaman,
};

export const IMBUE_CONFIG_MH = [
	{ config: ManaOil, stats: [Stat.StatHealingPower] },
	{ config: BrilWizardOil, stats: [Stat.StatSpellDamage] },
	{ config: SupWizardOil, stats: [Stat.StatSpellDamage] },
	{ config: AdamantiteSharpeningMH, stats: [Stat.StatAttackPower] },
	{ config: AdamantiteWeightMH, stats: [Stat.StatAttackPower] },
	{ config: ConsecratedSharpeningStoneMH, stats: [Stat.StatAttackPower] },
	{ config: RogueInstantPoison, stats: [] },
	{ config: RogueDeadlyPoison, stats: [] },
	{ config: RogueWoundPoison, stats: [] },
	{ config: ShamanImbueRockbiter, stats: [] },
	{ config: ShamanImbueFrostbrand, stats: [] },
	{ config: ShamanImbueFlametongue, stats: [] },
	{ config: ShamanImbueWindfury, stats: [] },
] as ConsumableStatOption<number>[];

export const IMBUE_CONFIG_OH = [
	{ config: ManaOil, stats: [Stat.StatHealingPower] },
	{ config: BrilWizardOil, stats: [Stat.StatSpellDamage] },
	{ config: SupWizardOil, stats: [Stat.StatSpellDamage] },
	{ config: AdamantiteSharpeningOH, stats: [Stat.StatAttackPower] },
	{ config: AdamantiteWeightOH, stats: [Stat.StatAttackPower] },
	{ config: ConsecratedSharpeningStoneOH, stats: [Stat.StatAttackPower] },
	{ config: RogueInstantPoison, stats: [] },
	{ config: RogueDeadlyPoison, stats: [] },
	{ config: RogueWoundPoison, stats: [] },
	{ config: ShamanImbueRockbiter, stats: [] },
	{ config: ShamanImbueFrostbrand, stats: [] },
	{ config: ShamanImbueFlametongue, stats: [] },
	{ config: ShamanImbueWindfury, stats: [] },
] as ConsumableStatOption<number>[];

// Specs that doesn't use Windfury and should always show imbues.
const specsWithoutWindfury = [Spec.SpecFeralBearDruid, Spec.SpecFeralCatDruid];

export const makeMHImbueInput = makeConsumeInputFactory({
	consumesFieldName: 'mhImbueId',
	showWhen: (player: Player<any>) =>
		specsWithoutWindfury.includes(player.getSpec()) ||
		!player.getParty() ||
		player.getParty()!.getBuffs().windfuryTotem == TristateEffect.TristateEffectMissing,
	changedEvent: (player: Player<any>) => TypedEvent.onAny([player.getParty()!.changeEmitter]),
});
export const makeOHImbueInput = makeConsumeInputFactory({
	consumesFieldName: 'ohImbueId',
	showWhen: (player: Player<any>) => player.getGear().getEquippedItem(ItemSlot.ItemSlotOffHand)?.item.weaponSpeed !== undefined,
});

///////////////////////////////////////////////////////////////////////////
//                               	DRUMS
///////////////////////////////////////////////////////////////////////////

export const DrumsBattle = {
	...(CURRENT_PHASE >= Phase.Phase4
		? { actionId: ActionId.fromItemId(185848), value: Drums.GreaterDrumsOfBattle }
		: {
				actionId: ActionId.fromItemId(29529),
				value: Drums.LesserDrumsOfBattle,
			}),
};

export const DrumsRestoration = {
	...(CURRENT_PHASE >= Phase.Phase4
		? { actionId: ActionId.fromItemId(185850), value: Drums.GreaterDrumsOfRestoration }
		: {
				actionId: ActionId.fromItemId(29531),
				value: Drums.LesserDrumsOfRestoration,
			}),
};

export const DrumsWar = {
	...(CURRENT_PHASE >= Phase.Phase4
		? { actionId: ActionId.fromItemId(185852), value: Drums.GreaterDrumsOfWar }
		: {
				actionId: ActionId.fromItemId(29528),
				value: Drums.LesserDrumsOfWar,
			}),
};

export const DRUMS_CONFIG = [
	{ config: DrumsBattle, stats: [] },
	{ config: DrumsRestoration, stats: [Stat.StatMana] },
	{ config: DrumsWar, stats: [Stat.StatAttackPower, Stat.StatSpellDamage] },
] as ConsumableStatOption<number>[];

export const makeDrumsInput = makeConsumeInputFactory({
	consumesFieldName: 'drumsId',
	showWhen: (player: Player<any>) => player.hasProfession(Profession.Leatherworking),
});

///////////////////////////////////////////////////////////////////////////
//                                   PET
///////////////////////////////////////////////////////////////////////////

export const PetScrollAgi = makeBooleanConsumeInput({
	actionId: () => ActionId.fromItemId(27498),
	fieldName: 'petScrollAgi',
	showWhen: (player: Player<any>) => [Spec.SpecHunter, Spec.SpecWarlock, Spec.SpecPriest].includes(player.getSpec()),
});

export const PetScrollStr = makeBooleanConsumeInput({
	actionId: () => ActionId.fromItemId(27503),
	fieldName: 'petScrollStr',
	showWhen: (player: Player<any>) => [Spec.SpecHunter, Spec.SpecWarlock, Spec.SpecPriest].includes(player.getSpec()),
});

///////////////////////////////////////////////////////////////////////////
//                                 SCROLLS
///////////////////////////////////////////////////////////////////////////

export const ScrollAgi = makeBooleanConsumeInput({
	actionId: () => ActionId.fromItemId(27498),
	fieldName: 'scrollAgi',
	showWhen: (player: Player<any>) => player.getEpWeights().getStat(Stat.StatAgility) > 0,
});

export const ScrollStr = makeBooleanConsumeInput({
	actionId: () => ActionId.fromItemId(27503),
	fieldName: 'scrollStr',
	showWhen: (player: Player<any>) => player.getEpWeights().getStat(Stat.StatStrength) > 0,
});

export const ScrollInt = makeBooleanConsumeInput({
	actionId: () => ActionId.fromItemId(27499),
	fieldName: 'scrollInt',
	showWhen: (player: Player<any>) => player.getEpWeights().getStat(Stat.StatIntellect) > 0,
});

export const ScrollSpi = makeBooleanConsumeInput({
	actionId: () => ActionId.fromItemId(27501),
	fieldName: 'scrollSpi',
	showWhen: (player: Player<any>) => player.getEpWeights().getStat(Stat.StatSpirit) > 0,
});

export const ScrollArm = makeBooleanConsumeInput({
	actionId: () => ActionId.fromItemId(27500),
	fieldName: 'scrollArm',
	showWhen: (player: Player<any>) => player.getEpWeights().getStat(Stat.StatArmor) > 0,
});

///////////////////////////////////////////////////////////////////////////
//                            MISCELLANEOUS
///////////////////////////////////////////////////////////////////////////

export const NightmareSeed = makeBooleanConsumeInput({
	actionId: () => ActionId.fromItemId(22797),
	fieldName: 'nightmareSeed',
	showWhen: (player: Player<any>) => player.getPlayerSpec().isTankSpec,
});

///////////////////////////////////////////////////////////////////////////

export interface ConsumableInputOptions {
	consumesFieldName: keyof ConsumesSpec;
	setValue?: (eventID: EventID, player: Player<any>, newValue: number) => void;
	showWhen?: (player: Player<any>) => boolean;
}

export function makeConsumableInput(
	items: Consumable[],
	options: ConsumableInputOptions,
	tooltip?: string,
): InputHelpers.TypedIconEnumPickerConfig<Player<any>, number> {
	const valueOptions = items.map(item => ({
		value: item.id,
		iconUrl: item.icon,
		actionId: ActionId.fromItemId(item.id),
		tooltip: item.name,
	}));
	return {
		type: 'iconEnum',
		tooltip: tooltip,
		numColumns: items.length > 10 ? 3 : items.length > 5 ? 2 : 1,
		values: [{ value: 0, iconUrl: '', tooltip: i18n.t('common.none') }].concat(valueOptions),
		equals: (a: number, b: number) => a === b,
		zeroValue: 0,
		changedEvent: (player: Player<any>) => player.consumesChangeEmitter,
		getValue: (player: Player<any>) => player.getConsumes()[options.consumesFieldName] as number,
		showWhen: (player: Player<any>) => !!valueOptions.length && (!options.showWhen || options.showWhen(player)),
		setValue: (eventID: EventID, player: Player<any>, newValue: number) => {
			if (options.setValue) {
				options.setValue(eventID, player, newValue);
			}

			const newConsumes = {
				...player.getConsumes(),
				[options.consumesFieldName]: newValue,
			};

			if (options.consumesFieldName === 'flaskId') {
				newConsumes.guardianElixirId = 0;
				newConsumes.battleElixirId = 0;
			}

			if ((options.consumesFieldName === 'battleElixirId' || options.consumesFieldName === 'guardianElixirId') && newValue != 0) {
				newConsumes.flaskId = 0;
			}
			player.setConsumes(eventID, newConsumes);
		},
	};
}
