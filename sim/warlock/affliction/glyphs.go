package affliction

import (
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/warlock"
)

func (affliction *AfflictionWarlock) registerGlyphs() {

	if affliction.HasMajorGlyph(proto.WarlockMajorGlyph_GlyphOfUnstableAffliction) {
		affliction.AddStaticMod(core.SpellModConfig{
			Kind:       core.SpellMod_CastTime_Pct,
			ClassMask:  warlock.WarlockSpellUnstableAffliction,
			FloatValue: -0.25,
		})
	}
}
