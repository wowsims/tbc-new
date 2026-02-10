# TBC Paladin Spell Implementation Checklist

Reference: [Wowhead TBC Paladin Abilities](https://www.wowhead.com/tbc/spells/abilities/paladin)

## Implementation Status Legend
- ✅ **Fully Implemented** - Spell is complete with damage/healing calculations and registered
- 🔶 **Stub/TODO** - Function exists & is called, but effect body is TODO
- ⚠️ **Not Wired** - Code exists in file but registration call is commented out or not called
- ❌ **Not Implemented** - No code exists in the codebase

### Registration Note
`registerSpells()` in `paladin.go` currently calls:
- `registerJudgement()`, `registerConsecration()`, `registerHammerOfWrath()`, `registerHolyWrath()`, `registerExorcism()`
- `registerSeals()`, `registerHealingSpells()`
- ~~`registerAuras()`~~ (commented out)
- ~~`registerBlessings()`~~ (commented out)

Talent abilities are registered via `registerTalentSpells()` (called at the start of `ApplyTalents()`) in `talents.go`.

---

## Seals (seals.go)

| Spell | Status | Notes |
|-------|--------|-------|
| Seal of Righteousness | ✅ | All ranks (1-9) with proc and judgement |
| Seal of Light | ✅ | All ranks with healing proc and JoL debuff |
| Seal of Wisdom | ✅ | All ranks with mana restore and JoW debuff |
| Seal of Justice | ✅ | All ranks with stun chance proc |
| Seal of the Crusader | ✅ | All ranks with AP buff, 1.4x attack speed, auto-attack damage reduction |
| Seal of Command | ✅ | Talent - All ranks with PPM proc |
| Seal of Blood | ✅ | Horde only - Implemented with self-damage |
| Seal of Vengeance | ✅ | Alliance only - Implemented with Holy Vengeance DoT stacking |

---

## Judgements (judgement.go + seals.go)

| Spell | Status | Notes |
|-------|--------|-------|
| Judgement (base spell) | ✅ | Core mechanic with seal twist support |
| Judgement of Righteousness | ✅ | Holy damage with spell batching |
| Judgement of Light | ✅ | Applies debuff for heal on hit |
| Judgement of Wisdom | ✅ | Applies debuff for mana on hit |
| Judgement of Justice | ✅ | Applies anti-flee debuff |
| Judgement of the Crusader | 🔶 | Debuff registered, but holy damage taken bonus on gain/expire is commented out |
| Judgement of Command | ✅ | Holy damage with spell batching |
| Judgement of Blood | ✅ | Holy damage with self-damage cost |
| Judgement of Vengeance | ✅ | Damage based on Holy Vengeance stacks |

---

## Healing Spells (healing.go)

| Spell | Status | Notes |
|-------|--------|-------|
| Holy Light | ✅ | All ranks (1-11) with scaling |
| Flash of Light | ✅ | All ranks (1-7) with scaling |
| Lay on Hands | ✅ | All ranks (1-4), drains caster mana, heals for max health |

---

## Offensive Abilities

| Spell | Status | File | Notes |
|-------|--------|------|-------|
| Consecration | ✅ | consecration.go | All ranks (1-6) with AoE DoT |
| Exorcism | ✅ | exorcism.go | All ranks (1-7) with Undead/Demon restriction, scaling, crit |
| Hammer of Wrath | ✅ | hammer_of_wrath.go | All ranks (1-4) with execute phase, scaling, crit |
| Holy Wrath | ✅ | holy_wrath.go | All ranks (1-3) with AoE vs Undead/Demons |
| Hammer of Justice | ⚠️ | abilities.go | Empty stub, **not called** from `registerSpells()` |

---

## Defensive/Utility Abilities

| Spell | Status | File | Notes |
|-------|--------|------|-------|
| Divine Shield | ⚠️ | abilities.go | Empty stub, **not called** from `registerSpells()` |
| Divine Protection | ⚠️ | abilities.go | Empty stub, **not called** from `registerSpells()` |
| Cleanse | ⚠️ | abilities.go | Empty stub, **not called** from `registerSpells()` |
| Forbearance | ⚠️ | forbearance.go | Aura defined, **not called** from `registerSpells()` |
| Righteous Fury | ❌ | - | Threat increase - NOT IMPLEMENTED |
| Purify | ❌ | - | Disease/Poison dispel - NOT IMPLEMENTED |
| Turn Undead / Turn Evil | ❌ | - | Fear undead - NOT IMPLEMENTED |
| Divine Intervention | ❌ | - | Party protection - NOT IMPLEMENTED |
| Righteous Defense | ❌ | - | Taunt - NOT IMPLEMENTED |

---

## Blessings (blessings.go) — ⚠️ `registerBlessings()` commented out in `registerSpells()`

| Spell | Status | Notes |
|-------|--------|-------|
| Blessing of Might | ⚠️ | Code exists, TODO buff application |
| Blessing of Wisdom | ⚠️ | Code exists, TODO mana regen buff |
| Blessing of Kings | ⚠️ | Talent - Code exists, TODO 10% stats buff |
| Blessing of Salvation | ⚠️ | Code exists, TODO threat reduction |
| Blessing of Sanctuary | ⚠️ | Talent - Code exists, TODO damage reduction |
| Blessing of Protection | ⚠️ | Code exists, TODO physical immunity + Forbearance |
| Blessing of Light | ❌ | Healing taken buff - NOT IMPLEMENTED |
| Blessing of Freedom | ❌ | Movement immunity - NOT IMPLEMENTED |
| Blessing of Sacrifice | ❌ | Damage transfer - NOT IMPLEMENTED |

---

## Auras (auras.go) — ⚠️ `registerAuras()` commented out in `registerSpells()`

| Spell | Status | Notes |
|-------|--------|-------|
| Devotion Aura | ⚠️ | Code exists, TODO armor buff activation |
| Retribution Aura | ⚠️ | Code exists, TODO damage reflect activation |
| Concentration Aura | ⚠️ | Code exists, TODO pushback resistance |
| Fire Resistance Aura | ⚠️ | Code exists, TODO resistance buff |
| Frost Resistance Aura | ⚠️ | Code exists, TODO resistance buff |
| Shadow Resistance Aura | ⚠️ | Code exists, TODO resistance buff |
| Sanctity Aura | 🔶 | Talent - Registered via `registerTalentSpells()`, but aura effect TODO |
| Crusader Aura | ❌ | Mounted speed - NOT IMPLEMENTED (low priority) |

---

## Talent Abilities

### Holy Tree
| Spell | Status | File | Notes |
|-------|--------|------|-------|
| Divine Favor | ✅ | divine_favor.go | 100% crit on next heal, fully working |
| Holy Shock | ✅ | holy_shock.go | All ranks, damage/healing dual-use |
| Divine Illumination | 🔶 | divine_illumination.go | Empty body, 50% mana reduction TODO |

### Protection Tree
| Spell | Status | File | Notes |
|-------|--------|------|-------|
| Holy Shield | ✅ | holy_shield.go | All ranks (1-4), block chance, proc damage, charges, threat |
| Avenger's Shield | ✅ | avengers_shield.go | All ranks (1-3), multi-target bounce, damage + coeff |

### Retribution Tree
| Spell | Status | File | Notes |
|-------|--------|------|-------|
| Crusader Strike | ✅ | crusader_strike.go | 110% weapon damage, normalized, working |
| Seal of Command | ✅ | seals.go | PPM proc system working |
| Repentance | 🔶 | abilities.go | Empty body, incapacitate TODO |

---

## Talent Passive Effects (talents.go)

### Holy Talents
| Talent | Status | Notes |
|--------|--------|-------|
| Divine Strength | ✅ | Strength % multiplier |
| Divine Intellect | ✅ | Intellect % multiplier |
| Spiritual Focus | ❌ | Pushback resistance - comment only, no stub |
| Improved Seal of Righteousness | 🔶 | Stub exists, TODO damage modifier |
| Healing Light | 🔶 | Stub exists, TODO healing modifier |
| Aura Mastery | ❌ | Aura range increase - comment only |
| Improved Lay on Hands | ❌ | Armor bonus + CD reduction - comment only |
| Unyielding Faith | ❌ | Fear/Disorient resistance - comment only |
| Illumination | 🔶 | Stub exists, TODO mana return on crit heal |
| Improved Blessing of Wisdom | 🔶 | Stub exists, TODO BoW modifier |
| Pure of Heart | ❌ | Curse/Disease resistance - comment only |
| Sanctified Light | 🔶 | Stub exists, TODO crit bonus for HL/HS |
| Purifying Power | 🔶 | Stub exists, TODO mana reduction + crit bonus |
| Holy Power | 🔶 | Stub exists, TODO holy spell crit bonus |
| Light's Grace | ❌ | HL cast time reduction - comment only |
| Blessed Life | ❌ | Damage reduction chance - comment only |
| Holy Guidance | 🔶 | Stub exists, TODO spell power from INT |

### Protection Talents
| Talent | Status | Notes |
|--------|--------|-------|
| Improved Devotion Aura | 🔶 | Stub exists, TODO aura modifier |
| Redoubt | ❌ | Block chance proc - comment only |
| Precision | ✅ | Melee + spell hit rating |
| Guardian's Favor | ❌ | BoP CD reduction - comment only |
| Toughness | 🔶 | Stub exists, TODO armor modifier |
| Improved Righteous Fury | 🔶 | Stub exists, TODO damage reduction |
| Shield Specialization | ❌ | Block value increase - comment only |
| Anticipation | ✅ | Defense rating |
| Stoicism | ❌ | Stun resistance - comment only |
| Improved Hammer of Justice | ❌ | HoJ CD reduction - comment only |
| Improved Concentration Aura | ❌ | Aura modifier - comment only |
| Spell Warding | 🔶 | Stub exists, TODO spell damage reduction |
| Reckoning | 🔶 | Stub exists, TODO extra attack proc |
| Sacred Duty | ✅ | Stamina % multiplier done; CD reduction TODO |
| One-Handed Weapon Spec | 🔶 | Stub exists, TODO damage modifier |
| Improved Holy Shield | ❌ | HS damage + charges - comment only |
| Ardent Defender | 🔶 | Stub exists, TODO low-health damage reduction |
| Combat Expertise | ✅ | Expertise + stamina + spell crit |

### Retribution Talents
| Talent | Status | Notes |
|--------|--------|-------|
| Improved Blessing of Might | 🔶 | Stub exists, TODO BoM modifier |
| Benediction | 🔶 | Stub exists, TODO mana cost reduction |
| Improved Judgement | 🔶 | Stub exists, TODO CD reduction |
| Improved Seal of the Crusader | 🔶 | Stub exists, TODO damage modifier |
| Deflection | ✅ | Parry rating |
| Vindication | ❌ | Target attribute debuff - comment only |
| Conviction | ✅ | Melee + spell crit rating |
| Pursuit of Justice | ❌ | Movement speed - comment only |
| Eye for an Eye | ❌ | Spell crit reflect - comment only |
| Improved Retribution Aura | 🔶 | Stub exists, TODO aura modifier |
| Crusade | 🔶 | Stub exists, TODO damage modifiers |
| Two-Handed Weapon Spec | 🔶 | Stub exists, TODO damage modifier |
| Improved Sanctity Aura | 🔶 | Stub exists, TODO aura modifier |
| Vengeance | 🔶 | Stub exists, TODO crit proc damage buff |
| Sanctified Judgement | 🔶 | Stub exists, TODO mana return |
| Sanctified Seals | ✅ | Crit bonus |
| Divine Purpose | 🔶 | Stub exists, TODO spell hit reduction |
| Fanaticism | 🔶 | Stub exists, TODO crit bonus + threat reduction |

---

## Summary Statistics

| Category | ✅ Implemented | 🔶 Stub/TODO | ⚠️ Not Wired | ❌ Missing |
|----------|---------------|-------------|-------------|-----------|
| Seals | 8 | 0 | 0 | 0 |
| Judgements | 8 | 1 | 0 | 0 |
| Healing | 3 | 0 | 0 | 0 |
| Offensive | 4 | 0 | 1 | 0 |
| Defensive/Utility | 0 | 0 | 4 | 5 |
| Blessings | 0 | 0 | 6 | 3 |
| Auras | 0 | 1 | 6 | 1 |
| Talent Abilities | 6 | 2 | 0 | 0 |
| Talent Passives (Holy) | 2 | 9 | 0 | 6 |
| Talent Passives (Prot) | 4 | 7 | 0 | 6 |
| Talent Passives (Ret) | 3 | 10 | 0 | 3 |
| **TOTAL** | **38** | **30** | **17** | **24** |

---

## Priority Implementation Order (Suggested)

### 🔴 High Priority — DPS/Tanking Core (missing effects on registered spells)

These are spells/talents that are already registered but have TODO effects that directly impact DPS/tanking sim accuracy:

1. [ ] **Judgement of the Crusader** — Implement holy damage taken debuff on gain/expire (seals.go)
2. [ ] **Vengeance** (talent) — 5% damage buff after crit, core Ret DPS talent
3. [ ] **Crusade** (talent) — Up to 6% damage increase, core Ret DPS talent
4. [ ] **Two-Handed Weapon Spec** (talent) — 6% 2H damage increase
5. [ ] **Fanaticism** (talent) — 25% Judgement crit + 30% threat reduction
6. [ ] **Improved Seal of Righteousness** (talent) — 15% SoR damage increase
7. [ ] **Benediction** (talent) — 15% Seal/Judgement mana cost reduction
8. [ ] **Improved Judgement** (talent) — 2 sec Judgement CD reduction
9. [ ] **Sanctified Judgement** (talent) — Mana return on Judgement
10. [ ] **Improved Seal of the Crusader** (talent) — 15% SotC AP/JotC bonus

### 🟡 Medium Priority — Tanking/Healing Accuracy

11. [ ] **Improved Holy Shield** — +20% HS damage and +4 charges
12. [ ] **Reckoning** (talent) — Extra attack on being crit
13. [ ] **One-Handed Weapon Spec** (talent) — 5% damage with 1H
14. [ ] **Ardent Defender** (talent) — Sub-35% damage reduction
15. [ ] **Spell Warding** (talent) — Spell damage reduction
16. [ ] **Illumination** (talent) — Mana return on heal crit (healer core)
17. [ ] **Holy Guidance** (talent) — SP from Intellect (healer core)
18. [ ] **Healing Light** (talent) — 12% more HL/FoL healing
19. [ ] **Sanctified Light** (talent) — 6% HL/HS crit bonus
20. [ ] **Divine Illumination** — 50% mana cost reduction activation
21. [ ] **Repentance** — Incapacitate effect implementation

### 🟢 Low Priority — Buffs/Auras/Utility

22. [ ] **Wire up Auras** — Uncomment `registerAuras()` and implement buff effects
23. [ ] **Wire up Blessings** — Uncomment `registerBlessings()` and implement buff effects
24. [ ] **Sanctity Aura** — 10% holy damage party buff
25. [ ] **Blessing of Might / Kings / Wisdom** — Core raid buffs
26. [ ] **Wire up abilities.go spells** — Call Hammer of Justice, Divine Shield, etc. from `registerSpells()`
27. [ ] **Righteous Fury** — Tank threat modifier
28. [ ] **Forbearance** — Wire up to Divine Shield / BoP

### ⚪ Very Low Priority — Situational/Non-Sim

29. [ ] Resistance Auras, CC abilities (Repentance, Turn Undead), Purify
30. [ ] Blessing of Light, Freedom, Sacrifice
31. [ ] Crusader Aura, Divine Intervention, Righteous Defense
32. [ ] Redemption, Sense Undead (not needed for sim)

---

## Recent Changes (since last update)

- ✅ **Seal of the Crusader** — AP buff, attack speed modifier, and auto-attack damage reduction are all working via `AttachMultiplyMeleeSpeed`, `AttachMultiplicativePseudoStatBuff`, and `AttachStatBuff` (was 🔶)
- ✅ **Exorcism** — `registerExorcism()` is now called from `registerSpells()` (was ⚠️)
- 🔧 **Refactored `ApplyTalents()`** — All talent spell registrations moved to new `registerTalentSpells()` method, called before passive talent applications
- 📝 Fixed **Divine Illumination** file reference from `abilities.go` to `divine_illumination.go`
- 📝 Updated Registration Note to reflect current `registerSpells()` calls

### Previous Changes

- ✅ **Holy Wrath** — Fully implemented in `holy_wrath.go` (was ❌)
- ✅ **Hammer of Wrath** — Fully implemented in `hammer_of_wrath.go` (was 🔶)
- ✅ **Holy Shield** — Fully implemented in `holy_shield.go` with block, charges, proc damage (was 🔶)
- ✅ **Avenger's Shield** — Fully implemented in `avengers_shield.go` with multi-target bounce (was 🔶)
- ⚠️ **Forbearance** — New file `forbearance.go` with aura definition (not yet wired)
- 📝 Added ⚠️ status to distinguish "code exists but not registered" from "registered but TODO"
- 📝 Added comprehensive **Talent Passive Effects** section tracking all talent stubs

---

*Last Updated: 2026-02-10*
