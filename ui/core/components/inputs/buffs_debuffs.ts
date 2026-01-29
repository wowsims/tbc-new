import { Drums, PseudoStat, Race, RaidBuffs, Stat, TristateEffect } from '../../proto/common';
import { ActionId } from '../../proto_utils/action_id';
import i18n from '../../../i18n/config';
import {
	makeBooleanDebuffInput,
	makeBooleanIndividualBuffInput,
	makeBooleanRaidBuffInput,
	makeEnumValuePartyBuffInput,
	makeMultistateIndividualBuffInput,
	makeMultistatePartyBuffInput,
	makeMultistateRaidBuffInput,
	makeTristateRaidBuffInput,
	makeTristatePartyBuffInput,
	makeTristateDebuffInput,
	makeTristateIndividualBuffInput,
	makeBooleanPartyBuffInput,
} from '../icon_inputs';
import * as InputHelpers from '../input_helpers';
import { IconPicker } from '../pickers/icon_picker';
import { MultiIconPicker } from '../pickers/multi_icon_picker';
import { IconPickerStatOption, PickerStatOptions } from './stat_options';
import { Player } from '../../player';
import { raceI18nKeys } from '../../../i18n/entity_mapping';
import { Party } from '../../party';

///////////////////////////////////////////////////////////////////////////
//                                 RAID BUFFS
///////////////////////////////////////////////////////////////////////////

// Raid Buffs
export const ArcaneBrilliance = makeBooleanRaidBuffInput({ actionId: ActionId.fromSpellId(27127), fieldName: 'arcaneBrilliance', label: 'Arcane Brilliance' });
export const Bloodlust = makeBooleanRaidBuffInput({ actionId: ActionId.fromSpellId(2825), fieldName: 'bloodlust', label: 'Bloodlust' });
export const DivineSpirit = makeTristateRaidBuffInput({
	actionId: ActionId.fromSpellId(25312),
	impId: ActionId.fromSpellId(33182),
	fieldName: 'divineSpirit',
	label: 'Divine Spirit',
});
export const GiftOfTheWild = makeTristateRaidBuffInput({
	actionId: ActionId.fromSpellId(26991),
	impId: ActionId.fromSpellId(17055),
	fieldName: 'giftOfTheWild',
	label: 'Gift of the Wild',
});
export const Thorns = makeTristateRaidBuffInput({
	actionId: ActionId.fromSpellId(26992),
	impId: ActionId.fromSpellId(16840),
	fieldName: 'thorns',
	label: 'Thorns',
});
export const PowerWordFortitude = makeTristateRaidBuffInput({
	actionId: ActionId.fromSpellId(25389),
	impId: ActionId.fromSpellId(14767),
	fieldName: 'powerWordFortitude',
	label: 'Power Word: Fortitude',
});
export const ShadowProtection = makeBooleanRaidBuffInput({ actionId: ActionId.fromSpellId(39374), fieldName: 'shadowProtection', label: 'Shadow Protection' });

// // Party Buffs
export const AtieshMage = makeMultistatePartyBuffInput(ActionId.fromSpellId(28142), 5, 'atieshMage', 'Atiesh - Mage');
export const AtieshWarlock = makeMultistatePartyBuffInput(ActionId.fromSpellId(28143), 5, 'atieshWarlock', 'Atiesh - Warlock');
export const BraidedEterniumChain = makeBooleanPartyBuffInput({
	actionId: ActionId.fromSpellId(31025),
	fieldName: 'braidedEterniumChain',
	label: 'Braided Eternium Chain',
});
export const ChainOfTheTwilightOwl = makeBooleanPartyBuffInput({
	actionId: ActionId.fromSpellId(31035),
	fieldName: 'chainOfTheTwilightOwl',
	label: 'Chain of the Twilight Owl',
});
export const CommandingShout = makeTristatePartyBuffInput({
	actionId: ActionId.fromSpellId(469),
	impId: ActionId.fromSpellId(12861),
	fieldName: 'commandingShout',
	label: 'Commanding Shout',
});
export const DevotionAura = makeTristatePartyBuffInput({
	actionId: ActionId.fromSpellId(27149),
	impId: ActionId.fromSpellId(20142),
	fieldName: 'devotionAura',
	label: 'Devotion Aura',
});
export const DraeneiRacialCaster = makeBooleanPartyBuffInput({
	actionId: ActionId.fromSpellId(28878),
	fieldName: 'draeneiRacialCaster',
	label: 'Inspiring Presense - Caster',
	showWhen: (party: Party) => {
		console.log('yayyy', party.getPlayer(0)!.getRace());
		return [Race.RaceDraenei, Race.RaceDwarf, Race.RaceGnome, Race.RaceHuman, Race.RaceNightElf].includes(party.getPlayer(0)!.getRace())
	},
});
export const DraeneiRacialMelee = makeBooleanPartyBuffInput({
	actionId: ActionId.fromSpellId(6562),
	fieldName: 'draeneiRacialMelee',
	label: 'Inspiring Presense - Melee',
	showWhen: (party: Party) => [Race.RaceDraenei, Race.RaceDwarf, Race.RaceGnome, Race.RaceHuman, Race.RaceNightElf].includes(party.getPlayer(0)!.getRace()),
});
export const EyeOfTheNight = makeBooleanPartyBuffInput({ actionId: ActionId.fromSpellId(31033), fieldName: 'eyeOfTheNight', label: 'Eye of the Night' });
export const FerociousInspiration = makeMultistatePartyBuffInput(ActionId.fromSpellId(34460), 5, 'ferociousInspiration', 'Ferocious Inspiratation');
export const GraceOfAirTotem = makeTristatePartyBuffInput({
	actionId: ActionId.fromSpellId(8835),
	impId: ActionId.fromSpellId(25359),
	fieldName: 'graceOfAirTotem',
	label: 'Grace of Air Totem',
});
export const JadePendantOfBlasting = makeBooleanPartyBuffInput({
	actionId: ActionId.fromSpellId(25607),
	fieldName: 'jadePendantOfBlasting',
	label: 'Jade Pendant of Blasting',
});
export const LeaderOfThePack = makeTristatePartyBuffInput({
	actionId: ActionId.fromSpellId(32387),
	impId: ActionId.fromItemId(34297),
	fieldName: 'leaderOfThePack',
	label: 'Leader of the Pack',
});
export const ManaSpringTotem = makeTristatePartyBuffInput({
	actionId: ActionId.fromSpellId(25570),
	impId: ActionId.fromSpellId(16208),
	fieldName: 'manaSpringTotem',
	label: 'Mana Spring Totem',
});
export const ManaTideTotem = makeMultistatePartyBuffInput(ActionId.fromSpellId(16190), 5, 'manaTideTotems', 'Mana Tide Totem');
export const MoonkinAura = makeTristatePartyBuffInput({
	actionId: ActionId.fromSpellId(24907),
	impId: ActionId.fromItemId(32387),
	fieldName: 'moonkinAura',
	label: 'Moonkin Aura',
});
export const RetributionAura = makeTristatePartyBuffInput({
	actionId: ActionId.fromSpellId(27150),
	impId: ActionId.fromSpellId(20092),
	fieldName: 'retributionAura',
	label: 'Retribution Aura',
});
export const SanctityAura = makeTristatePartyBuffInput({
	actionId: ActionId.fromSpellId(20218),
	impId: ActionId.fromSpellId(31870),
	fieldName: 'sanctityAura',
	label: 'Sanctity Aura',
});
export const StrengthOfEarthTotem = makeTristatePartyBuffInput({
	actionId: ActionId.fromSpellId(8076),
	impId: ActionId.fromSpellId(25527),
	fieldName: 'strengthOfEarthTotem',
	label: 'Strength of Earth Totem',
});
export const TotemOfWrath = makeMultistatePartyBuffInput(ActionId.fromSpellId(30706), 5, 'totemOfWrath', 'Totem of Wrath');
export const TrueshotAura = makeBooleanPartyBuffInput({ actionId: ActionId.fromSpellId(27066), fieldName: 'trueshotAura', label: 'Trueshot Aura' });
export const WrathOfAirTotem = makeTristatePartyBuffInput({
	actionId: ActionId.fromSpellId(3738),
	impId: ActionId.fromSpellId(37212),
	fieldName: 'wrathOfAirTotem',
	label: 'Wrath of Air Totem',
});
export const BloodPact = makeTristatePartyBuffInput({
	actionId: ActionId.fromSpellId(27268),
	impId: ActionId.fromSpellId(18696),
	fieldName: 'bloodPact',
	label: 'Bloodpact',
});

export const DrumsOfBattleBuff = makeEnumValuePartyBuffInput(ActionId.fromItemId(185848), 'drums', Drums.DrumsOfBattle);
export const DrumsOfRestorationBuff = makeEnumValuePartyBuffInput(ActionId.fromItemId(185850), 'drums', Drums.DrumsOfRestoration);

// Individual Buffs
export const BlessingOfKings = makeBooleanIndividualBuffInput({
	actionId: ActionId.fromSpellId(25898),
	fieldName: 'blessingOfKings',
	label: 'Blessing of Kings',
});
export const BlessingOfMight = makeTristateIndividualBuffInput({
	actionId: ActionId.fromSpellId(27140),
	impId: ActionId.fromSpellId(20048),
	fieldName: 'blessingOfMight',
	label: 'Blessing of Might',
});
export const BlessingOfSalvation = makeBooleanIndividualBuffInput({
	actionId: ActionId.fromSpellId(25895),
	fieldName: 'blessingOfSalvation',
	label: 'Blessing of Salvation',
});
export const BlessingOfSanctuary = makeBooleanIndividualBuffInput({
	actionId: ActionId.fromSpellId(27169),
	fieldName: 'blessingOfSanctuary',
	label: 'BlessingOfSanctuary',
});
export const BlessingOfWisdom = makeTristateIndividualBuffInput({
	actionId: ActionId.fromSpellId(27143),
	impId: ActionId.fromSpellId(20245),
	fieldName: 'blessingOfWisdom',
	label: 'Blessing of Wisdom',
});
export const Innervate = makeMultistateIndividualBuffInput({
	actionId: ActionId.fromSpellId(29166),
	numStates: 11,
	fieldName: 'innervates',
	label: 'Innervates',
});
export const PowerInfusion = makeMultistateIndividualBuffInput({
	actionId: ActionId.fromSpellId(10060),
	numStates: 11,
	fieldName: 'powerInfusions',
	label: 'Power Infusions',
});
export const UnleashedRage = makeBooleanIndividualBuffInput({ actionId: ActionId.fromSpellId(30811), fieldName: 'unleashedRage', label: 'Unleashed Rage' });

export const PARTY_BUFFS_CONFIG = [
	{
		config: BloodPact,
		picker: IconPicker,
		stats: [Stat.StatStamina],
	},
	{
		config: CommandingShout,
		picker: IconPicker,
		stats: [Stat.StatHealth],
	},
	{
		config: DevotionAura,
		picker: IconPicker,
		stats: [Stat.StatArmor],
	},
	{
		config: FerociousInspiration,
		picker: IconPicker,
		stats: [Stat.StatAttackPower, Stat.StatRangedAttackPower],
	},
	{
		config: LeaderOfThePack,
		picker: IconPicker,
		stats: [Stat.StatAllPhysCritRating],
	},
	{
		config: ManaSpringTotem,
		picker: IconPicker,
		stats: [Stat.StatMP5],
	},
	{
		config: ManaTideTotem,
		picker: IconPicker,
		stats: [Stat.StatMP5],
	},
	{
		config: MoonkinAura,
		picker: IconPicker,
		stats: [Stat.StatSpellCritRating],
	},
	{
		config: RetributionAura,
		picker: IconPicker,
		stats: [Stat.StatResilienceRating, Stat.StatArmor, Stat.StatDefenseRating],
	},
	{
		config: SanctityAura,
		picker: IconPicker,
		stats: [Stat.StatHolyDamage],
	},
	{
		config: StrengthOfEarthTotem,
		picker: IconPicker,
		stats: [Stat.StatStrength],
	},
	{
		config: TotemOfWrath,
		picker: IconPicker,
		stats: [Stat.StatSpellCritRating, Stat.StatSpellHitRating],
	},
	{
		config: TrueshotAura,
		picker: IconPicker,
		stats: [],
	},
	{
		config: WrathOfAirTotem,
		picker: IconPicker,
		stats: [Stat.StatSpellDamage],
	},
	{
		config: UnleashedRage,
		picker: IconPicker,
		stats: [Stat.StatAttackPower],
	},
	{
		config: AtieshMage,
		picker: IconPicker,
		stats: [Stat.StatSpellDamage, Stat.StatHealingPower],
	},
	{
		config: AtieshWarlock,
		picker: IconPicker,
		stats: [Stat.StatSpellDamage, Stat.StatHealingPower],
	},
	{
		config: BraidedEterniumChain,
		picker: IconPicker,
		stats: [Stat.StatSpellCritRating],
	},
	{
		config: ChainOfTheTwilightOwl,
		picker: IconPicker,
		stats: [Stat.StatSpellCritRating],
	},
	{
		config: DraeneiRacialCaster,
		picker: IconPicker,
		stats: [Stat.StatSpellHitRating],
	},
	{
		config: DraeneiRacialMelee,
		picker: IconPicker,
		stats: [Stat.StatMeleeHitRating],
	},
	{
		config: EyeOfTheNight,
		picker: IconPicker,
		stats: [Stat.StatSpellDamage],
	},
	{
		config: JadePendantOfBlasting,
		picker: IconPicker,
		stats: [Stat.StatSpellDamage],
	},
	{
		config: GraceOfAirTotem,
		picker: IconPicker,
		stats: [Stat.StatAgility],
	},
] as PickerStatOptions[];

export const BUFFS_CONFIG = [
	// Raid Buffs
	{
		config: ArcaneBrilliance,
		picker: IconPicker,
		stats: [Stat.StatIntellect],
	},
	{
		config: BlessingOfKings,
		picker: IconPicker,
		stats: [Stat.StatAgility, Stat.StatIntellect, Stat.StatSpirit, Stat.StatStamina, Stat.StatStrength],
	},
	{
		config: Bloodlust,
		picker: IconPicker,
		stats: [],
	},
	{
		config: DivineSpirit,
		picker: IconPicker,
		stats: [Stat.StatSpirit],
	},
	{
		config: GiftOfTheWild,
		picker: IconPicker,
		stats: [Stat.StatArmor, Stat.StatStrength, Stat.StatAgility, Stat.StatIntellect, Stat.StatSpirit, Stat.StatStamina],
	},
	{
		config: Thorns,
		picker: IconPicker,
		stats: [Stat.StatResilienceRating, Stat.StatDefenseRating, Stat.StatStamina],
	},
	{
		config: PowerWordFortitude,
		picker: IconPicker,
		stats: [Stat.StatStamina],
	},
	{
		config: BlessingOfMight,
		picker: IconPicker,
		stats: [Stat.StatAttackPower],
	},
	{
		config: BlessingOfWisdom,
		picker: IconPicker,
		stats: [Stat.StatMP5],
	},
	{
		config: Innervate,
		picker: IconPicker,
		stats: [Stat.StatMP5],
	},
	{
		config: PowerInfusion,
		picker: IconPicker,
		stats: [Stat.StatSpellHasteRating],
	},
] as PickerStatOptions[];

// Debuffs
export const BloodFrenzy = makeBooleanDebuffInput({ actionId: ActionId.fromSpellId(29859), fieldName: 'bloodFrenzy', label: 'Blood Frenzy' });
export const HuntersMark = makeTristateDebuffInput({
	actionId: ActionId.fromSpellId(14325),
	impId: ActionId.fromSpellId(19425),
	fieldName: 'huntersMark',
	label: "Hunter's Mark",
});
export const ImprovedScorch = makeBooleanDebuffInput({ actionId: ActionId.fromSpellId(12873), fieldName: 'improvedScorch', label: 'Improved Scorch' });
export const ImprovedSealOfTheCrusader = makeBooleanDebuffInput({
	actionId: ActionId.fromSpellId(20337),
	fieldName: 'improvedSealOfTheCrusader',
	label: 'Improved Seal of the Crusader',
});
export const JudgementOfWisdom = makeBooleanDebuffInput({
	actionId: ActionId.fromSpellId(27164),
	fieldName: 'judgementOfWisdom',
	label: 'Judgement of Wisdom',
});
export const JudgementOfLight = makeBooleanDebuffInput({ actionId: ActionId.fromSpellId(27163), fieldName: 'judgementOfLight', label: 'Judgement of Light' });
export const Mangle = makeBooleanDebuffInput({ actionId: ActionId.fromSpellId(33876), fieldName: 'mangle', label: 'Mangle' });
export const Misery = makeBooleanDebuffInput({ actionId: ActionId.fromSpellId(33195), fieldName: 'misery', label: 'Misery' });
export const ShadowWeaving = makeBooleanDebuffInput({ actionId: ActionId.fromSpellId(15334), fieldName: 'shadowWeaving', label: 'Shadow Weaving' });
export const CurseOfElements = makeTristateDebuffInput({
	actionId: ActionId.fromSpellId(27228),
	impId: ActionId.fromSpellId(32484),
	fieldName: 'curseOfElements',
	label: 'Curse of Elements',
});
export const CurseOfRecklessness = makeBooleanDebuffInput({
	actionId: ActionId.fromSpellId(27226),
	fieldName: 'curseOfRecklessness',
	label: 'Curse of Recklessness',
});
export const FaerieFire = makeTristateDebuffInput({
	actionId: ActionId.fromSpellId(26993),
	impId: ActionId.fromSpellId(33602),
	fieldName: 'faerieFire',
	label: 'Faerie Fire',
});
export const ExposeArmor = makeTristateDebuffInput({
	actionId: ActionId.fromSpellId(26866),
	impId: ActionId.fromSpellId(14169),
	fieldName: 'exposeArmor',
	label: 'Expose Armor',
});
export const SunderArmor = makeBooleanDebuffInput({ actionId: ActionId.fromSpellId(25225), fieldName: 'sunderArmor', label: 'Sunder Armor' });
export const WintersChill = makeBooleanDebuffInput({ actionId: ActionId.fromSpellId(28595), fieldName: 'wintersChill', label: "Winter's Chill" });
export const GiftOfArthas = makeBooleanDebuffInput({ actionId: ActionId.fromSpellId(11374), fieldName: 'giftOfArthas', label: 'Gift of Arthas' });
export const DemoralizingRoar = makeTristateDebuffInput({
	actionId: ActionId.fromSpellId(26998),
	impId: ActionId.fromSpellId(16862),
	fieldName: 'demoralizingRoar',
	label: 'Demoralizing Roar',
});
export const DemoralizingShout = makeTristateDebuffInput({
	actionId: ActionId.fromSpellId(25203),
	impId: ActionId.fromSpellId(12879),
	fieldName: 'demoralizingShout',
	label: 'Demoralizing Shout',
});
export const Screech = makeBooleanDebuffInput({ actionId: ActionId.fromSpellId(27051), fieldName: 'screech', label: 'Screech' });
export const ThunderClap = makeTristateDebuffInput({
	actionId: ActionId.fromSpellId(25264),
	impId: ActionId.fromSpellId(12666),
	fieldName: 'thunderClap',
	label: 'Thunder Clap',
});
export const InsectSwarm = makeBooleanDebuffInput({ actionId: ActionId.fromSpellId(27013), fieldName: 'insectSwarm', label: 'Insect Swarm' });
export const ScorpidSting = makeBooleanDebuffInput({ actionId: ActionId.fromSpellId(3043), fieldName: 'scorpidSting', label: 'Scorpid Sting' });
export const ShadowEmbrace = makeBooleanDebuffInput({ actionId: ActionId.fromSpellId(32394), fieldName: 'shadowEmbrace', label: 'Shadow Embrace' });

export const DEBUFFS_CONFIG = [
	{
		config: BloodFrenzy,
		picker: IconPicker,
		stats: [Stat.StatAttackPower],
	},
	{
		config: HuntersMark,
		picker: IconPicker,
		stats: [Stat.StatRangedAttackPower, Stat.StatAttackPower],
	},
	{
		config: ImprovedScorch,
		picker: IconPicker,
		stats: [Stat.StatFireDamage],
	},
	{
		config: ImprovedSealOfTheCrusader,
		picker: IconPicker,
		stats: [Stat.StatMeleeCritRating, Stat.StatSpellCritRating],
	},
	{
		config: JudgementOfLight,
		picker: IconPicker,
		stats: [Stat.StatResilienceRating],
	},
	{
		config: JudgementOfWisdom,
		picker: IconPicker,
		stats: [Stat.StatMP5],
	},
	{
		config: Mangle,
		picker: IconPicker,
		stats: [],
	},
	{
		config: Misery,
		picker: IconPicker,
		stats: [Stat.StatSpellDamage],
	},
	{
		config: ShadowWeaving,
		picker: IconPicker,
		stats: [Stat.StatShadowDamage],
	},
	{
		config: CurseOfElements,
		picker: IconPicker,
		stats: [Stat.StatSpellDamage],
	},
	{
		config: CurseOfRecklessness,
		picker: IconPicker,
		stats: [Stat.StatAttackPower],
	},
	{
		config: FaerieFire,
		picker: IconPicker,
		stats: [Stat.StatAttackPower, Stat.StatAllPhysHitRating],
	},
	{
		config: ExposeArmor,
		picker: IconPicker,
		stats: [Stat.StatAttackPower],
	},
	{
		config: SunderArmor,
		picker: IconPicker,
		stats: [Stat.StatAttackPower],
	},
	{
		config: WintersChill,
		picker: IconPicker,
		stats: [Stat.StatFrostDamage],
	},
	{
		config: GiftOfArthas,
		picker: IconPicker,
		stats: [Stat.StatAttackPower],
	},
	{
		config: DemoralizingRoar,
		picker: IconPicker,
		stats: [Stat.StatResilienceRating],
	},
	{
		config: DemoralizingShout,
		picker: IconPicker,
		stats: [Stat.StatResilienceRating],
	},
	{
		config: Screech,
		picker: IconPicker,
		stats: [Stat.StatResilienceRating],
	},
	{
		config: ThunderClap,
		picker: IconPicker,
		stats: [Stat.StatResilienceRating],
	},
	{
		config: InsectSwarm,
		picker: IconPicker,
		stats: [Stat.StatResilienceRating],
	},
	{
		config: ScorpidSting,
		picker: IconPicker,
		stats: [Stat.StatResilienceRating],
	},
	{
		config: ShadowEmbrace,
		picker: IconPicker,
		stats: [Stat.StatResilienceRating],
	},
] as PickerStatOptions[];

export const DEBUFFS_MISC_CONFIG = [] as IconPickerStatOption[];
