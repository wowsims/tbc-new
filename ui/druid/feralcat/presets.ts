import * as PresetUtils from '../../core/preset_utils';
import { ConsumesSpec, Drums, Profession, Race, Spec, Stat } from '../../core/proto/common';
import {
	FeralCatDruid_Options as FeralDruidOptions,
	FeralCatDruid_Rotation as FeralCatDruidRotation,
	FeralCatDruid_Rotation_FinishingMove as FinishingMove,
} from '../../core/proto/druid';
import { SavedTalents } from '../../core/proto/ui';
import { Stats } from '../../core/proto_utils/stats';
import DefaultApl from './apls/default.apl.json';
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

export const DefaultOptions = FeralDruidOptions.create({});

export const DefaultConsumables = ConsumesSpec.create({
	potId: 22838,              // Haste Potion
	battleElixirId: 22831,     // Elixir of Major Agility
	guardianElixirId: 32067,   // Elixir of Draenic Wisdom
	foodId: 27664,             // Grilled Mudfish (+20 Agility)
	mhImbueId: 34340,          // Adamantite Weightstone
	conjuredId: 12662,         // Demonic Rune
	drumsId: Drums.GreaterDrumsOfBattle,
	superSapper: true,
	goblinSapper: true,
	scrollAgi: true,
	scrollStr: true,
});

export const OtherDefaults = {
	distanceFromTarget: 0,
	profession1: Profession.Engineering,
	profession2: Profession.Enchanting,
	race: Race.RaceNightElf,
};


export const P1_EP_PRESET = PresetUtils.makePresetEpWeights(
	'P1',
	Stats.fromMap(
		{
			[Stat.StatAgility]: 1.17,
			[Stat.StatStrength]: 0.79,
			[Stat.StatAttackPower]: 0.35,
			[Stat.StatFeralAttackPower]: 0.35,
			[Stat.StatMeleeHitRating]: 0.99,
			[Stat.StatExpertiseRating]: 1.00,
			[Stat.StatMeleeCritRating]: 0.78,
			[Stat.StatMeleeHasteRating]: 0.45,
			[Stat.StatArmorPenetration]: 0.17,
		},
		{},
	),
);

export const DefaultRotation = FeralCatDruidRotation.create({
	finishingMove: FinishingMove.Rip,
	biteweave: true,
	ripMinComboPoints: 5,
	biteMinComboPoints: 5,
	mangleTrick: true,
	maintainFaerieFire: false,
});

export const SIMPLE = PresetUtils.makePresetSimpleRotation('Simple', Spec.SpecFeralCatDruid, DefaultRotation);

export const APL = PresetUtils.makePresetAPLRotation('APL', DefaultApl);
