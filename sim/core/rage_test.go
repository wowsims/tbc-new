package core

import (
	"testing"

	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/simsignals"
)

func init() {
	RegisterAgentFactory(
		proto.Player_DpsWarrior{},
		proto.Spec_SpecDpsWarrior,
		NewFakeRageWarrior,
		func(player *proto.Player, spec interface{}) {
			playerSpec, ok := spec.(*proto.Player_DpsWarrior)
			if !ok {
				panic("Invalid spec value for Dps Warrior!")
			}
			player.Spec = playerSpec
		},
	)
}

const (
	fakeMHSwingSpeed = 2.6
	fakeOHSwingSpeed = 1.8
)

type FakeRageWarrior struct {
	Character
}

func (fw *FakeRageWarrior) GetCharacter() *Character { return &fw.Character }

func (fw *FakeRageWarrior) Initialize()                    {}
func (fw *FakeRageWarrior) ApplyTalents()                  {}
func (fw *FakeRageWarrior) Reset(_ *Simulation)            {}
func (fw *FakeRageWarrior) OnGCDReady(_ *Simulation)       {}
func (fw *FakeRageWarrior) OnEncounterStart(_ *Simulation) {}

func NewFakeRageWarrior(char *Character, _ *proto.Player, _ *proto.Raid) Agent {
	fw := &FakeRageWarrior{
		Character: *char,
	}

	fw.EnableRageBar(RageBarOptions{
		MaxRage:            100,
		BaseRageMultiplier: 1,
		StartingRage:       0,
	})

	fw.EnableAutoAttacks(fw, AutoAttackOptions{
		MainHand:       Weapon{SwingSpeed: fakeMHSwingSpeed, CritMultiplier: 2},
		OffHand:        Weapon{SwingSpeed: fakeOHSwingSpeed, CritMultiplier: 2},
		AutoSwingMelee: true,
	})

	return fw
}

func SetupFakeRageSim() *Simulation {
	sim := NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{
			RandomSeed: 100,
		},
		Raid: &proto.Raid{
			Parties: []*proto.Party{
				{
					Players: []*proto.Player{
						{
							Name:      "Warrior",
							Class:     proto.Class_ClassWarrior,
							Buffs:     &proto.IndividualBuffs{},
							Spec:      &proto.Player_DpsWarrior{},
							Equipment: &proto.EquipmentSpec{},
							// AddRage() pokes the rotation, so it needs to be non-nil.
							Rotation: &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
						},
					},
					Buffs: &proto.PartyBuffs{},
				},
			},
		},
		Encounter: &proto.Encounter{
			Targets: []*proto.Target{
				{Name: "target", Level: 73, MobType: proto.MobType_MobTypeHumanoid},
			},
			Duration: 180,
		},
	}, simsignals.CreateSignals())
	sim.Reset()

	return sim
}

// Feeds a hand-crafted auto attack result to the Rage bar and returns the Rage gained from it.
// PostArmorAndResistanceMultiplier is always populated, because the sim fills it in before the
// outcome is applied, even for swings which end up doing no damage.
func rageFromAutoAttack(sim *Simulation, fw *FakeRageWarrior, spell *Spell, outcome HitOutcome, preOutcomeDamage float64) float64 {
	result := &SpellResult{
		Target:                           sim.Encounter.ActiveTargetUnits[0],
		Outcome:                          outcome,
		PostArmorAndResistanceMultiplier: preOutcomeDamage,
	}
	if outcome.Matches(OutcomeLanded) {
		result.Damage = preOutcomeDamage
	}

	rageBefore := fw.CurrentRage()
	fw.Unit.OnSpellHitDealt(sim, spell, result)

	return fw.CurrentRage() - rageBefore
}

func TestAutoAttackRageGeneration(t *testing.T) {
	// Rage for a 500 damage MH swing:
	//   min((500*7.5/274.7 + 3.5*2.6)/2, 500*15/274.7) = min(11.376, 27.303) = 11.376
	// A crit doubles the hit factor:
	//   min((500*7.5/274.7 + 7.0*2.6)/2, 500*15/274.7) = min(15.926, 27.303) = 15.926
	const swingDamage = 500.0

	tests := []struct {
		name     string
		outcome  HitOutcome
		wantRage float64
	}{
		{
			name:     "hit",
			outcome:  OutcomeHit,
			wantRage: 11.376,
		},
		{
			name:     "crit",
			outcome:  OutcomeCrit,
			wantRage: 15.926,
		},
		{
			// Dodges and parries are not "landed" outcomes, but they still generate
			// Rage based on the damage the swing would have done.
			name:     "dodge",
			outcome:  OutcomeDodge,
			wantRage: 11.376,
		},
		{
			name:     "parry",
			outcome:  OutcomeParry,
			wantRage: 11.376,
		},
		{
			name:     "miss",
			outcome:  OutcomeMiss,
			wantRage: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sim := SetupFakeRageSim()
			fw := sim.Raid.Parties[0].Players[0].(*FakeRageWarrior)

			gained := rageFromAutoAttack(sim, fw, fw.AutoAttacks.MHAuto(), test.outcome, swingDamage)

			if !WithinToleranceFloat64(test.wantRage, gained, 0.01) {
				t.Fatalf("Incorrect Rage generated on %s: Expected: %0.3f, Actual: %0.3f", test.name, test.wantRage, gained)
			}
		})
	}
}

func TestOffHandAutoAttackRageGeneration(t *testing.T) {
	// OH swings use half the hit factor:
	//   min((500*7.5/274.7 + 1.75*1.8)/2, 500*15/274.7) = min(8.401, 27.303) = 8.401
	const swingDamage = 500.0

	sim := SetupFakeRageSim()
	fw := sim.Raid.Parties[0].Players[0].(*FakeRageWarrior)
	ohAuto := fw.AutoAttacks.OHAuto()

	hitRage := rageFromAutoAttack(sim, fw, ohAuto, OutcomeHit, swingDamage)
	if !WithinToleranceFloat64(8.401, hitRage, 0.01) {
		t.Fatalf("Incorrect Rage generated on OH hit: Expected: %0.3f, Actual: %0.3f", 8.401, hitRage)
	}

	dodgeRage := rageFromAutoAttack(sim, fw, ohAuto, OutcomeDodge, swingDamage)
	if !WithinToleranceFloat64(hitRage, dodgeRage, 0.01) {
		t.Fatalf("Dodged OH swing generated %0.3f Rage, expected the same as a hit (%0.3f)", dodgeRage, hitRage)
	}
}
