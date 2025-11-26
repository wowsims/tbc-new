package frost

import (
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/mage"
)

func (frost *FrostMage) registerGlyphs() {
	if frost.HasMajorGlyph(proto.MageMajorGlyph_GlyphOfWaterElemental) {
		frost.waterElemental.AddStaticMod(core.SpellModConfig{
			Kind:      core.SpellMod_AllowCastWhileMoving,
			ClassMask: mage.MageWaterElementalSpellWaterBolt,
		})
	}
}
