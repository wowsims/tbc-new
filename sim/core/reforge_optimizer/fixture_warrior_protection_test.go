//go:build with_db

package reforgeoptimizer

import (
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

// Sources (see the canonical template in fixture_rogue_dps_test.go):
//   - EP weights / stat caps / soft caps: the spec's FRONTEND config.
//     EP vector = ui/warrior/protection/presets.ts (P1_EP_PRESET, used by defaults.epWeights)
//     statCaps  = ui/warrior/protection/sim.ts     (defaults.statCaps)
//     softCaps  = ui/warrior/protection/sim.ts     (none)
//   - race / professions / talents / consumables / specOptions: the spec's Go test-suite
//     config, copied verbatim from sim/warrior/protection/protection_test.go.
//   - gearDir/gearName: the FE default gear preset (defaults.gear = P2_PRESET.gear ->
//     ui/warrior/protection/gear_sets/p2_bis.gear.json); loaded from origin/master.
func init() {
	// Frontend mechanics constant (ui/core/constants/mechanics.ts), used to build the
	// stat cap with the same arithmetic the frontend uses.
	const (
		expertisePerQuarterPercentReduction = 3.942308 // EXPERTISE_PER_QUARTER_PERCENT_REDUCTION
	)

	registerSpecFixtures(specFixture{
		fileName: "warrior-protection.test.json",
		class:    proto.Class_ClassWarrior,
		race:     proto.Race_RaceOrc,
		prof1:    proto.Profession_Engineering,
		prof2:    proto.Profession_Blacksmithing,
		gearDir:  "warrior/protection/gear_sets",
		gearName: "p2_bis",
		aplDir:   "warrior/protection/apls",
		aplName:  "default",

		// Verbatim from sim/warrior/protection/protection_test.go.
		talents: "35000301302-03-0055511033001101501351",
		consumables: &proto.ConsumesSpec{
			PotId:            22849,
			FoodId:           27667,
			ConjuredId:       22105,
			ExplosiveId:      30217,
			SuperSapper:      true,
			OhImbueId:        29453,
			ScrollAgi:        true,
			ScrollStr:        true,
			ScrollArm:        true,
			BattleElixirId:   22831,
			GuardianElixirId: 9088,
			NightmareSeed:    true,
		},
		specOptions: &proto.Player_ProtectionWarrior{
			ProtectionWarrior: &proto.ProtectionWarrior{
				Options: &proto.ProtectionWarrior_Options{
					ClassOptions: &proto.WarriorOptions{
						StartingRage:  100,
						DefaultShout:  proto.WarriorShout_WarriorShoutCommanding,
						DefaultStance: proto.WarriorStance_WarriorStanceDefensive,
					},
				},
			},
		},

		isTank: true,

		// EP weights: P1_EP_PRESET ("P1 - Default") in ui/warrior/protection/presets.ts, full vector.
		epStats: map[stats.Stat]float64{
			stats.Strength:         0.61,
			stats.Agility:          0.83,
			stats.Stamina:          1.15,
			stats.AttackPower:      0.25,
			stats.MeleeHitRating:   0.35,
			stats.MeleeCritRating:  0.5,
			stats.MeleeHasteRating: 0.41,
			stats.ArmorPenetration: 0.09,
			stats.ExpertiseRating:  2.01,
			stats.DefenseRating:    0.41,
			stats.BlockRating:      0.01,
			stats.BlockValue:       0.57,
			stats.ParryRating:      0.51,
			stats.ResilienceRating: 0.02,
			stats.Armor:            0.06,
			stats.BonusArmor:       0.06,
		},
		epPseudo: map[proto.PseudoStat]float64{
			proto.PseudoStat_PseudoStatMainHandDps: 3.15,
		},

		// Hard stat cap: Expertise (defaults.statCaps in ui/warrior/protection/sim.ts).
		statCapsStats: map[stats.Stat]float64{
			stats.ExpertiseRating: 6.5 * 4 * expertisePerQuarterPercentReduction,
		},
		// Hard stat caps: melee hit (9%) and crit immunity (5.6% reduced crit taken).
		statCapsPseudo: map[proto.PseudoStat]float64{
			proto.PseudoStat_PseudoStatMeleeHitPercent:         9,
			proto.PseudoStat_PseudoStatReducedCritTakenPercent: 5.6,
		},

		// Gear set is P2; frontend default gem quality is Epic.
		maxGemPhase:   2,
		maxGemQuality: proto.ItemQuality_ItemQualityEpic,
	})
}
