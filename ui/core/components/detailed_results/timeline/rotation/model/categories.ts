import { OtherAction } from '../../../../../proto/common';
import type { ActionId } from '../../../../../proto_utils/action_id';

export const MELEE_ACTION_CATEGORY = 1;
export const SPELL_ACTION_CATEGORY = 2;
export const DEFAULT_ACTION_CATEGORY = 3;

// A fixed sort category for casts whose default MELEE/SPELL/DEFAULT bucket would order them
// badly. Audited 2026-09-04 against five signals - a spellId in a committed APL, a `SpellID:`
// literal in sim/**/*.go, an id the client DB's spellIcons knows, a reference in ui/*.ts(x), and
// one in a committed ui/ JSON - which dropped 206 of the 242 inherited entries as unreachable in
// TBC. Most were later-expansion ids for spells TBC registers under a rank id instead (Arcane
// Shot is 27019 here, not 3044), plus every Death Knight and Monk entry, for classes TBC's Class
// enum does not have. Flat Record<number, number> keyed by anyId(), so an OtherAction id and a
// spellId of the same numeric value are indistinguishable; safe only because OtherAction values
// are small and spell ids are large.
export const rotationCategoryOverrides: Record<number, number> = {
	[OtherAction.OtherActionMove]: 0,
	[OtherAction.OtherActionAttack]: 0.01,
	[OtherAction.OtherActionShoot]: 0.5,

	// Hunter
	[34490]: MELEE_ACTION_CATEGORY + 0.29, // Silencing Shot

	// Paladin
	[35395]: MELEE_ACTION_CATEGORY + 0.02, // Crusader Strike
	[20271]: MELEE_ACTION_CATEGORY + 0.08, // Judgment
	[42463]: MELEE_ACTION_CATEGORY + 0.09, // Seal of Truth (on-hit)
	[31803]: MELEE_ACTION_CATEGORY + 0.1, // Censure (Seal of Truth)
	[879]: MELEE_ACTION_CATEGORY + 0.15, // Exorcism
	[26573]: MELEE_ACTION_CATEGORY + 0.16, // Consecration
	[24275]: MELEE_ACTION_CATEGORY + 0.18, // Hammer of Wrath
	[31884]: SPELL_ACTION_CATEGORY + 0.06, // Avenging Wrath
	[20925]: SPELL_ACTION_CATEGORY + 0.11, // Sacred Shield (Ret / Prot)

	// Rogue
	[6774]: MELEE_ACTION_CATEGORY + 0.1, // Slice and Dice
	[8647]: MELEE_ACTION_CATEGORY + 0.2, // Expose Armor

	// Shaman
	[8232]: 0.11, // Windfury Weapon
	[8024]: 0.12, // Flametongue Weapon
	[8033]: 0.12, // Frostbrand Weapon
	[17364]: MELEE_ACTION_CATEGORY + 0.1, // Stormstrike
	[2825]: DEFAULT_ACTION_CATEGORY + 0.1, // Bloodlust

	// Warlock
	[603]: SPELL_ACTION_CATEGORY + 0.01, // Curse of Doom
	[30108]: SPELL_ACTION_CATEGORY + 0.3, // Unstable Affliction
	[17962]: SPELL_ACTION_CATEGORY + 0.32, // Conflagrate

	// Mage
	[12472]: SPELL_ACTION_CATEGORY + 0.4, // Icy Veins
	[11129]: SPELL_ACTION_CATEGORY + 0.4, // Combustion
	[12042]: SPELL_ACTION_CATEGORY + 0.4, // Arcane Power
	[11958]: SPELL_ACTION_CATEGORY + 0.41, // Cold Snap
	[12043]: SPELL_ACTION_CATEGORY + 0.41, // Presence of Mind
	[31687]: SPELL_ACTION_CATEGORY + 0.41, // Water Elemental
	[12051]: SPELL_ACTION_CATEGORY + 0.52, // Evocate
	[12536]: SPELL_ACTION_CATEGORY + 0.61, // Clearcasting

	// Warrior
	[23881]: MELEE_ACTION_CATEGORY + 0.1, // Bloodthirst
	[30356]: MELEE_ACTION_CATEGORY + 0.1, // Shield Slam
	[1680]: MELEE_ACTION_CATEGORY + 0.24, // Whirlwind
	[12867]: SPELL_ACTION_CATEGORY + 0.51, // Deep Wounds
	[2565]: SPELL_ACTION_CATEGORY + 0.62, // Shield Block
	[71]: DEFAULT_ACTION_CATEGORY + 0.1, // Defensive Stance
	[2457]: DEFAULT_ACTION_CATEGORY + 0.1, // Battle Stance
	[469]: DEFAULT_ACTION_CATEGORY + 0.1, // Commanding Shout
};

export function actionCategory(actionId: ActionId, meleeKeys: ReadonlySet<string>, spellKeys: ReadonlySet<string>): number {
	const fixedCategory = rotationCategoryOverrides[actionId.anyId()];
	if (fixedCategory != null) return fixedCategory;
	const key = actionId.equalityKey();
	if (meleeKeys.has(key)) return MELEE_ACTION_CATEGORY;
	if (spellKeys.has(key)) return SPELL_ACTION_CATEGORY;
	return DEFAULT_ACTION_CATEGORY;
}
