package shaman

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

var earthElementalTotemRank = genRanks.EarthElementalTotem.BySpellID(2062)

func (shaman *Shaman) registerEarthElementalTotem() {
	actionID := core.ActionID{SpellID: earthElementalTotemRank.SpellID}
	totalDuration := time.Second * 120

	earthElementalAura := shaman.RegisterAura(core.Aura{
		Label:    "Earth Elemental Totem",
		ActionID: actionID,
		Duration: totalDuration,
	})

	shaman.EarthElementalTotem = shaman.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		DefenseType:    core.DefenseTypeMagic,
		Flags:          core.SpellFlagAPL | SpellFlagInstant,
		ClassSpellMask: SpellMaskEarthElementalTotem,
		ManaCost: core.ManaCostOptions{
			FlatCost: earthElementalTotemRank.Cost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: earthElementalTotemRank.GCD,
			},
			CD: core.Cooldown{
				Timer:    shaman.NewTimer(),
				Duration: earthElementalTotemRank.Cooldown,
			},
			SharedCD: core.Cooldown{
				Timer:    shaman.GetOrInitTimer(&shaman.ElementalSharedCDTimer),
				Duration: time.Minute * 1,
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, _ *core.Spell) {
			if shaman.Totems.Earth != proto.EarthTotem_NoEarthTotem {
				shaman.TotemExpirations[EarthTotem] = sim.CurrentTime + totalDuration
			}

			shaman.EarthElemental.EnableWithTimeout(sim, shaman.EarthElemental, totalDuration)

			// Add a dummy aura to show in metrics
			earthElementalAura.Activate(sim)
		},
	})

	shaman.AddMajorCooldown(core.MajorCooldown{
		Spell: shaman.EarthElementalTotem,
		Type:  core.CooldownTypeDPS,
		ShouldActivate: func(sim *core.Simulation, character *core.Character) bool {
			// Fele should only be cast by manual APL intervention
			return false
		},
	})
}
