package hunter

import (
	"github.com/wowsims/mop/sim/core"
	"github.com/wowsims/mop/sim/core/proto"
)

func (hunt *Hunter) ApplyHotfixes() {
	// -- From the general Hotfix Passive, present before tunings --
	// https://www.wowhead.com/mop-classic/spell=137014/hotfix-passive
	if hunt.Spec == proto.Spec_SpecSurvivalHunter {
		hunt.AddStaticMod(core.SpellModConfig{
			Kind:       core.SpellMod_DamageDone_Pct,
			ClassMask:  HunterSpellExplosiveShot,
			FloatValue: 0.1,
		})
	}
	if hunt.Spec == proto.Spec_SpecMarksmanshipHunter {
		hunt.AddStaticMod(core.SpellModConfig{
			Kind:       core.SpellMod_DamageDone_Pct,
			ClassMask:  HunterSpellChimeraShot,
			FloatValue: 0.5,
		})
	}

	// -- From the spec-specific Hotfix Passives, can be removed at any time --

	// SV: https://www.wowhead.com/mop-classic/spell=137017/hotfix-passive
	if hunt.Spec == proto.Spec_SpecSurvivalHunter {
		hunt.AddStaticMod(core.SpellModConfig{
			Kind:       core.SpellMod_DamageDone_Pct,
			ClassMask:  HunterSpellExplosiveShot,
			FloatValue: -0.05,
		})
	}
	// MM: https://www.wowhead.com/mop-classic/spell=137016/hotfix-passive
	if hunt.Spec == proto.Spec_SpecMarksmanshipHunter {
		hunt.AddStaticMod(core.SpellModConfig{
			Kind:       core.SpellMod_DamageDone_Pct,
			ClassMask:  HunterSpellAimedShot,
			FloatValue: 0.05,
		})
		hunt.AddStaticMod(core.SpellModConfig{
			Kind:       core.SpellMod_DamageDone_Pct,
			ClassMask:  HunterSpellSteadyShot | HunterSpellChimeraShot,
			FloatValue: 0.08,
		})
		hunt.AddStaticMod(core.SpellModConfig{
			Kind:       core.SpellMod_DamageDone_Pct,
			ClassMask:  HunterSpellBarrage,
			FloatValue: 0.15,
		})
	}
	// BM: https://www.wowhead.com/mop-classic/spell=137015/hotfix-passive
	// Nothing...
}
