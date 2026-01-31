import * as PresetUtils from '../../core/preset_utils';
import { Debuffs, PseudoStat, RaidBuffs, Stat, ConsumesSpec, TristateEffect, PartyBuffs, IndividualBuffs } from '../../core/proto/common';
import { defaultRaidBuffMajorDamageCooldowns } from '../../core/proto_utils/utils';
import { Stats } from '../../core/proto_utils/stats';
import { SavedTalents } from '../../core/proto/ui';
import { MageArmor, Mage_Options as MageOptions } from '../../core/proto/mage';
import BlankAPL from './apls/blank.apl.json'
import BlankGear from './gear_sets/blank.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

export const BLANK_APL = PresetUtils.makePresetAPLRotation('Blank', BlankAPL)

export const BLANK_GEARSET = PresetUtils.makePresetGear('Blank', BlankGear);

// Preset options for EP weights
export const P1_EP_PRESET = PresetUtils.makePresetEpWeights(
	'A',
	Stats.fromMap(
		{
			[Stat.StatAgility]: 1.0,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 1.43,
			[PseudoStat.PseudoStatOffHandDps]: 0.26,
		},
	),
);

export const Talents = {
	name: 'A',
	data: SavedTalents.create({
		talentsString: '',
	}),
};

export const DefaultOptions = MageOptions.create({
	classOptions: {
	},
});

export const OtherDefaults = {
	distanceFromTarget: 20,
};

export const DefaultConsumables = ConsumesSpec.create({
	guardianElixirId:32067, // Elixir of Draenic Wisdom
	battleElixirId:28103, // Adept's Elixir
	foodId: 27657, // Blackened Basilisk
	mhImbueId: 20749, // Brilliant Wizard Oil
	prepotId: 22839, // Destruction Potion
	potId: 22839, // Destruction Potion
	conjuredId: 12662, // Demonic Rune
});

export const DefaultRaidBuffs = RaidBuffs.create({
	bloodlust: true,
	divineSpirit: 2,
	arcaneBrilliance: true,
	giftOfTheWild: 2,
	powerWordFortitude: 2,
	shadowProtection: true,
});

export const DefaultPartyBuffs = PartyBuffs.create({
})

export const DefaultIndividualBuffs = IndividualBuffs.create({
	blessingOfKings: true,
	blessingOfWisdom: 2,
})

export const DefaultDebuffs = Debuffs.create({
	misery: true,
	curseOfElements: 2,
});
