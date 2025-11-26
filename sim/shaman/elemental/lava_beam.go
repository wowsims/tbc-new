package elemental

import (
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/shaman"
)

func (ele *ElementalShaman) registerLavaBeamSpell() {
	maxHits := min(5, ele.Env.TotalTargetCount())
	ele.LavaBeam = ele.newLavaBeamSpell(false)
	ele.LavaBeamOverloads = [2][]*core.Spell{}

	for range maxHits {
		ele.LavaBeamOverloads[0] = append(ele.LavaBeamOverloads[0], ele.newLavaBeamSpell(true))
		ele.LavaBeamOverloads[1] = append(ele.LavaBeamOverloads[1], ele.newLavaBeamSpell(true))
	}
}

func (ele *ElementalShaman) newLavaBeamSpell(isElementalOverload bool) *core.Spell {
	shamConfig := shaman.ShamSpellConfig{
		ActionID:            core.ActionID{SpellID: 114074},
		IsElementalOverload: isElementalOverload,
		BaseCostPercent:     8.3,
		BonusCoefficient:    0.57099997997,
		Coeff:               1.08800005913,
		Variance:            0.13300000131,
		SpellSchool:         core.SpellSchoolFire,
		Overloads:           &ele.LavaBeamOverloads,
		BounceReduction:     1.1,
		ClassSpellMask:      core.TernaryInt64(isElementalOverload, shaman.SpellMaskLavaBeamOverload, shaman.SpellMaskLavaBeam),
	}
	spellConfig := ele.NewChainSpellConfig(shamConfig)
	spellConfig.ExtraCastCondition = func(sim *core.Simulation, target *core.Unit) bool {
		return ele.AscendanceAura.IsActive()
	}
	return ele.RegisterSpell(spellConfig)
}
