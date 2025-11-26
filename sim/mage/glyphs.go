package mage

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

func (mage *Mage) registerGlyphs() {
	// Majors MOP

	// Glyph of Frostfire Bolt
	if mage.HasMajorGlyph(proto.MageMajorGlyph_GlyphOfFrostfireBolt) {
		mage.AddStaticMod(core.SpellModConfig{
			ClassMask: MageSpellFrostfireBolt,
			TimeValue: time.Millisecond * -500,
			Kind:      core.SpellMod_CastTime_Flat,
		})
	}

	// Glyph of Cone of Cold
	if mage.HasMajorGlyph(proto.MageMajorGlyph_GlyphOfConeOfCold) {
		mage.AddStaticMod(core.SpellModConfig{
			ClassMask:  MageSpellConeOfCold,
			FloatValue: 2.0,
			Kind:       core.SpellMod_DamageDone_Pct,
		})
	}

	// Glyph of Water Elemental
	if mage.HasMajorGlyph(proto.MageMajorGlyph_GlyphOfWaterElemental) {
		mage.AddStaticMod(core.SpellModConfig{
			Kind:      core.SpellMod_AllowCastWhileMoving,
			ClassMask: MageWaterElementalSpellWaterBolt,
		})
	}

	// Glyph of Armors
	if mage.HasMajorGlyph(proto.MageMajorGlyph_GlyphOfArmors) {
		mage.AddStaticMod(core.SpellModConfig{
			Kind:      core.SpellMod_CastTime_Flat,
			ClassMask: MageSpellFrostArmor | MageSpellMageArmor | MageSpellMoltenArmor,
			TimeValue: -time.Millisecond * 1500,
		})
	}

}
