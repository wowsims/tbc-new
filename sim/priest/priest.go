package priest

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

var TalentTreeSizes = [3]int{22, 21, 21}

type Priest struct {
	core.Character
	SelfBuffs
	Talents *proto.PriestTalents

	Latency float64

	ShadowfiendAura *core.Aura
	// ShadowfiendPet  *Shadowfiend

	Shadowfiend *core.Spell

	ShadowWordPain  []*core.Spell
	MindBlast       []*core.Spell
	MindFlay        []*core.Spell
	ShadowWordDeath []*core.Spell
	DevouringPlague *core.Spell
	VampiricEmbrace *core.Spell
	VampiricTouch   []*core.Spell
}

type TargetDoTInfo struct {
	Swp time.Duration
	VT  time.Duration
}

type SelfBuffs struct {
	UseShadowfiend bool
}

func (priest *Priest) GetCharacter() *core.Character {
	return &priest.Character
}

func (priest *Priest) AddPartyBuffs(_ *proto.PartyBuffs) {
}

func (priest *Priest) Initialize() {
	MindBlastRankMap.RegisterAll(priest.registerMindBlastSpell)
	MindFlayRankMap.RegisterAll(priest.registerMindFlaySpell)
	ShadowWordPainRankMap.RegisterAll(priest.registerShadowWordPainSpell)
	ShadowWordDeathRankMap.RegisterAll(priest.registerShadowWordDeathSpell)
	VampiricTouchRankMap.RegisterAll(priest.registerVampiricTouchSpell)
	// priest.registerShadowfiendSpell()
	// priest.registerVampiricTouchSpell()
	// priest.registerPowerInfusionSpell()
}

func (priest *Priest) ApplyTalents() {
}

func (priest *Priest) Reset(_ *core.Simulation) {

}

func (priest *Priest) OnEncounterStart(sim *core.Simulation) {
}

func New(char *core.Character, selfBuffs SelfBuffs, talents string) *Priest {
	priest := &Priest{
		Character: *char,
		SelfBuffs: selfBuffs,
		Talents:   &proto.PriestTalents{},
	}

	core.FillTalentsProto(priest.Talents.ProtoReflect(), talents, TalentTreeSizes)
	priest.EnableManaBar()
	// priest.ShadowfiendPet = priest.NewShadowfiend()

	return priest
}

// Agent is a generic way to access underlying priest on any of the agents.
type PriestAgent interface {
	GetPriest() *Priest
}

func NewPriest(character *core.Character, options *proto.Player) *Priest {
	selfBuffs := SelfBuffs{
		UseShadowfiend: true,
	}

	basePriest := New(character, selfBuffs, options.TalentsString)
	basePriest.Latency = float64(basePriest.ChannelClipDelay.Milliseconds())
	/*	priest := &Priest{
			Priest:  basePriest,
			options: priestOptions.Options,
		}

		return priest*/
	return basePriest
}

func RegisterPriest() {
	core.RegisterAgentFactory(
		proto.Player_Priest{},
		proto.Spec_SpecPriest,
		func(character *core.Character, options *proto.Player) core.Agent {
			return NewPriest(character, options)
		},
		func(player *proto.Player, spec interface{}) {
			playerSpec, ok := spec.(*proto.Player_Priest)
			if !ok {
				panic("Invalid spec value for Priest!")
			}
			player.Spec = playerSpec
		},
	)
}

const (
	PriestSpellFlagNone        int64 = 0
	PriestSpellDevouringPlague int64 = 1 << iota
	PriestSpellDevouringPlagueDoT
	PriestSpellDevouringPlagueHeal
	PriestSpellHolyNova
	PriestSpellHolyFire
	PriestSpellMindBlast
	PriestSpellMindFlay
	PriestSpellPowerInfusion
	PriestSpellShadowform
	PriestSpellShadowWordDeath
	PriestSpellShadowWordPain
	PriestSpellShadowFiend
	PriestSpellVampiricEmbrace
	PriestSpellVampiricTouch
	PriestSpellFade

	PriestSpellLast
	PriestSpellsAll    = PriestSpellLast<<1 - 1
	PriestSpellDoT     = PriestSpellDevouringPlague | PriestSpellHolyFire | PriestSpellMindFlay | PriestSpellShadowWordPain | PriestSpellVampiricTouch
	PriestSpellInstant = PriestSpellDevouringPlague |
		PriestSpellFade |
		PriestSpellHolyNova |
		PriestSpellPowerInfusion |
		PriestSpellShadowWordDeath |
		PriestSpellShadowWordPain |
		PriestSpellVampiricEmbrace
	PriestShadowSpells = PriestSpellDevouringPlague |
		PriestSpellShadowWordDeath |
		PriestSpellShadowform |
		PriestSpellShadowWordPain |
		PriestSpellMindFlay |
		PriestSpellMindBlast |
		PriestSpellVampiricTouch |
		PriestSpellShadowFiend |
		PriestSpellVampiricEmbrace
)
