import * as PresetUtils from '../../core/preset_utils';
import { ConsumesSpec, Profession, Race } from '../../core/proto/common';
import {
	FeralCatDruid_Options as FeralDruidOptions,
} from '../../core/proto/druid';
import { SavedTalents } from '../../core/proto/ui';
import P1Gear from './gear_sets/p1.gear.json';
import P2Gear from './gear_sets/p2.gear.json';
import P3Gear from './gear_sets/p3.gear.json';
import P4Gear from './gear_sets/p4.gear.json';
import P5Gear from './gear_sets/p5.gear.json';

// Default talents. Uses the wowhead calculator format, make the talents on
// https://wowhead.com/tbc/talent-calc and copy the numbers in the url.
export const StandardTalents = {
	name: 'Standard',
	data: SavedTalents.create({
		talentsString: '-503032132322105301251-05503301',
	}),
};

export const MonocatTalents = {
	name: 'Monocat',
	data: SavedTalents.create({
		talentsString: '-553002132322105301051-05503301',
	}),
};

export const P1_GEARSET = PresetUtils.makePresetGear('P1', P1Gear);
export const P2_GEARSET = PresetUtils.makePresetGear('P2', P2Gear);
export const P3_GEARSET = PresetUtils.makePresetGear('P3', P3Gear);
export const P4_GEARSET = PresetUtils.makePresetGear('P4', P4Gear);
export const P5_GEARSET = PresetUtils.makePresetGear('P5', P5Gear);

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
