package priest

import (
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

type Shadowfiend struct {
	core.Pet
	Priest          *Priest
	ManaRestoreAura *core.Aura
}

var baseStats = stats.Stats{
	stats.Strength:    153,
	stats.Agility:     108,
	stats.Stamina:     297,
	stats.Intellect:   175,
	stats.Spirit:      122,
	stats.AttackPower: -20, // Negative base offset; Str*2 dependency brings displayed AP to 286
	stats.Armor:       5290,
}

func (priest *Priest) NewShadowfiend() *Shadowfiend {
	shadowfiend := &Shadowfiend{
		Pet: core.NewPet(core.PetConfig{
			Name:            "Shadowfiend",
			Owner:           &priest.Character,
			BaseStats:       baseStats,
			StatInheritance: priest.shadowfiendStatInheritance(),
		}),
		Priest: priest,
	}

	manaMetric := priest.NewManaMetrics(core.ActionID{SpellID: 34433})
	shadowfiend.ManaRestoreAura = shadowfiend.GetOrRegisterAura(core.Aura{
		Label:    "Shadowfiend Mana Restore",
		Duration: core.NeverExpires,
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			priest.AddMana(sim, result.Damage*2.5, manaMetric)
		},
	})

	shadowfiend.EnableAutoAttacks(shadowfiend, core.AutoAttackOptions{
		MainHand: core.Weapon{
			BaseDamageMin:        68,
			BaseDamageMax:        92,
			SwingSpeed:           1.5,
			NormalizedSwingSpeed: 1.5,
			CritMultiplier:       2,
			SpellSchool:          core.SpellSchoolShadow,
			AttackPowerPerDPS:    core.DefaultAttackPowerPerDPS,
		},
		AutoSwingMelee: true,
	})

	shadowfiend.AutoAttacks.MHConfig().BonusCoefficient = 1.0
	priest.AddPet(shadowfiend)

	return shadowfiend
}

func (priest *Priest) shadowfiendStatInheritance() core.PetStatInheritance {
	return func(ownerStats stats.Stats) stats.Stats {
		return stats.Stats{
			stats.AttackPower: (ownerStats[stats.SpellDamage] + ownerStats[stats.ShadowDamage]) * 0.57,
		}
	}
}

func (shadowfiend *Shadowfiend) Initialize() {
	shadowfiend.AddStatDependency(stats.Strength, stats.AttackPower, 2)
}

func (shadowfiend *Shadowfiend) ExecuteCustomRotation(sim *core.Simulation) {
}

func (shadowfiend *Shadowfiend) Reset(sim *core.Simulation) {
	shadowfiend.Disable(sim)
}

func (shadowfiend *Shadowfiend) OnEncounterStart(_ *core.Simulation) {
}
func (shadowfiend *Shadowfiend) OnPetEnable(sim *core.Simulation) {
	shadowfiend.ManaRestoreAura.Activate(sim)
}

func (shadowfiend *Shadowfiend) OnPetDisable(sim *core.Simulation) {
	if shadowfiend.ManaRestoreAura.IsActive() {
		shadowfiend.ManaRestoreAura.Deactivate(sim)
	}
}

func (shadowfiend *Shadowfiend) GetPet() *core.Pet {
	return &shadowfiend.Pet
}
