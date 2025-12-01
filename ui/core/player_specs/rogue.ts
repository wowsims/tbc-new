import { IconSize } from '../player_class';
import { PlayerSpec } from '../player_spec';
import { Class, Spec } from '../proto/common';
import { getSpecSiteUrl } from '../proto_utils/utils';

export class Rogue extends PlayerSpec<Spec.SpecRogue> {
	static specIndex = 0;
	static specID = Spec.SpecRogue as Spec.SpecRogue;
	static classID = Class.ClassRogue as Class.ClassRogue;
	static friendlyName = 'Rogue';
	static simLink = getSpecSiteUrl('rogue', 'dps');

	static isTankSpec = false;
	static isHealingSpec = false;
	static isRangedDpsSpec = false;
	static isMeleeDpsSpec = true;

	static canDualWield = true;

	readonly specIndex = Rogue.specIndex;
	readonly specID = Rogue.specID;
	readonly classID = Rogue.classID;
	readonly friendlyName = Rogue.friendlyName;
	readonly simLink = Rogue.simLink;

	readonly isTankSpec = Rogue.isTankSpec;
	readonly isHealingSpec = Rogue.isHealingSpec;
	readonly isRangedDpsSpec = Rogue.isRangedDpsSpec;
	readonly isMeleeDpsSpec = Rogue.isMeleeDpsSpec;

	readonly canDualWield = Rogue.canDualWield;

	static getIcon = (size: IconSize): string => {
		return `https://wow.zamimg.com/images/wow/icons/${size}/class_rogue.jpg`;
	};

	getIcon = (size: IconSize): string => {
		return Rogue.getIcon(size);
	};
}
