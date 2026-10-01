import { ItemSlot } from '@generated/proto/common';
import { BULK_SIM_ITEM_SLOT_TO_ITEM_SLOT_PAIRS, BulkSimItemSlot } from '@sim/bulk/utils';
import type { Player } from '@sim/player/player';
import type { EquippedItem } from '@sim/proto/equipped_item';
import { loadStoredBulkSettings, seedBulkSettings, storeBulkSettings } from '@sim/settings/bulk_settings';
import type { IndividualSimHost } from '@sim/sim_host';
import { subscribeBulkChange, subscribePlayerField, subscribeSimField } from '@sim/state/subscriptions';

import { addBulkItems, loadEquippedBulkItems, seedBulkPickerGroups } from './items';
import { refreshBulkCombinations } from './run';
import {
	type BulkFrozenSlot,
	createBulkSettingsProto,
	setBulkFrozenItem,
	setBulkFrozenWeaponSlot,
	setBulkStatConstraints,
	setBulkUseLegacyBulkSim,
	setBulkWeaponTypeFilter,
} from './settings';

const equippedItemForFrozenSlot = (player: Player<any>, bulkSlot: BulkFrozenSlot, itemSlot: number): EquippedItem | null => {
	const slots = BULK_SIM_ITEM_SLOT_TO_ITEM_SLOT_PAIRS.get(bulkSlot);
	if (!slots?.includes(itemSlot)) {
		return null;
	}

	return player.getGear().getEquippedItem(itemSlot) ?? null;
};

const loadSettings = (player: Player<any>) => {
	const settings = loadStoredBulkSettings(player);
	if (settings != null) {
		addBulkItems(player, settings.items, true);
		setBulkFrozenItem(player, BulkSimItemSlot.ItemSlotFinger, equippedItemForFrozenSlot(player, BulkSimItemSlot.ItemSlotFinger, settings.freezeRingSlot));
		setBulkFrozenItem(
			player,
			BulkSimItemSlot.ItemSlotTrinket,
			equippedItemForFrozenSlot(player, BulkSimItemSlot.ItemSlotTrinket, settings.freezeTrinketSlot),
		);
		setBulkFrozenWeaponSlot(player, settings.freezeWeaponSlot);
		setBulkWeaponTypeFilter(player, ItemSlot.ItemSlotMainHand, settings.freezeMainhandWeaponSlots);
		setBulkWeaponTypeFilter(player, ItemSlot.ItemSlotOffHand, settings.freezeOffhandWeaponSlots);
		setBulkUseLegacyBulkSim(player, settings.useLegacyBulkSim);
		setBulkStatConstraints(player, settings.statConstraints);
	}
};

/**
 * The batch belongs to the page, not to its tab: the gear tab adds items to it and the autosave
 * has to keep them, so none of this can wait for `BulkTabBody` to mount. The host calls it once.
 */
export const initBulk = (host: IndividualSimHost<any>) => {
	const { sim, player } = host;
	seedBulkSettings(player);
	seedBulkPickerGroups(player);

	sim.waitForInit().then(() => {
		loadSettings(player);

		subscribePlayerField(player, 'gear')(() => loadEquippedBulkItems(player));
		subscribeBulkChange(player)(() => storeBulkSettings(player, createBulkSettingsProto(player)));
		subscribeBulkChange(player)(() => void refreshBulkCombinations(host));
		subscribeSimField(sim, 'iterations')(() => void refreshBulkCombinations(host));

		loadEquippedBulkItems(player);
		void refreshBulkCombinations(host);
	});
};
