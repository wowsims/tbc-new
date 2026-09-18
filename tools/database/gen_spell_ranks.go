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
)

var rankSubtext = regexp.MustCompile(`^Rank (\d+)$`)

type generatedRow struct {
	Rank     int32
	SpellID  int32
	Cost     int32
	Direct   *generatedAmount
	Heal     *generatedAmount
	Periodic *generatedAmount
	Energize float64
}

type generatedAmount struct {
	Min  float64
	Max  float64
	Coef float64
}

type rankCandidate struct {
	SpellID   int32
	Rank      int32
	ClassMask int
	SkillLine int32
}

// The skill lines the anchor spell itself belongs to. A talent rank carries ClassMask 0, so the class
// bit alone cannot find it; sharing a skill line with the max rank is what identifies it as part of the
// same family rather than some other class's spell of the same name.
func anchorSkillLines(db *sql.DB, anchor int32) (map[int32]bool, error) {
	rows, err := db.Query(`SELECT DISTINCT SkillLine FROM SkillLineAbility WHERE Spell = ?`, anchor)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lines := map[int32]bool{}
	for rows.Next() {
		var line int32
		if err := rows.Scan(&line); err != nil {
			return nil, err
		}
		lines[line] = true
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(lines) == 0 {
		return nil, fmt.Errorf("anchor spell %d is in no skill line", anchor)
	}
	return lines, nil
}

// Rebuilds a family's ladder as rank number -> spell ID.
//
// Two candidates can share a name and a rank. Lightning Bolt's Elemental Overload twins (45284-45293)
// share the name, the skill line, the spell class set and even have SkillLineAbility rows; they are
// separated only by carrying ClassMask 0 where the real ladder carries the class bit. So the class bit
// wins wherever it exists, and a ClassMask-0 candidate is accepted only when no real one was found -
// which is what makes talent ranks like Holy Shield 1-3 resolvable.
func resolveLadder(db *sql.DB, fam RankFamily) (map[int32]int32, error) {
	lines, err := anchorSkillLines(db, fam.Anchor)
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(`
		SELECT sla.Spell, s.NameSubtext_lang, sla.ClassMask, sla.SkillLine
		FROM SkillLineAbility sla
		JOIN SpellName n ON n.ID = sla.Spell
		JOIN Spell s ON s.ID = sla.Spell
		WHERE n.Name_lang = ?`, fam.Name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	byRank := map[int32][]rankCandidate{}
	for rows.Next() {
		var c rankCandidate
		var subtext string
		if err := rows.Scan(&c.SpellID, &subtext, &c.ClassMask, &c.SkillLine); err != nil {
			return nil, err
		}
		m := rankSubtext.FindStringSubmatch(subtext)
		if m == nil || !lines[c.SkillLine] {
			continue
		}
		rank, _ := strconv.Atoi(m[1])
		c.Rank = int32(rank)
		byRank[c.Rank] = append(byRank[c.Rank], c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	ladder := map[int32]int32{}
	for rank, cands := range byRank {
		if pinned, ok := fam.Pin[rank]; ok {
			ladder[rank] = pinned
			continue
		}

		var chosen []rankCandidate
		for _, c := range cands {
			if c.ClassMask&fam.ClassBit != 0 {
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
			return nil, fmt.Errorf("%s rank %d is ambiguous between spells %s - pin it in the manifest",
				fam.Name, rank, strings.Join(list, ", "))
		}
		ladder[rank] = chosen[0].SpellID
	}

	return ladder, validateLadder(fam, ladder)
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

func validateLadder(fam RankFamily, ladder map[int32]int32) error {
	if len(ladder) == 0 {
		return fmt.Errorf("%s resolved no ranks", fam.Name)
	}

	maxRank := int32(0)
	for rank := range ladder {
		if rank > maxRank {
			maxRank = rank
		}
	}
	for rank := int32(1); rank <= maxRank; rank++ {
		if _, ok := ladder[rank]; !ok {
			return fmt.Errorf("%s is missing rank %d of %d", fam.Name, rank, maxRank)
		}
	}
	if ladder[maxRank] != fam.Anchor {
		return fmt.Errorf("%s: manifest anchor is %d but the highest resolved rank (%d) is spell %d",
			fam.Name, fam.Anchor, maxRank, ladder[maxRank])
	}
	return nil
}

// Which effect supplies which field. Derived from the effect types rather than declared per family,
// because the client data already says it: a SCHOOL_DAMAGE effect is direct damage, a HEAL effect is a
// heal, an ENERGIZE effect is Lay on Hands' mana restore, a periodic aura is a tick.
//
// The fallback matters for exactly one shape in the sim today: Holy Shield keeps its per-block damage
// on an aura effect (EffectAura 43) that is none of the above.
func buildRow(db *sql.DB, fam RankFamily, rank int32, spellID int32) (generatedRow, error) {
	spell, candidates, err := RankCandidates(db, spellID, fam.ClassBit)
	if err != nil {
		return generatedRow{}, err
	}

	row := generatedRow{Rank: rank, SpellID: spellID}
	if spell.ManaCost.Valid {
		row.Cost = int32(spell.ManaCost.Int64)
	}

	amountOf := func(e RankEffect) *generatedAmount {
		min, max := DeriveRankAmount(e, spell.SpellLevel, spell.MaxLevel)
		return &generatedAmount{Min: min, Max: max, Coef: e.Coefficient}
	}

	for _, e := range candidates {
		switch {
		case e.Effect == effSchoolDamage && row.Direct == nil:
			row.Direct = amountOf(e)
		case e.Effect == effHeal && row.Heal == nil:
			row.Heal = amountOf(e)
		case e.Effect == effEnergize && row.Energize == 0:
			min, _ := DeriveRankAmount(e, spell.SpellLevel, spell.MaxLevel)
			row.Energize = min
		case IsPeriodicAura(e.Aura) && row.Periodic == nil:
			row.Periodic = amountOf(e)
		}
	}

	// Holy Shield keeps its per-block damage on an aura effect that is none of the roles above, and it
	// is not the only aura effect on the spell: index 0 is the block value and index 1 is the damage.
	// The damage is the one that scales with spell power, so a nonzero coefficient is what picks it.
	if row.Direct == nil && row.Heal == nil && row.Periodic == nil && row.Energize == 0 {
		for _, e := range candidates {
			if e.Aura != 0 && e.BasePoints > 0 && e.Coefficient > 0 {
				row.Direct = amountOf(e)
				break
			}
		}
	}
	if row.Direct == nil && row.Heal == nil && row.Periodic == nil && row.Energize == 0 {
		for _, e := range candidates {
			if e.Aura != 0 && e.BasePoints > 0 {
				row.Direct = amountOf(e)
				break
			}
		}
	}

	return row, nil
}

// A low rank can legitimately carry no numbers at all - Lay on Hands rank 1 heals a share of max health
// and restores no mana, so it has no ENERGIZE effect where ranks 2-4 do. An empty MAX rank is the real
// failure, because it means the payload was never found.
func (row generatedRow) hasValue() bool {
	return row.Direct != nil || row.Heal != nil || row.Periodic != nil || row.Energize != 0
}

func GenerateSpellRankFiles(helper *DBHelper) error {
	byClass := map[string][]RankFamily{}
	for _, fam := range SpellRankManifest {
		byClass[fam.Class] = append(byClass[fam.Class], fam)
	}

	classes := make([]string, 0, len(byClass))
	for class := range byClass {
		classes = append(classes, class)
	}
	sort.Strings(classes)

	rendered := map[string][]byte{}
	for _, class := range classes {
		out, err := renderClassFile(helper.db, class, byClass[class])
		if err != nil {
			return err
		}
		rendered[class] = out
	}

	for _, class := range classes {
		if err := os.WriteFile(fmt.Sprintf("sim/%s/spell_ranks_auto_gen.go", class), rendered[class], 0644); err != nil {
			return err
		}
	}
	return nil
}

func renderClassFile(db *sql.DB, class string, families []RankFamily) ([]byte, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "// Code generated by tools/database/gen_db -gen=spellranks. DO NOT EDIT.\n\n")
	fmt.Fprintf(&b, "package %s\n\n", class)
	fmt.Fprintf(&b, "import \"github.com/wowsims/tbc/sim/common/shared\"\n\n")

	// One typed struct holding every table, rather than a package-level var per family: a spell reaches
	// its ranks through genRanks.Consecration, so a family renamed or dropped in the manifest breaks the
	// build at the use site instead of leaving an orphaned global behind.
	b.WriteString("type generatedRanks struct {\n")
	for _, fam := range families {
		fmt.Fprintf(&b, "\t%s shared.RankTable\n", fam.Field())
	}
	b.WriteString("}\n\nvar genRanks = generatedRanks{\n")

	for _, fam := range families {
		ladder, err := resolveLadder(db, fam)
		if err != nil {
			return nil, err
		}

		ranks := make([]int32, 0, len(ladder))
		for rank := range ladder {
			ranks = append(ranks, rank)
		}
		sort.Slice(ranks, func(i, j int) bool { return ranks[i] < ranks[j] })

		fmt.Fprintf(&b, "\t%s: shared.RankTable{\n", fam.Field())
		maxRankHasValue := false
		for _, rank := range ranks {
			row, err := buildRow(db, fam, rank, ladder[rank])
			if err != nil {
				return nil, err
			}
			if rank == ranks[len(ranks)-1] {
				maxRankHasValue = row.hasValue()
			}
			fmt.Fprintf(&b, "\t\t%s\n", formatRow(row))
		}
		if !maxRankHasValue {
			return nil, fmt.Errorf("%s: the max rank carries no value - the payload effect was not found", fam.Name)
		}
		b.WriteString("\t},\n")
	}
	b.WriteString("}\n")

	out, err := format.Source([]byte(b.String()))
	if err != nil {
		return nil, fmt.Errorf("generated %s file does not parse, refusing to write it: %w", class, err)
	}
	return out, nil
}

func formatRow(row generatedRow) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Rank: %d", row.Rank), fmt.Sprintf("SpellID: %d", row.SpellID))
	if row.Cost > 0 {
		parts = append(parts, fmt.Sprintf("Cost: %d", row.Cost))
	}
	if row.Direct != nil {
		parts = append(parts, "Direct: "+formatAmount(*row.Direct))
	}
	if row.Heal != nil {
		parts = append(parts, "Heal: "+formatAmount(*row.Heal))
	}
	if row.Periodic != nil {
		parts = append(parts, fmt.Sprintf("Periodic: &shared.Periodic{Tick: %s, Coef: %s}",
			num(row.Periodic.Min), num(row.Periodic.Coef)))
	}
	if row.Energize > 0 {
		parts = append(parts, fmt.Sprintf("Energize: %s", num(row.Energize)))
	}
	return "{" + strings.Join(parts, ", ") + "},"
}

func formatAmount(a generatedAmount) string {
	return fmt.Sprintf("&shared.Amount{Min: %s, Max: %s, Coef: %s}", num(a.Min), num(a.Max), num(a.Coef))
}

func num(f float64) string {
	return strconv.FormatFloat(f, 'g', -1, 64)
}
