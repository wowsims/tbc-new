import { IconSize } from '../player_class';
import { PlayerSpec } from '../player_spec';
import { Class, Spec } from '../proto/common';
import { getSpecSiteUrl } from '../proto_utils/utils';

export class Warlock extends PlayerSpec<Spec.SpecWarlock> {
	static specIndex = 0;
	static specID = Spec.SpecWarlock as Spec.SpecWarlock;
	static classID = Class.ClassWarlock as Class.ClassWarlock;
	static friendlyName = 'Warlock';
	static simLink = getSpecSiteUrl('warlock', 'warlock');

	static isTankSpec = false;
	static isHealingSpec = false;
	static isRangedDpsSpec = true;
	static isMeleeDpsSpec = false;

	static canDualWield = false;

	readonly specIndex = Warlock.specIndex;
	readonly specID = Warlock.specID;
	readonly classID = Warlock.classID;
	readonly friendlyName = Warlock.friendlyName;
	readonly simLink = Warlock.simLink;

	readonly isTankSpec = Warlock.isTankSpec;
	readonly isHealingSpec = Warlock.isHealingSpec;
	readonly isRangedDpsSpec = Warlock.isRangedDpsSpec;
	readonly isMeleeDpsSpec = Warlock.isMeleeDpsSpec;

	readonly canDualWield = Warlock.canDualWield;

	static getIcon = (size: IconSize): string => {
		return `https://wow.zamimg.com/images/wow/icons/${size}/class_warlock.jpg`;
	};

	getIcon = (size: IconSize): string => {
		return Warlock.getIcon(size);
	};
}
