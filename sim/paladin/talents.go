package paladin

import (
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

func (paladin *Paladin) ApplyTalents() {
	// ==================
	// Holy Talents
	// ==================

	// Divine Strength (Tier 1) - Increases your total Strength by 2/4/6/8/10%
	if paladin.Talents.DivineStrength > 0 {
		paladin.applyDivineStrength()
	}

	// Divine Intellect (Tier 1) - Increases your total Intellect by 2/4/6/8/10%
	if paladin.Talents.DivineIntellect > 0 {
		paladin.applyDivineIntellect()
	}

	// Spiritual Focus (Tier 1) - Gives your Flash of Light and Holy Light spells a 14/28/42/56/70% chance to not lose casting time when you take damage
	// TODO: Implement pushback resistance

	// Improved Seal of Righteousness (Tier 2) - Increases the damage done by your Seal of Righteousness and its Judgement by 3/6/9/12/15%
	if paladin.Talents.ImprovedSealOfRighteousness > 0 {
		paladin.applyImprovedSealOfRighteousness()
	}

	// Healing Light (Tier 2) - Increases the amount healed by your Holy Light and Flash of Light spells by 4/8/12%
	if paladin.Talents.HealingLight > 0 {
		paladin.applyHealingLight()
	}

	// Aura Mastery (Tier 3) - Increases the radius of your Auras to 40 yards
	// TODO: Implement aura range increase

	// Improved Lay on Hands (Tier 3) - Gives the target of your Lay on Hands spell a 15/30% bonus to their armor value from items for 2 min
	// Also reduces cooldown by 10/20 min
	// TODO: Implement

	// Unyielding Faith (Tier 3) - Increases your chance to resist Fear and Disorient effects by an additional 5/10%
	// TODO: Implement resistance

	// Illumination (Tier 4) - After getting a critical effect from your Flash of Light, Holy Light, or Holy Shock heal spell,
	// you have a 20/40/60/80/100% chance to gain mana equal to 60% of the base cost of the spell
	if paladin.Talents.Illumination > 0 {
		paladin.applyIllumination()
	}

	// Improved Blessing of Wisdom (Tier 4) - Increases the effect of your Blessing of Wisdom spell by 10/20%
	if paladin.Talents.ImprovedBlessingOfWisdom > 0 {
		paladin.applyImprovedBlessingOfWisdom()
	}

	// Pure of Heart (Tier 5) - Increases your resistance to Curse and Disease effects by 5/10/15%
	// TODO: Implement resistance

	// Divine Favor (Tier 5) - When activated, gives your next Flash of Light, Holy Light, or Holy Shock spell a 100% critical effect chance
	// https://www.wowhead.com/tbc/spell=20216
	if paladin.Talents.DivineFavor {
		paladin.registerDivineFavor()
	}

	// Sanctified Light (Tier 6) - Increases the critical effect chance of your Holy Light and Holy Shock spells by 2/4/6%
	if paladin.Talents.SanctifiedLight > 0 {
		paladin.applySanctifiedLight()
	}

	// Purifying Power (Tier 6) - Reduces the mana cost of your Cleanse and Consecration spells by 5/10%, and increases the critical strike chance of your Exorcism and Holy Wrath spells by 10/20%
	if paladin.Talents.PurifyingPower > 0 {
		paladin.applyPurifyingPower()
	}

	// Holy Power (Tier 7) - Increases the critical effect chance of your Holy spells by 1/2/3/4/5%
	if paladin.Talents.HolyPower > 0 {
		paladin.applyHolyPowerTalent()
	}

	// Light's Grace (Tier 7) - Your Holy Light spell reduces the cast time of your next Holy Light spell by 0.15/0.30/0.50 sec
	// TODO: Implement

	// Holy Shock (Tier 8) - Blasts the target with Holy energy
	// https://www.wowhead.com/tbc/spell=33072
	if paladin.Talents.HolyShock {
		paladin.registerHolyShock()
	}

	// Blessed Life (Tier 8) - All attacks against you have a 4/7/10% chance to cause half damage
	// TODO: Implement damage reduction

	// Holy Guidance (Tier 9) - Increases your spell damage and healing by 7/14/21/28/35% of your total Intellect
	if paladin.Talents.HolyGuidance > 0 {
		paladin.applyHolyGuidance()
	}

	// Divine Illumination (Tier 9) - Reduces the mana cost of all spells by 50% for 15 sec
	// https://www.wowhead.com/tbc/spell=31842
	if paladin.Talents.DivineIllumination {
		paladin.registerDivineIllumination()
	}

	// ==================
	// Protection Talents
	// ==================

	// Improved Devotion Aura (Tier 1) - Increases the armor bonus of your Devotion Aura by 8/16/24/32/40%
	if paladin.Talents.ImprovedDevotionAura > 0 {
		paladin.applyImprovedDevotionAura()
	}

	// Redoubt (Tier 1) - Increases your chance to block by 6/12/18/24/30% after being the victim of a critical strike
	// TODO: Implement proc

	// Precision (Tier 2) - Increases your chance to hit with melee weapons and spells by 1/2/3%
	if paladin.Talents.Precision > 0 {
		paladin.applyPrecision()
	}

	// Guardian's Favor (Tier 2) - Reduces the cooldown of your Blessing of Protection by 60/120 sec and increases the duration by 1/2 sec
	// TODO: Implement cooldown reduction

	// Toughness (Tier 2) - Increases your armor value from items by 2/4/6/8/10%
	if paladin.Talents.Toughness > 0 {
		paladin.applyToughness()
	}

	// Blessing of Kings (Tier 3) - Places a Blessing on the friendly target, increasing total stats by 10% for 10 min
	// https://www.wowhead.com/tbc/spell=25898
	// NOTE: This is a trainable spell in TBC if you have the talent point
	if paladin.Talents.BlessingOfKings {
		// Registered in blessings.go when talent is taken
	}

	// Improved Righteous Fury (Tier 3) - While Righteous Fury is active, all damage taken is reduced by 2/4/6%
	if paladin.Talents.ImprovedRighteousFury > 0 {
		paladin.applyImprovedRighteousFury()
	}

	// Shield Specialization (Tier 3) - Increases the amount of damage absorbed by your shield by 10/20/30%
	// TODO: Implement block value increase

	// Anticipation (Tier 4) - Increases your Defense skill by 4/8/12/16/20
	if paladin.Talents.Anticipation > 0 {
		paladin.applyAnticipation()
	}

	// Stoicism (Tier 4) - Increases your resistance to Stun effects by an additional 2/4/6/8/10% and reduces the duration of all Stun effects used against you by 10/20/30%
	// TODO: Implement stun resistance

	// Improved Hammer of Justice (Tier 5) - Decreases the cooldown of your Hammer of Justice spell by 10/20/30 sec
	// TODO: Implement cooldown reduction

	// Improved Concentration Aura (Tier 5) - Increases the effect of your Concentration Aura by an additional 5/10/15% and reduces the duration of all Silence and Interrupt effects used against the group by 10/20/30%
	// TODO: Implement

	// Spell Warding (Tier 5) - All spell damage taken is reduced by 2/4%
	if paladin.Talents.SpellWarding > 0 {
		paladin.applySpellWarding()
	}

	// Blessing of Sanctuary (Tier 6) - When the target blocks, parries, or dodges a melee attack the target will gain 10 rage, 20 runic power, or 2% of maximum mana
	// https://www.wowhead.com/tbc/spell=25899
	if paladin.Talents.BlessingOfSanctuary {
		// Registered in blessings.go when talent is taken
	}

	// Reckoning (Tier 6) - Gives you a 2/4/6/8/10% chance after being hit by any damaging attack that the next 4 weapon swings within 8 sec will generate an additional attack
	if paladin.Talents.Reckoning > 0 {
		paladin.applyReckoning()
	}

	// Sacred Duty (Tier 7) - Increases your total Stamina by 3/6% and reduces the cooldown of your Divine Shield and Divine Protection by 30/60 sec
	if paladin.Talents.SacredDuty > 0 {
		paladin.applySacredDuty()
	}

	// One-Handed Weapon Specialization (Tier 7) - Increases all damage you deal when a one-handed melee weapon is equipped by 1/2/3/4/5%
	if paladin.Talents.OneHandedWeaponSpecialization > 0 {
		paladin.applyOneHandedWeaponSpecialization()
	}

	// Improved Holy Shield (Tier 8) - Increases damage caused by Holy Shield by 10/20% and increases the number of charges by 2/4
	// TODO: Implement when Holy Shield is added

	// Holy Shield (Tier 8) - Increases chance to block by 30% for 10 sec, and deals Holy damage for each attack blocked while active
	// https://www.wowhead.com/tbc/spell=27179
	if paladin.Talents.HolyShield {
		paladin.registerHolyShield()
	}

	// Ardent Defender (Tier 8) - When you have less than 35% health, all damage taken is reduced by 6/12/18/24/30%
	if paladin.Talents.ArdentDefender > 0 {
		paladin.applyArdentDefender()
	}

	// Combat Expertise (Tier 9) - Increases your expertise by 1/2/3/4/5, total Stamina by 2/4/6/8/10% and spell critical strike chance by 1/2/3/4/5%
	if paladin.Talents.CombatExpertise > 0 {
		paladin.applyCombatExpertise()
	}

	// Avenger's Shield (Tier 9) - Hurls a holy shield at the enemy, dealing damage and silencing
	// https://www.wowhead.com/tbc/spell=32700
	if paladin.Talents.AvengersShield {
		paladin.registerAvengersShield()
	}

	// ==================
	// Retribution Talents
	// ==================

	// Improved Blessing of Might (Tier 1) - Increases the melee attack power bonus of your Blessing of Might by 4/8/12/16/20%
	if paladin.Talents.ImprovedBlessingOfMight > 0 {
		paladin.applyImprovedBlessingOfMight()
	}

	// Benediction (Tier 1) - Reduces the mana cost of your Judgement and Seal spells by 3/6/9/12/15%
	if paladin.Talents.Benediction > 0 {
		paladin.applyBenediction()
	}

	// Improved Judgement (Tier 2) - Decreases the cooldown of your Judgement spell by 1/2 sec
	if paladin.Talents.ImprovedJudgement > 0 {
		paladin.applyImprovedJudgement()
	}

	// Improved Seal of the Crusader (Tier 2) - Increases the melee attack power bonus of your Seal of the Crusader and increases the Holy damage bonus of Judgement of the Crusader by 5/10/15%
	if paladin.Talents.ImprovedSealOfTheCrusader > 0 {
		paladin.applyImprovedSealOfTheCrusader()
	}

	// Deflection (Tier 2) - Increases your Parry chance by 1/2/3/4/5%
	if paladin.Talents.Deflection > 0 {
		paladin.applyDeflection()
	}

	// Vindication (Tier 3) - Gives the Paladin's damaging attacks a chance to reduce the target's attributes by 5/10/15% for 10 sec
	// TODO: Implement debuff

	// Conviction (Tier 3) - Increases your chance to get a critical strike with all spells and attacks by 1/2/3/4/5%
	if paladin.Talents.Conviction > 0 {
		paladin.applyConviction()
	}

	// Seal of Command (Tier 3) - Gives the Paladin a chance to deal additional Holy damage
	// https://www.wowhead.com/tbc/spell=20375
	if paladin.Talents.SealOfCommand {
		paladin.registerSealOfCommand()
	}

	// Pursuit of Justice (Tier 4) - Increases movement and mounted movement speed by 5/10%. This does not stack with other movement speed increasing effects
	// TODO: Implement movement speed

	// Eye for an Eye (Tier 4) - All spell criticals against you cause 15/30% of the damage taken to the caster as well
	// TODO: Implement reflect

	// Improved Retribution Aura (Tier 5) - Increases the damage done by your Retribution Aura by 25/50%
	if paladin.Talents.ImprovedRetributionAura > 0 {
		paladin.applyImprovedRetributionAura()
	}

	// Crusade (Tier 5) - Increases all damage caused by 1/2/3% and all damage caused against Humanoids, Demons, Undead and Elementals by an additional 1/2/3%
	if paladin.Talents.Crusade > 0 {
		paladin.applyCrusade()
	}

	// Two-Handed Weapon Specialization (Tier 6) - Increases the damage you deal with two-handed melee weapons by 2/4/6%
	if paladin.Talents.TwoHandedWeaponSpecialization > 0 {
		paladin.applyTwoHandedWeaponSpecialization()
	}

	// Sanctity Aura (Tier 6) - Increases Holy damage done by party members within 30 yards by 10%
	// https://www.wowhead.com/tbc/spell=20218
	if paladin.Talents.SanctityAura {
		paladin.registerSanctityAura()
	}

	// Improved Sanctity Aura (Tier 7) - Increases the damage caused by all party members within 30 yards of the Paladin with Sanctity Aura active by 1/2%
	// Note: This modifies Sanctity Aura if talented
	if paladin.Talents.ImprovedSanctityAura > 0 {
		paladin.applyImprovedSanctityAura()
	}

	// Vengeance (Tier 7) - Gives you a 1/2/3/4/5% bonus to Physical and Holy damage you deal for 15 sec after dealing a critical strike from a weapon swing, spell, or ability
	if paladin.Talents.Vengeance > 0 {
		paladin.applyVengeance()
	}

	// Sanctified Judgement (Tier 8) - Gives your Judgement spell a 33/66/100% chance to return 50% of the mana cost of the Judgement
	if paladin.Talents.SanctifiedJudgement > 0 {
		paladin.applySanctifiedJudgement()
	}

	// Sanctified Seals (Tier 8) - Increases your chance to critically hit with all spells and attacks by 1/2/3% and reduces the chance your Seals will be dispelled by 33/67/100%
	if paladin.Talents.SanctifiedSeals > 0 {
		paladin.applySanctifiedSeals()
	}

	// Repentance (Tier 8) - Puts the enemy target in a state of meditation, incapacitating them for up to 1 min
	// https://www.wowhead.com/tbc/spell=20066
	if paladin.Talents.Repentance {
		paladin.registerRepentance()
	}

	// Divine Purpose (Tier 9) - Reduces your chance to be hit by spells and ranged attacks by 1/2/3%
	// Also decreases the duration of Stun effects by 10/20/30% and gives your Hand of Freedom a 50/100% chance to remove Stun effects
	if paladin.Talents.DivinePurpose > 0 {
		paladin.applyDivinePurposeTalent()
	}

	// Fanaticism (Tier 9) - Increases the critical strike chance of all Judgements capable of a critical hit by 5/10/15/18/25% and reduces threat caused by all actions by 6/12/18/24/30% except when under the effects of Righteous Fury
	if paladin.Talents.Fanaticism > 0 {
		paladin.applyFanaticism()
	}

	// Crusader Strike (Tier 9) - An instant strike that causes weapon damage plus Holy damage
	// https://www.wowhead.com/tbc/spell=35395
	if paladin.Talents.CrusaderStrike {
		paladin.registerCrusaderStrike()
	}
}

// ==================
// Holy Talent Implementations
// ==================

// Divine Strength - Increases your total Strength by 2/4/6/8/10%
func (paladin *Paladin) applyDivineStrength() {
	bonus := 1.0 + 0.02*float64(paladin.Talents.DivineStrength)
	paladin.MultiplyStat(stats.Strength, bonus)
}

// Divine Intellect - Increases your total Intellect by 2/4/6/8/10%
func (paladin *Paladin) applyDivineIntellect() {
	bonus := 1.0 + 0.02*float64(paladin.Talents.DivineIntellect)
	paladin.MultiplyStat(stats.Intellect, bonus)
}

// Improved Seal of Righteousness - Increases the damage done by your Seal of Righteousness and its Judgement by 3/6/9/12/15%
func (paladin *Paladin) applyImprovedSealOfRighteousness() {
	// TODO: Implement damage modifier for SoR and Judgement of Righteousness
}

// Healing Light - Increases the amount healed by your Holy Light and Flash of Light spells by 4/8/12%
func (paladin *Paladin) applyHealingLight() {
	// TODO: Implement healing modifier
}

// Illumination - After getting a critical effect from your Flash of Light, Holy Light, or Holy Shock heal spell, you have a 20/40/60/80/100% chance to gain mana equal to 60% of the base cost of the spell
func (paladin *Paladin) applyIllumination() {
	// TODO: Implement mana return on crit
}

// Improved Blessing of Wisdom - Increases the effect of your Blessing of Wisdom spell by 10/20%
func (paladin *Paladin) applyImprovedBlessingOfWisdom() {
	// TODO: Implement Blessing of Wisdom modifier
}

// Sanctified Light - Increases the critical effect chance of your Holy Light and Holy Shock spells by 2/4/6%
func (paladin *Paladin) applySanctifiedLight() {
	// TODO: Implement crit bonus
}

// Purifying Power - Reduces the mana cost of your Cleanse and Consecration spells by 5/10%, and increases the critical strike chance of your Exorcism and Holy Wrath spells by 10/20%
func (paladin *Paladin) applyPurifyingPower() {
	// TODO: Implement mana cost reduction and crit bonus
}

// Holy Power (talent) - Increases the critical effect chance of your Holy spells by 1/2/3/4/5%
func (paladin *Paladin) applyHolyPowerTalent() {
	// TODO: Implement holy spell crit bonus
}

// Holy Guidance - Increases your spell damage and healing by 7/14/21/28/35% of your total Intellect
func (paladin *Paladin) applyHolyGuidance() {
	// TODO: Implement spell power from intellect
}

// ==================
// Protection Talent Implementations
// ==================

// Improved Devotion Aura - Increases the armor bonus of your Devotion Aura by 8/16/24/32/40%
func (paladin *Paladin) applyImprovedDevotionAura() {
	// TODO: Implement aura modifier
}

// Precision - Increases your chance to hit with melee weapons and spells by 1/2/3%
func (paladin *Paladin) applyPrecision() {
	hitBonus := float64(paladin.Talents.Precision) * core.PhysicalHitRatingPerHitPercent
	paladin.AddStat(stats.MeleeHitRating, hitBonus)
	paladin.AddStat(stats.SpellHitRating, hitBonus)
}

// Toughness - Increases your armor value from items by 2/4/6/8/10%
func (paladin *Paladin) applyToughness() {
	// TODO: Implement armor modifier
}

// Improved Righteous Fury - While Righteous Fury is active, all damage taken is reduced by 2/4/6%
func (paladin *Paladin) applyImprovedRighteousFury() {
	// TODO: Implement damage reduction
}

// Anticipation - Increases your Defense skill by 4/8/12/16/20
func (paladin *Paladin) applyAnticipation() {
	defenseBonus := float64(paladin.Talents.Anticipation) * 4
	paladin.AddStat(stats.DefenseRating, defenseBonus)
}

// Spell Warding - All spell damage taken is reduced by 2/4%
func (paladin *Paladin) applySpellWarding() {
	// TODO: Implement spell damage reduction
}

// Reckoning - Gives you a 2/4/6/8/10% chance after being hit by any damaging attack that the next 4 weapon swings within 8 sec will generate an additional attack
func (paladin *Paladin) applyReckoning() {
	// TODO: Implement reckoning proc
}

// Sacred Duty - Increases your total Stamina by 3/6% and reduces the cooldown of your Divine Shield and Divine Protection by 30/60 sec
func (paladin *Paladin) applySacredDuty() {
	bonus := 1.0 + 0.03*float64(paladin.Talents.SacredDuty)
	paladin.MultiplyStat(stats.Stamina, bonus)
	// TODO: Implement cooldown reduction
}

// One-Handed Weapon Specialization - Increases all damage you deal when a one-handed melee weapon is equipped by 1/2/3/4/5%
func (paladin *Paladin) applyOneHandedWeaponSpecialization() {
	// TODO: Implement damage modifier for one-handed weapons
}

// Ardent Defender - When you have less than 35% health, all damage taken is reduced by 6/12/18/24/30%
func (paladin *Paladin) applyArdentDefender() {
	// TODO: Implement low health damage reduction
}

// Combat Expertise - Increases your expertise by 1/2/3/4/5, total Stamina by 2/4/6/8/10% and spell critical strike chance by 1/2/3/4/5%
func (paladin *Paladin) applyCombatExpertise() {
	expertiseBonus := float64(paladin.Talents.CombatExpertise)
	paladin.AddStat(stats.ExpertiseRating, expertiseBonus*core.ExpertisePerQuarterPercentReduction*4)

	staminaBonus := 1.0 + 0.02*float64(paladin.Talents.CombatExpertise)
	paladin.MultiplyStat(stats.Stamina, staminaBonus)

	spellCritBonus := float64(paladin.Talents.CombatExpertise) * core.SpellCritRatingPerCritPercent
	paladin.AddStat(stats.SpellCritRating, spellCritBonus)
}

// ==================
// Retribution Talent Implementations
// ==================

// Improved Blessing of Might - Increases the melee attack power bonus of your Blessing of Might by 4/8/12/16/20%
func (paladin *Paladin) applyImprovedBlessingOfMight() {
	// TODO: Implement Blessing of Might modifier
}

// Benediction - Reduces the mana cost of your Judgement and Seal spells by 3/6/9/12/15%
func (paladin *Paladin) applyBenediction() {
	// TODO: Implement mana cost reduction
}

// Improved Judgement - Decreases the cooldown of your Judgement spell by 1/2 sec
func (paladin *Paladin) applyImprovedJudgement() {
	// TODO: Implement cooldown reduction
}

// Improved Seal of the Crusader - Increases the melee attack power bonus of your Seal of the Crusader and increases the Holy damage bonus of Judgement of the Crusader by 5/10/15%
func (paladin *Paladin) applyImprovedSealOfTheCrusader() {
	// TODO: Implement damage modifier
}

// Deflection - Increases your Parry chance by 1/2/3/4/5%
func (paladin *Paladin) applyDeflection() {
	parryBonus := float64(paladin.Talents.Deflection) * core.ParryRatingPerParryPercent
	paladin.AddStat(stats.ParryRating, parryBonus)
}

// Conviction - Increases your chance to get a critical strike with all spells and attacks by 1/2/3/4/5%
func (paladin *Paladin) applyConviction() {
	critBonus := float64(paladin.Talents.Conviction) * core.PhysicalCritRatingPerCritPercent
	paladin.AddStat(stats.MeleeCritRating, critBonus)
	paladin.AddStat(stats.SpellCritRating, critBonus)
}

// Improved Retribution Aura - Increases the damage done by your Retribution Aura by 25/50%
func (paladin *Paladin) applyImprovedRetributionAura() {
	// TODO: Implement aura modifier
}

// Crusade - Increases all damage caused by 1/2/3% and all damage caused against Humanoids, Demons, Undead and Elementals by an additional 1/2/3%
func (paladin *Paladin) applyCrusade() {
	// TODO: Implement damage modifiers
}

// Two-Handed Weapon Specialization - Increases the damage you deal with two-handed melee weapons by 2/4/6%
func (paladin *Paladin) applyTwoHandedWeaponSpecialization() {
	// TODO: Implement damage modifier for two-handed weapons
}

// Improved Sanctity Aura - Increases the damage caused by all party members within 30 yards of the Paladin with Sanctity Aura active by 1/2%
func (paladin *Paladin) applyImprovedSanctityAura() {
	// TODO: Implement aura modifier (combined with Sanctity Aura registration)
}

// Vengeance - Gives you a 1/2/3/4/5% bonus to Physical and Holy damage you deal for 15 sec after dealing a critical strike from a weapon swing, spell, or ability
func (paladin *Paladin) applyVengeance() {
	// TODO: Implement vengeance proc
}

// Sanctified Judgement - Gives your Judgement spell a 33/66/100% chance to return 50% of the mana cost of the Judgement
func (paladin *Paladin) applySanctifiedJudgement() {
	// TODO: Implement mana return
}

// Sanctified Seals - Increases your chance to critically hit with all spells and attacks by 1/2/3% and reduces the chance your Seals will be dispelled by 33/67/100%
func (paladin *Paladin) applySanctifiedSeals() {
	critBonus := float64(paladin.Talents.SanctifiedSeals) * core.PhysicalCritRatingPerCritPercent
	paladin.AddStat(stats.MeleeCritRating, critBonus)
	paladin.AddStat(stats.SpellCritRating, critBonus)
}

// Divine Purpose (talent) - Reduces your chance to be hit by spells and ranged attacks by 1/2/3%
func (paladin *Paladin) applyDivinePurposeTalent() {
	// TODO: Implement spell hit reduction
}

// Fanaticism - Increases the critical strike chance of all Judgements capable of a critical hit by 5/10/15/18/25% and reduces threat caused by all actions by 6/12/18/24/30%
func (paladin *Paladin) applyFanaticism() {
	// TODO: Implement crit bonus and threat reduction
}
