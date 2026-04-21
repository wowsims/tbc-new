import { Phase } from '../../core/constants/other';
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
import PreRaidGear from './gear_sets/pre_raid.gear.json';
import P1_Realistic_6P_Gear from './gear_sets/p1_realistic_6p.gear.json';
import P1_Realistic_9P_Gear from './gear_sets/p1_realistic_9p.gear.json';
import P1_BiS_6P_Gear from './gear_sets/p1_bis_6p.gear.json';
import P1_BiS_9P_Gear from './gear_sets/p1_bis_9p.gear.json';
import P1_Alt_6P_Gear from './gear_sets/p1_alt_6p.gear.json';
import P1_Alt_9P_Gear from './gear_sets/p1_alt_9p.gear.json';
import P2_6P_Gear from './gear_sets/p2_6p.gear.json';
import P2_Alt_6P_Gear from './gear_sets/p2_alt_6p.gear.json';
import P2_9P_Gear from './gear_sets/p2_9p.gear.json';
import P2_Alt_9P_Gear from './gear_sets/p2_alt_9p.gear.json';
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

// Phase 1
export const PRE_RAID_GEARSET = PresetUtils.makePresetGear('Pre-Raid', PreRaidGear, { phase: Phase.Phase1 });
export const P1_REALISTIC_6P_GEARSET = PresetUtils.makePresetGear('P1 Realistic 6%', P1_Realistic_6P_Gear, { phase: Phase.Phase1 });
export const P1_REALISTIC_9P_GEARSET = PresetUtils.makePresetGear('P1 Realistic 9%', P1_Realistic_9P_Gear, { phase: Phase.Phase1 });
export const P1_BIS_6P_GEARSET = PresetUtils.makePresetGear('P1 BiS 6%', P1_BiS_6P_Gear, { phase: Phase.Phase1 });
export const P1_BIS_9P_GEARSET = PresetUtils.makePresetGear('P1 BiS 9%', P1_BiS_9P_Gear, { phase: Phase.Phase1 });
export const P1_ALT_6P_GEARSET = PresetUtils.makePresetGear('P1 Alt 6%', P1_Alt_6P_Gear, { phase: Phase.Phase1 });
export const P1_ALT_9P_GEARSET = PresetUtils.makePresetGear('P1 Alt 9%', P1_Alt_9P_Gear, { phase: Phase.Phase1 });

// Phase 2
export const P2_6P_GEARSET = PresetUtils.makePresetGear('P2 6%', P2_6P_Gear, { phase: Phase.Phase2 });
export const P2_ALT_6P_GEARSET = PresetUtils.makePresetGear('P2 Alt 6%', P2_Alt_6P_Gear, { phase: Phase.Phase2 });
export const P2_9P_GEARSET = PresetUtils.makePresetGear('P2 9%', P2_9P_Gear, { phase: Phase.Phase2 });
export const P2_ALT_9P_GEARSET = PresetUtils.makePresetGear('P2 Alt 9%', P2_Alt_9P_Gear, { phase: Phase.Phase2 });

// Phase 3
export const P3_GEARSET = PresetUtils.makePresetGear('P3', P3Gear, { phase: Phase.Phase3 });

// Phase 4
export const P4_GEARSET = PresetUtils.makePresetGear('P4', P4Gear, { phase: Phase.Phase4 });

// Phase 5
export const P5_GEARSET = PresetUtils.makePresetGear('P5', P5Gear, { phase: Phase.Phase5 });

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
			[Stat.StatAgility]: 1.16,
			[Stat.StatStrength]: 0.78,
			[Stat.StatAttackPower]: 0.35,
			[Stat.StatFeralAttackPower]: 0.35,
			[Stat.StatMeleeHitRating]: 1.02,
			[Stat.StatExpertiseRating]: 1.02,
			[Stat.StatMeleeCritRating]: 0.77,
			[Stat.StatMeleeHasteRating]: 0.41,
			[Stat.StatArmorPenetration]: 0.16,
            [Stat.StatPhysicalDamage]: 3.13,
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
	maintainFaerieFire: true,
});

export const SIMPLE = PresetUtils.makePresetSimpleRotation('Simple', Spec.SpecFeralCatDruid, DefaultRotation);

export const APL = PresetUtils.makePresetAPLRotation('APL', DefaultApl);
