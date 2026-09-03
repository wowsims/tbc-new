import { GemColor, ItemSlot } from '../../proto/common';
import { UIGem as Gem } from '../../proto/ui';
import { ActionId } from '../../proto_utils/action_id';
import { getEmptyGemSocketIconUrl } from '../../proto_utils/gems';

const emptySlotIcons: Record<ItemSlot, string> = {
	[ItemSlot.ItemSlotHead]: '/tbc/assets/item_slots/head.jpg',
	[ItemSlot.ItemSlotNeck]: '/tbc/assets/item_slots/neck.jpg',
	[ItemSlot.ItemSlotShoulder]: '/tbc/assets/item_slots/shoulders.jpg',
	[ItemSlot.ItemSlotBack]: '/tbc/assets/item_slots/shirt.jpg',
	[ItemSlot.ItemSlotChest]: '/tbc/assets/item_slots/chest.jpg',
	[ItemSlot.ItemSlotWrist]: '/tbc/assets/item_slots/wrists.jpg',
	[ItemSlot.ItemSlotHands]: '/tbc/assets/item_slots/hands.jpg',
	[ItemSlot.ItemSlotWaist]: '/tbc/assets/item_slots/waist.jpg',
	[ItemSlot.ItemSlotLegs]: '/tbc/assets/item_slots/legs.jpg',
	[ItemSlot.ItemSlotFeet]: '/tbc/assets/item_slots/feet.jpg',
	[ItemSlot.ItemSlotFinger1]: '/tbc/assets/item_slots/finger.jpg',
	[ItemSlot.ItemSlotFinger2]: '/tbc/assets/item_slots/finger.jpg',
	[ItemSlot.ItemSlotTrinket1]: '/tbc/assets/item_slots/trinket.jpg',
	[ItemSlot.ItemSlotTrinket2]: '/tbc/assets/item_slots/trinket.jpg',
	[ItemSlot.ItemSlotMainHand]: '/tbc/assets/item_slots/mainhand.jpg',
	[ItemSlot.ItemSlotOffHand]: '/tbc/assets/item_slots/offhand.jpg',
	[ItemSlot.ItemSlotRanged]: '/tbc/assets/item_slots/ranged.jpg',
};
export function getEmptySlotIconUrl(slot: ItemSlot): string {
	return emptySlotIcons[slot];
}

export const createNameDescriptionLabel = (nameDesc: string) => {
	return <small className="heroic-label">({nameDesc})</small>;
};

// Points the gem icon inside a container built by createGemContainer at `gem`, or hides it when
// the socket is empty. Resolves with the filled ActionId so callers can wire up extra state.
export const setGemInContainer = async (container: HTMLElement, gem: Gem | null, emptySocketIconUrl: string): Promise<ActionId | null> => {
	const gemIconElem = container.querySelector<HTMLImageElement>('.gem-icon')!;
	if (!gem) {
		gemIconElem.classList.add('hide');
		gemIconElem.src = emptySocketIconUrl;
		return null;
	}

	gemIconElem.classList.remove('hide');
	const filledId = await ActionId.fromItemId(gem.id).fill();
	gemIconElem.src = filledId.iconUrl;
	return filledId;
};

export const createGemContainer = (socketColor: GemColor, gem: Gem | null, index: number) => {
	const gemContainer = (
		<a className="gem-socket-container" href="javascript:void(0)" dataset={{ socketIdx: index }}>
			<img className="gem-icon hide" />
			<img className="socket-icon" src={getEmptyGemSocketIconUrl(socketColor)} />
		</a>
	) as HTMLAnchorElement;

	setGemInContainer(gemContainer, gem, getEmptyGemSocketIconUrl(socketColor)).then(filledId => filledId?.setWowheadHref(gemContainer));
	return gemContainer;
};
