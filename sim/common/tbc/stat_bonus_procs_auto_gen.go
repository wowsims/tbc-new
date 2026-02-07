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
	// Increases your pet's resistances by 130 and increases your spell damage by up to 48.
	// shared.NewProcStatBonusEffect(shared.ProcStatBonusEffect{
	//	Name:               "Void Star Talisman",
	//	ItemID:             30449,
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskEmpty,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// })

	// Reduces an enemy's armor by 200. Stacks up to 3 times.
	shared.NewStackingStatBonusEffect(shared.StackingStatBonusEffect{
		Name:               "Annihilator",
		ItemID:             12798,
		MaxStacks:          3,
		Callback:           core.CallbackOnSpellHitDealt,
		ProcMask:           core.ProcMaskMeleeMHAuto | core.ProcMaskMeleeOHAuto | core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeOHSpecial,
		Outcome:            core.OutcomeLanded,
		RequireDamageDealt: true,
	})

	// Increases Strength by 100 for 10s.
	shared.NewProcStatBonusEffect(shared.ProcStatBonusEffect{
		Name:               "Lionheart Champion",
		ItemID:             28429,
		Callback:           core.CallbackOnSpellHitDealt,
		ProcMask:           core.ProcMaskMeleeMHAuto | core.ProcMaskMeleeOHAuto | core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeOHSpecial,
		Outcome:            core.OutcomeLanded,
		RequireDamageDealt: true,
	})

	// Increases Strength by 100 for 10s.
	shared.NewProcStatBonusEffect(shared.ProcStatBonusEffect{
		Name:               "Lionheart Executioner",
		ItemID:             28430,
		Callback:           core.CallbackOnSpellHitDealt,
		ProcMask:           core.ProcMaskMeleeMHAuto | core.ProcMaskMeleeOHAuto | core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeOHSpecial,
		Outcome:            core.OutcomeLanded,
		RequireDamageDealt: true,
	})

	// Increases your haste rating by 212 for 10s.
	shared.NewProcStatBonusEffect(shared.ProcStatBonusEffect{
		Name:               "Drakefist Hammer",
		ItemID:             28437,
		Callback:           core.CallbackOnSpellHitDealt,
		ProcMask:           core.ProcMaskMeleeMHAuto | core.ProcMaskMeleeOHAuto | core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeOHSpecial,
		Outcome:            core.OutcomeLanded,
		RequireDamageDealt: true,
	})

	// Increases your haste rating by 212 for 10s.
	shared.NewProcStatBonusEffect(shared.ProcStatBonusEffect{
		Name:               "Dragonmaw",
		ItemID:             28438,
		Callback:           core.CallbackOnSpellHitDealt,
		ProcMask:           core.ProcMaskMeleeMHAuto | core.ProcMaskMeleeOHAuto | core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeOHSpecial,
		Outcome:            core.OutcomeLanded,
		RequireDamageDealt: true,
	})

	// Increases your haste rating by 212 for 10s.
	shared.NewProcStatBonusEffect(shared.ProcStatBonusEffect{
		Name:               "Dragonstrike",
		ItemID:             28439,
		Callback:           core.CallbackOnSpellHitDealt,
		ProcMask:           core.ProcMaskMeleeMHAuto | core.ProcMaskMeleeOHAuto | core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeOHSpecial,
		Outcome:            core.OutcomeLanded,
		RequireDamageDealt: true,
	})

	// Increases your haste rating by 180 for 10s.
	shared.NewProcStatBonusEffect(shared.ProcStatBonusEffect{
		Name:               "The Bladefist",
		ItemID:             29348,
		Callback:           core.CallbackOnSpellHitDealt,
		ProcMask:           core.ProcMaskMeleeMHAuto | core.ProcMaskMeleeOHAuto | core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeOHSpecial,
		Outcome:            core.OutcomeLanded,
		RequireDamageDealt: true,
	})

	// Increases attack power by 270 for 10s.
	shared.NewProcStatBonusEffect(shared.ProcStatBonusEffect{
		Name:               "Heartrazor",
		ItemID:             29962,
		Callback:           core.CallbackOnSpellHitDealt,
		ProcMask:           core.ProcMaskMeleeMHAuto | core.ProcMaskMeleeOHAuto | core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeOHSpecial,
		Outcome:            core.OutcomeLanded,
		RequireDamageDealt: true,
	})

	// Each time you cast a spell, there is chance you will gain up to 290 spell damage and healing.
	shared.NewProcStatBonusEffect(shared.ProcStatBonusEffect{
		Name:               "Tome of Fiery Redemption",
		ItemID:             30447,
		Callback:           core.CallbackOnSpellHitDealt | core.CallbackOnHealDealt,
		ProcMask:           core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeOHSpecial | core.ProcMaskRangedSpecial | core.ProcMaskSpellDamage | core.ProcMaskSpellHealing | core.ProcMaskMeleeProc | core.ProcMaskRangedProc | core.ProcMaskSpellDamageProc,
		Outcome:            core.OutcomeLanded,
		RequireDamageDealt: false,
	})

	// Your special attacks have a chance to give you 1000 armor penetration for 15s.
	shared.NewProcStatBonusEffect(shared.ProcStatBonusEffect{
		Name:               "Warp-Spring Coil",
		ItemID:             30450,
		Callback:           core.CallbackOnSpellHitDealt,
		ProcMask:           core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeOHSpecial | core.ProcMaskMeleeProc,
		Outcome:            core.OutcomeLanded,
		RequireDamageDealt: true,
	})

	// Your attacks ignore 435 of your enemies' armor for 10s. This effect stacks up to 3 times.
	shared.NewStackingStatBonusEffect(shared.StackingStatBonusEffect{
		Name:               "The Night Blade",
		ItemID:             31331,
		MaxStacks:          3,
		Callback:           core.CallbackOnSpellHitDealt,
		ProcMask:           core.ProcMaskMeleeMHAuto | core.ProcMaskMeleeOHAuto | core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeOHSpecial,
		Outcome:            core.OutcomeLanded,
		RequireDamageDealt: true,
	})

	// Your Mortal Strike, Bloodthirst, and Shield Slam attacks have a 25% chance to heal you for 330 and grant
	// 55 Strength for 12s.
	shared.NewProcStatBonusEffect(shared.ProcStatBonusEffect{
		Name:               "Ashtongue Talisman of Valor",
		ItemID:             32485,
		Callback:           core.CallbackOnSpellHitDealt,
		ProcMask:           core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeOHSpecial,
		Outcome:            core.OutcomeLanded,
		RequireDamageDealt: true,
	})

	// Your Steady Shot has a 15% chance to grant you 275 attack power for 8s.
	shared.NewProcStatBonusEffect(shared.ProcStatBonusEffect{
		Name:               "Ashtongue Talisman of Swiftness",
		ItemID:             32487,
		Callback:           core.CallbackOnSpellHitDealt,
		ProcMask:           core.ProcMaskRangedSpecial,
		Outcome:            core.OutcomeLanded,
		RequireDamageDealt: true,
	})

	// Your spell critical strikes have a 50% chance to grant you 145 spell haste rating for 5s.
	shared.NewProcStatBonusEffect(shared.ProcStatBonusEffect{
		Name:               "Ashtongue Talisman of Insight",
		ItemID:             32488,
		Callback:           core.CallbackOnSpellHitDealt,
		ProcMask:           core.ProcMaskSpellDamage,
		Outcome:            core.OutcomeCrit,
		RequireDamageDealt: false,
	})

	// Each time your Corruption deals damage, it has a 20% chance to grant you 220 spell damage for 5s.
	shared.NewProcStatBonusEffect(shared.ProcStatBonusEffect{
		Name:               "Ashtongue Talisman of Shadows",
		ItemID:             32493,
		Callback:           core.CallbackOnPeriodicDamageDealt,
		ProcMask:           core.ProcMaskSpellDamage,
		Outcome:            core.OutcomeLanded,
		RequireDamageDealt: true,
	})
}
