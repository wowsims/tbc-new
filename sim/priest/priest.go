package priest

import (
	"github.com/wowsims/tbc/sim/common/shared"
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
	ShadowfiendPet  *Shadowfiend

	Shadowfiend    *core.Spell
	InnerFocusAura *core.Aura

	ShadowWordPain  []*core.Spell
	MindBlast       []*core.Spell
	MindFlay        []*core.Spell
	ShadowWordDeath []*core.Spell
	DevouringPlague []*core.Spell
	VampiricEmbrace *core.Spell
	VampiricTouch   []*core.Spell
	Smite           []*core.Spell
	Starshards      []*core.Spell
	HolyNova        []*core.Spell
}

type SelfBuffs struct {
	UseShadowfiend bool
	PreShadowform  bool
}

func (priest *Priest) GetCharacter() *core.Character {
	return &priest.Character
}
func (priest *Priest) GetPriest() *Priest {
	return priest
}

func (priest *Priest) AddPartyBuffs(_ *proto.PartyBuffs) {
}

func (priest *Priest) Initialize() {
	mindblastCDTimer := priest.NewTimer()
	shadowWordDeathCDTimer := priest.NewTimer()

	MindBlastRankMap.RegisterAll(func(rankConfig shared.SpellRankConfig) {
		priest.registerMindBlastSpell(rankConfig, mindblastCDTimer)
	})
	ShadowWordPainRankMap.RegisterAll(priest.registerShadowWordPainSpell)
	ShadowWordDeathRankMap.RegisterAll(func(rankConfig shared.SpellRankConfig) {
		priest.registerShadowWordDeathSpell(rankConfig, shadowWordDeathCDTimer)
	})
	SmiteRankMap.RegisterAll(priest.registerSmiteSpell)
	priest.registerShadowfiendSpell()

	if priest.Race == proto.Race_RaceNightElf {
		starshardsCDTimer := priest.NewTimer()
		StarshardsRankMap.RegisterAll(func(rankConfig shared.SpellRankConfig) {
			priest.registerStarshardsSpell(rankConfig, starshardsCDTimer)
		})
	}
	if priest.Race == proto.Race_RaceUndead {
		devouringPlagueCDTimer := priest.NewTimer()
		DevouringPlagueRankMap.RegisterAll(func(rankConfig shared.SpellRankConfig) {
			priest.registerDevouringPlagueSpell(rankConfig, devouringPlagueCDTimer)
		})
	}
}

func (priest *Priest) ApplyTalents() {
	// Discipline
	priest.applyInnerFocus()
	priest.applyMeditation()
	priest.applyMentalAgility()
	priest.applyMentalStrength()
	priest.applyEnlightenment()
	priest.applyFocusedPower()
	priest.applyForceOfWill()
	priest.applySilentResolve()
	priest.applyPowerInfusion()

	// Holy
	priest.applyHolyNova()
	priest.applyDivineFury()
	priest.applySearingLight()
	priest.applySurgeOfLight()
	priest.applySpiritualGuidance()

	// Shadow
	priest.applyMindFlay()
	priest.applyDarkness()
	priest.applyShadowFocus()
	priest.applyImprovedShadowWordPain()
	priest.applyFocusedMind()
	priest.applyShadowAffinity()
	priest.applyShadowPower()
	priest.applyShadowWeaving()
	priest.applyMisery()
	priest.applyShadowform()
	priest.applyVampiricEmbrace()
	priest.applyVampiricTouch()
	priest.applyImprovedMindBlast()
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
	priest.ShadowfiendPet = priest.NewShadowfiend()

	return priest
}

// Agent is a generic way to access underlying priest on any of the agents.
type PriestAgent interface {
	GetPriest() *Priest
}

func NewPriest(character *core.Character, options *proto.Player) *Priest {
	classOptions := options.GetPriest().GetOptions().GetClassOptions()
	selfBuffs := SelfBuffs{
		UseShadowfiend: true,
		PreShadowform:  classOptions.GetPreShadowform(),
	}

	basePriest := New(character, selfBuffs, options.TalentsString)
	basePriest.Latency = float64(basePriest.ChannelClipDelay.Milliseconds())

	return basePriest
}

func RegisterPriest() {
	core.RegisterAgentFactory(
		proto.Player_Priest{},
		proto.Spec_SpecPriest,
		func(character *core.Character, options *proto.Player, _ *proto.Raid) core.Agent {
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
	PriestSpellStarshards
	PriestSpellShadowform
	PriestSpellShadowWordDeath
	PriestSpellShadowWordPain
	PriestSpellShadowFiend
	PriestSpellVampiricEmbrace
	PriestSpellVampiricTouch
	PriestSpellFade
	PriestSpellSmite

	PriestSpellLast
	PriestSpellsAll    = PriestSpellLast<<1 - 1
	PriestSpellDoT     = PriestSpellDevouringPlague | PriestSpellHolyFire | PriestSpellMindFlay | PriestSpellShadowWordPain | PriestSpellVampiricTouch | PriestSpellStarshards
	PriestSpellInstant = PriestSpellDevouringPlague |
		PriestSpellFade |
		PriestSpellHolyNova |
		PriestSpellPowerInfusion |
		PriestSpellShadowWordDeath |
		PriestSpellShadowWordPain |
		PriestSpellVampiricEmbrace |
		PriestSpellShadowFiend |
		PriestSpellStarshards |
		PriestSpellShadowform |
		PriestSpellPowerInfusion
	PriestShadowSpells = PriestSpellDevouringPlague |
		PriestSpellShadowWordDeath |
		PriestSpellShadowform |
		PriestSpellShadowWordPain |
		PriestSpellMindFlay |
		PriestSpellMindBlast |
		PriestSpellVampiricTouch |
		PriestSpellShadowFiend |
		PriestSpellVampiricEmbrace
	PriestHolySpells = PriestSpellSmite | PriestSpellHolyFire | PriestSpellHolyNova
)
