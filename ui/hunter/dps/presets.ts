import { Phase } from '../../core/constants/other';
import * as PresetUtils from '../../core/preset_utils';
import { APLRotation_Type } from '../../core/proto/apl';
import {
	Class,
	ConsumesSpec,
	Debuffs,
	IndividualBuffs,
	PartyBuffs,
	Profession,
	PseudoStat,
	Race,
	RaidBuffs,
	Stat,
	TristateEffect,
} from '../../core/proto/common';
import { HunterOptions_PetType as PetType, Hunter_Options as HunterOptions, HunterOptions_Ammo, HunterOptions_QuiverBonus } from '../../core/proto/hunter';
import { SavedTalents } from '../../core/proto/ui';
import { Stats } from '../../core/proto_utils/stats';
import { defaultRaidBuffMajorDamageCooldowns } from '../../core/proto_utils/utils';
import TurretAPL from './apls/turret.apl.json';
import WeaveAPL from './apls/weave.apl.json';
import P1PreRaidGear from './gear_sets/p1_pre_raid.gear.json';
import P1BMDW6PGear from './gear_sets/p1_bm_dw_6p.gear.json';
import P1BMDW9PGear from './gear_sets/p1_bm_dw_9p.gear.json';
import P1BM2H6PGear from './gear_sets/p1_bm_2h_6p.gear.json';
import P1BM2H9PGear from './gear_sets/p1_bm_2h_9p.gear.json';
import P1SVDW3PGear from './gear_sets/p1_sv_dw_3p.gear.json';
import P1SVDW6PGear from './gear_sets/p1_sv_dw_6p.gear.json';
import P1SV2H3PGear from './gear_sets/p1_sv_2h_3p.gear.json';
import P1SV2H6PGear from './gear_sets/p1_sv_2h_6p.gear.json';

export const TURRET_APL = PresetUtils.makePresetAPLRotation('Turret', TurretAPL);
export const WEAVE_APL = PresetUtils.makePresetAPLRotation('Weave', WeaveAPL);

export const P1_PreRaid_GEARSET = PresetUtils.makePresetGear('P1 - Pre-raid', P1PreRaidGear);
export const P1_BM_DW_6P_GEARSET = PresetUtils.makePresetGear('P1 - BM - DW (imp. FF)', P1BMDW6PGear);
export const P1_BM_DW_9P_GEARSET = PresetUtils.makePresetGear('P1 - BM - DW (no imp. FF)', P1BMDW9PGear);
export const P1_BM_2H_6P_GEARSET = PresetUtils.makePresetGear('P1 - BM - 2H (imp. FF)', P1BM2H6PGear);
export const P1_BM_2H_9P_GEARSET = PresetUtils.makePresetGear('P1 - BM - 2H (no imp. FF)', P1BM2H9PGear);
export const P1_SV_DW_3P_GEARSET = PresetUtils.makePresetGear('P1 - SV - DW (imp. FF)', P1SVDW3PGear);
export const P1_SV_DW_6P_GEARSET = PresetUtils.makePresetGear('P1 - SV - DW (no imp. FF)', P1SVDW6PGear);
export const P1_SV_2H_3P_GEARSET = PresetUtils.makePresetGear('P1 - SV - 2H (imp. FF)', P1SV2H3PGear);
export const P1_SV_2H_6P_GEARSET = PresetUtils.makePresetGear('P1 - SV - 2H (no imp. FF)', P1SV2H6PGear);

export const P1_BM_EP_PRESET = PresetUtils.makePresetEpWeights(
	'P1 BM',
	Stats.fromMap(
		{
			[Stat.StatAgility]: 1,
			[Stat.StatStrength]: 0.06,
			[Stat.StatIntellect]: 0.01,
			[Stat.StatAttackPower]: 0.06,
			[Stat.StatRangedAttackPower]: 0.4,
			[Stat.StatMeleeHitRating]: 0.12,
			[Stat.StatMeleeCritRating]: 0.92,
			[Stat.StatMeleeHasteRating]: 0.788,
			[Stat.StatArmorPenetration]: 0.16,
		},
		{
			[PseudoStat.PseudoStatRangedDps]: 1.75,
		},
	),
);

export const P1_SV_EP_PRESET = PresetUtils.makePresetEpWeights(
	'P1 SV',
	Stats.fromMap(
		{
			[Stat.StatAgility]: 1,
			[Stat.StatStrength]: 0.06,
			[Stat.StatIntellect]: 0.01,
			[Stat.StatAttackPower]: 0.06,
			[Stat.StatRangedAttackPower]: 0.4,
			[Stat.StatMeleeHitRating]: 0.12,
			[Stat.StatMeleeCritRating]: 0.92,
			[Stat.StatMeleeHasteRating]: 0.788,
			[Stat.StatArmorPenetration]: 0.16,
		},
		{
			[PseudoStat.PseudoStatRangedDps]: 1.75,
		},
	),
);

// Default talents. Uses the wowhead calculator format, make the talents on
// https://wowhead.com/wotlk/talent-calc and copy the numbers in the url.

export const BMTalents = {
	name: 'BM',
	data: SavedTalents.create({
		talentsString: '522002005150122431051-0550201205',
	}),
};
export const SVTalents = {
	name: 'SV',
	data: SavedTalents.create({
		talentsString: '502-0550201205-333200022003223005103',
	}),
};

export const DefaultOptions = HunterOptions.create({
	classOptions: {
		ammo: HunterOptions_Ammo.WardensArrow,
		quiverBonus: HunterOptions_QuiverBonus.Speed15,
		petType: PetType.Ravager,
		petUptime: 1,
		petSingleAbility: false,
	},
});

export const DefaultIndividualBuffs = IndividualBuffs.create({
	blessingOfKings: true,
	blessingOfMight: TristateEffect.TristateEffectImproved,
	blessingOfWisdom: TristateEffect.TristateEffectImproved,
	unleashedRage: true,
});

export const DefaultPartyBuffs = PartyBuffs.create({
	battleShout: TristateEffect.TristateEffectImproved,
	braidedEterniumChain: true,
	ferociousInspiration: 1,
	graceOfAirTotem: TristateEffect.TristateEffectImproved,
	leaderOfThePack: TristateEffect.TristateEffectImproved,
	strengthOfEarthTotem: TristateEffect.TristateEffectImproved,
	totemTwisting: true,
	windfuryTotem: TristateEffect.TristateEffectImproved,
});

export const DefaultPartyBuffsDW = PartyBuffs.create({
	...DefaultPartyBuffs,
	totemTwisting: false,
	windfuryTotem: TristateEffect.TristateEffectMissing,
});

export const DefaultRaidBuffs = RaidBuffs.create({
	...defaultRaidBuffMajorDamageCooldowns(Class.ClassWarrior),
	arcaneBrilliance: true,
	divineSpirit: TristateEffect.TristateEffectImproved,
	giftOfTheWild: TristateEffect.TristateEffectImproved,
	powerWordFortitude: TristateEffect.TristateEffectImproved,
	shadowProtection: true,
});

export const DefaultDebuffsNoImpFF = Debuffs.create({
	bloodFrenzy: true,
	curseOfRecklessness: true,
	exposeArmor: TristateEffect.TristateEffectImproved,
	exposeWeaknessUptime: 0.9,
	exposeWeaknessHunterAgility: 1080,
	faerieFire: TristateEffect.TristateEffectRegular,
	giftOfArthas: true,
	huntersMark: TristateEffect.TristateEffectImproved,
	improvedSealOfTheCrusader: true,
	insectSwarm: true,
	judgementOfLight: true,
	judgementOfWisdom: true,
	mangle: true,
	misery: true,
	sunderArmor: true,
});

export const DefaultDebuffs = Debuffs.create({
	...DefaultDebuffsNoImpFF,
	faerieFire: TristateEffect.TristateEffectImproved,
});

export const DefaultConsumables = ConsumesSpec.create({
	battleElixirId: 22831, // Elixir of Major Agility
	foodId: 27659, // Warp Burger
	potId: 22838, // Haste Potion
	conjuredId: 12662,
	explosiveId: 30217,
	drumsId: 351355,
	petFoodId: 33874, // Kibler's Bits
	petScrollAgi: true,
	petScrollStr: true,
	superSapper: true,
	goblinSapper: true,
	scrollAgi: true,
	scrollStr: true,
});

export const DefaultConsumablesDW = ConsumesSpec.create({
	...DefaultConsumables,
	// Doesn't work right now, requires pressing the preset twice...
	// mhImbueId: 34340,
	// ohImbueId: 29453,
});

export const OtherDefaults = {
	distanceFromTarget: 7,
	iterationCount: 25000,
	profession1: Profession.Engineering,
	profession2: Profession.Blacksmithing,
	race: Race.RaceOrc,
};

export const P1_PLAYER_SETTINGS_2H: PresetUtils.PresetSettings = {
	name: 'P1 - 2H',
	consumables: DefaultConsumables,
	debuffs: DefaultDebuffsNoImpFF,
	partyBuffs: DefaultPartyBuffs,
	playerOptions: OtherDefaults,
	reforgeSettings: {
		maxGemPhase: Phase.Phase1,
	},
};

export const P1_PLAYER_SETTINGS_IMP_FF_2H: PresetUtils.PresetSettings = {
	name: 'P1 (Improved Faerie Fire) - 2H',
	consumables: DefaultConsumables,
	debuffs: DefaultDebuffs,
	partyBuffs: DefaultPartyBuffs,
	playerOptions: OtherDefaults,
	reforgeSettings: {
		maxGemPhase: Phase.Phase1,
	},
};

export const P1_PLAYER_SETTINGS_DW: PresetUtils.PresetSettings = {
	name: 'P1 - DW',
	consumables: DefaultConsumablesDW,
	debuffs: DefaultDebuffsNoImpFF,
	partyBuffs: DefaultPartyBuffsDW,
	playerOptions: OtherDefaults,
	reforgeSettings: {
		maxGemPhase: Phase.Phase1,
	},
};

export const P1_PLAYER_SETTINGS_IMP_FF_DW: PresetUtils.PresetSettings = {
	name: 'P1 (Improved Faerie Fire) - DW',
	consumables: DefaultConsumablesDW,
	debuffs: DefaultDebuffs,
	partyBuffs: DefaultPartyBuffsDW,
	playerOptions: OtherDefaults,
	reforgeSettings: {
		maxGemPhase: Phase.Phase1,
	},
};

export const P1_PRESET_BUILD_PRE_RAID = PresetUtils.makePresetBuild('P1 - Pre-Raid', {
	gear: P1_PreRaid_GEARSET,
	epWeights: P1_BM_EP_PRESET,
	rotationType: APLRotation_Type.TypeAuto,
	rotation: WEAVE_APL,
	settings: P1_PLAYER_SETTINGS_2H,
	talents: BMTalents,
});

export const P1_PRESET_BUILD_BM_2H = PresetUtils.makePresetBuild('P1 - BM - 2H', {
	gear: P1_BM_2H_6P_GEARSET,
	epWeights: P1_BM_EP_PRESET,
	rotationType: APLRotation_Type.TypeAuto,
	rotation: WEAVE_APL,
	settings: P1_PLAYER_SETTINGS_IMP_FF_2H,
	talents: BMTalents,
});

export const P1_PRESET_BUILD_BM_DW = PresetUtils.makePresetBuild('P1 - BM - DW', {
	gear: P1_BM_DW_6P_GEARSET,
	epWeights: P1_BM_EP_PRESET,
	rotationType: APLRotation_Type.TypeAuto,
	rotation: TURRET_APL,
	settings: P1_PLAYER_SETTINGS_IMP_FF_DW,
	talents: BMTalents,
});

const Custom_Surefooted_Talents = {
	name: 'SV',
	data: SavedTalents.create({
		talentsString: '502-0550201205-333200023002223005103',
	}),
};

export const P1_PRESET_BUILD_SV_2H = PresetUtils.makePresetBuild('P1 - SV - 2H', {
	gear: P1_SV_2H_3P_GEARSET,
	epWeights: P1_SV_EP_PRESET,
	rotationType: APLRotation_Type.TypeAuto,
	rotation: WEAVE_APL,
	settings: P1_PLAYER_SETTINGS_IMP_FF_2H,
	talents: Custom_Surefooted_Talents,
});

export const P1_PRESET_BUILD_SV_DW = PresetUtils.makePresetBuild('P1 - SV - DW', {
	gear: P1_SV_DW_3P_GEARSET,
	epWeights: P1_SV_EP_PRESET,
	rotationType: APLRotation_Type.TypeAuto,
	rotation: TURRET_APL,
	settings: P1_PLAYER_SETTINGS_IMP_FF_DW,
	talents: Custom_Surefooted_Talents,
});
