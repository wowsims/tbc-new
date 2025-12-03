import { IconSize } from '../player_class';
import { PlayerSpec } from '../player_spec';
import { Class, Spec } from '../proto/common';
import { getSpecSiteUrl } from '../proto_utils/utils';

export class DPSWarrior extends PlayerSpec<Spec.SpecDPSWarrior> {
	static specIndex = 0;
	static specID = Spec.SpecDPSWarrior as Spec.SpecDPSWarrior;
	static classID = Class.ClassWarrior as Class.ClassWarrior;
	static friendlyName = 'DPS Warrior';
	static simLink = getSpecSiteUrl('warrior', 'dps');

	static isTankSpec = false;
	static isHealingSpec = false;
	static isRangedDpsSpec = false;
	static isMeleeDpsSpec = true;

	static canDualWield = true;

	readonly specIndex = DPSWarrior.specIndex;
	readonly specID = DPSWarrior.specID;
	readonly classID = DPSWarrior.classID;
	readonly friendlyName = DPSWarrior.friendlyName;
	readonly simLink = DPSWarrior.simLink;

	readonly isTankSpec = DPSWarrior.isTankSpec;
	readonly isHealingSpec = DPSWarrior.isHealingSpec;
	readonly isRangedDpsSpec = DPSWarrior.isRangedDpsSpec;
	readonly isMeleeDpsSpec = DPSWarrior.isMeleeDpsSpec;

	readonly canDualWield = DPSWarrior.canDualWield;

	static getIcon = (size: IconSize): string => {
		return `https://wow.zamimg.com/images/wow/icons/${size}/spell_nature_bloodlust.jpg`;
	};

	getIcon = (size: IconSize): string => {
		return DPSWarrior.getIcon(size);
	};
}

export class ProtectionWarrior extends PlayerSpec<Spec.SpecProtectionWarrior> {
	static specIndex = 2;
	static specID = Spec.SpecProtectionWarrior as Spec.SpecProtectionWarrior;
	static classID = Class.ClassWarrior as Class.ClassWarrior;
	static friendlyName = 'Protection';
	static simLink = getSpecSiteUrl('warrior', 'protection');

	static isTankSpec = true;
	static isHealingSpec = false;
	static isRangedDpsSpec = false;
	static isMeleeDpsSpec = false;

	static canDualWield = true;

	readonly specIndex = ProtectionWarrior.specIndex;
	readonly specID = ProtectionWarrior.specID;
	readonly classID = ProtectionWarrior.classID;
	readonly friendlyName = ProtectionWarrior.friendlyName;
	readonly simLink = ProtectionWarrior.simLink;

	readonly isTankSpec = ProtectionWarrior.isTankSpec;
	readonly isHealingSpec = ProtectionWarrior.isHealingSpec;
	readonly isRangedDpsSpec = ProtectionWarrior.isRangedDpsSpec;
	readonly isMeleeDpsSpec = ProtectionWarrior.isMeleeDpsSpec;

	readonly canDualWield = ProtectionWarrior.canDualWield;

	static getIcon = (size: IconSize): string => {
		return `https://wow.zamimg.com/images/wow/icons/${size}/ability_warrior_defensivestance.jpg`;
	};

	getIcon = (size: IconSize): string => {
		return ProtectionWarrior.getIcon(size);
	};
}
