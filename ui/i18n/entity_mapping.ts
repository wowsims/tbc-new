import { LaunchStatus } from '../core/launched_sims';
import { ArmorType, Class, MobType, PseudoStat, Race, Profession, Spec, Stat, SpellSchool, WeaponType, RangedWeaponType, ItemSlot } from '../core/proto/common';
import { ResourceType } from '../core/proto/spell';
import { RaidFilterOption, SourceFilterOption } from '../core/proto/ui';
import { BulkSimItemSlot } from '../core/components/individual_sim_ui/bulk/utils';
import { PresetConfigurationCategory } from '../core/components/individual_sim_ui/preset_configuration_picker';

export const statI18nKeys: Record<Stat, string> = {
	[Stat.StatStrength]: 'strength',
	[Stat.StatAgility]: 'agility',
	[Stat.StatStamina]: 'stamina',
	[Stat.StatIntellect]: 'intellect',
	[Stat.StatSpirit]: 'spirit',
	[Stat.StatExpertiseRating]: 'expertise',
	[Stat.StatDodgeRating]: 'dodge',
	[Stat.StatParryRating]: 'parry',
	[Stat.StatAttackPower]: 'attack_power',
	[Stat.StatRangedAttackPower]: 'ranged_attack_power',
	[Stat.StatSpellPower]: 'spell_power',
	[Stat.StatArmor]: 'armor',
	[Stat.StatBonusArmor]: 'bonus_armor',
	[Stat.StatHealth]: 'health',
	[Stat.StatMana]: 'mana',
	[Stat.StatMP5]: 'mp5',
	[Stat.StatHealingPower]: 'healing_power',
	[Stat.StatSpellDamage]: 'spell_damage',
	[Stat.StatArcaneDamage]: 'arcane_damage',
	[Stat.StatFireDamage]: 'fire_damage',
	[Stat.StatFrostDamage]: 'frost_damage',
	[Stat.StatHolyDamage]: 'holy_damage',
	[Stat.StatNatureDamage]: 'nature_damage',
	[Stat.StatShadowDamage]: 'shadow_damage',
	[Stat.StatSpellHitRating]: 'spell_hit_rating',
	[Stat.StatSpellCritRating]: 'spell_crit_rating',
	[Stat.StatSpellHasteRating]: 'spell_haste_rating',
	[Stat.StatSpellPenetration]: 'spell_penetration',
	[Stat.StatFeralAttackPower]: 'feral_attack_power',
	[Stat.StatMeleeHitRating]: 'melee_hit_rating',
	[Stat.StatMeleeCritRating]: 'melee_crit_rating',
	[Stat.StatMeleeHasteRating]: 'melee_haste_rating',
	[Stat.StatArmorPenetration]: 'armor_penetration',
	[Stat.StatAllPhysHitRating]: 'all_hit_rating',
	[Stat.StatAllPhysCritRating]: 'all_crit_rating',
	[Stat.StatAllPhysHasteRating]: 'all_haste_rating',
	[Stat.StatDefenseRating]: 'defense_rating',
	[Stat.StatBlockRating]: 'block_rating',
	[Stat.StatBlockValue]: 'block_value',
	[Stat.StatResilienceRating]: 'resilience',
	[Stat.StatArcaneResistance]: 'arcane_resistance',
	[Stat.StatFireResistance]: 'fire_resistance',
	[Stat.StatFrostResistance]: 'frost_resistance',
	[Stat.StatNatureResistance]: 'nature_resistance',
	[Stat.StatShadowResistance]: 'shadow_resistance',
};

export const protoStatNameI18nKeys: Record<string, string> = {
	['Strength']: 'strength',
	['Agility']: 'agility',
	['Stamina']: 'stamina',
	['Intellect']: 'intellect',
	['Spirit']: 'spirit',
	['HitRating']: 'hit',
	['CritRating']: 'crit',
	['HasteRating']: 'haste',
	['ExpertiseRating']: 'expertise',
	['DodgeRating']: 'dodge',
	['ParryRating']: 'parry',
	['MasteryRating']: 'mastery',
	['AttackPower']: 'attack_power',
	['RangedAttackPower']: 'ranged_attack_power',
	['SpellPower']: 'spell_power',
	['PvpResilienceRating']: 'pvp_resilience',
	['PvpPowerRating']: 'pvp_power',
	['Armor']: 'armor',
	['BonusArmor']: 'bonus_armor',
	['Health']: 'health',
	['Mana']: 'mana',
	['MP5']: 'mp5',
};

export const pseudoStatI18nKeys: Record<PseudoStat, string> = {
	[PseudoStat.PseudoStatMainHandDps]: 'main_hand_dps',
	[PseudoStat.PseudoStatOffHandDps]: 'off_hand_dps',
	[PseudoStat.PseudoStatRangedDps]: 'ranged_dps',
	[PseudoStat.PseudoStatDodgePercent]: 'dodge',
	[PseudoStat.PseudoStatParryPercent]: 'parry',
	[PseudoStat.PseudoStatBlockPercent]: 'block',
	[PseudoStat.PseudoStatMeleeSpeedMultiplier]: 'melee_speed_multiplier',
	[PseudoStat.PseudoStatRangedSpeedMultiplier]: 'ranged_speed_multiplier',
	[PseudoStat.PseudoStatCastSpeedMultiplier]: 'cast_speed_multiplier',
	[PseudoStat.PseudoStatMeleeHastePercent]: 'melee_haste',
	[PseudoStat.PseudoStatRangedHastePercent]: 'ranged_haste',
	[PseudoStat.PseudoStatSpellHastePercent]: 'spell_haste',
	[PseudoStat.PseudoStatMeleeHitPercent]: 'melee_hit',
	[PseudoStat.PseudoStatSpellHitPercent]: 'spell_hit',
	[PseudoStat.PseudoStatMeleeCritPercent]: 'melee_crit',
	[PseudoStat.PseudoStatSpellCritPercent]: 'spell_crit',
	[PseudoStat.PseudoStatBlockValueMultiplier]: 'block_value_multiplier',
	[PseudoStat.BonusPhysicalDamage]: 'bonus_physical_damage',
	[PseudoStat.PseudoStatSchoolHitArcane]: 'arcane_hit',
	[PseudoStat.PseudoStatSchoolHitFire]: 'fire_hit',
	[PseudoStat.PseudoStatSchoolHitFrost]: 'frost_hit',
	[PseudoStat.PseudoStatSchoolHitHoly]: 'holy_hit',
	[PseudoStat.PseudoStatSchoolHitNature]: 'nature_hit',
	[PseudoStat.PseudoStatSchoolHitShadow]: 'shadow_hit',
	[PseudoStat.PseudoStatBlockValuePerStrength]: 'block_per_strength',
	[PseudoStat.PseudoStatRangedHitPercent]: 'ranged_hit',
	[PseudoStat.PseudoStatRangedCritPercent]: 'ranged_crit'
};

export const spellSchoolI18nKeys: Record<SpellSchool, string> = {
	[SpellSchool.SpellSchoolPhysical]: 'physical',
	[SpellSchool.SpellSchoolArcane]: 'arcane',
	[SpellSchool.SpellSchoolFire]: 'fire',
	[SpellSchool.SpellSchoolFrost]: 'frost',
	[SpellSchool.SpellSchoolHoly]: 'holy',
	[SpellSchool.SpellSchoolNature]: 'nature',
	[SpellSchool.SpellSchoolShadow]: 'shadow',
};

export const classI18nKeys: Record<Class, string> = {
	[Class.ClassUnknown]: 'unknown',
	[Class.ClassWarrior]: 'warrior',
	[Class.ClassPaladin]: 'paladin',
	[Class.ClassHunter]: 'hunter',
	[Class.ClassRogue]: 'rogue',
	[Class.ClassPriest]: 'priest',
	[Class.ClassShaman]: 'shaman',
	[Class.ClassMage]: 'mage',
	[Class.ClassWarlock]: 'warlock',
	[Class.ClassDruid]: 'druid',
	[Class.ClassExtra1]: 'extra1',
	[Class.ClassExtra2]: 'extra2',
	[Class.ClassExtra3]: 'extra3',
	[Class.ClassExtra4]: 'extra4',
	[Class.ClassExtra5]: 'extra5',
	[Class.ClassExtra6]: 'extra6',
};

export const aplItemLabelI18nKeys: Record<string, string> = {
	action: 'rotation.apl.priority_list.item_label',
	'prepull-action': 'rotation.apl.prepull_actions.item_label',
	value: 'rotation.apl.values.item_label',
};

export const resourceTypeI18nKeys: Record<ResourceType, string> = {
	[ResourceType.ResourceTypeNone]: 'none',
	[ResourceType.ResourceTypeHealth]: 'health',
	[ResourceType.ResourceTypeMana]: 'mana',
	[ResourceType.ResourceTypeEnergy]: 'energy',
	[ResourceType.ResourceTypeRage]: 'rage',
	[ResourceType.ResourceTypeComboPoints]: 'combo_points',
	[ResourceType.ResourceTypeFocus]: 'focus',
	[ResourceType.ResourceTypeGenericResource]: 'generic_resource',
};

// standardize keys regardless they are from backend or frontend
export const backendMetricI18nKeys: Record<string, string> = {
	'Chance of Death': 'cod',
	DTPS: 'dtps',
	TMI: 'tmi',
	DPS: 'dps',
	HPS: 'hps',
	TPS: 'tps',
	DUR: 'dur',
	TTO: 'tto',
	OOM: 'oom',
};

export const specI18nKeys: Record<Spec, string> = {
	[Spec.SpecUnknown]: 'unknown',
	// Druid
	[Spec.SpecBalanceDruid]: 'balance',
	[Spec.SpecFeralCatDruid]: 'feralcat',
	[Spec.SpecFeralBearDruid]: 'feralbear',
	[Spec.SpecRestorationDruid]: 'restoration',
	// Hunter
	[Spec.SpecHunter]: 'hunter',
	// Mage
	[Spec.SpecMage]: 'mage',
	// Paladin
	[Spec.SpecHolyPaladin]: 'holy',
	[Spec.SpecProtectionPaladin]: 'protection',
	[Spec.SpecRetributionPaladin]: 'retribution',
	// Priest
	[Spec.SpecDisciplinePriest]: 'discipline',
	[Spec.SpecHolyPriest]: 'holy',
	[Spec.SpecShadowPriest]: 'shadow',
	// Rogue
	[Spec.SpecRogue]: 'rogue',
	// Shaman
	[Spec.SpecElementalShaman]: 'elemental',
	[Spec.SpecEnhancementShaman]: 'enhancement',
	[Spec.SpecRestorationShaman]: 'restoration',
	// Warlock
	[Spec.SpecWarlock]: 'warlock',
	// Warrior
	[Spec.SpecDPSWarrior]: 'dps',
	[Spec.SpecProtectionWarrior]: 'protection',
};

export const statusI18nKeys: Record<LaunchStatus, string> = {
	[LaunchStatus.Unlaunched]: 'unlaunched',
	[LaunchStatus.Alpha]: 'alpha',
	[LaunchStatus.Beta]: 'beta',
	[LaunchStatus.Launched]: 'launched',
};

export const targetInputI18nKeys: Record<string, string> = {
	'Frenzy time': 'frenzy_time',
	'Spiritual Grasp frequency': 'spiritual_grasp_frequency',
};

export const mobTypeI18nKeys: Record<MobType, string> = {
	[MobType.MobTypeUnknown]: 'unknown',
	[MobType.MobTypeBeast]: 'beast',
	[MobType.MobTypeDemon]: 'demon',
	[MobType.MobTypeDragonkin]: 'dragonkin',
	[MobType.MobTypeElemental]: 'elemental',
	[MobType.MobTypeGiant]: 'giant',
	[MobType.MobTypeHumanoid]: 'humanoid',
	[MobType.MobTypeMechanical]: 'mechanical',
	[MobType.MobTypeUndead]: 'undead',
};

export const raceI18nKeys: Record<Race, string> = {
	[Race.RaceUnknown]: 'unknown',
	[Race.RaceBloodElf]: 'blood_elf',
	[Race.RaceDraenei]: 'draenei',
	[Race.RaceDwarf]: 'dwarf',
	[Race.RaceGnome]: 'gnome',
	[Race.RaceHuman]: 'human',
	[Race.RaceNightElf]: 'night_elf',
	[Race.RaceOrc]: 'orc',
	[Race.RaceTauren]: 'tauren',
	[Race.RaceTroll]: 'troll',
	[Race.RaceUndead]: 'undead',
};

export const professionI18nKeys: Record<Profession, string> = {
	[Profession.ProfessionUnknown]: 'unknown',
	[Profession.Alchemy]: 'alchemy',
	[Profession.Blacksmithing]: 'blacksmithing',
	[Profession.Enchanting]: 'enchanting',
	[Profession.Engineering]: 'engineering',
	[Profession.Herbalism]: 'herbalism',
	[Profession.Inscription]: 'inscription',
	[Profession.Jewelcrafting]: 'jewelcrafting',
	[Profession.Leatherworking]: 'leatherworking',
	[Profession.Mining]: 'mining',
	[Profession.Skinning]: 'skinning',
	[Profession.Tailoring]: 'tailoring',
};

export const sourceFilterI18nKeys: Record<SourceFilterOption, string> = {
	[SourceFilterOption.SourceUnknown]: 'unknown',
	[SourceFilterOption.SourceCrafting]: 'crafting',
	[SourceFilterOption.SourceQuest]: 'quest',
	[SourceFilterOption.SourceReputation]: 'reputation',
	[SourceFilterOption.SourceSoldBy]: 'sold_by',
	[SourceFilterOption.SourcePvp]: 'pvp',
	[SourceFilterOption.SourceDungeon]: 'dungeon',
	[SourceFilterOption.SourceDungeonH]: 'dungeon_h',
	[SourceFilterOption.SourceRaid]: 'raid',
	[SourceFilterOption.SourceRaidH]: 'raid_h',
	[SourceFilterOption.SourceRaidRF]: 'raid_rf',
	[SourceFilterOption.SourceRaidFlex]: 'raid_flex',
};

export const raidFilterI18nKeys: Record<RaidFilterOption, string> = {
	[RaidFilterOption.RaidUnknown]: 'unknown',
	[RaidFilterOption.RaidMogushanVaults]: 'mogushan_vaults',
	[RaidFilterOption.RaidHeartOfFear]: 'heart_of_fear',
	[RaidFilterOption.RaidTerraceOfEndlessSpring]: 'terrace_of_endless_spring',
	[RaidFilterOption.RaidThroneOfThunder]: 'throne_of_thunder',
	[RaidFilterOption.RaidSiegeOfOrgrimmar]: 'siege_of_orgrimmar',
};

export const armorTypeI18nKeys: Record<ArmorType, string> = {
	[ArmorType.ArmorTypeUnknown]: 'unknown',
	[ArmorType.ArmorTypeCloth]: 'cloth',
	[ArmorType.ArmorTypeLeather]: 'leather',
	[ArmorType.ArmorTypeMail]: 'mail',
	[ArmorType.ArmorTypePlate]: 'plate',
};

export const weaponTypeI18nKeys: Record<WeaponType, string> = {
	[WeaponType.WeaponTypeUnknown]: 'unknown',
	[WeaponType.WeaponTypeAxe]: 'axe',
	[WeaponType.WeaponTypeDagger]: 'dagger',
	[WeaponType.WeaponTypeFist]: 'fist',
	[WeaponType.WeaponTypeMace]: 'mace',
	[WeaponType.WeaponTypeOffHand]: 'off_hand',
	[WeaponType.WeaponTypePolearm]: 'polearm',
	[WeaponType.WeaponTypeShield]: 'shield',
	[WeaponType.WeaponTypeStaff]: 'staff',
	[WeaponType.WeaponTypeSword]: 'sword',
};

export const rangedWeaponTypeI18nKeys: Record<RangedWeaponType, string> = {
	[RangedWeaponType.RangedWeaponTypeUnknown]: 'unknown',
	[RangedWeaponType.RangedWeaponTypeBow]: 'bow',
	[RangedWeaponType.RangedWeaponTypeCrossbow]: 'crossbow',
	[RangedWeaponType.RangedWeaponTypeGun]: 'gun',
	[RangedWeaponType.RangedWeaponTypeThrown]: 'thrown',
	[RangedWeaponType.RangedWeaponTypeWand]: 'wand',
	[RangedWeaponType.RangedWeaponTypeIdol]: 'idol',
	[RangedWeaponType.RangedWeaponTypeLibram]: 'libram',
	[RangedWeaponType.RangedWeaponTypeTotem]: 'totem',
	[RangedWeaponType.RangedWeaponTypeSigil]: 'sigil',
};

export const slotNamesI18nKeys: Record<ItemSlot, string> = {
	[ItemSlot.ItemSlotHead]: 'head',
	[ItemSlot.ItemSlotNeck]: 'neck',
	[ItemSlot.ItemSlotShoulder]: 'shoulder',
	[ItemSlot.ItemSlotBack]: 'back',
	[ItemSlot.ItemSlotChest]: 'chest',
	[ItemSlot.ItemSlotWrist]: 'wrist',
	[ItemSlot.ItemSlotHands]: 'hands',
	[ItemSlot.ItemSlotWaist]: 'waist',
	[ItemSlot.ItemSlotLegs]: 'legs',
	[ItemSlot.ItemSlotFeet]: 'feet',
	[ItemSlot.ItemSlotFinger1]: 'finger_1',
	[ItemSlot.ItemSlotFinger2]: 'finger_2',
	[ItemSlot.ItemSlotTrinket1]: 'trinket_1',
	[ItemSlot.ItemSlotTrinket2]: 'trinket_2',
	[ItemSlot.ItemSlotMainHand]: 'main_hand',
	[ItemSlot.ItemSlotOffHand]: 'off_hand',
	[ItemSlot.ItemSlotRanged]: 'ranged',
};

export const bulkSlotNamesI18nKeys: Record<BulkSimItemSlot, string> = {
	[BulkSimItemSlot.ItemSlotHead]: 'head',
	[BulkSimItemSlot.ItemSlotNeck]: 'neck',
	[BulkSimItemSlot.ItemSlotShoulder]: 'shoulder',
	[BulkSimItemSlot.ItemSlotBack]: 'back',
	[BulkSimItemSlot.ItemSlotChest]: 'chest',
	[BulkSimItemSlot.ItemSlotWrist]: 'wrist',
	[BulkSimItemSlot.ItemSlotHands]: 'hands',
	[BulkSimItemSlot.ItemSlotWaist]: 'waist',
	[BulkSimItemSlot.ItemSlotLegs]: 'legs',
	[BulkSimItemSlot.ItemSlotFeet]: 'feet',
	[BulkSimItemSlot.ItemSlotFinger]: 'rings',
	[BulkSimItemSlot.ItemSlotTrinket]: 'trinkets',
	[BulkSimItemSlot.ItemSlotMainHand]: 'main_hand',
	[BulkSimItemSlot.ItemSlotOffHand]: 'off_hand',
	[BulkSimItemSlot.ItemSlotRanged]: 'ranged',
	[BulkSimItemSlot.ItemSlotHandWeapon]: 'weapons',
};

export const presetConfigurationCategoryI18nKeys: Record<PresetConfigurationCategory, string> = {
	[PresetConfigurationCategory.EPWeights]: 'ep_weights',
	[PresetConfigurationCategory.Gear]: 'gear',
	[PresetConfigurationCategory.Talents]: 'talents',
	[PresetConfigurationCategory.Rotation]: 'rotation',
	[PresetConfigurationCategory.Encounter]: 'encounter',
	[PresetConfigurationCategory.Settings]: 'settings',
};

export const getClassI18nKey = (classID: Class): string => classI18nKeys[classID] || Class[classID].toLowerCase();

export const getSpecI18nKey = (specID: Spec): string => specI18nKeys[specID] || Spec[specID].toLowerCase();

export const getStatusI18nKey = (status: LaunchStatus): string => statusI18nKeys[status] || LaunchStatus[status].toLowerCase();

export const getTargetInputI18nKey = (label: string): string => targetInputI18nKeys[label] || label.toLowerCase().replace(/[()]/g, '').replace(/\s+/g, '_');

export const getMobTypeI18nKey = (mobType: MobType): string => mobTypeI18nKeys[mobType] || MobType[mobType].toLowerCase();

export const getRaceI18nKey = (race: Race): string => raceI18nKeys[race] || Race[race].toLowerCase();

export const getProfessionI18nKey = (profession: Profession): string => professionI18nKeys[profession] || Profession[profession].toLowerCase();

export const getSourceFilterI18nKey = (source: SourceFilterOption): string => sourceFilterI18nKeys[source] || SourceFilterOption[source].toLowerCase();

export const getRaidFilterI18nKey = (raid: RaidFilterOption): string => raidFilterI18nKeys[raid] || RaidFilterOption[raid].toLowerCase();

export const getBulkSlotI18nKey = (slot: BulkSimItemSlot): string => bulkSlotNamesI18nKeys[slot] || '';

export const getArmorTypeI18nKey = (armorType: ArmorType): string => armorTypeI18nKeys[armorType] || ArmorType[armorType].toLowerCase();

export const getWeaponTypeI18nKey = (weaponType: WeaponType): string => weaponTypeI18nKeys[weaponType] || WeaponType[weaponType].toLowerCase();

export const getRangedWeaponTypeI18nKey = (rangedWeaponType: RangedWeaponType): string =>
	rangedWeaponTypeI18nKeys[rangedWeaponType] || RangedWeaponType[rangedWeaponType].toLowerCase();

export const getSlotNameI18nKey = (slot: ItemSlot): string => slotNamesI18nKeys[slot] || ItemSlot[slot].toLowerCase();

export const getPresetConfigurationCategoryI18nKey = (category: PresetConfigurationCategory): string =>
	presetConfigurationCategoryI18nKeys[category] || category.toLowerCase();

export const classNameToClassKey = (className: string): string => {
	return className.toLowerCase().replace(/_/g, '');
};
