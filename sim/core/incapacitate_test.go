package core

import (
	"testing"
	"time"

	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/simsignals"
)

func init() {
	RegisterAgentFactory(
		proto.Player_ProtectionWarrior{},
		proto.Spec_SpecProtectionWarrior,
		NewFakeFearWarrior,
		func(player *proto.Player, spec interface{}) {
			playerSpec, ok := spec.(*proto.Player_ProtectionWarrior)
			if !ok {
				panic("Invalid spec value for Protection Warrior!")
			}
			player.Spec = playerSpec
		},
	)
}

const fearTestDuration = time.Millisecond * 500
const stunTestDuration = time.Second * 2

type FakeFearWarrior struct {
	Character

	Fear      *Aura
	OtherFear *Aura
	Stun      *Aura

	ShortImmunity *Aura
	LongImmunity  *Aura
}

func (fw *FakeFearWarrior) GetCharacter() *Character { return &fw.Character }

func (fw *FakeFearWarrior) ApplyTalents()                  {}
func (fw *FakeFearWarrior) Reset(_ *Simulation)            {}
func (fw *FakeFearWarrior) OnGCDReady(_ *Simulation)       {}
func (fw *FakeFearWarrior) OnEncounterStart(_ *Simulation) {}

// Auras have to be registered before the environment is finalized.
func (fw *FakeFearWarrior) Initialize() {
	fw.Fear = fw.RegisterFearAura("Fear", ActionID{SpellID: 31970}, fearTestDuration)
	fw.OtherFear = fw.RegisterFearAura("Other Fear", ActionID{SpellID: 5782}, fearTestDuration*4)
	fw.Stun = fw.RegisterStunAura("Stun", ActionID{SpellID: 5211}, stunTestDuration)

	// Stand-ins for Berserker Rage and Death Wish, which overlap in real play.
	fw.ShortImmunity = fw.RegisterAura(Aura{
		Label:    "Short Immunity",
		ActionID: ActionID{SpellID: 18499},
		Duration: time.Second * 10,
	}).AttachFearImmunity()
	fw.LongImmunity = fw.RegisterAura(Aura{
		Label:    "Long Immunity",
		ActionID: ActionID{SpellID: 12292},
		Duration: time.Second * 30,
	}).AttachFearImmunity()
}

func NewFakeFearWarrior(char *Character, _ *proto.Player, _ *proto.Raid) Agent {
	fw := &FakeFearWarrior{
		Character: *char,
	}

	fw.EnableAutoAttacks(fw, AutoAttackOptions{
		MainHand:       Weapon{SwingSpeed: fakeMHSwingSpeed, CritMultiplier: 2},
		OffHand:        Weapon{SwingSpeed: fakeOHSwingSpeed, CritMultiplier: 2},
		AutoSwingMelee: true,
	})

	return fw
}

func setupFakeFearSim() (*Simulation, *FakeFearWarrior) {
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
							Spec:      &proto.Player_ProtectionWarrior{},
							Equipment: &proto.EquipmentSpec{},
							Rotation:  &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
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

	return sim, sim.Raid.Parties[0].Players[0].(*FakeFearWarrior)
}

// Being feared holds swings back, it does not restart the swing timer, so a
// swing that was already due lands as soon as the Fear ends and one that was
// not yet due keeps the progress it had.
func TestFearPausesSwingTimerWithoutResetting(t *testing.T) {
	t.Run("holds due swings and leaves later ones alone", func(t *testing.T) {
		sim, fw := setupFakeFearSim()
		aa := &fw.AutoAttacks

		// MH is due immediately, OH lands after the Fear ends. Both are well
		// inside a full swing, so a reset would push them out much further.
		ohSwingAt := aa.oh.swingAt
		if ohSwingAt <= fearTestDuration {
			t.Fatalf("test setup: OH swing at %v should be later than the Fear (%v)", ohSwingAt, fearTestDuration)
		}

		ApplyFear(sim, fw.Fear)

		if aa.mh.swingAt != fearTestDuration {
			t.Fatalf("MH swing: expected it held to the end of the Fear (%v), got %v", fearTestDuration, aa.mh.swingAt)
		}
		if aa.oh.swingAt != ohSwingAt {
			t.Fatalf("OH swing: expected it untouched at %v, got %v", ohSwingAt, aa.oh.swingAt)
		}
	})

	t.Run("breaking early does not restart the timer", func(t *testing.T) {
		sim, fw := setupFakeFearSim()
		aa := &fw.AutoAttacks

		ApplyFear(sim, fw.Fear)

		sim.CurrentTime = fearTestDuration / 2
		fw.Unit.BreakFear(sim)

		// Swings stay held until the Fear would have ended rather than being
		// restarted from the break, which is what a reset would have done.
		if aa.mh.swingAt != fearTestDuration {
			t.Fatalf("MH swing: expected %v after an early break, got %v", fearTestDuration, aa.mh.swingAt)
		}
	})
}

// The PseudoStats flags are shared between effects, so an effect ending must not
// clear a flag another still needs.
func TestIncapacitateFlagsSurviveOverlappingSources(t *testing.T) {
	t.Run("a fading fear leaves a longer one in charge", func(t *testing.T) {
		sim, fw := setupFakeFearSim()

		ApplyFear(sim, fw.OtherFear)
		ApplyFear(sim, fw.Fear)
		fw.Fear.Deactivate(sim)

		if !fw.PseudoStats.Incapacitated {
			t.Fatal("expected the unit to still be incapacitated by the longer Fear")
		}
	})

	t.Run("a fading stun leaves an active fear in charge", func(t *testing.T) {
		sim, fw := setupFakeFearSim()

		ApplyFear(sim, fw.Fear)
		ApplyStun(sim, fw.Stun)
		fw.Stun.Deactivate(sim)

		if !fw.PseudoStats.Incapacitated {
			t.Fatal("expected the unit to still be incapacitated by the Fear")
		}
		if fw.PseudoStats.Stunned {
			t.Fatal("expected Stunned to clear once the Stun ended")
		}
	})

	t.Run("a fading immunity leaves a longer one in charge", func(t *testing.T) {
		sim, fw := setupFakeFearSim()

		fw.LongImmunity.Activate(sim)
		fw.ShortImmunity.Activate(sim)
		fw.ShortImmunity.Deactivate(sim)

		if !fw.PseudoStats.FearImmune {
			t.Fatal("expected the longer immunity to keep the unit Fear immune")
		}
		if ApplyFear(sim, fw.Fear) {
			t.Fatal("expected the Fear to be refused while still immune")
		}
	})

	t.Run("immunity clears once every source is gone", func(t *testing.T) {
		sim, fw := setupFakeFearSim()

		fw.LongImmunity.Activate(sim)
		fw.ShortImmunity.Activate(sim)
		fw.ShortImmunity.Deactivate(sim)
		fw.LongImmunity.Deactivate(sim)

		if fw.PseudoStats.FearImmune {
			t.Fatal("expected immunity to clear once both sources ended")
		}
		if !ApplyFear(sim, fw.Fear) {
			t.Fatal("expected the Fear to land once immunity was gone")
		}
	})
}
