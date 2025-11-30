import { IconSize } from '../player_class';
import { PlayerSpec } from '../player_spec';
import { Class, Spec } from '../proto/common';
import { getSpecSiteUrl } from '../proto_utils/utils';

export class Hunter extends PlayerSpec<Spec.SpecHunter> {
	static specIndex = 0;
	static specID = Spec.SpecHunter as Spec.SpecHunter;
	static classID = Class.ClassHunter as Class.ClassHunter;
	static friendlyName = 'Hunter';
	static simLink = getSpecSiteUrl('hunter', 'hunter');

	static isTankSpec = false;
	static isHealingSpec = false;
	static isRangedDpsSpec = true;
	static isMeleeDpsSpec = false;

	static canDualWield = true;

	readonly specIndex = Hunter.specIndex;
	readonly specID = Hunter.specID;
	readonly classID = Hunter.classID;
	readonly friendlyName = Hunter.friendlyName;
	readonly simLink = Hunter.simLink;

	readonly isTankSpec = Hunter.isTankSpec;
	readonly isHealingSpec = Hunter.isHealingSpec;
	readonly isRangedDpsSpec = Hunter.isRangedDpsSpec;
	readonly isMeleeDpsSpec = Hunter.isMeleeDpsSpec;

	readonly canDualWield = Hunter.canDualWield;

	static getIcon = (size: IconSize): string => {
		return `https://wow.zamimg.com/images/wow/icons/${size}/class_hunter.jpg`;
	};

	getIcon = (size: IconSize): string => {
		return Hunter.getIcon(size);
	};
}
