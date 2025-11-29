package tbc

import (
	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

func RegisterAllProcs() {

	// Procs

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Increases healing done by up to 43 and damage done by up to 14 for all magical spells and effects.
	// shared.NewProcStatBonusEffect(shared.ProcStatBonusEffect{
	//	Name:               "Eye of Gruul",
	//	ItemID:             28823,
	//	Callback:           core.CallbackOnHealDealt,
	//	ProcMask:           core.ProcMaskSpellHealing,
	//	Outcome:            core.OutcomeLanded,
	//	RequireDamageDealt: false
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Increases your pet's resistances by 129 and increases your spell damage by up to 47.
	// shared.NewProcStatBonusEffect(shared.ProcStatBonusEffect{
	//	Name:               "Void Star Talisman",
	//	ItemID:             30449,
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskEmpty,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// })

	// Each time you cast a spell, there is chance you will gain up to 289 spell damage and healing.
	shared.NewProcStatBonusEffect(shared.ProcStatBonusEffect{
		Name:               "Tome of Fiery Redemption",
		ItemID:             30447,
		Callback:           core.CallbackOnSpellHitDealt | core.CallbackOnHealDealt,
		ProcMask:           core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeOHSpecial | core.ProcMaskRangedSpecial | core.ProcMaskSpellDamage | core.ProcMaskSpellHealing | core.ProcMaskMeleeProc | core.ProcMaskRangedProc | core.ProcMaskSpellDamageProc,
		Outcome:            core.OutcomeLanded,
		RequireDamageDealt: false,
	})

	// Your special attacks have a chance to give you 1001 armor penetration for 15s.
	shared.NewProcStatBonusEffect(shared.ProcStatBonusEffect{
		Name:               "Warp-Spring Coil",
		ItemID:             30450,
		Callback:           core.CallbackOnSpellHitDealt,
		ProcMask:           core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeOHSpecial | core.ProcMaskMeleeProc,
		Outcome:            core.OutcomeLanded,
		RequireDamageDealt: true,
	})

	// Your Nature spells have a chance to restore 334 mana.
	shared.NewProcStatBonusEffect(shared.ProcStatBonusEffect{
		Name:               "Fathom-Brooch of the Tidewalker",
		ItemID:             30663,
		Callback:           core.CallbackOnSpellHitDealt,
		ProcMask:           core.ProcMaskSpellDamage,
		Outcome:            core.OutcomeLanded,
		RequireDamageDealt: false,
	})

	// Your Mortal Strike, Bloodthirst, and Shield Slam attacks have a 25% chance to heal you for 329 and grant
	// 54 Strength for 12s.
	shared.NewProcStatBonusEffect(shared.ProcStatBonusEffect{
		Name:               "Ashtongue Talisman of Valor",
		ItemID:             32485,
		Callback:           core.CallbackOnSpellHitDealt,
		ProcMask:           core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeOHSpecial,
		Outcome:            core.OutcomeLanded,
		RequireDamageDealt: true,
	})

	// Your Steady Shot has a 15% chance to grant you 274 attack power for 8s.
	shared.NewProcStatBonusEffect(shared.ProcStatBonusEffect{
		Name:               "Ashtongue Talisman of Swiftness",
		ItemID:             32487,
		Callback:           core.CallbackOnSpellHitDealt,
		ProcMask:           core.ProcMaskRangedSpecial,
		Outcome:            core.OutcomeLanded,
		RequireDamageDealt: true,
	})

	// Your spell critical strikes have a 50% chance to grant you 144 spell haste rating for 5s.
	shared.NewProcStatBonusEffect(shared.ProcStatBonusEffect{
		Name:               "Ashtongue Talisman of Insight",
		ItemID:             32488,
		Callback:           core.CallbackOnSpellHitDealt,
		ProcMask:           core.ProcMaskSpellDamage,
		Outcome:            core.OutcomeCrit,
		RequireDamageDealt: false,
	})

	// Each time your Corruption deals damage, it has a 20% chance to grant you 219 spell damage for 5s.
	shared.NewProcStatBonusEffect(shared.ProcStatBonusEffect{
		Name:               "Ashtongue Talisman of Shadows",
		ItemID:             32493,
		Callback:           core.CallbackOnPeriodicDamageDealt,
		ProcMask:           core.ProcMaskSpellDamage,
		Outcome:            core.OutcomeLanded,
		RequireDamageDealt: true,
	})
}
