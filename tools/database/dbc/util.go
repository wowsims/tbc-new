package dbc

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"

	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

func GetProfession(id int) proto.Profession {
	if profession, ok := MapProfessionIdToProfession[id]; ok {
		return profession
	}
	return 0
}

func GetRepLevel(minReputation int) proto.RepLevel {
	if repLevel, ok := MapMinReputationToRepLevel[minReputation]; ok {
		return repLevel
	}
	return proto.RepLevel_RepLevelUnknown
}
func NullFloat(arr []float64) []float64 {
	for _, v := range arr {
		if v > 0 {
			return arr
		}
	}

	return nil
}
func GetClassesFromClassMask(mask int) []proto.Class {
	var result []proto.Class

	allClasses := (1 << len(Classes)) - 1
	if mask&allClasses == allClasses {
		return result
	}

	for _, class := range Classes {
		if mask&(1<<(class.ID-1)) != 0 {
			result = append(result, class.ProtoClass)
		}
	}
	slices.Sort(result)
	return result
}

func WriteGzipFile(filePath string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directories for %s: %w", filePath, err)
	}
	// Create the file
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Create a gzip writer on top of the file writer
	gw := gzip.NewWriter(f)
	defer gw.Close()

	// Write the data to the gzip writer
	_, err = gw.Write(data)
	return err
}
func ReadGzipFile(filename string) ([]byte, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, DataLoadError{
			Source:   filename,
			DataType: "gzip file",
			Reason:   err.Error(),
		}
	}
	defer f.Close()

	gzReader, err := gzip.NewReader(f)
	if err != nil {
		return nil, DataLoadError{
			Source:   filename,
			DataType: "gzip",
			Reason:   err.Error(),
		}
	}
	defer gzReader.Close()

	data, err := io.ReadAll(gzReader)
	if err != nil {
		return nil, DataLoadError{
			Source:   filename,
			DataType: "decompression",
			Reason:   err.Error(),
		}
	}

	return data, nil
}

func processEnchantmentEffects(
	effects []int,
	effectArgs []int,
	effectPoints []int,
	spellEffectPoints []int,
	outStats *stats.Stats,
	addRanged bool,
) {
	for i, effect := range effects {
		switch effect {
		case ITEM_ENCHANTMENT_RESISTANCE:
			stat, match := MapResistanceToStat(effectArgs[i])
			if !match {
				continue
			}
			outStats[stat] = float64(effectPoints[i])
		case ITEM_ENCHANTMENT_STAT:
			stat, success := MapBonusStatIndexToStat(effectArgs[i])
			if !success {
				continue
			}
			if effectPoints[i] == 0 && spellEffectPoints != nil {
				// This might be stored in a SpellEffect row
				outStats[stat] = float64(spellEffectPoints[i] + 1)
			} else {
				outStats[stat] = float64(effectPoints[i])

				// If the bonus stat is attack power, copy it to ranged attack power
				if addRanged && stat == proto.Stat_StatAttackPower {
					outStats[proto.Stat_StatRangedAttackPower] = float64(effectPoints[i])
				}
				// If it's All Hit then boost both
				if stat == proto.Stat_StatAllHitRating {
					outStats[proto.Stat_StatMeleeHitRating] = float64(effectPoints[i])
					outStats[proto.Stat_StatSpellHitRating] = float64(effectPoints[i])
				}
			}
		case ITEM_ENCHANTMENT_EQUIP_SPELL: //Buff
			spellEffects := dbcInstance.SpellEffects[effectArgs[i]]
			for _, spellEffect := range spellEffects {
				if spellEffect.EffectMiscValues[0] == -1 &&
					spellEffect.EffectType == E_APPLY_AURA &&
					spellEffect.EffectAura == A_MOD_STAT {
					// Apply bonus to all stats
					outStats[proto.Stat_StatAgility] += float64(spellEffect.EffectBasePoints + 1)
					outStats[proto.Stat_StatIntellect] += float64(spellEffect.EffectBasePoints + 1)
					outStats[proto.Stat_StatSpirit] += float64(spellEffect.EffectBasePoints + 1)
					outStats[proto.Stat_StatStamina] += float64(spellEffect.EffectBasePoints + 1)
					outStats[proto.Stat_StatStrength] += float64(spellEffect.EffectBasePoints + 1)
					continue
				}
				if spellEffect.EffectType == E_APPLY_AURA && spellEffect.EffectAura == A_MOD_STAT {
					outStats[spellEffect.EffectMiscValues[0]] += float64(spellEffect.EffectBasePoints + 1)
				} else {
					stat := ConvertEffectAuraToStatIndex(int(spellEffect.EffectAura), spellEffect.EffectMiscValues[0])
					if stat >= 0 {
						outStats[stat] += float64(spellEffect.EffectBasePoints + 1)
					}
				}
			}
		case ITEM_ENCHANTMENT_COMBAT_SPELL:
			// Not processed (chance on hit, ignore for now)
		case ITEM_ENCHANTMENT_USE_SPELL:
			// Not processed
		}
	}
}

func ConvertEffectAuraToStatIndex(effectAura int, effectMisc int) proto.Stat {
	switch effectAura {
	case 99: // MOD_ATTACK_POWER
		return proto.Stat_StatAttackPower
	case 124: // MOD_RANGED_ATTACK_POWER
		return proto.Stat_StatRangedAttackPower
	case 13: // MOD_DAMAGE_DONE
		return proto.Stat_StatSpellDamage
	case 135: // MOD_HEALING_DONE
		return proto.Stat_StatHealingPower
	case 34: // MOD_INCREASE_HEALTH
		return proto.Stat_StatHealth
	case 123: // MOD_TARGET_RESISTANCE
		return ConvertTargetResistanceFlagToPenetrationStat(effectMisc)
	case 189: // MOD_RATING (Stat Ratings but as Auras; includes mostly Vanilla items, but also some socket bonuses and random one-offs)
		return ConvertModRatingFlagToRatingStat(effectMisc)
	default:
		return -1
	}
}

func ConvertTargetResistanceFlagToPenetrationStat(flag int) proto.Stat {
	switch flag {
	case 1:
		return proto.Stat_StatArmorPenetration
	default:
		return proto.Stat_StatSpellPenetration
	}
}

func ConvertModRatingFlagToRatingStat(flag int) proto.Stat {
	switch flag {
	case 2:
		return proto.Stat_StatDefenseRating
	case 4:
		return proto.Stat_StatDodgeRating
	case 8:
		return proto.Stat_StatParryRating
	case 16:
		return proto.Stat_StatBlockRating
	case 64:
		// The forbidden "Only Ranged Hit". There's a single instance of this (Enchant 2523, SpellID 22780).
		return proto.Stat_StatRage
	case 96:
		return proto.Stat_StatAllHitRating
	case 128:
		return proto.Stat_StatSpellHitRating
	case 512:
		// The forbidden "Only Ranged Crit". Only two of these exist, and they're not valid sim items.
		return proto.Stat_StatRage
	case 768:
		return proto.Stat_StatMeleeCritRating
	case 1024:
		return proto.Stat_StatSpellCritRating
	case 131072:
		return proto.Stat_StatMeleeHasteRating
	case 393216:
		return proto.Stat_StatMeleeHasteRating
	case 49152:
		return proto.Stat_StatResilience
	default:
		println("UNHANDLED RATING FLAG: " + strconv.Itoa(flag))
		return proto.Stat_StatRage
	}
}
