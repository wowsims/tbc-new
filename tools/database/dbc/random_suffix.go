package dbc

import (
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

type RandomSuffix struct {
	ID            int
	Name          string
	AllocationPct []int // AllocationPct_0-4
	EffectArgs    []int // EffectArg_0-4
	Effects       []int // Effect_0-4
}

func (raw RandomSuffix) ToProto() *proto.ItemRandomSuffix {
	suffix := &proto.ItemRandomSuffix{
		Name:  raw.Name,
		Id:    int32(raw.ID),
		Stats: stats.Stats{}.ToProtoArray(),
	}

	for i, effect := range raw.Effects {
		amount := float64(raw.AllocationPct[i])
		switch effect {
		case ITEM_ENCHANTMENT_RESISTANCE:
			stat, match := MapResistanceToStat(raw.EffectArgs[i])
			if !match {
				continue
			}

			suffix.Stats[stat] = amount
			if suffix.Name == "" {
				suffix.Name = stats.Stat(stat).StatName()
			}
		case ITEM_ENCHANTMENT_STAT:
			stat, match := MapBonusStatIndexToStat(raw.EffectArgs[i])
			if !match {
				continue
			}
			suffix.Stats[stat] = amount
			if suffix.Name == "" {
				suffix.Name = stats.Stat(stat).StatName()
			}
		case ITEM_ENCHANTMENT_EQUIP_SPELL: //Buff
			spellEffects := dbcInstance.SpellEffects[raw.EffectArgs[i]]
			for _, spellEffect := range spellEffects {
				if spellEffect.EffectMiscValues[0] == -1 &&
					spellEffect.EffectType == E_APPLY_AURA &&
					spellEffect.EffectAura == A_MOD_STAT {
					// Apply bonus to all stats
					suffix.Stats[proto.Stat_StatAgility] += amount
					suffix.Stats[proto.Stat_StatIntellect] += amount
					suffix.Stats[proto.Stat_StatSpirit] += amount
					suffix.Stats[proto.Stat_StatStamina] += amount
					suffix.Stats[proto.Stat_StatStrength] += amount
					continue
				}
				if spellEffect.EffectType == E_APPLY_AURA && spellEffect.EffectAura == A_MOD_STAT {
					suffix.Stats[spellEffect.EffectMiscValues[0]] += amount
				} else if spellEffect.EffectType == E_APPLY_AURA && spellEffect.EffectAura == A_MOD_RESISTANCE && (spellEffect.EffectMiscValues[0] == 126 || spellEffect.EffectMiscValues[0] == 124) {
					suffix.Stats[proto.Stat_StatArcaneResistance] += amount
					suffix.Stats[proto.Stat_StatFireResistance] += amount
					suffix.Stats[proto.Stat_StatFrostResistance] += amount
					suffix.Stats[proto.Stat_StatNatureResistance] += amount
					suffix.Stats[proto.Stat_StatShadowResistance] += amount
				} else {
					stat := ConvertEffectAuraToStatIndex(spellEffect.EffectAura, spellEffect.EffectMiscValues[0])
					if stat >= 0 {
						suffix.Stats[stat] += amount
					}
				}
			}

		}
	}
	return suffix
}
