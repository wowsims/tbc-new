import { BulkSettings, BulkStatConstraint } from '@generated/proto/api';
import { ItemSlot, ItemSpec, WeaponType } from '@generated/proto/common';
import { BULK_SIM_ITEM_SLOT_TO_ITEM_SLOT_PAIRS, BulkSimItemSlot } from '@sim/bulk/utils';
import type { Player } from '@sim/player/player';
import type { EquippedItem } from '@sim/proto/equipped_item';
import { bulkState, patchBulkState } from '@sim/settings/bulk_settings';

import { frozenItemSlot } from './picker_groups';

export type BulkWeaponSlot = ItemSlot.ItemSlotMainHand | ItemSlot.ItemSlotOffHand;
export type BulkFrozenSlot = BulkSimItemSlot.ItemSlotFinger | BulkSimItemSlot.ItemSlotTrinket;

export const bulkFrozenItemSlot = (player: Player<any>, bulkSlot: BulkFrozenSlot): ItemSlot | undefined => {
	const slots = BULK_SIM_ITEM_SLOT_TO_ITEM_SLOT_PAIRS.get(bulkSlot);
	return frozenItemSlot(player.getGear(), slots, bulkState(player).frozenItems.get(bulkSlot)) ?? undefined;
};

export const createBulkSettingsProto = (player: Player<any>): BulkSettings => {
	const current = bulkState(player);
	return BulkSettings.create({
		items: current.items.flatMap(spec => (spec ? [ItemSpec.clone(spec)] : [])),
		useLegacyBulkSim: current.useLegacyBulkSim,
		statConstraints: current.statConstraints.map(constraint => BulkStatConstraint.clone(constraint)),
		iterationsPerCombo: player.sim.getIterations(),
		freezeRingSlot: bulkFrozenItemSlot(player, BulkSimItemSlot.ItemSlotFinger),
		freezeTrinketSlot: bulkFrozenItemSlot(player, BulkSimItemSlot.ItemSlotTrinket),
		freezeWeaponSlot: current.frozenWeaponSlot,
		freezeMainhandWeaponSlots: current.weaponTypeFilters.get(ItemSlot.ItemSlotMainHand)?.slice(),
		freezeOffhandWeaponSlots: current.weaponTypeFilters.get(ItemSlot.ItemSlotOffHand)?.slice(),
	});
};

export const setBulkUseLegacyBulkSim = (player: Player<any>, newValue: boolean) => patchBulkState(player, { useLegacyBulkSim: newValue }, ['settings']);

// A change that alters nothing must not emit: every settings emit refreshes the combination
// count, which disables Simulate while it loads, so a no-op emit on blur would swallow a click.
export const setBulkStatConstraints = (player: Player<any>, constraints: ReadonlyArray<BulkStatConstraint>) => {
	const current = bulkState(player).statConstraints;
	if (current.length === constraints.length && current.every((constraint, idx) => BulkStatConstraint.equals(constraint, constraints[idx]))) {
		return;
	}

	patchBulkState(player, { statConstraints: constraints.map(constraint => BulkStatConstraint.clone(constraint)) }, ['settings']);
};

export const setBulkFrozenItem = (player: Player<any>, bulkSlot: BulkFrozenSlot, item: EquippedItem | null) => {
	const frozenItems = bulkState(player).frozenItems;
	if (item === frozenItems.get(bulkSlot)) {
		return;
	}

	patchBulkState(player, { frozenItems: new Map(frozenItems).set(bulkSlot, item) }, ['settings']);
};

export const setBulkWeaponTypeFilter = (player: Player<any>, slot: BulkWeaponSlot, newFilter: WeaponType[], shouldEmit = true): boolean => {
	const weaponTypeFilters = bulkState(player).weaponTypeFilters;
	const currentFilter = weaponTypeFilters.get(slot)!;
	const hasChanged = currentFilter.length !== newFilter.length || currentFilter.some((weaponType, idx) => weaponType !== newFilter[idx]);

	if (!hasChanged) {
		return false;
	}

	patchBulkState(player, { weaponTypeFilters: new Map(weaponTypeFilters).set(slot, newFilter) }, shouldEmit ? ['settings'] : []);
	return true;
};

export const setBulkFrozenWeaponSlot = (player: Player<any>, itemSlot: number | null): boolean => {
	const newSlot = [ItemSlot.ItemSlotMainHand, ItemSlot.ItemSlotOffHand].includes(itemSlot ?? -1) ? (itemSlot as BulkWeaponSlot) : undefined;
	const filtersChanged = newSlot !== undefined && setBulkWeaponTypeFilter(player, newSlot, [], false);

	if (newSlot === bulkState(player).frozenWeaponSlot && !filtersChanged) {
		return false;
	}

	patchBulkState(player, { frozenWeaponSlot: newSlot }, ['settings']);
	return true;
};
