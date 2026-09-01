import clsx from 'clsx';
import tippy from 'tippy.js';
import { ref } from 'tsx-vanilla';

import { translateSlotName } from '../../i18n/localization';
import { Player } from '../player';
import { ItemSlot } from '../proto/common';
import { EquippedItem } from '../proto_utils/equipped_item';
import { getEmptyGemSocketIconUrl } from '../proto_utils/gems';
import { getEmptySlotIconUrl } from './gear_picker/utils';

export const buildGearChangeIcon = (
	player: Player<any>,
	slot: ItemSlot,
	item: EquippedItem | undefined,
	previousItem: EquippedItem | undefined,
): HTMLElement => {
	const slotName = translateSlotName(slot);
	const iconRef = ref<HTMLDivElement>();
	const linkRef = ref<HTMLAnchorElement>();
	const socketsContainerRef = ref<HTMLDivElement>();
	const itemElement = (
		<div className="item-picker-root gear-change-icon">
			<div className="gear-change-icon-frame">
				<div
					ref={iconRef}
					className="item-picker-icon-wrapper"
					style={{
						backgroundImage: `url('${getEmptySlotIconUrl(slot)}')`,
					}}
				/>
				<a ref={linkRef} className="gear-change-icon-link" />
				<div ref={socketsContainerRef} className="item-picker-sockets-container"></div>
			</div>
		</div>
	) as HTMLElement;

	if (item) {
		item.asActionId()
			.fill(undefined)
			.then(filledId => {
				filledId.setBackground(iconRef.value!);
				filledId.setWowheadHref(linkRef.value!);
			});
		player.setWowheadData(item, linkRef.value!);

		const previousGems = previousItem?.gems;

		const { gems } = item;

		if (gems || previousGems) {
			const changedGems: number[] = [];
			previousItem?.gemSockets.forEach((_, socketIdx) => {
				const previousGem = previousGems ? previousGems[socketIdx] : undefined;
				const currentGem = gems ? gems[socketIdx] : undefined;
				if (previousGem?.id !== currentGem?.id) {
					changedGems.push(socketIdx);
				}
			});

			item.allSocketColors().forEach((socketColor, gemIdx) => {
				const hasChangedSocket = changedGems.includes(gemIdx);
				const socketRef = ref<HTMLDivElement>();
				const gemName = gems[gemIdx]?.name;
				socketsContainerRef.value?.appendChild(
					<div
						ref={socketRef}
						className={clsx('gem-socket-container', hasChangedSocket && 'interactive')}
						style={{
							backgroundImage: `url(${getEmptyGemSocketIconUrl(socketColor)})`,
						}}>
						{hasChangedSocket && (
							<>
								<i className={'d-block fas fa-exclamation-circle'}></i>
							</>
						)}
					</div>,
				);
				if (hasChangedSocket && gemName)
					tippy(socketRef.value!, {
						content: (
							<>
								<strong>
									{slotName} - Socket {gemIdx + 1}
								</strong>
								<br />
								{gemName}
							</>
						),
					});
			});
		}
	}

	return itemElement;
};
