import { ConsumesSpec, Profession, Race } from '../../core/proto/common';
import {
	FeralCatDruid_Options as FeralDruidOptions,
} from '../../core/proto/druid';
import { SavedTalents } from '../../core/proto/ui';

// Default talents. Uses the wowhead calculator format, make the talents on
// https://wowhead.com/tbc/talent-calc and copy the numbers in the url.
export const StandardTalents = {
	name: 'Standard',
	data: SavedTalents.create({
		talentsString: '',
	}),
};

export const DefaultOptions = FeralDruidOptions.create({
	assumeBleedActive: false,
});

export const DefaultConsumables = ConsumesSpec.create({});

export const OtherDefaults = {
	distanceFromTarget: 0,
	profession1: Profession.ProfessionUnknown,
	profession2: Profession.ProfessionUnknown,
	race: Race.RaceTauren,
};
