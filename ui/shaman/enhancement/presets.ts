import * as Mechanics from '../../core/constants/mechanics';
import * as PresetUtils from '../../core/preset_utils.js';
import { Class, ConsumesSpec, Debuffs, Profession, PseudoStat, Race, RaidBuffs, Stat } from '../../core/proto/common.js';
import {
	EnhancementShaman_Options as EnhancementShamanOptions,
	FeleAutocastSettings,
	ShamanImbue,
	ShamanShield,
	ShamanSyncType,
} from '../../core/proto/shaman.js';
import { SavedTalents } from '../../core/proto/ui.js';
import { Stats } from '../../core/proto_utils/stats';
import { defaultRaidBuffMajorDamageCooldowns } from '../../core/proto_utils/utils';
import DefaultApl from './apls/default.apl.json';
import P3Apl from './apls/p3.apl.json';
import P1Gear from './gear_sets/p1.gear.json';
import P2Gear from './gear_sets/p2.gear.json';
import P3Gear from './gear_sets/p3.gear.json';
import PreraidGear from './gear_sets/preraid.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

export const PRERAID_PRESET = PresetUtils.makePresetGear('Pre-raid', PreraidGear);

export const P1_PRESET = PresetUtils.makePresetGear('P1', P1Gear);
export const P2_PRESET = PresetUtils.makePresetGear('P2', P2Gear);
export const P3_PRESET = PresetUtils.makePresetGear('P3 (WiP)', P3Gear);

export const ROTATION_PRESET_DEFAULT = PresetUtils.makePresetAPLRotation('Default', DefaultApl);
export const ROTATION_PRESET_P3 = PresetUtils.makePresetAPLRotation('P3 (WiP)', P3Apl);

// Preset options for EP weights
export const P1_EP_PRESET = PresetUtils.makePresetEpWeights(
	'Default',
	Stats.fromMap(
		{
			[Stat.StatIntellect]: 0.04,
			[Stat.StatAgility]: 1.0,
			[Stat.StatAttackPower]: 0.4,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 0.88,
			[PseudoStat.PseudoStatOffHandDps]: 0.76,
			[PseudoStat.PseudoStatSpellHitPercent]: 0.57 * Mechanics.SPELL_HIT_RATING_PER_HIT_PERCENT,
			[PseudoStat.PseudoStatMeleeHitPercent]: 0.39 * Mechanics.PHYSICAL_HIT_RATING_PER_HIT_PERCENT,
		},
	),
);

export const P3_EP_PRESET = PresetUtils.makePresetEpWeights(
	'P3 (WiP)',
	Stats.fromMap(
		{
			[Stat.StatIntellect]: 0.04,
			[Stat.StatAgility]: 1.0,
			[Stat.StatAttackPower]: 0.4,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 0.67,
			[PseudoStat.PseudoStatOffHandDps]: 0.52,
			[PseudoStat.PseudoStatSpellHitPercent]: 0.32 * Mechanics.SPELL_HIT_RATING_PER_HIT_PERCENT,
			[PseudoStat.PseudoStatMeleeHitPercent]: 0.565 * Mechanics.PHYSICAL_HIT_RATING_PER_HIT_PERCENT,
		},
	),
);

// Default talents. Uses the wowhead calculator format, make the talents on
// https://wowhead.com/tbc/talent-calc and copy the numbers in the url.
export const StandardTalents = {
	name: 'Standard',
	data: SavedTalents.create({
		talentsString: '313133',
	}),
};

export const P3Talents = {
	name: 'P3 (WiP)',
	data: SavedTalents.create({
		talentsString: '313122',
	}),
};

export const DefaultOptions = EnhancementShamanOptions.create({
	classOptions: {
		shield: ShamanShield.LightningShield,
		imbueMh: ShamanImbue.WindfuryWeapon,
		imbueMhSwap: ShamanImbue.WindfuryWeapon,
		feleAutocast: FeleAutocastSettings.create({
			autocastFireblast: true,
			autocastFirenova: true,
			autocastImmolate: true,
			autocastEmpower: false,
			noImmolateWfunleash: false,
			noImmolateDuration: 0,
		}),
	},
	imbueOh: ShamanImbue.FlametongueWeapon,
	imbueOhSwap: ShamanImbue.FlametongueWeapon,
	syncType: ShamanSyncType.Auto,
});

export const OtherDefaults = {
	distanceFromTarget: 5,
	profession1: Profession.Engineering,
	profession2: Profession.Herbalism,
	race: Race.RaceTroll,
};

export const DefaultConsumables = ConsumesSpec.create({
	flaskId: 76084, // Flask of Spring Blossoms
	foodId: 74648, // Sea Mist Rice Noodles
	potId: 76089, // Virmen's Bite
	prepotId: 76089, // Virmen's Bite
});

export const DefaultRaidBuffs = RaidBuffs.create({
	...defaultRaidBuffMajorDamageCooldowns(Class.ClassShaman),

});

export const DefaultDebuffs = Debuffs.create({

});
