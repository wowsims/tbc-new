package frost

import (
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
	"github.com/wowsims/tbc/sim/mage"
)

func RegisterFrostMage() {
	core.RegisterAgentFactory(
		proto.Player_FrostMage{},
		proto.Spec_SpecFrostMage,
		func(character *core.Character, options *proto.Player) core.Agent {
			return NewFrostMage(character, options)
		},
		func(player *proto.Player, spec interface{}) {
			playerSpec, ok := spec.(*proto.Player_FrostMage)
			if !ok {
				panic("Invalid spec value for Frost Mage!")
			}
			player.Spec = playerSpec
		},
	)
}

type FrostMage struct {
	*mage.Mage

	waterElemental             *WaterElemental
	frostfireFrozenCritBuffMod *core.SpellMod
	iceLanceFrozenCritBuffMod  *core.SpellMod
}

func NewFrostMage(character *core.Character, options *proto.Player) *FrostMage {
	frostOptions := options.GetFrostMage().Options

	frostMage := &FrostMage{
		Mage: mage.NewMage(character, options, frostOptions.ClassOptions),
	}
	frostMage.waterElemental = frostMage.NewWaterElemental()

	return frostMage
}

func (frost *FrostMage) GetMage() *mage.Mage {
	return frost.Mage
}

func (frost *FrostMage) Reset(sim *core.Simulation) {
	frost.Mage.Reset(sim)
}

func (frost *FrostMage) Initialize() {
	frost.Mage.Initialize()

	frost.registerGlyphs()
	frost.registerPassives()
	frost.registerSpells()
	frost.registerHotfixes()
}

func (frost *FrostMage) registerPassives() {
	frost.registerMastery()
	frost.registerFingersOfFrost()
	frost.registerBrainFreeze()
}

func (frost *FrostMage) registerSpells() {
	frost.registerSummonWaterElementalSpell()
	frost.registerFrostboltSpell()
	frost.registerFrozenOrbSpell()
}

func (frost *FrostMage) GetFrozenCritPercentage() float64 {
	baseCritPercent := frost.GetStat(stats.SpellCritPercent)

	suppressionPercent := 0.0
	if frost.CurrentTarget != nil &&
		int(frost.CurrentTarget.UnitIndex) < len(frost.AttackTables) &&
		frost.AttackTables[frost.CurrentTarget.UnitIndex] != nil {
		attackTable := frost.AttackTables[frost.CurrentTarget.UnitIndex]
		suppressionPercent = attackTable.SpellCritSuppression * 100
	}

	return baseCritPercent - suppressionPercent + 50
}

func (frost *FrostMage) registerMastery() {
	/*
		Shatter doubles the crit chance of spells against frozen targets and then adds an additional 50%, hence critChance * 2 + 50
		https://www.wowhead.com/mop-classic/spell=12982/shatter for more information.
	*/
	frost.frostfireFrozenCritBuffMod = frost.Mage.AddDynamicMod(core.SpellModConfig{
		ClassMask: mage.MageSpellFrostfireBolt,
		Kind:      core.SpellMod_BonusCrit_Percent,
	})

	frost.iceLanceFrozenCritBuffMod = frost.Mage.AddDynamicMod(core.SpellModConfig{
		ClassMask: mage.MageSpellIceLance,
		Kind:      core.SpellMod_BonusCrit_Percent,
	})

	frost.AddOnTemporaryStatsChange(func(sim *core.Simulation, buffAura *core.Aura, statsChangeWithoutDeps stats.Stats) {
		frozenCritPercentage := frost.GetFrozenCritPercentage()
		frost.frostfireFrozenCritBuffMod.UpdateFloatValue(frozenCritPercentage)
		frost.iceLanceFrozenCritBuffMod.UpdateFloatValue(frozenCritPercentage)
	})

	frostMasteryMod := frost.waterElemental.AddDynamicMod(core.SpellModConfig{
		ClassMask:  mage.MageWaterElementalSpellWaterBolt,
		FloatValue: frost.GetFrostMasteryBonus(),
		Kind:       core.SpellMod_DamageDone_Pct,
	})

	frost.AddOnMasteryStatChanged(func(sim *core.Simulation, oldMastery, newMastery float64) {
		masteryBonus := frost.GetFrostMasteryBonus()
		frostMasteryMod.UpdateFloatValue(masteryBonus)
	})

	core.MakePermanent(frost.RegisterAura(core.Aura{
		Label: "Mastery: Icicles - Water Elemental",
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			frostMasteryMod.Activate()
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			frostMasteryMod.Deactivate()
		},
	}))
}
