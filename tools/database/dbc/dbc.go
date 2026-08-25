package dbc

import (
	"encoding/json"
	"fmt"
	"log"
	"slices"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"
)

type DBC struct {
	Items                  map[int]Item                       // Item ID
	Gems                   map[int]Gem                        // Item ID
	Enchants               map[int]Enchant                    // Index in the raw SpellItemEnchantment list
	EnchantsByEffectId     map[int]Enchant                    // ItemEchantment ID
	ItemStatEffects        map[int]ItemStatEffect             // ItemID? something anyway
	SpellEffects           map[int]map[int]SpellEffect        // Search by spellID and effect index
	SpellEffectsById       map[int]SpellEffect                // Search by effectid
	Spells                 map[int]Spell                      // Search by spellId
	RandomSuffix           map[int]RandomSuffix               // Item level
	ItemDamageTable        map[string]map[int]ItemDamageTable // By Table name and item level
	RandomPropertiesByIlvl map[int]RandomPropAllocationMap
	ItemArmorQuality       map[int]ItemArmorQuality
	//ItemArmorShield        map[int]ItemArmorShield
	ItemArmorTotal        map[int]ItemArmorTotal
	ArmorLocation         map[int]ArmorLocation
	SpellScalings         map[int]SpellScaling
	Consumables           map[int]Consumable   // Item ID
	ItemEffects           map[int]ItemEffect   // Effect ID
	ItemEffectsByParentID map[int][]ItemEffect // ParentItemID
	ShieldBlockValues     map[int]ShieldBlock

	// Every spell's effects flattened into effect index order. Built once by indexSpellEffects and
	// served by SpellEffectsInOrder, which every trigger chain walk goes through - sorting per call
	// meant two allocations and a sort for each spell visited.
	spellEffectsOrdered map[int][]SpellEffect
}

func NewDBC() *DBC {
	return &DBC{
		Items:                  make(map[int]Item),
		Gems:                   make(map[int]Gem),
		Enchants:               make(map[int]Enchant),
		EnchantsByEffectId:     make(map[int]Enchant),
		ItemStatEffects:        make(map[int]ItemStatEffect),
		SpellEffects:           make(map[int]map[int]SpellEffect),
		SpellEffectsById:       make(map[int]SpellEffect),
		Spells:                 make(map[int]Spell),
		RandomSuffix:           make(map[int]RandomSuffix),
		ItemDamageTable:        make(map[string]map[int]ItemDamageTable),
		RandomPropertiesByIlvl: make(map[int]RandomPropAllocationMap),
		ItemArmorQuality:       make(map[int]ItemArmorQuality),
		//ItemArmorShield:        make(map[int]ItemArmorShield),
		ItemArmorTotal:        make(map[int]ItemArmorTotal),
		ArmorLocation:         make(map[int]ArmorLocation),
		Consumables:           make(map[int]Consumable),
		ItemEffects:           make(map[int]ItemEffect),
		SpellScalings:         make(map[int]SpellScaling),
		ItemEffectsByParentID: make(map[int][]ItemEffect),
		ShieldBlockValues:     make(map[int]ShieldBlock),
	}
}

var (
	dbcInstance *DBC
	once        sync.Once
)

func InitDBC() error {
	dbcInstance = NewDBC()

	// The inputs are independent files folding into disjoint DBC fields, so they load
	// concurrently.
	var g errgroup.Group
	load := func(name string, loadFn func(string) error, path string) {
		g.Go(func() error {
			if err := loadFn(path); err != nil {
				return fmt.Errorf("loading %s: %w", name, err)
			}
			return nil
		})
	}
	load("items", dbcInstance.loadItems, "./assets/db_inputs/dbc/items.json")
	load("gems", dbcInstance.loadGems, "./assets/db_inputs/dbc/gems.json")
	load("enchants", dbcInstance.loadEnchants, "./assets/db_inputs/dbc/enchants.json")
	load("item stat effects", dbcInstance.loadItemStatEffects, "./assets/db_inputs/dbc/item_stat_effects.json")
	load("spell effects", dbcInstance.loadSpellEffects, "./assets/db_inputs/dbc/spell_effects.json")
	load("random suffixes", dbcInstance.loadRandomSuffix, "./assets/db_inputs/dbc/random_suffix.json")
	load("random properties", dbcInstance.loadRandomPropertiesByIlvl, "./assets/db_inputs/dbc/rand_prop_points.json")
	load("item damage tables", dbcInstance.loadItemDamageTables, "./assets/db_inputs/dbc/item_damage_tables.json")
	load("item armor quality", dbcInstance.LoadItemArmorQuality, "./assets/db_inputs/dbc/item_armor_quality.json")
	load("item armor total", dbcInstance.LoadItemArmorTotal, "./assets/db_inputs/dbc/item_armor_total.json")
	// load("item armor shield", dbcInstance.LoadItemArmorShield, "./assets/db_inputs/dbc/item_armor_shield.json")
	load("armor location", dbcInstance.LoadArmorLocation, "./assets/db_inputs/dbc/armor_location.json")
	load("consumables", dbcInstance.loadConsumables, "./assets/db_inputs/dbc/consumables.json")
	load("item effects", dbcInstance.loadItemEffects, "./assets/db_inputs/dbc/item_effects.json")
	load("spells", dbcInstance.loadSpells, "./assets/db_inputs/dbc/spells.json")
	if err := g.Wait(); err != nil {
		return err
	}
	dbcInstance.LoadSpellScaling()
	dbcInstance.LoadShieldBlockValues()
	dbcInstance.indexSpellEffects()
	return nil
}

// Flattens every spell's effects into effect index order, once, after all inputs are loaded.
// SpellEffects is keyed by index, so ranging over it directly yields a random order and makes any
// traversal that stops at the first match non-deterministic.
func (d *DBC) indexSpellEffects() {
	d.spellEffectsOrdered = make(map[int][]SpellEffect, len(d.SpellEffects))

	for spellID, effects := range d.SpellEffects {
		indices := make([]int, 0, len(effects))
		for idx := range effects {
			indices = append(indices, idx)
		}
		slices.Sort(indices)

		ordered := make([]SpellEffect, 0, len(effects))
		for _, idx := range indices {
			ordered = append(ordered, effects[idx])
		}
		d.spellEffectsOrdered[spellID] = ordered
	}
}

// Returns the effects of a spell ordered by effect index, or nil if it has none. The slice is
// shared with every other caller, so it is to be read and not written to.
func (d *DBC) SpellEffectsInOrder(spellID int) []SpellEffect {
	return d.spellEffectsOrdered[spellID]
}

// GetDBC returns the DBC singleton instance
func GetDBC() *DBC {
	once.Do(func() {
		if err := InitDBC(); err != nil {
			log.Fatalf("Failed to initialize DBC: %v", err)
		}
	})
	return dbcInstance
}

func (d *DBC) loadConsumables(filename string) error {
	data, err := ReadGzipFile(filename)
	if err != nil {
		return err
	}

	var consumables []Consumable
	if err = json.Unmarshal(data, &consumables); err != nil {
		return ParseError{
			Source: filename,
			Field:  "Consumable",
			Reason: err.Error(),
		}
	}

	for i := range consumables {
		consumable := consumables[i]
		d.Consumables[consumable.Id] = consumable
	}
	return nil
}
func (d *DBC) loadSpells(filename string) error {
	data, err := ReadGzipFile(filename)
	if err != nil {
		return err
	}

	var spells []Spell
	if err = json.Unmarshal(data, &spells); err != nil {
		return ParseError{
			Source: filename,
			Field:  "Spell",
			Reason: err.Error(),
		}
	}

	for i := range spells {
		spell := spells[i]
		d.Spells[int(spell.ID)] = spell
	}
	return nil
}
func (d *DBC) loadItemEffects(filename string) error {
	data, err := ReadGzipFile(filename)
	if err != nil {
		return err
	}

	var effects []ItemEffect
	if err = json.Unmarshal(data, &effects); err != nil {
		return ParseError{
			Source: filename,
			Field:  "ItemEffect",
			Reason: err.Error(),
		}
	}

	// Populate both maps
	for _, effect := range effects {
		// Single lookup by effect ID
		d.ItemEffects[effect.ID] = effect
		// Grouping by parent item ID
		d.ItemEffectsByParentID[effect.ParentItemID] = append(
			d.ItemEffectsByParentID[effect.ParentItemID],
			effect,
		)
	}

	return nil
}

func (d *DBC) loadRandomPropertiesByIlvl(filename string) error {
	data, err := ReadGzipFile(filename)
	if err != nil {
		return err
	}

	var properties RandomPropAllocationsByIlvl
	if err = json.Unmarshal(data, &properties); err != nil {
		return ParseError{
			Source: filename,
			Field:  "RandomProps",
			Reason: err.Error(),
		}
	}

	d.RandomPropertiesByIlvl = properties
	return nil
}

func (d *DBC) loadItems(filename string) error {
	data, err := ReadGzipFile(filename)
	if err != nil {
		return err
	}

	var items []Item
	if err = json.Unmarshal(data, &items); err != nil {
		return ParseError{
			Source: filename,
			Field:  "Item",
			Reason: err.Error(),
		}
	}

	for i := range items {
		item := items[i]
		d.Items[item.Id] = item
	}
	return nil
}

func (d *DBC) loadGems(filename string) error {
	data, err := ReadGzipFile(filename)
	if err != nil {
		return err
	}

	var gems []Gem
	if err = json.Unmarshal(data, &gems); err != nil {
		return ParseError{
			Source: filename,
			Field:  "Gem",
			Reason: err.Error(),
		}
	}

	for i := range gems {
		gem := gems[i]
		d.Gems[gem.ItemId] = gem
	}
	return nil
}

func (d *DBC) loadEnchants(filename string) error {
	data, err := ReadGzipFile(filename)
	if err != nil {
		return err
	}

	var enchants []Enchant
	if err = json.Unmarshal(data, &enchants); err != nil {
		return ParseError{
			Source: filename,
			Field:  "Enchant",
			Reason: err.Error(),
		}
	}

	for i, ench := range enchants {
		if strings.Contains(ench.Name, "QASpell") {
			continue
		}
		d.Enchants[i] = ench
		d.EnchantsByEffectId[ench.EffectId] = ench
	}
	return nil
}

func (d *DBC) loadItemStatEffects(filename string) error {
	data, err := ReadGzipFile(filename)
	if err != nil {
		return err
	}

	var effects []ItemStatEffect
	if err = json.Unmarshal(data, &effects); err != nil {
		return ParseError{
			Source: filename,
			Field:  "ItemStatEffect",
			Reason: err.Error(),
		}
	}

	for i := range effects {
		effect := effects[i]
		d.ItemStatEffects[effect.ID] = effect
	}
	return nil
}

func (d *DBC) loadSpellEffects(filename string) error {
	data, err := ReadGzipFile(filename)
	if err != nil {
		return err
	}

	var effects map[int]map[int]SpellEffect
	if err = json.Unmarshal(data, &effects); err != nil {
		return ParseError{
			Source: filename,
			Field:  "SpellEffect",
			Reason: err.Error(),
		}
	}

	d.SpellEffects = effects
	for _, spell := range effects {
		for _, effect := range spell {
			d.SpellEffectsById[effect.ID] = effect
		}
	}
	return nil
}

func (d *DBC) loadRandomSuffix(filename string) error {
	data, err := ReadGzipFile(filename)
	if err != nil {
		return err
	}

	var suffixes []RandomSuffix
	if err = json.Unmarshal(data, &suffixes); err != nil {
		return ParseError{
			Source: filename,
			Field:  "RandomSuffix",
			Reason: err.Error(),
		}
	}

	for i := range suffixes {
		suffix := suffixes[i]
		d.RandomSuffix[suffix.ID] = suffix
	}
	return nil
}

func (d *DBC) LoadItemArmorQuality(filename string) error {
	data, err := ReadGzipFile(filename)
	if err != nil {
		return err
	}

	var tables map[int]ItemArmorQuality
	if err = json.Unmarshal(data, &tables); err != nil {
		return ParseError{
			Source: filename,
			Field:  "ItemArmorQuality",
			Reason: err.Error(),
		}
	}

	d.ItemArmorQuality = tables
	return nil
}
func (d *DBC) LoadArmorLocation(filename string) error {
	data, err := ReadGzipFile(filename)
	if err != nil {
		return err
	}

	var tables map[int]ArmorLocation
	if err = json.Unmarshal(data, &tables); err != nil {
		return ParseError{
			Source: filename,
			Field:  "ArmorLocation",
			Reason: err.Error(),
		}
	}

	d.ArmorLocation = tables
	return nil
}

// func (d *DBC) LoadItemArmorShield(filename string) error {
// 	data, err := ReadGzipFile(filename)
// 	if err != nil {
// 		return err
// 	}

// 	var tables map[int]ItemArmorShield
// 	if err = json.Unmarshal(data, &tables); err != nil {
// 		return ParseError{
// 			Source: filename,
// 			Field:  "ItemArmorShield",
// 			Reason: err.Error(),
// 		}
// 	}

// 	d.ItemArmorShield = tables
// 	return nil
// }

func (d *DBC) LoadItemArmorTotal(filename string) error {
	data, err := ReadGzipFile(filename)
	if err != nil {
		return err
	}

	var tables map[int]ItemArmorTotal
	if err = json.Unmarshal(data, &tables); err != nil {
		return ParseError{
			Source: filename,
			Field:  "ItemArmorTotal",
			Reason: err.Error(),
		}
	}

	d.ItemArmorTotal = tables
	return nil
}

func (d *DBC) loadItemDamageTables(filename string) error {
	data, err := ReadGzipFile(filename)
	if err != nil {
		return err
	}

	var tables map[string]map[int]ItemDamageTable
	if err = json.Unmarshal(data, &tables); err != nil {
		return ParseError{
			Source: filename,
			Field:  "ItemDamage",
			Reason: err.Error(),
		}
	}

	d.ItemDamageTable = tables
	return nil
}
