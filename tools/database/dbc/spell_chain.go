package dbc

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/wowsims/tbc/sim/core/proto"
)

// Walks the chain of spells reachable through SpellEffect.EffectTriggerSpell, visiting each
// spell at most once.
//
// The traversal policy differs per caller - some examine a whole level before descending, some
// interleave, some fall back to siblings - so each keeps its own recursion. What they share, and
// what used to be re-implemented per caller, is this: proc chains can loop back on themselves,
// and the effects of a spell have to be read in index order or any traversal that stops at the
// first match becomes non-deterministic across runs.
type chainWalker struct {
	visited map[int]bool
}

func newChainWalker() *chainWalker {
	return &chainWalker{visited: map[int]bool{}}
}

// Returns spellID's effects in index order the first time spellID is reached, and nil on every
// later visit so a cyclic chain terminates. Callers treat that the same way they treat a spell
// with no effects.
func (w *chainWalker) effects(spellID int) []SpellEffect {
	if w.visited[spellID] {
		return nil
	}
	w.visited[spellID] = true
	return dbcInstance.SpellEffectsInOrder(spellID)
}

// A container aura and the stat aura it accumulates, with what makes a stack land.
//
// The stat aura is itself a proto.ItemEffect: it has a buff, a duration, a stack cap and scaled
// stats, which is exactly that shape. The container is reported separately because the window's
// duration and its proc flags belong to it rather than to the stat aura.
type stackingAura struct {
	Aura        *proto.ItemEffect
	ContainerID int
	Proc        *proto.ProcEffect
	// Set for the trinkets that start at full stacks and lose one per event instead of building
	// up. Read from the container's description, because the decrement is server script.
	Decays bool
}

// Resolves the stacking shape reachable from an item effect's spell, or nil when there is none.
//
// For example: Blackened Naaru Sliver is the canonical case: the item's effect triggers 45040 Battle Trance,
// a 20s aura that grants nothing, which in turn triggers 45041 Combat Insight - 10 stacks of
// attack power with no duration of its own, because the client drops it when the window ends.
// resolveStatsSpell walks straight past the container to reach the stats, so the container has
// to be recovered by walking the chain a second time.
func buildStackingAura(rootSpellID, statsSpellID, itemLevel, parentItemID int) *stackingAura {
	containerID := newChainWalker().findStackingContainer(rootSpellID, statsSpellID)
	if containerID == 0 {
		return nil
	}

	// The aura the edge leads to is not always the one holding the stats.
	//
	// For Example: Zandalarian Hero Badge triggers a script marker with no effects at all and
	// name the real aura only in the container's description text, so fall back to that
	// when the edge resolves nothing worth having.
	statAuraID := statsSpellID
	if len(collectStats(statAuraID, itemLevel).ToProtoMap()) == 0 {
		if referenced := referencedStatAura(containerID, itemLevel); referenced != 0 {
			statAuraID = referenced
		}
	}

	statAura := dbcInstance.Spells[statAuraID]
	if statAura.MaxCumulativeStacks < 2 {
		// One stack is not a stacking trinket. Syphon of the Nathrezim reaches here with a leech
		// effect that happens to sit behind a window aura.
		return nil
	}

	// Petrified Scarab is the case that ends here: its container triggers a script marker with no
	// effects and writes its numbers into the tooltip as literals, so nothing in the data names
	// the aura that actually carries the stats. Returning nil leaves the effect with no stats,
	// which sends it to the missing-effects report to be implemented by hand rather than guessed
	// at.
	if len(collectStats(statAuraID, itemLevel).ToProtoMap()) == 0 {
		return nil
	}

	// A container with ProcChance above 100 is the "rate lives somewhere else" sentinel - Badge of
	// the Swarmguard reads 101 and really procs at 10 PPM. That number is not in the spell data,
	// so it comes from MapItemIdToPPM, the same table assignTrigger uses for chance-on-hit items.
	// With no entry there the rate is unknown, and emitting the shape anyway would leave the stack
	// trigger with no rate at all - worse than not generating it, so abandon and let the effect
	// reach the missing-effects report.
	container := dbcInstance.Spells[containerID]
	proc := &proto.ProcEffect{IcdMs: procIcdMs(container, dbcInstance.Spells[statAuraID])}
	switch {
	case container.ProcChance > 100:
		ppm := getPPMForItemID(int32(parentItemID))
		if ppm == 0 {
			ReportMissingPPM(int32(parentItemID), containerID)
			return nil
		}
		proc.ProcRate = &proto.ProcEffect_Ppm{Ppm: ppm}
	case container.ProcChance > 0:
		proc.ProcRate = &proto.ProcEffect_ProcChance{ProcChance: float64(container.ProcChance) / 100}
	}

	return &stackingAura{
		Aura: &proto.ItemEffect{
			BuffId:              int32(statAuraID),
			BuffName:            statAura.NameLang,
			MaxCumulativeStacks: statAura.MaxCumulativeStacks,
			ScalingOptions:      map[int32]*proto.ScalingItemEffectProperties{},
		},
		ContainerID: containerID,
		Proc:        proc,
		Decays:      stacksDecay.MatchString(container.Description),
	}
}

// A $<spellID><token> reference inside a spell description, e.g. the "$24575m1" in Zandalarian
// Hero Badge's "Increases your armor by ${$24575m1*10}". Four digits is the shortest real spell
// ID, which keeps the plain $s1 / $t1 / $d tokens out.
var descriptionSpellRef = regexp.MustCompile(`\$(\d{4,7})[a-zA-Z]`)

// Trinkets that start full and lose a stack per event say so in prose and nowhere else - the
// decrement is server script, with no effect or flag behind it.
var stacksDecay = regexp.MustCompile(`(?i)(bonus|this) is reduced by`)

// Resolves the stat aura a container names only in its description text.
//
// Returns 0 when the description references nothing that resolves to stats, which is the
// overwhelming majority. A description can mention spells for all sorts of reasons, so the one
// that actually carries stats is the one being looked for.
func referencedStatAura(containerSpellID int, itemLevel int) int {
	// Mining the English description is the only link there is, so an ambiguous one has to be
	// visible rather than silently resolved by ordering: a reworded or hotfixed description that
	// happens to name another stat-resolving spell first would otherwise write the wrong buff ID
	// and the wrong stats into the database with no diagnostic at all.
	chosen := 0
	// Counted per distinct spell: one aura referenced as both $24575m1 and $24575m2 is one
	// candidate, not two.
	resolved := map[int]bool{}

	for _, match := range descriptionSpellRef.FindAllStringSubmatch(dbcInstance.Spells[containerSpellID].Description, -1) {
		referenced, err := strconv.Atoi(match[1])
		if err != nil || referenced == containerSpellID || resolved[referenced] {
			continue
		}
		if dbcInstance.Spells[referenced].ID == 0 {
			continue
		}
		if len(collectStats(referenced, itemLevel).ToProtoMap()) > 0 {
			resolved[referenced] = true
			if chosen == 0 {
				chosen = referenced
			}
		}
	}

	if len(resolved) > 1 {
		fmt.Printf("WARN: spell %d's description names %d stat-resolving spells, taking the first (%d)\n", containerSpellID, len(resolved), chosen)
	}

	return chosen
}

// Finds the aura that bounds a stacking stat aura: the spell in the chain above statAuraID that
// triggers it and carries the duration it lives inside.
//
// The stat aura having no duration of its own is what separates this shape from an ordinary proc
// that happens to stack, where the buff expires on its own schedule and refreshing it is correct.
func (w *chainWalker) findStackingContainer(spellID, statAuraID int) int {
	statAura := dbcInstance.Spells[statAuraID]
	if statAura.MaxCumulativeStacks <= 0 || statAura.Duration > 0 {
		return 0
	}

	for _, se := range w.effects(spellID) {
		if se.EffectTriggerSpell == 0 {
			continue
		}

		if se.EffectTriggerSpell == statAuraID &&
			isProcTriggerAura(se.EffectAura) &&
			dbcInstance.Spells[spellID].Duration > 0 {
			return spellID
		}

		if found := w.findStackingContainer(se.EffectTriggerSpell, statAuraID); found != 0 {
			return found
		}
	}
	return 0
}

func isProcTriggerAura(aura EffectAuraType) bool {
	return aura == A_PROC_TRIGGER_SPELL || aura == A_PROC_TRIGGER_SPELL_WITH_VALUE
}

// Attaches an accumulating aura to the effect that bounds it.
//
// The effect is rebased onto the container: resolveStatsSpell walked past it to reach the stats,
// leaving buff_id and the duration pointing at an aura that grants nothing by itself and never
// expires on its own.
func applyStackingAura(pe *proto.ItemEffect, stacking *stackingAura) {
	pe.StackingAura = stacking.Aura
	pe.StackTrigger = &proto.ItemEffect_StackProc{StackProc: stacking.Proc}
	pe.StacksDecay = stacking.Decays

	container := dbcInstance.Spells[stacking.ContainerID]
	pe.BuffId = int32(stacking.ContainerID)
	pe.BuffName = container.NameLang
	pe.EffectDurationMs = container.Duration
	// The stacks belong to the child; the container carries none and would otherwise leave the
	// flat stacking path in the sim competing with this one.
	pe.MaxCumulativeStacks = 0
}
