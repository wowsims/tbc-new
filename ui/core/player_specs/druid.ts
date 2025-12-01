import { IconSize } from '../player_class';
import { PlayerSpec } from '../player_spec';
import { Class, Spec } from '../proto/common';
import { getSpecSiteUrl } from '../proto_utils/utils';

export class BalanceDruid extends PlayerSpec<Spec.SpecBalanceDruid> {
	static specIndex = 0;
	static specID = Spec.SpecBalanceDruid as Spec.SpecBalanceDruid;
	static classID = Class.ClassDruid as Class.ClassDruid;
	static friendlyName = 'Balance';
	static simLink = getSpecSiteUrl('druid', 'balance');

	static isTankSpec = false;
	static isHealingSpec = false;
	static isRangedDpsSpec = true;
	static isMeleeDpsSpec = false;

	static canDualWield = false;

	readonly specIndex = BalanceDruid.specIndex;
	readonly specID = BalanceDruid.specID;
	readonly classID = BalanceDruid.classID;
	readonly friendlyName = BalanceDruid.friendlyName;
	readonly simLink = BalanceDruid.simLink;

	readonly isTankSpec = BalanceDruid.isTankSpec;
	readonly isHealingSpec = BalanceDruid.isHealingSpec;
	readonly isRangedDpsSpec = BalanceDruid.isRangedDpsSpec;
	readonly isMeleeDpsSpec = BalanceDruid.isMeleeDpsSpec;

	readonly canDualWield = BalanceDruid.canDualWield;

	static getIcon = (size: IconSize): string => {
		return `https://wow.zamimg.com/images/wow/icons/${size}/spell_nature_starfall.jpg`;
	};

	getIcon = (size: IconSize): string => {
		return BalanceDruid.getIcon(size);
	};
}

export class FeralCatDruid extends PlayerSpec<Spec.SpecFeralCatDruid> {
	static specIndex = 1;
	static specID = Spec.SpecFeralCatDruid as Spec.SpecFeralCatDruid;
	static classID = Class.ClassDruid as Class.ClassDruid;
	static friendlyName = 'Feral Cat';
	static simLink = getSpecSiteUrl('druid', 'feralcat');

	static isTankSpec = false;
	static isHealingSpec = false;
	static isRangedDpsSpec = false;
	static isMeleeDpsSpec = true;

	static canDualWield = false;

	readonly specIndex = FeralCatDruid.specIndex;
	readonly specID = FeralCatDruid.specID;
	readonly classID = FeralCatDruid.classID;
	readonly friendlyName = FeralCatDruid.friendlyName;
	readonly simLink = FeralCatDruid.simLink;

	readonly isTankSpec = FeralCatDruid.isTankSpec;
	readonly isHealingSpec = FeralCatDruid.isHealingSpec;
	readonly isRangedDpsSpec = FeralCatDruid.isRangedDpsSpec;
	readonly isMeleeDpsSpec = FeralCatDruid.isMeleeDpsSpec;

	readonly canDualWield = FeralCatDruid.canDualWield;

	static getIcon = (size: IconSize): string => {
		return `https://wow.zamimg.com/images/wow/icons/${size}/ability_druid_catform.jpg`;
	};

	getIcon = (size: IconSize): string => {
		return FeralCatDruid.getIcon(size);
	};
}

export class FeralBearDruid extends PlayerSpec<Spec.SpecFeralBearDruid> {
	static specIndex = 2;
	static specID = Spec.SpecFeralBearDruid as Spec.SpecFeralBearDruid;
	static classID = Class.ClassDruid as Class.ClassDruid;
	static friendlyName = 'Feral Bear';
	static simLink = getSpecSiteUrl('druid', 'feralbear');

	static isTankSpec = true;
	static isHealingSpec = false;
	static isRangedDpsSpec = false;
	static isMeleeDpsSpec = true;

	static canDualWield = false;

	readonly specIndex = FeralBearDruid.specIndex;
	readonly specID = FeralBearDruid.specID;
	readonly classID = FeralBearDruid.classID;
	readonly friendlyName = FeralBearDruid.friendlyName;
	readonly simLink = FeralBearDruid.simLink;

	readonly isTankSpec = FeralBearDruid.isTankSpec;
	readonly isHealingSpec = FeralBearDruid.isHealingSpec;
	readonly isRangedDpsSpec = FeralBearDruid.isRangedDpsSpec;
	readonly isMeleeDpsSpec = FeralBearDruid.isMeleeDpsSpec;

	readonly canDualWield = FeralBearDruid.canDualWield;

	static getIcon = (size: IconSize): string => {
		return `https://wow.zamimg.com/images/wow/icons/${size}/ability_racial_bearform.jpg`;
	};

	getIcon = (size: IconSize): string => {
		return FeralBearDruid.getIcon(size);
	};
}

export class RestorationDruid extends PlayerSpec<Spec.SpecRestorationDruid> {
	static specIndex = 3;
	static specID = Spec.SpecRestorationDruid as Spec.SpecRestorationDruid;
	static classID = Class.ClassDruid as Class.ClassDruid;
	static friendlyName = 'Restoration';
	static simLink = getSpecSiteUrl('druid', 'restoration');

	static isTankSpec = false;
	static isHealingSpec = true;
	static isRangedDpsSpec = false;
	static isMeleeDpsSpec = false;

	static canDualWield = false;

	readonly specIndex = RestorationDruid.specIndex;
	readonly specID = RestorationDruid.specID;
	readonly classID = RestorationDruid.classID;
	readonly friendlyName = RestorationDruid.friendlyName;
	readonly simLink = RestorationDruid.simLink;

	readonly isTankSpec = RestorationDruid.isTankSpec;
	readonly isHealingSpec = RestorationDruid.isHealingSpec;
	readonly isRangedDpsSpec = RestorationDruid.isRangedDpsSpec;
	readonly isMeleeDpsSpec = RestorationDruid.isMeleeDpsSpec;

	readonly canDualWield = RestorationDruid.canDualWield;

	static getIcon = (size: IconSize): string => {
		return `https://wow.zamimg.com/images/wow/icons/${size}/spell_nature_healingtouch.jpg`;
	};

	getIcon = (size: IconSize): string => {
		return RestorationDruid.getIcon(size);
	};
}
