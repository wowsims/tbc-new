import * as PresetUtils from '../../core/preset_utils';
import { Debuffs, PseudoStat, RaidBuffs, Stat, ConsumesSpec, TristateEffect, PartyBuffs, IndividualBuffs, Profession } from '../../core/proto/common';
import { defaultRaidBuffMajorDamageCooldowns } from '../../core/proto_utils/utils';
import { Stats } from '../../core/proto_utils/stats';
import { SavedTalents } from '../../core/proto/ui';
import { MageArmor, Mage_Options as MageOptions } from '../../core/proto/mage';
import BlankAPL from './apls/blank.apl.json';
import BlankGear from './gear_sets/blank.gear.json';
import ArcaneApl from './apls/arcane.apl.json';
import PreBISArcaneGear from './gear_sets/preBisArcane.gear.json';
import P1BISArcaneGear from './gear_sets/p1Arcane.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

export const BLANK_APL = PresetUtils.makePresetAPLRotation('Blank', BlankAPL);
export const PREBIS_ARCANE = PresetUtils.makePresetGear('Arcane PreRaid - BIS', PreBISArcaneGear);
export const P1_BIS_ARCANE = PresetUtils.makePresetGear('Arcane P1 - BIS', P1BISArcaneGear);
//export const P2_BIS_ARCANE = PresetUtils.makePresetGear('Arcane P2 - BIS', P2BISArcaneGear);
//export const P3_BIS_ARCANE = PresetUtils.makePresetGear('Arcane P3 - BIS', P3BISArcaneGear);

export const ARCANE_TALENTS = PresetUtils.makePresetTalents('Arcane', SavedTalents.create({ talentsString: '2500052300030150330125--053500031003001' }));

export const ROTATION_PRESET_ARCANE = PresetUtils.makePresetAPLRotation('Arcane', ArcaneApl);

export const BLANK_GEARSET = PresetUtils.makePresetGear('Blank', BlankGear);

// Preset options for EP weights
export const P1_EP_PRESET = PresetUtils.makePresetEpWeights(
	'A',
	Stats.fromMap({
		[Stat.StatIntellect]: 1,
		[Stat.StatSpirit]: 1,
		[Stat.StatSpellDamage]: 1,
		[Stat.StatFrostDamage]: 1,
		[Stat.StatFireDamage]: 1,
		[Stat.StatArcaneDamage]: 1,
		[Stat.StatSpellHitRating]: 1,
		[Stat.StatSpellCritRating]: 1,
		[Stat.StatSpellHasteRating]: 1,
		[Stat.StatSpellPenetration]: 1,
		[Stat.StatMana]: 1,
	}),
);

export const Talents = {
	name: 'A',
	data: SavedTalents.create({
		talentsString: '',
	}),
};

export const DefaultOptions = MageOptions.create({
	classOptions: {
		defaultMageArmor: MageArmor.MageArmorMageArmor,
	},
});

export const OtherDefaults = {
	distanceFromTarget: 20,
	profession1: Profession.Engineering,
	profession2: Profession.Tailoring,
};

export const DefaultConsumables = ConsumesSpec.create({
	guardianElixirId: 32067, // Elixir of Draenic Wisdom
	battleElixirId: 28103, // Adept's Elixir
	foodId: 27657, // Blackened Basilisk
	mhImbueId: 25122, // Brilliant Wizard Oil
	potId: 22839, // Destruction Potion
	conjuredId: 12662, // Demonic Rune
	drumsId: 351355, // Greater Drums of Battle
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
	manaSpringTotem: 2,
	manaTideTotems: 1,
	wrathOfAirTotem: 1,
});

export const DefaultIndividualBuffs = IndividualBuffs.create({
	blessingOfKings: true,
	blessingOfWisdom: 2,
	innervates: 1,
	powerInfusions: 1,
	shadowPriestDps: 800,
});

export const DefaultDebuffs = Debuffs.create({
	misery: true,
	curseOfElements: 2,
	improvedSealOfTheCrusader: true,
	judgementOfWisdom: true,
});
