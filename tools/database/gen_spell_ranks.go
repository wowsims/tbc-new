package database

import (
	"database/sql"
	"fmt"
	"go/format"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/wowsims/tbc/tools/database/dbc"
)

var rankSubtext = regexp.MustCompile(`^Rank (\d+)$`)

type generatedRow struct {
	Rank       int32
	SpellID    int32
	Cost       int32
	CastTimeMs int32
	GCDMs      int32
	CooldownMs int32
	MaxRange   float64
	Direct     *generatedAmount
	Heal       *generatedAmount
	Periodic   *generatedAmount
	Energize   *generatedAmount
}

type generatedAmount struct {
	Min    float64
	Max    float64
	Coef   float64
	APCoef float64

	PeriodMs int32
	Ticks    int32
}

type rankCandidate struct {
	SpellID   int32
	Rank      int32
	ClassMask int
	SkillLine int32
}

type rankLadder struct {
	Name  string
	Field string
	Ranks map[int32]int32
}

// SkillLineAbility.ClassMask is a bitmask over the class index dbc.Classes already carries, so the bit
// is that index shifted rather than a second table to keep in step with it.
func classMaskOf(class dbc.DbcClass) int {
	return 1 << (class.ID - 1)
}

// The Go identifier a family is reached by: "Shadow Word: Pain" -> ShadowWordPain.
func fieldNameOf(spellName string) string {
	var b strings.Builder
	upper := true
	for _, r := range spellName {
		switch {
		case r == '\'' || r == '\u2019':
			// Dropped without breaking the word, so "Avenger's Shield" is AvengersShield rather than
			// AvengerSShield.
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			if upper {
				b.WriteRune(unicode.ToUpper(r))
				upper = false
			} else {
				b.WriteRune(r)
			}
		default:
			upper = true
		}
	}

	name := b.String()
	if name == "" || unicode.IsDigit(rune(name[0])) {
		return ""
	}
	return name
}

// Every spell family a class can learn that has more than one rank.
//
// Driven entirely off the client data - the class's own skill lines, every spell in them whose subtext
// reads "Rank N", grouped by name - so nothing here is hand-maintained and a family the sim has not
// implemented yet still gets a table, ready for whoever wants it.
func discoverLadders(db *sql.DB, class dbc.DbcClass) ([]rankLadder, []string, error) {
	mask := classMaskOf(class)

	exclusive, err := exclusiveSkillLines(db, mask)
	if err != nil {
		return nil, nil, err
	}

	rows, err := db.Query(`
		SELECT n.Name_lang, sla.Spell, s.NameSubtext_lang, sla.ClassMask, sla.SkillLine
		FROM SkillLineAbility sla
		JOIN SpellName n ON n.ID = sla.Spell
		JOIN Spell s ON s.ID = sla.Spell
		WHERE sla.SkillLine IN (
			SELECT DISTINCT sla2.SkillLine
			FROM SkillLineAbility sla2
			JOIN SkillLine sl2 ON sl2.ID = sla2.SkillLine AND sl2.CategoryID = 7
			WHERE (sla2.ClassMask & ?) != 0
		)
		AND s.NameSubtext_lang LIKE 'Rank %'
		ORDER BY n.Name_lang, sla.Spell`, mask)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	byName := map[string]map[int32][]rankCandidate{}
	for rows.Next() {
		var name, subtext string
		var c rankCandidate
		if err := rows.Scan(&name, &c.SpellID, &subtext, &c.ClassMask, &c.SkillLine); err != nil {
			return nil, nil, err
		}
		m := rankSubtext.FindStringSubmatch(subtext)
		if m == nil {
			continue
		}
		rank, _ := strconv.Atoi(m[1])
		c.Rank = int32(rank)
		if byName[name] == nil {
			byName[name] = map[int32][]rankCandidate{}
		}
		byName[name][c.Rank] = append(byName[name][c.Rank], c)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	names := make([]string, 0, len(byName))
	for name := range byName {
		names = append(names, name)
	}
	sort.Strings(names)

	var ladders []rankLadder
	var skipped []string
	taken := map[string]string{}

	for _, name := range names {
		field := fieldNameOf(name)
		if field == "" {
			skipped = append(skipped, fmt.Sprintf("%s: no usable Go identifier", name))
			continue
		}
		if owner, clash := taken[field]; clash {
			skipped = append(skipped, fmt.Sprintf("%s: identifier %s already taken by %s", name, field, owner))
			continue
		}

		if !claimedByClass(byName[name], mask, exclusive) {
			continue
		}

		ladder, err := resolveLadder(db, byName[name], mask)
		if err != nil {
			skipped = append(skipped, fmt.Sprintf("%s: %s", name, err))
			continue
		}
		if len(ladder) == 0 {
			continue
		}

		taken[field] = name
		ladders = append(ladders, rankLadder{Name: name, Field: field, Ranks: ladder})
	}

	return ladders, skipped, nil
}

// The skill lines only this class appears in.
//
// Most are exclusive - "Fire" carries the mage bit and nothing else - but 11 are shared, and "Holy"
// holding both the paladin and the priest bit is how paladin Holy Shock first reached the priest file.
// Inside an exclusive line a ClassMask-0 spell can only be this class's, which is what makes a pure
// talent ladder like Ignite - every rank ClassMask 0 - attributable at all.
func exclusiveSkillLines(db *sql.DB, mask int) (map[int32]bool, error) {
	rows, err := db.Query(`
		SELECT SkillLine, group_concat(DISTINCT ClassMask)
		FROM SkillLineAbility WHERE ClassMask != 0 GROUP BY SkillLine`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lines := map[int32]bool{}
	for rows.Next() {
		var line int32
		var masks string
		if err := rows.Scan(&line, &masks); err != nil {
			return nil, err
		}
		union := 0
		for _, part := range strings.Split(masks, ",") {
			m, err := strconv.Atoi(part)
			if err != nil {
				continue
			}
			union |= m
		}
		if union == mask {
			lines[line] = true
		}
	}
	return lines, rows.Err()
}

// A family is this class's if one of its ranks carries the class bit, or if its ranks sit in a skill
// line no other class appears in.
func claimedByClass(byRank map[int32][]rankCandidate, mask int, exclusive map[int32]bool) bool {
	for _, cands := range byRank {
		for _, c := range cands {
			if c.ClassMask&mask != 0 || exclusive[c.SkillLine] {
				return true
			}
		}
	}
	return false
}

// Picks one spell per rank number.
//
// Two candidates can share a name and a rank. Lightning Bolt's Elemental Overload twins (45284-45293)
// share the name, the skill line and the spell class set, and even have SkillLineAbility rows; they are
// separated only by carrying ClassMask 0 where the real ladder carries the class bit. So the class bit
// wins wherever it exists, and a ClassMask-0 candidate is accepted only when no real one was found -
// which is what makes talent ranks like Holy Shield 1-3 resolvable.
func resolveLadder(db *sql.DB, byRank map[int32][]rankCandidate, mask int) (map[int32]int32, error) {
	ladder := map[int32]int32{}

	// Sorted, so that when several ranks are ambiguous the one named in the skipped list is always the
	// lowest rather than whichever the map handed over first.
	rankNums := make([]int32, 0, len(byRank))
	for rank := range byRank {
		rankNums = append(rankNums, rank)
	}
	sort.Slice(rankNums, func(i, j int) bool { return rankNums[i] < rankNums[j] })

	for _, rank := range rankNums {
		cands := byRank[rank]
		var chosen []rankCandidate
		for _, c := range cands {
			if c.ClassMask&mask != 0 {
				chosen = append(chosen, c)
			}
		}
		if len(chosen) == 0 {
			chosen = cands
		}

		ids := map[int32]bool{}
		for _, c := range chosen {
			ids[c.SpellID] = true
		}

		// Holy Shock is one name over three spells per rank: a dummy the player casts, plus a damage
		// and a heal spell the client never exposes. Only the castable one carries a SpellPower row,
		// and it is the one the sim registers, so that is the tie-break.
		if len(ids) > 1 {
			castable, err := castableOf(db, ids)
			if err != nil {
				return nil, err
			}
			if castable != 0 {
				ladder[rank] = castable
				continue
			}
		}

		if len(ids) > 1 {
			var list []string
			for id := range ids {
				list = append(list, strconv.Itoa(int(id)))
			}
			sort.Strings(list)
			return nil, fmt.Errorf("rank %d is ambiguous between spells %s", rank, strings.Join(list, ", "))
		}
		ladder[rank] = chosen[0].SpellID
	}

	maxRank := int32(0)
	for rank := range ladder {
		if rank > maxRank {
			maxRank = rank
		}
	}
	for rank := int32(1); rank <= maxRank; rank++ {
		if _, ok := ladder[rank]; !ok {
			return nil, fmt.Errorf("missing rank %d of %d", rank, maxRank)
		}
	}
	return ladder, nil
}

// The one spell among these that has a mana cost, or 0 when that does not single one out.
func castableOf(db *sql.DB, ids map[int32]bool) (int32, error) {
	var found int32
	for id := range ids {
		var n int
		if err := db.QueryRow(`SELECT count(*) FROM SpellPower WHERE SpellID = ?`, id).Scan(&n); err != nil {
			return 0, err
		}
		if n == 0 {
			continue
		}
		if found != 0 {
			return 0, nil
		}
		found = id
	}
	return found, nil
}

// Which effect supplies which field. Derived from the effect types rather than declared per family,
// because the client data already says it: a SCHOOL_DAMAGE effect is direct damage, a HEAL effect is a
// heal, an ENERGIZE effect is Lay on Hands' mana restore, a periodic aura is a tick.
func buildRow(db *sql.DB, rank int32, spellID int32, mask int) (generatedRow, error) {
	spell, candidates, err := RankCandidates(db, spellID, mask)
	if err != nil {
		return generatedRow{}, err
	}

	row := generatedRow{
		Rank: rank, SpellID: spellID,
		CastTimeMs: spell.CastTimeMs, GCDMs: spell.GCDMs, CooldownMs: spell.CooldownMs,
		MaxRange: spell.MaxRange,
	}
	if spell.ManaCost.Valid {
		row.Cost = NormalizePowerCost(int32(spell.ManaCost.Int64), spell.PowerType)
	}

	amountOf := func(e RankEffect) *generatedAmount {
		min, max := DeriveRankAmount(e, spell.SpellLevel, spell.MaxLevel)
		a := &generatedAmount{Min: min, Max: max, Coef: e.Coefficient, APCoef: e.APCoef}

		// A tick count is the duration over the period, which is how every hand-written
		// NumberOfTicks in the sim was arrived at.
		if e.AuraPeriod > 0 {
			a.PeriodMs = e.AuraPeriod
			if spell.DurationMs > 0 {
				a.Ticks = spell.DurationMs / e.AuraPeriod
			}
		}
		return a
	}

	for _, e := range candidates {
		switch {
		case (e.Effect == effSchoolDamage || IsWeaponDamageEffect(e.Effect)) && row.Direct == nil:
			row.Direct = amountOf(e)
		case e.Effect == effHeal && row.Heal == nil:
			row.Heal = amountOf(e)
		case e.Effect == effEnergize && row.Energize == nil:
			row.Energize = amountOf(e)
		case IsPeriodicAura(e.Aura) && row.Periodic == nil:
			row.Periodic = amountOf(e)
		}
	}

	// An aura effect reached by the fallbacks below still lands in the role its shape says it has:
	// Frenzied Regeneration's aura ticks every second, and filing that under Direct would hand the
	// call site a periodic value through a field that promises a direct one.
	fallback := func(e RankEffect) {
		if e.AuraPeriod > 0 {
			row.Periodic = amountOf(e)
		} else {
			row.Direct = amountOf(e)
		}
	}

	// Holy Shield keeps its per-block damage on an aura effect that is none of the roles above, and it
	// is not the only aura effect on the spell: one index holds the block value and another the damage.
	// The damage is the one that scales with spell power, so a nonzero coefficient is what picks it.
	if !row.hasValue() {
		for _, e := range candidates {
			if e.Aura != 0 && e.BasePoints > 0 && e.Coefficient > 0 {
				fallback(e)
				break
			}
		}
	}
	if !row.hasValue() {
		for _, e := range candidates {
			if e.Aura != 0 && e.BasePoints > 0 {
				fallback(e)
				break
			}
		}
	}

	return row, nil
}

// A low rank can legitimately carry no numbers at all - Lay on Hands rank 1 heals a share of max health
// and restores no mana, so it has no ENERGIZE effect where ranks 2-4 do.
func (row generatedRow) hasValue() bool {
	return row.Direct != nil || row.Heal != nil || row.Periodic != nil || row.Energize != nil
}

func GenerateSpellRankFiles(helper *DBHelper) error {
	// Rendered in full before anything is written, so a class that fails validation cannot leave half
	// the packages regenerated and half stale.
	rendered := map[string][]byte{}
	for _, class := range dbc.Classes {
		pkg := strings.ToLower(dbc.ClassNameFromDBC(class))
		out, err := renderClassFile(helper.db, pkg, class)
		if err != nil {
			return fmt.Errorf("%s: %w", pkg, err)
		}
		rendered[pkg] = out
	}

	for pkg, out := range rendered {
		if err := os.WriteFile(fmt.Sprintf("sim/%s/spell_ranks_auto_gen.go", pkg), out, 0644); err != nil {
			return err
		}
	}
	return nil
}

func renderClassFile(db *sql.DB, pkg string, class dbc.DbcClass) ([]byte, error) {
	ladders, skipped, err := discoverLadders(db, class)
	if err != nil {
		return nil, err
	}

	var b strings.Builder

	// Named rather than dropped silently, so a family the resolver could not make sense of is visible
	// here instead of merely absent.
	if len(skipped) > 0 {
		b.WriteString("// Not generated:\n")
		for _, s := range skipped {
			fmt.Fprintf(&b, "//   %s\n", s)
		}
		b.WriteString("\n")
	}

	b.WriteString("type generatedRanks struct {\n")
	for _, l := range ladders {
		fmt.Fprintf(&b, "\t%s shared.SpellRankTable\n", l.Field)
	}
	b.WriteString("}\n\nvar genRanks = generatedRanks{\n")

	mask := classMaskOf(class)
	for _, l := range ladders {
		ranks := make([]int32, 0, len(l.Ranks))
		for rank := range l.Ranks {
			ranks = append(ranks, rank)
		}
		sort.Slice(ranks, func(i, j int) bool { return ranks[i] < ranks[j] })

		fmt.Fprintf(&b, "\t%s: shared.SpellRankTable{\n", l.Field)
		for _, rank := range ranks {
			row, err := buildRow(db, rank, l.Ranks[rank], mask)
			if err != nil {
				return nil, fmt.Errorf("%s rank %d: %w", l.Name, rank, err)
			}
			fmt.Fprintf(&b, "\t\t%s\n", formatRow(row))
		}
		b.WriteString("\t},\n")
	}
	b.WriteString("}\n")

	body := b.String()

	// Imports are gated on the body actually using them: an unused import does not compile, and this
	// file is one the generator itself needs in order to run again.
	var head strings.Builder
	fmt.Fprintf(&head, "// Code generated by tools/database/gen_db -gen=spellranks. DO NOT EDIT.\n\n")
	fmt.Fprintf(&head, "package %s\n\n", pkg)
	if strings.Contains(body, "time.") {
		head.WriteString("import (\n\t\"time\"\n\n\t\"github.com/wowsims/tbc/sim/common/shared\"\n)\n\n")
	} else {
		head.WriteString("import \"github.com/wowsims/tbc/sim/common/shared\"\n\n")
	}

	out, err := format.Source([]byte(head.String() + body))
	if err != nil {
		return nil, fmt.Errorf("generated %s file does not parse, refusing to write it: %w", pkg, err)
	}
	return out, nil
}

func formatRow(row generatedRow) string {
	parts := []string{fmt.Sprintf("Rank: %d", row.Rank), fmt.Sprintf("SpellID: %d", row.SpellID)}
	if row.Cost > 0 {
		parts = append(parts, fmt.Sprintf("Cost: %d", row.Cost))
	}
	if row.CastTimeMs > 0 {
		parts = append(parts, fmt.Sprintf("CastTime: %s", millis(row.CastTimeMs)))
	}
	if row.GCDMs > 0 {
		parts = append(parts, fmt.Sprintf("GCD: %s", millis(row.GCDMs)))
	}
	if row.CooldownMs > 0 {
		parts = append(parts, fmt.Sprintf("Cooldown: %s", millis(row.CooldownMs)))
	}
	if row.MaxRange > 0 {
		parts = append(parts, fmt.Sprintf("MaxRange: %s", num(row.MaxRange)))
	}
	for _, role := range []struct {
		name  string
		value *generatedAmount
	}{
		{"Direct", row.Direct},
		{"Heal", row.Heal},
		{"Periodic", row.Periodic},
		{"Energize", row.Energize},
	} {
		if role.value != nil {
			parts = append(parts, role.name+": "+formatValue(*role.value))
		}
	}
	return "{" + strings.Join(parts, ", ") + "},"
}

// Picks the variant from the shape of the data: a periodic value if it ticks, a range if the client
// rolls it, a flat number otherwise.
func formatValue(a generatedAmount) string {
	tail := fmt.Sprintf("Coef: %s", num(a.Coef))
	if a.APCoef > 0 {
		tail += fmt.Sprintf(", APCoef: %s", num(a.APCoef))
	}

	switch {
	case a.PeriodMs > 0:
		out := fmt.Sprintf("shared.SpellRankPeriodic{Tick: %s, %s, Period: %s", num(a.Min), tail, millis(a.PeriodMs))
		if a.Ticks > 0 {
			out += fmt.Sprintf(", Ticks: %d", a.Ticks)
		}
		return out + "}"
	case a.Max > a.Min:
		return fmt.Sprintf("shared.SpellRankRange{Min: %s, Max: %s, %s}", num(a.Min), num(a.Max), tail)
	default:
		return fmt.Sprintf("shared.SpellRankFlat{Value: %s, %s}", num(a.Min), tail)
	}
}

// Written as a time.Duration expression rather than a bare number, so the generated file reads the way
// a hand-written cast time or tick length does.
func millis(ms int32) string {
	return fmt.Sprintf("%d * time.Millisecond", ms)
}

func num(f float64) string {
	return strconv.FormatFloat(f, 'g', -1, 64)
}
