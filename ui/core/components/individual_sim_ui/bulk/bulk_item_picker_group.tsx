import i18n from '../../../../i18n/config';
import { IndividualSimUI } from '../../../individual_sim_ui';
import { HandType } from '../../../proto/common';
import { EquippedItem } from '../../../proto_utils/equipped_item';
import { ContentBlock } from '../../content_block';
import Toast from '../../toast';
import { BulkTab } from '../bulk_tab';
import BulkItemPicker from './bulk_item_picker';
import { translateBulkSlotName } from '../../../../i18n/localization';
import { getBulkSlotI18nKey } from '../../../../i18n/entity_mapping';
import { BulkSimItemSlot } from './constants_auto_gen';

export default class BulkItemPickerGroup extends ContentBlock {
	readonly simUI: IndividualSimUI<any>;
	readonly bulkUI: BulkTab;
	readonly bulkSlot: BulkSimItemSlot;

	readonly pickers: Map<number, BulkItemPicker> = new Map();

	constructor(parent: HTMLElement, simUI: IndividualSimUI<any>, bulkUI: BulkTab, bulkSlot: BulkSimItemSlot) {
		const slotName = translateBulkSlotName(bulkSlot);
		super(parent, 'bulk-item-picker-group-root', { header: { title: slotName } });
		const slotKey = getBulkSlotI18nKey(bulkSlot);
		this.rootElem.classList.add(`gear-group-${slotKey.replace(/_/g, '-')}`);
		this.simUI = simUI;
		this.bulkUI = bulkUI;
		this.bulkSlot = bulkSlot;

		this.addEmptyElement();
	}

	has(idx: number) {
		return !!this.pickers.get(idx);
	}

	// Whether both of this bulk slot's physical slots can hold the same item at once, which is
	// what makes a same-item combo (two identical rings, one weapon in each hand) a valid input.
	// Mirrors canStackTwoCopies in the backend.
	private canStackTwoCopies(item: EquippedItem): boolean {
		if (item._item.unique || item._item.limitCategory != 0) return false;
		switch (this.bulkSlot) {
			case BulkSimItemSlot.ItemSlotFinger:
			case BulkSimItemSlot.ItemSlotTrinket:
				return true;
			case BulkSimItemSlot.ItemSlotHandWeapon:
				return item._item.handType == HandType.HandTypeOneHand;
			default:
				return false;
		}
	}

	// True when the slot already lists as many copies of this exact item as can be worn.
	// Items sharing a limit category are NOT rejected: only one of them can be worn at a time,
	// but listing several is how you compare them, and the candidate generator drops the
	// conflicting pairings itself.
	private isDuplicateOfExisting(item: EquippedItem): boolean {
		const pickers = Array.from(this.pickers.values());
		const maxCopies = this.canStackTwoCopies(item) ? 2 : 1;
		return pickers.filter(picker => picker.item.id === item.id).length >= maxCopies;
	}

	// An equipped item is already part of every candidate, so a batch entry for the same item is
	// redundant and renders as a phantom duplicate. Mirrors the backend, which drops the
	// user-added copy in initSelectedItems - except where equipped + added is what makes a
	// same-item-in-both-slots combo possible.
	private evictRedundantAddedCopies(equippedItem: EquippedItem) {
		if (this.canStackTwoCopies(equippedItem)) {
			return;
		}
		for (const [idx, picker] of Array.from(this.pickers.entries())) {
			if (idx >= 0 && picker.item.id === equippedItem.id) {
				this.bulkUI.removeItemByIndex(idx, true);
			}
		}
	}

	// Returns false if the item was rejected, so callers can undo the entry they pushed onto
	// the batch list; a stale one sims and counts toward combinations with no picker to remove it.
	add(idx: number, item: EquippedItem, silent = false): boolean {
		// Equipped pickers (idx < 0) report what is worn rather than offering a choice, so they
		// always render - the guard must never hide one. They evict redundant batch entries instead.
		if (idx < 0) {
			this.evictRedundantAddedCopies(item);
		}

		// After eviction: a group emptied by it re-rendered its "no items" placeholder.
		if (!this.pickers.size) this.bodyElement.replaceChildren();

		if (idx >= 0 && this.isDuplicateOfExisting(item)) {
			if (!silent)
				new Toast({
					delay: 1000,
					variant: 'error',
					body: <>{i18n.t('bulk_tab.search.item_unique', { itemName: item._item.name })}</>,
				});
			return false;
		}

		if (this.pickers.has(idx)) {
			const picker = this.pickers.get(idx);
			picker!.dispose();
			this.pickers.delete(idx);
		}

		this.pickers.set(idx, new BulkItemPicker(this.bodyElement, this.simUI, this.bulkUI, item, this.bulkSlot, idx));

		if (!silent)
			new Toast({
				delay: 1000,
				variant: 'success',
				body: <>{i18n.t('bulk_tab.search.item_added', { itemName: item._item.name })}</>,
			});

		return true;
	}

	update(idx: number, newItem: EquippedItem) {
		const picker = this.pickers.get(idx);
		if (!picker) {
			new Toast({
				variant: 'error',
				body: i18n.t('bulk_tab.picker.failed_update'),
			});
			return;
		}

		picker.setItem(newItem);
	}

	remove(idx: number, silent = false) {
		const picker = this.pickers.get(idx);
		if (!picker) {
			if (!silent)
				new Toast({
					variant: 'error',
					body: i18n.t('bulk_tab.picker.failed_remove'),
				});
			return;
		}

		picker.dispose();
		this.pickers.delete(idx);

		if (!this.pickers.size) this.addEmptyElement();

		if (!silent)
			new Toast({
				delay: 1000,
				variant: 'success',
				body: <>{i18n.t('bulk_tab.search.item_removed', { itemName: picker.item._item.name })}</>,
			});
	}

	private addEmptyElement() {
		this.bodyElement.appendChild(<span>{i18n.t('bulk_tab.picker.no_items')}</span>);
	}
}
