// Regenerates sim/<class>/spell_ranks_auto_gen.go from the client database.
//
// Deliberately its own binary rather than a mode of gen_db: gen_db imports the sim, and the sim reads
// the tables this writes, so a stale or missing generated file would stop the generator that fixes it
// from compiling. Importing only tools/database keeps regeneration possible from any state.
//
//	go run ./tools/database/gen_spellranks
package main

import (
	"flag"
	"log"

	"github.com/wowsims/tbc/tools/database"
)

var dbPath = flag.String("dbPath", "./tools/database/wowsims.db", "Location of the wowsims.db file produced by tools/db2tool")

func main() {
	flag.Parse()
	database.DatabasePath = *dbPath

	helper, err := database.NewDBHelper()
	if err != nil {
		log.Fatalf("failed to open %s: %v", *dbPath, err)
	}
	defer helper.Close()

	if err := database.GenerateSpellRankFiles(helper); err != nil {
		log.Fatalf("failed to generate spell rank tables: %v", err)
	}
}
