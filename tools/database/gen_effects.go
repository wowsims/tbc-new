package database

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"text/template"

	_ "github.com/wowsims/tbc/sim/common"
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/tools/database/dbc"
	"github.com/wowsims/tbc/tools/tooltip"
)

// Sets the minimum itemlevel that should be considered for this expansions
const MIN_EFFECT_ILVL = 50

// Enchantment IDs at or below this are not generated.
const MIN_ENCHANT_EFFECT_ID = 0

func isGeneratableEnchant(effectID int32) bool {
	return effectID > MIN_ENCHANT_EFFECT_ID
}

type ProcInfo struct {
	Outcome             core.HitOutcome
	Callback            core.AuraCallback
	ProcMask            core.ProcMask
	MaxCumulativeStacks int32
	RequireDamageDealt  bool
	ClassSpellsOnly     bool
}

// Entry represents a effect with its Item ID, Spell ID and display name.
type Variant struct {
	ID      int
	SpellID int
	Name    string
}

type Entry struct {
	Variants  []*Variant
	Tooltip   []string
	ProcInfo  ProcInfo
	Supported bool
	// What adds a stack while the window is open, for the trinkets whose effect carries a
	// separate accumulating aura. Nil for everything else.
	StackProcInfo *ProcInfo
	// Set for an on-use whose window accumulates a separate aura, which needs the stacking
	// helper rather than the flat one. Carries what that helper cannot read from the database.
	StackingOnUse *StackingOnUse
	// Set for effects an ignore list deliberately excludes. These emit a comment only, so that
	// skipping them is visible in the generated file rather than silent.
	Skipped bool
}

// The literals a stacking on-use needs in the generated call. Everything else - stacks,
// per-stack stats, which aura is which - the helper reads from the database at runtime.
type StackingOnUse struct {
	Name       string
	DurationMs int32
	CooldownMs int32
}

// Group holds a category of effects.
type Group struct {
	Name    string
	Entries []*Entry
}

type MissingItemEffect struct {
	ItemID  int32
	Name    string
	Effects []Variant
}

var missingEffectsMap = map[string]map[int32]MissingItemEffect{
	"EnchantEffects": {},
	"ItemEffects":    {},
}

type EffectParseResult byte

const (
	EffectParseResultInvalid     EffectParseResult = iota // Returned when the effect is invalid for the current parameters
	EffectParseResultUnsupported                          // Returned when the effect could be parsed but is not supported for effect generation
	EffectParseResultSuccess                              // Returned when the effect was parsed successfuly
)

func GenerateEffectsFile(groups []*Group, outFile string, templateString string) error {
	if _, err := os.Stat(outFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("unable to check file %s: %w", outFile, err)
	}

	// Ensure groups and entries are sorted
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Name < groups[j].Name
	})

	for _, grp := range groups {
		sort.Slice(grp.Entries, func(i, j int) bool {
			if grp.Entries[i].Supported != grp.Entries[j].Supported {
				return !grp.Entries[i].Supported
			}

			return entryOrder(grp.Entries[i], grp.Entries[j])
		})
	}

	funcMap := map[string]any{
		"asCoreCallback": asCoreCallback,
		"asCoreProcMask": asCoreProcMask,
		"asCoreOutcome":  asCoreOutcome,
		"formatStrings":  formatStrings,
	}
	tmpl := template.Must(template.New("effects").Funcs(funcMap).Parse(templateString))

	// An empty generated file must not import anything: gen_db links sim/common, so unused
	// imports in a file it just wrote break the very build the next run needs.
	hasEntries := false
	for _, grp := range groups {
		if len(grp.Entries) > 0 {
			hasEntries = true
			break
		}
	}

	hasStacking := false
	for _, grp := range groups {
		for _, entry := range grp.Entries {
			if entry.StackingOnUse != nil {
				hasStacking = true
			}
		}
	}

	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, map[string]interface{}{"Groups": groups, "HasEntries": hasEntries, "HasStacking": hasStacking}); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// The template cannot indent commented-out blocks or blank lines the way gofmt wants, so
	// format the result. Otherwise every regeneration reverts whatever formatted the file last
	// and the diff is hundreds of whitespace-only lines.
	out := rendered.Bytes()
	if formatted, err := format.Source(out); err != nil {
		fmt.Printf("WARN: generated %s is not valid Go, writing unformatted: %v\n", outFile, err)
	} else {
		out = formatted
	}

	if err := os.WriteFile(outFile, out, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", outFile, err)
	}

	return nil
}

// A total order over entries. Sorting on the item or enchant ID alone is not one: an item with
// two effects yields two entries sharing that ID, and sort.Slice is not stable, so their order
// in the generated file flipped between runs.
func entryOrder(a *Entry, b *Entry) bool {
	if a.Variants[0].ID != b.Variants[0].ID {
		return a.Variants[0].ID < b.Variants[0].ID
	}
	return a.Variants[0].SpellID < b.Variants[0].SpellID
}

// Escapes a rendered tooltip for use inside a double-quoted TypeScript string. Tooltips
// routinely span several lines, which would otherwise produce a file that does not parse.
func jsString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\r\n", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	return s
}

const missingEffectsFileName = "ui/core/constants/missing_effects_auto_gen.ts"

func GenerateMissingEffectsFile() error {
	if _, err := os.Stat(missingEffectsFileName); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("unable to check file %s: %w", missingEffectsFileName, err)
	}

	funcMap := map[string]any{
		"asCoreCallback": asCoreCallback,
		"asCoreProcMask": asCoreProcMask,
		"asCoreOutcome":  asCoreOutcome,
		"formatStrings":  formatStrings,
		"jsString":       jsString,
	}
	tmpl := template.Must(template.New("missingEffects").Funcs(funcMap).Parse(TmplStrMissingEffects))
	f, err := os.Create(missingEffectsFileName)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", missingEffectsFileName, err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, missingEffectsMap); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

func GenerateEnchantEffects(instance *dbc.DBC, db *WowDatabase) {
	groupMapProc := map[string]Group{}
	enchantSpellEffects := map[int]*dbc.SpellEffect{}

	for _, effect := range instance.SpellEffectsById {
		if effect.EffectType == dbc.E_ENCHANT_ITEM {
			enchantSpellEffects[effect.EffectMiscValues[0]] = &effect
		}
	}

	for _, enchant := range instance.Enchants {
		parsed := enchant.ToProto()
		if _, ok := db.Enchants[EnchantToDBKey(parsed)]; !ok {
			continue
		}

		for _, enchantEffect := range parsed.EnchantEffects {
			TryParseEnchantEffect(parsed, enchantEffect, groupMapProc, instance, enchantSpellEffects)
		}
	}

	var procGroups []*Group
	for _, grp := range groupMapProc {
		procGroups = append(procGroups, &grp)
	}

	GenerateEffectsFile(procGroups, "sim/common/tbc/enchants_auto_gen.go", TmplStrEnchant)
}

// Names the ignore-list rule that excluded an effect, for the comment emitted in the generated
// file. Returns "" when nothing excludes it.
func ignoredEffectReason(instance *dbc.DBC, effectID int) string {
	for _, effect := range instance.SpellEffectsInOrder(effectID) {
		if params, ok := IgnoreSpellEffectByAuraType[effect.EffectAura]; ok {
			if len(params) == 0 || slices.Contains(params, effect.EffectMiscValues[0]) {
				return fmt.Sprintf("ignored aura type %d", effect.EffectAura)
			}
		}

		if params, ok := IgnoreSpellEffectBySpellEffectType[effect.EffectType]; ok {
			if len(params) == 0 || slices.Contains(params, effect.EffectMiscValues[0]) {
				return fmt.Sprintf("ignored effect type %d", effect.EffectType)
			}
		}
	}

	return ""
}

// Records an effect excluded by an ignore list so the generated file documents it. Kept in its
// own group: variant merging is per-group, so these cannot affect whether a real effect's
// variant set is emitted live or commented.
func storeSkippedEffect(id int32, name string, buffID int32, instance *dbc.DBC, groupMap map[string]Group) {
	grp, exists := groupMap["Skipped"]
	if !exists {
		grp = Group{Name: "Skipped"}
	}

	buffName := instance.Spells[int(buffID)].NameLang
	grp.Entries = append(grp.Entries, &Entry{
		Skipped:  true,
		Variants: []*Variant{{ID: int(id), Name: name, SpellID: int(buffID)}},
		Tooltip: []string{fmt.Sprintf("%s: %q (%d) - %s",
			name, buffName, buffID, ignoredEffectReason(instance, int(buffID)))},
	})
	groupMap["Skipped"] = grp
}

func ItemEffectIsSupported(instance *dbc.DBC, effectID int) bool {
	supported := true
	if effects, ok := instance.SpellEffects[effectID]; ok {
		for _, effect := range effects {
			if params, ok := IgnoreSpellEffectByAuraType[effect.EffectAura]; ok {
				if len(params) == 0 {
					supported = false
					break
				} else {
					if slices.Contains(params, effect.EffectMiscValues[0]) {
						supported = false
					}
				}
			}

			if params, ok := IgnoreSpellEffectBySpellEffectType[effect.EffectType]; ok {
				if len(params) == 0 {
					supported = false
					break
				} else {
					if slices.Contains(params, effect.EffectMiscValues[0]) {
						supported = false
					}
				}
			}
		}
	}
	return supported
}

func GenerateItemEffects(instance *dbc.DBC, db *WowDatabase, itemSources map[int][]*proto.DropSource) {
	groupMapOnUse := map[string]Group{}
	groupMapProc := map[string]Group{}

	// Example loop over your items
	for _, parsed := range db.Items {
		parsed.ItemEffects = dbc.MergeItemEffectsForAllStates(parsed)

		for _, itemEffect := range parsed.ItemEffects {
			if !ItemEffectIsSupported(instance, int(itemEffect.BuffId)) {
				// Commented into the generated file rather than dropped. These are deliberately
				// out of scope - summons, teleports, created items - but an item whose only
				// effect is skipped otherwise vanished with no trace, while a sibling marker
				// aura on the same item got reported as missing instead.
				skippedGroup := groupMapProc
				if itemEffect.GetOnUse() != nil {
					skippedGroup = groupMapOnUse
				}
				storeSkippedEffect(parsed.Id, parsed.Name, itemEffect.BuffId, instance, skippedGroup)
				continue
			}

			if TryParseOnUseEffect(parsed, itemEffect, instance, groupMapOnUse) != EffectParseResultSuccess &&
				TryParseProcEffect(parsed, itemEffect, instance, groupMapProc) != EffectParseResultSuccess {
				ParseTooltipForMissingEffect(parsed, itemEffect, instance, groupMapProc, "Procs")
			}
		}
	}

	// Sorting done in GenerateEffectsFile
	var onUseGroups []*Group
	for _, grp := range groupMapOnUse {
		onUseGroups = append(onUseGroups, &grp)
	}

	// Merge variants
	var procGroups []*Group
	needsStatPostfix := map[string]bool{}
	for _, grp := range groupMapProc {
		newEntries := []*Entry{}
		entryGroupings := map[string]*Entry{}

		// sort entries first to make tooltip generation consistent for variants
		sort.Slice(grp.Entries, func(i, j int) bool {
			return entryOrder(grp.Entries[i], grp.Entries[j])
		})

		for _, entry := range grp.Entries {
			var idx int64 = 0
			added := false

			// Make sure to only group by name and proc mask, each proc mask will create it's own sub group
			for _, group := range entryGroupings {
				if group.Variants[0].Name == entry.Variants[0].Name {
					idx++
					if group.ProcInfo.ProcMask == entry.ProcInfo.ProcMask {
						group.AddVariant(entry.Variants[0])
						added = true
						break
					}
				}
			}

			if !added {
				groupName := entry.Variants[0].Name
				if idx > 0 {
					needsStatPostfix[groupName] = true
					groupName += "(" + strconv.FormatInt(idx, 10) + ")"
				}

				newEntries = append(newEntries, entry)
				entryGroupings[entry.Variants[0].Name] = entry
			}
		}

		grp.Entries = newEntries
		procGroups = append(procGroups, &grp)
	}

	updateNames := func(entries []*Entry) {
		for _, entry := range entries {
			for _, variant := range entry.Variants {
				if _, ok := needsStatPostfix[variant.Name]; ok {
					item := db.Items[int32(variant.ID)]
					for _, itemEffect := range item.ItemEffects {
						variant.Name += " - " + GetEffectStatString(itemEffect)
					}
				}

				variant.Name += BuildItemDifficultyPostfix(itemSources, variant.ID, instance)
			}
		}
	}

	// Update Item names
	for _, grp := range onUseGroups {
		updateNames(grp.Entries)
	}

	for _, grp := range procGroups {
		updateNames(grp.Entries)
	}

	GenerateEffectsFile(onUseGroups, "sim/common/tbc/stat_bonus_cds_auto_gen.go", TmplStrOnUse)
	GenerateEffectsFile(procGroups, "sim/common/tbc/stat_bonus_procs_auto_gen.go", TmplStrProc)
}

func GenerateItemEffectRandomPropPoints(instance *dbc.DBC, db *WowDatabase) {
	for id, allocMap := range instance.RandomPropertiesByIlvl {
		ilvl := int32(id)
		if ilvl < core.MinIlvl || ilvl > core.MaxIlvl {
			continue
		}
		db.ItemEffectRandPropPoints[ilvl] = &proto.ItemEffectRandPropPoints{
			Ilvl:           ilvl,
			RandPropPoints: allocMap[proto.ItemQuality_ItemQualityEpic][0],
		}
	}
}

func BuildItemDifficultyPostfix(itemSources map[int][]*proto.DropSource, itemId int, instance *dbc.DBC) string {
	difficultyPostfix := ""
	if sources, ok := itemSources[itemId]; ok {
		name := DifficultyToShortName(sources[0].Difficulty)
		if len(name) > 0 {
			difficultyPostfix += " " + name
		}
	}

	if item, ok := instance.Items[itemId]; ok {
		if len(item.NameDescription) > 0 && item.NameDescription != "Heroic" {
			difficultyPostfix += " (" + item.NameDescription + ")"
		}

		if item.Flags1.Has(dbc.HORDE_SPECIFIC) {
			difficultyPostfix += " (Horde)"
		}

		if item.Flags1.Has(dbc.ALLIANCE_SPECIFIC) {
			difficultyPostfix += " (Alliance)"
		}
	}

	return difficultyPostfix
}

func TryParseProcEffect(parsed *proto.UIItem, itemEffect *proto.ItemEffect, instance *dbc.DBC, groupMapProc map[string]Group) EffectParseResult {
	if itemEffect.GetProc() != nil && parsed.ScalingOptions[0].Ilvl > MIN_EFFECT_ILVL {
		// Effect was already manually implemented
		if core.HasItemEffect(parsed.Id) {
			return EffectParseResultSuccess
		}

		tooltipString, id := dbc.GetItemEffectSpellTooltip(int(parsed.Id), int(itemEffect.BuffId))
		tooltip, _ := tooltip.ParseTooltip(tooltipString, tooltip.DBCTooltipDataProvider{DBC: instance}, int64(id))

		grp, exists := groupMapProc["Procs"]
		if !exists {
			grp = Group{Name: "Procs"}
		}

		if tooltip != nil {
			renderedTooltip := tooltip.String()
			entry := Entry{Tooltip: strings.Split(renderedTooltip, "\n"), Variants: []*Variant{{ID: int(parsed.Id), Name: parsed.Name, SpellID: int(itemEffect.BuffId)}}}
			entry.ProcInfo, entry.Supported = BuildProcInfo(parsed, int(itemEffect.BuffId), instance, renderedTooltip)
			entry.StackProcInfo = buildStackProcInfo(itemEffect, instance, renderedTooltip)

			if len(itemEffect.ScalingOptions[0].Stats) == 0 || !entry.Supported {
				StoreMissingEffect("ItemEffects", parsed.Name, Variant{
					ID:      int(parsed.Id),
					Name:    renderedTooltip,
					SpellID: int(itemEffect.BuffId),
				})
				return EffectParseResultUnsupported
			}

			grp.Entries = append(grp.Entries, &entry)
			groupMapProc["Procs"] = grp

			return EffectParseResultSuccess
		} else {
			return EffectParseResultUnsupported
		}
	}

	// check if the item has any kind of proc as we only support stat proc parsing right now
	if effects, ok := instance.ItemEffectsByParentID[int(parsed.Id)]; ok && parsed.ScalingOptions[0].Ilvl > MIN_EFFECT_ILVL {
		for _, effect := range effects {
			if SpellHasTriggerEffect(effect.SpellID, instance) {
				return EffectParseResultUnsupported
			}
		}
	}

	return EffectParseResultInvalid
}

func TryParseOnUseEffect(parsed *proto.UIItem, itemEffect *proto.ItemEffect, instance *dbc.DBC, groupMap map[string]Group) EffectParseResult {
	// Effect was already manually implemented
	if core.HasItemEffect(parsed.Id) {
		return EffectParseResultSuccess
	}

	if itemEffect.GetOnUse() != nil && parsed.ScalingOptions[0].Ilvl > MIN_EFFECT_ILVL {
		if itemEffect.GetOnUse().CooldownMs < 0 && itemEffect.GetOnUse().CategoryCooldownMs < 0 {
			return EffectParseResultUnsupported
		}

		groupName := GetEffectStatString(itemEffect)
		grp, exists := groupMap[groupName]
		if !exists {
			grp = Group{Name: groupName}
		}

		entry := &Entry{Variants: []*Variant{{ID: int(parsed.Id), Name: parsed.Name, SpellID: int(itemEffect.BuffId)}}, Supported: true}
		grp.Entries = append(grp.Entries, entry)
		groupMap[groupName] = grp

		// A stacking on-use keeps its stats on the accumulating aura, so the flat check below
		// would call it unsupported, and the flat helper would grant nothing.
		stacking := itemEffect.StackingAura
		if stacking != nil && len(stacking.ScalingOptions[0].Stats) > 0 {
			entry.StackProcInfo = buildStackProcInfo(itemEffect, instance, "")
			if entry.StackProcInfo == nil {
				entry.Supported = false
				return EffectParseResultUnsupported
			}
			entry.StackingOnUse = &StackingOnUse{
				Name:       parsed.Name,
				DurationMs: itemEffect.EffectDurationMs,
				CooldownMs: itemEffect.GetOnUse().CooldownMs,
			}
			return EffectParseResultSuccess
		}

		if len(itemEffect.ScalingOptions[0].Stats) == 0 {
			entry.Supported = false
			return EffectParseResultUnsupported
		}

		return EffectParseResultSuccess
	}

	return EffectParseResultInvalid
}

func TryParseEnchantEffect(enchant *proto.UIEnchant, enchantEffect *proto.ItemEffect, groupMapProc map[string]Group, instance *dbc.DBC, enchantSpellEffects map[int]*dbc.SpellEffect) EffectParseResult {
	if (enchantEffect.GetProc() != nil || EnchantHasDummyEffect(enchant, instance)) && isGeneratableEnchant(enchant.EffectId) {

		// Effect was already manually implemented
		if core.HasEnchantEffect(enchant.EffectId) {
			return EffectParseResultSuccess
		}

		if enchantingSpell, ok := enchantSpellEffects[int(enchant.EffectId)]; ok {
			tooltipString := instance.Spells[enchantingSpell.SpellID].Description
			tooltip, _ := tooltip.ParseTooltip(tooltipString, tooltip.DBCTooltipDataProvider{DBC: instance}, int64(enchantingSpell.SpellID))

			grp, exists := groupMapProc["Enchants"]
			if !exists {
				grp = Group{Name: "Enchants"}
			}

			renderedTooltip := tooltip.String()
			entry := Entry{Tooltip: strings.Split(renderedTooltip, "\n"), Variants: []*Variant{{ID: int(enchant.EffectId), Name: enchant.Name, SpellID: int(enchantingSpell.SpellID)}}}
			entry.ProcInfo, entry.Supported = BuildEnchantProcInfo(enchant, instance, renderedTooltip)
			grp.Entries = append(grp.Entries, &entry)
			groupMapProc["Enchants"] = grp

			if !entry.Supported {
				StoreMissingEffect("EnchantEffects", enchant.Name, Variant{
					ID:      int(enchant.EffectId),
					Name:    renderedTooltip,
					SpellID: int(enchant.SpellId),
				})
				return EffectParseResultUnsupported
			}

			return EffectParseResultSuccess
		}
	}

	return EffectParseResultInvalid
}

func ParseTooltipForMissingEffect(parsed *proto.UIItem, itemEffect *proto.ItemEffect, instance *dbc.DBC, groupMap map[string]Group, groupMapName string) {
	if parsed.ScalingOptions[0].Ilvl > MIN_EFFECT_ILVL {
		// Effect was already manually implemented
		if core.HasItemEffect(parsed.Id) {
			return
		}

		tooltipString, id := dbc.GetItemEffectSpellTooltip(int(parsed.Id), int(itemEffect.BuffId))
		tooltip, _ := tooltip.ParseTooltip(tooltipString, tooltip.DBCTooltipDataProvider{DBC: instance}, int64(id))

		grp, exists := groupMap[groupMapName]
		if !exists {
			grp = Group{Name: groupMapName}
		}

		if tooltip != nil {
			renderedTooltip := tooltip.String()
			entry := Entry{
				Tooltip:   strings.Split(renderedTooltip, "\n"),
				Supported: false,
				Variants: []*Variant{
					{
						ID:      int(parsed.Id),
						Name:    parsed.Name,
						SpellID: int(itemEffect.BuffId),
					},
				},
			}

			grp.Entries = append(grp.Entries, &entry)
			groupMap[groupMapName] = grp

			// Flavour auras carry no mechanic worth implementing and only add noise to the
			// report. Suppressing the report only, never the group entry: those
			// Supported: false entries take part in the variant grouping that decides whether
			// a whole variant set is emitted live or commented, so dropping one can flip a
			// real registration.
			if _, ignored := IgnoreMissingEffectBySpellID[int(itemEffect.BuffId)]; ignored {
				return
			}

			if len(itemEffect.ScalingOptions[0].Stats) == 0 || !entry.Supported {
				StoreMissingEffect("ItemEffects", parsed.Name, Variant{
					ID:      int(parsed.Id),
					Name:    renderedTooltip,
					SpellID: int(itemEffect.BuffId),
				})
			}
		}
	}
}

var critMatcher = regexp.MustCompile(`critical ([^\s]+|damage,?)( chance)? [^fbc]`)
var pureHealMatcher = regexp.MustCompile(`healing spells`)
var hasHealMatcher = regexp.MustCompile(`heal(ing)?[^,]`)
var hasGenericMatcher = regexp.MustCompile(`a spell`)

// A trigger clause restricted to one named ability: "Your Shock spells", "Your Moonfire ability".
// The capital is what carries the meaning - an unrestricted trigger reads "your spell critical
// strikes" or "each time you cast a spell", with nothing capitalized to name.
var namedAbilityMatcher = regexp.MustCompile(`[Yy]our [A-Z][A-Za-z']*( [A-Z][A-Za-z']*)* (spell|spells|ability|abilities)`)

// Derives what adds a stack to an accumulating aura, from the container spell rather than from
// the one that opens the window. buff_id is the container by then: the parser rebases the effect
// onto it precisely because that is where the duration and these proc flags live.
func buildStackProcInfo(itemEffect *proto.ItemEffect, instance *dbc.DBC, tooltip string) *ProcInfo {
	if itemEffect.StackingAura == nil || itemEffect.GetStackProc() == nil {
		return nil
	}

	container, ok := instance.Spells[int(itemEffect.BuffId)]
	if !ok {
		return nil
	}

	info, supported := BuildSpellProcInfo(&container, tooltip, proto.ItemType_ItemTypeUnknown)
	if !supported {
		return nil
	}
	return &info
}

func BuildProcInfo(parsed *proto.UIItem, itemEffectID int, instance *dbc.DBC, tooltip string) (ProcInfo, bool) {
	itemEffect := dbc.GetItemEffectForBuffID(int(parsed.Id), itemEffectID)
	if itemEffect == nil {
		return ProcInfo{}, false
	}

	// if we have multiple spells find the first that has a proc aura assigned
	procId := itemEffect.SpellID
	procSpell, ok := instance.Spells[int(procId)]
	if !ok {
		panic(fmt.Sprintf("Could not find proc aura %d spell for item effect %d.\n", procId, parsed.Id))
	}

	itemType := proto.ItemType_ItemTypeUnknown
	if itemEffect.TriggerType == 2 {
		itemType = proto.ItemType_ItemTypeWeapon
	}

	procInfo, supported := BuildSpellProcInfo(&procSpell, tooltip, itemType)

	if SpellHasDummyEffect(int(procId), instance) {
		return procInfo, false
	}

	return procInfo, supported
}

func BuildEnchantProcInfo(enchant *proto.UIEnchant, instance *dbc.DBC, tooltip string) (ProcInfo, bool) {
	procSpellID := enchant.SpellId
	if procSpellID == 0 {
		fmt.Printf("WARN: Enchant %d with no spell id", enchant.EffectId)
		return ProcInfo{}, false
	}

	procSpell, ok := instance.Spells[int(procSpellID)]
	if !ok {
		panic(fmt.Sprintf("Could not find proc aura %d spell for item effect %d.\n", procSpellID, enchant.EffectId))
	}

	procInfo, supported := BuildSpellProcInfo(&procSpell, tooltip, enchant.Type)
	if SpellHasDummyEffect(int(procSpellID), instance) {
		return procInfo, false
	}

	return procInfo, supported
}

func BuildSpellProcInfo(procSpell *dbc.Spell, tooltip string, itemType proto.ItemType) (ProcInfo, bool) {
	var info = ProcInfo{
		RequireDamageDealt:  true,
		MaxCumulativeStacks: procSpell.MaxCumulativeStacks,
	}

	requiresOutcome := true
	onHitProc := false

	// On hit proc
	if itemType == proto.ItemType_ItemTypeWeapon {
		onHitProc = true
		info.Callback |= core.CallbackOnSpellHitDealt
		info.ProcMask |= core.ProcMaskUnknown
	}

	if itemType == proto.ItemType_ItemTypeRanged {
		info.Callback |= core.CallbackOnSpellHitDealt
		info.ProcMask |= core.ProcMaskRanged
	}

	if procSpell.OnlyProcsFromClassAbilities() {
		info.ClassSpellsOnly = true
	}

	// A spell-family filter the generated proc cannot reproduce. Two sources of evidence for the
	// same thing: the spell names a family in SpellClassMask, or its tooltip names one. On item
	// procs TBC stores the mask as either nil or all-zero and keeps the real filter server-side,
	// so for those the tooltip is the only evidence there is - "Your Shock spells", "Your Moonfire
	// ability". Generating one anyway would proc it off every spell instead of that one.
	if slices.ContainsFunc(procSpell.SpellClassMask, func(mask int) bool { return mask != 0 }) ||
		namedAbilityMatcher.MatchString(tooltip) {
		return info, false
	}

	if !onHitProc && len(procSpell.ProcTypeMask) > 0 {
		if procSpell.ProcTypeMask[0]&dbc.PROC_FLAG_DEAL_MELEE_SWING > 0 {
			info.ProcMask |= core.ProcMaskMeleeWhiteHit
		}

		if procSpell.ProcTypeMask[0]&dbc.PROC_FLAG_DEAL_MELEE_ABILITY > 0 {
			info.ProcMask |= core.ProcMaskMeleeSpecial
		}

		if procSpell.ProcTypeMask[0]&dbc.PROC_FLAG_DEAL_RANGED_ATTACK > 0 {
			info.ProcMask |= core.ProcMaskRangedAuto
		}

		if procSpell.ProcTypeMask[0]&dbc.PROC_FLAG_DEAL_RANGED_ABILITY > 0 {
			info.ProcMask |= core.ProcMaskRangedSpecial
		}

		if procSpell.ProcTypeMask[0]&dbc.PROC_FLAG_DEAL_HARMFUL_PERIODIC > 0 {
			info.ProcMask |= core.ProcMaskSpellDamage
		}

		if procSpell.ProcTypeMask[0]&dbc.PROC_FLAG_DEAL_HARMFUL_SPELL > 0 {
			info.ProcMask |= core.ProcMaskSpellDamage
		}

		if procSpell.ProcTypeMask[0]&dbc.PROC_FLAG_ANY_DIRECT_TAKEN > 0 {
			info.Callback |= core.CallbackOnSpellHitTaken
			info.Outcome = core.OutcomeLanded

			if procSpell.ProcTypeMask[0]&dbc.PROC_FLAG_TAKE_MELEE_SWING > 0 {
				info.ProcMask |= core.ProcMaskMeleeWhiteHit
			}

			if procSpell.ProcTypeMask[0]&dbc.PROC_FLAG_TAKE_MELEE_ABILITY > 0 {
				info.ProcMask |= core.ProcMaskMeleeSpecial
			}

			if procSpell.ProcTypeMask[0]&dbc.PROC_FLAG_TAKE_RANGED_ATTACK > 0 {
				info.ProcMask |= core.ProcMaskRangedAuto
			}

			if procSpell.ProcTypeMask[0]&dbc.PROC_FLAG_TAKE_RANGED_ABILITY > 0 {
				info.ProcMask |= core.ProcMaskRangedSpecial
			}

			if procSpell.ProcTypeMask[0]&dbc.PROC_FLAG_TAKE_HARMFUL_SPELL > 0 {
				info.ProcMask |= core.ProcMaskSpellDamage
			}
		}

		// In TBC spells whose mask carries only the spell-cast bits seem to not care about landing
		// or not, harmful and helpful alike. A tooltip naming an outcome is the exception: a crit
		// is only known once the hit resolves, so those stay on hit-dealt.
		// The harmful bit has to be one of them. A helpful-only mask carries no evidence that
		// casting is the trigger at all, and the helpful branch below already demands tooltip
		// evidence before it believes one - the PvP Librams that buff a heal target read
		// "Causes your Flash of Light to increase the target's Resilience" and are neither a
		// self buff nor unrestricted.
		castOnly := procSpell.ProcTypeMask[0]&dbc.PROC_FLAG_DEAL_HARMFUL_SPELL > 0 &&
			procSpell.ProcTypeMask[0]&^(dbc.PROC_FLAG_DEAL_HARMFUL_SPELL|dbc.PROC_FLAG_DEAL_HELPFUL_SPELL) == 0

		if castOnly && !critMatcher.MatchString(tooltip) {
			info.Callback |= core.CallbackOnCastComplete
			info.RequireDamageDealt = false
			requiresOutcome = false
		} else if procSpell.ProcTypeMask[0]&dbc.PROC_FLAG_ANY_DIRECT_DEALT > 0 {
			info.Callback |= core.CallbackOnSpellHitDealt

			if procSpell.ProcTypeMask[0]&dbc.PROC_FLAG_DEAL_HARMFUL_SPELL > 0 {
				info.RequireDamageDealt = false
			}
		}

		if procSpell.ProcTypeMask[0]&dbc.PROC_FLAG_DEAL_HARMFUL_PERIODIC > 0 {
			info.Callback |= core.CallbackOnPeriodicDamageDealt
		}

		if procSpell.ProcTypeMask[0]&dbc.PROC_FLAG_DEAL_HELPFUL_SPELL > 0 &&
			(hasHealMatcher.MatchString(tooltip) || hasGenericMatcher.MatchString(tooltip)) {
			info.RequireDamageDealt = false
			info.ProcMask |= core.ProcMaskSpellHealing

			// Casting the heal is already the trigger above, so adding heal-dealt on top would
			// proc twice for one heal.
			if !info.Callback.Matches(core.CallbackOnCastComplete) {
				info.Callback |= core.CallbackOnHealDealt

				// handle HoTs only with direct heals for now, there are some odd cases with HoT / DoT overlaps
				if procSpell.ProcTypeMask[0]&dbc.PROC_FLAG_DEAL_HELPFUL_PERIODIC > 0 {
					info.Callback |= core.CallbackOnPeriodicHealDealt
				}

				// Check if we have periodic damage flag but only heal paired with it
				// This usually indicates a pure heal proc mask
				if procSpell.ProcTypeMask[0]&dbc.PROC_FLAG_ANY_DIRECT_DEALT == 0 {
					info.Callback &= ^core.CallbackOnPeriodicDamageDealt
					info.Callback &= ^core.CallbackOnSpellHitDealt
					info.ProcMask &= ^core.ProcMaskSpellDamage
				}
			}
		}
	}

	if info.ProcMask.Matches(core.ProcMaskMelee) && procSpell.CanProcFromProcs() {
		info.ProcMask |= core.ProcMaskMeleeProc
	}

	if info.ProcMask.Matches(core.ProcMaskRanged) && procSpell.CanProcFromProcs() {
		info.ProcMask |= core.ProcMaskRangedProc
	}

	if info.ProcMask.Matches(core.ProcMaskSpellDamage) && procSpell.CanProcFromProcs() {
		info.ProcMask |= core.ProcMaskSpellProc
	}

	if requiresOutcome {
		if critMatcher.MatchString(tooltip) {
			info.Outcome = core.OutcomeCrit
		} else {
			info.Outcome = core.OutcomeLanded
		}
	}

	// check for pure healing spell
	if pureHealMatcher.MatchString(tooltip) {
		info.Callback &= ^core.CallbackOnSpellHitDealt
		info.Callback &= ^core.CallbackOnPeriodicDamageDealt
	}

	// A trigger with no callback never fires, and factory_ProcStatBonusEffect returns early on
	// exactly that, so such an effect cannot be generated. The test used to also require an
	// empty Outcome and ProcMask, which it can never have: when requiresOutcome is true the
	// Outcome is set a few lines up, and when it is false the whole term is false, so nothing
	// was ever refused here and empty-callback effects were emitted as live registrations that
	// silently did nothing.
	return info, info.Callback != core.CallbackEmpty
}

func StoreMissingEffect(effectType string, name string, variant Variant) {
	if missingEffectsMap[effectType] == nil {
		missingEffectsMap[effectType] = map[int32]MissingItemEffect{}
	}
	id := int32(variant.ID)
	if missingEffectsMap[effectType][id].Effects == nil {
		missingEffectsMap[effectType][id] = MissingItemEffect{
			ItemID:  id,
			Name:    name,
			Effects: []Variant{},
		}
	}
	itemEntry := missingEffectsMap[effectType][id]
	haveEffect := false
	for _, effect := range itemEntry.Effects {
		if effect.SpellID == variant.SpellID {
			haveEffect = true
			break
		}
	}
	if haveEffect {
		return
	}

	itemEntry.Effects = append(
		itemEntry.Effects,
		variant,
	)
	missingEffectsMap[effectType][id] = itemEntry
}

func asCoreCallback(callback core.AuraCallback) string {
	callbacks := []string{}
	for i := range 32 {
		callbackFlag := core.AuraCallback(1 << i)
		if callbackFlag >= core.CallbackLast {
			break
		}

		if callback.Matches(callbackFlag) {
			callbacks = append(callbacks, "core."+callbackFlag.String())
		}
	}

	if len(callbacks) == 0 {
		return "core.CallbackEmpty"
	}

	return strings.Join(callbacks, " | ")
}

func asCoreProcMask(procMask core.ProcMask) string {
	procs := []string{}
	for i := range 32 {
		procFlag := core.ProcMask(1 << i)
		if procFlag >= core.ProcMaskLast {
			break
		}

		if procMask.Matches(procFlag) {
			procs = append(procs, "core."+procFlag.String())
		}
	}

	if len(procs) == 0 {
		return "core.ProcMaskUnknown"
	}
	return strings.Join(procs, " | ")
}

func asCoreOutcome(outcome core.HitOutcome) string {
	if outcome == core.OutcomeCrit {
		return "core.OutcomeCrit"
	}

	if outcome.Matches(core.OutcomeLanded) {
		return "core.OutcomeLanded"
	}

	return "core.OutcomeEmpty"
}

func (entry *Entry) AddVariant(variant *Variant) {
	entry.Variants = append(entry.Variants, variant)
	sort.Slice(entry.Variants, func(i, j int) bool {
		return entry.Variants[i].ID < entry.Variants[j].ID
	})
}
