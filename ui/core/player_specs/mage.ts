import { IconSize } from '../player_class';
import { PlayerSpec } from '../player_spec';
import { Class, Spec } from '../proto/common';
import { getSpecSiteUrl } from '../proto_utils/utils';

export class Mage extends PlayerSpec<Spec.SpecMage> {
	static specIndex = 0;
	static specID = Spec.SpecMage as Spec.SpecMage;
	static classID = Class.ClassMage as Class.ClassMage;
	static friendlyName = 'Mage';
	static simLink = getSpecSiteUrl('mage', 'mage');

	static isTankSpec = false;
	static isHealingSpec = false;
	static isRangedDpsSpec = true;
	static isMeleeDpsSpec = false;

	static canDualWield = false;

	readonly specIndex = Mage.specIndex;
	readonly specID = Mage.specID;
	readonly classID = Mage.classID;
	readonly friendlyName = Mage.friendlyName;
	readonly simLink = Mage.simLink;

	readonly isTankSpec = Mage.isTankSpec;
	readonly isHealingSpec = Mage.isHealingSpec;
	readonly isRangedDpsSpec = Mage.isRangedDpsSpec;
	readonly isMeleeDpsSpec = Mage.isMeleeDpsSpec;

	readonly canDualWield = Mage.canDualWield;

	static getIcon = (size: IconSize): string => {
		return `https://wow.zamimg.com/images/wow/icons/${size}/class_mage.jpg`;
	};

	getIcon = (size: IconSize): string => {
		return Mage.getIcon(size);
	};
}
