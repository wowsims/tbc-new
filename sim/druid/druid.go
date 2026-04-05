package druid

import (
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

var TalentTreeSizes = [3]int{21, 21, 20}

type Druid struct {
	core.Character
	SelfBuffs

	Talents *proto.DruidTalents

	StartingForm DruidForm

	Treants Treants

	BleedsActive      map[*core.Unit]int32
	CannotShredTarget bool
	RipBaseNumTicks   int32
	RipMaxNumTicks    int32

	ShredFlatBonus    float64 // Nordrassil Harness 4P: +75
	LacerateTickBonus float64 // Nordrassil Harness 4P: +15 per stack per tick

	MHAutoSpell *core.Spell

	Barkskin             *DruidSpell
	Dash                 *DruidSpell
	FaerieFire           *DruidSpell
	FaerieFireFeral      *DruidSpell
	FerociousBite        *DruidSpell
	ForceOfNature        *DruidSpell
	FrenziedRegeneration *DruidSpell
	Hurricane            *DruidSpell
	Innervate            *DruidSpell
	InsectSwarm          *DruidSpell
	Lacerate             *DruidSpell
	MangleBear           *DruidSpell
	MangleCat            *DruidSpell
	Maul                 *DruidSpell
	Moonfire             *DruidSpell
	NaturesSwiftness     *DruidSpell
	Prowl                *DruidSpell
	Rake                 *DruidSpell
	Ravage               *DruidSpell
	Rejuvenation         *DruidSpell
	Rip                  *DruidSpell
	Shred                *DruidSpell
	Starfire             *DruidSpell
	TigersFury           *DruidSpell
	Swipe                *DruidSpell
	Wrath                *DruidSpell

	CatForm     *DruidSpell
	BearForm    *DruidSpell
	MoonkinForm *DruidSpell

	BearFormAura             *core.Aura
	CatFormAura              *core.Aura
	ClearcastingAura         *core.Aura
	DashAura                 *core.Aura
	FrenziedRegenerationAura *core.Aura
	FaerieFireAuras          core.AuraArray
	MangleAuras              core.AuraArray
	MoonkinFormAura          *core.Aura
	ProwlAura                *core.Aura
	TigersFuryAura           *core.Aura

	form DruidForm

	IntensityEnrageRageBonus float64
}

const (
	DruidSpellFlagNone        int64 = 0
	DruidSpellEntanglingRoots int64 = 1 << iota
	DruidSpellFaerieFire
	DruidSpellFaerieFireFeral
	DruidSpellForceOfNature
	DruidSpellHurricane
	DruidSpellFerociousBite
	DruidSpellFrenziedRegeneration
	DruidSpellInnervate
	DruidSpellInsectSwarm
	DruidSpellLacerate
	DruidSpellMangleBear
	DruidSpellMangleCat
	DruidSpellMaul
	DruidSpellMoonfireInitial
	DruidSpellMoonfireDoT
	DruidSpellRake
	DruidSpellRavage
	DruidSpellRip
	DruidSpellShred
	DruidSpellStarfire
	DruidSpellSwipe
	DruidSpellThorns
	DruidSpellWrath
	DruidSpellTigersFury
	DruidSpellCatForm
	DruidSpellBearForm

	DruidSpellHealingTouch
	DruidSpellRegrowth
	DruidSpellLifebloom
	DruidSpellRejuvenation
	DruidSpellTranquility
	DruidSpellMarkOfTheWild
	DruidSpellSwiftmend
	DruidSpellCenarionWard

	DruidSpellLast
	DruidSpellsAll = DruidSpellLast<<1 - 1

	DruidSpellMoonfire           = DruidSpellMoonfireInitial | DruidSpellMoonfireDoT
	DruidSpellDoT                = DruidSpellMoonfireDoT | DruidSpellInsectSwarm
	DruidSpellHoT                = DruidSpellRejuvenation | DruidSpellLifebloom | DruidSpellRegrowth
	DruidSpellInstant            = DruidSpellMoonfire | DruidSpellFaerieFire
	DruidSpellMangle             = DruidSpellMangleBear | DruidSpellMangleCat
	DruidSpellBuilder            = DruidSpellMangleCat | DruidSpellShred | DruidSpellRake | DruidSpellRavage
	DruidSpellFinisher           = DruidSpellFerociousBite | DruidSpellRip
	DruidArcaneSpells            = DruidSpellMoonfire | DruidSpellMoonfireDoT | DruidSpellStarfire
	DruidNatureSpells            = DruidSpellWrath | DruidSpellHurricane | DruidSpellInsectSwarm
	DruidHealingNonInstantSpells = DruidSpellHealingTouch | DruidSpellRegrowth
	DruidHealingSpells           = DruidHealingNonInstantSpells | DruidSpellRejuvenation | DruidSpellLifebloom | DruidSpellSwiftmend
	DruidDamagingSpells          = DruidArcaneSpells | DruidNatureSpells
)

type SelfBuffs struct {
	InnervateTarget *proto.UnitReference
}

func (druid *Druid) GetCharacter() *core.Character {
	return &druid.Character
}

func (druid *Druid) AddPartyBuffs(partyBuffs *proto.PartyBuffs) {
	if druid.InForm(Cat|Bear) && druid.Talents.LeaderOfThePack {
		partyBuffs.LeaderOfThePack = core.Ternary(druid.HasItemEquipped(32387, []proto.ItemSlot{proto.ItemSlot_ItemSlotRanged}), proto.TristateEffect_TristateEffectImproved, proto.TristateEffect_TristateEffectRegular)
	} else if druid.InForm(Moonkin) && druid.Talents.MoonkinForm {
		partyBuffs.MoonkinAura = core.Ternary(druid.HasItemEquipped(32387, []proto.ItemSlot{proto.ItemSlot_ItemSlotRanged}), proto.TristateEffect_TristateEffectImproved, proto.TristateEffect_TristateEffectRegular)
	}
}

func (druid *Druid) RegisterSpell(formMask DruidForm, config core.SpellConfig) *DruidSpell {
	prev := config.ExtraCastCondition
	prevModify := config.Cast.ModifyCast

	ds := &DruidSpell{FormMask: formMask}
	config.ExtraCastCondition = func(sim *core.Simulation, target *core.Unit) bool {
		// Check if we're in allowed form to cast
		// Allow 'humanoid' auto unshift casts
		if (ds.FormMask != Any && !druid.InForm(ds.FormMask)) && !ds.FormMask.Matches(Humanoid) {
			if sim.Log != nil {
				sim.Log("Failed cast to spell %s, wrong form", ds.ActionID)
			}
			return false
		}
		return prev == nil || prev(sim, target)
	}
	config.Cast.ModifyCast = func(sim *core.Simulation, s *core.Spell, c *core.Cast) {
		if !druid.InForm(ds.FormMask) && ds.FormMask.Matches(Humanoid) {
			druid.ClearForm(sim)
		}
		if prevModify != nil {
			prevModify(sim, s, c)
		}
	}

	ds.Spell = druid.Unit.RegisterSpell(config)

	return ds
}

func (druid *Druid) Initialize() {
	druid.form = druid.StartingForm

	druid.Env.RegisterPostFinalizeEffect(func() {
		druid.MHAutoSpell = druid.AutoAttacks.MHAuto()
	})

	druid.RegisterItemSwapCallback([]proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand}, func(sim *core.Simulation, slot proto.ItemSlot) {
		switch {
		case druid.InForm(Cat):
			druid.AutoAttacks.SetMH(druid.GetCatWeapon())
		case druid.InForm(Bear):
			druid.AutoAttacks.SetMH(druid.GetBearWeapon())
		}
	})

	druid.RegisterBaselineSpells()
}

func (druid *Druid) RegisterBaselineSpells() {
	// Balance

	druid.registerStarfireSpell()
	druid.registerMoonfireSpell()
	druid.registerWrathSpell()

	druid.registerHurricaneSpell()
	druid.registerFaerieFireSpell()
	druid.registerInnervateCD()
}

func (druid *Druid) RegisterFeralCatSpells() {
	druid.registerCatFormSpell()

	druid.registerMangleCatSpell()
	druid.registerRakeSpell()
	druid.registerRipSpell()
	druid.registerFerociousBiteSpell()
	druid.registerFaerieFireFeralSpell()
	druid.registerShredSpell()
	druid.registerTigersFurySpell()
	druid.applyOmenOfClarity()
}

func (druid *Druid) RegisterFeralTankSpells() {
	druid.registerBearFormSpell()
	druid.registerBarkskin()
	// druid.registerBerserkCD()
	// druid.registerCatFormSpell()
	// druid.registerFrenziedRegenerationSpell()
	// druid.registerMangleBearSpell()
	// druid.registerMangleCatSpell()
	// druid.registerMaulSpell()
	// druid.registerMightOfUrsocCD()
	//druid.registerLacerateSpell()
	// druid.registerRakeSpell()
	// druid.registerRipSpell()
	// druid.registerSurvivalInstinctsCD()
	// druid.registerSwipeBearSpell()
	// druid.registerThrashBearSpell()
}

func (druid *Druid) Reset(_ *core.Simulation) {
	druid.form = druid.StartingForm

	for target := range druid.BleedsActive {
		druid.BleedsActive[target] = 0
	}
}

func (druid *Druid) OnEncounterStart(sim *core.Simulation) {
}

func New(char *core.Character, form DruidForm, selfBuffs SelfBuffs, talents string) *Druid {
	druid := &Druid{
		Character:       *char,
		SelfBuffs:       selfBuffs,
		Talents:         &proto.DruidTalents{},
		StartingForm:    form,
		form:            form,
		BleedsActive:    make(map[*core.Unit]int32),
		RipBaseNumTicks: 8,
	}

	druid.RipMaxNumTicks = druid.RipBaseNumTicks + 3

	core.FillTalentsProto(druid.Talents.ProtoReflect(), talents, TalentTreeSizes)
	druid.EnableManaBar()

	druid.AddStatDependency(stats.Strength, stats.AttackPower, 1)
	druid.AddStatDependency(stats.BonusArmor, stats.Armor, 1)
	druid.AddStatDependency(stats.Agility, stats.PhysicalCritPercent, core.CritPerAgiMaxLevel[char.Class])
	druid.AddStatDependency(stats.Agility, stats.DodgeRating, 1.0/14.7059*core.DodgeRatingPerDodgePercent)

	if druid.Talents.ForceOfNature {
		druid.registerTreants()
	}

	return druid
}

type DruidSpell struct {
	*core.Spell
	FormMask DruidForm

	// Optional fields used in snapshotting calculations
	CurrentSnapshotPower float64
	NewSnapshotPower     float64
	ShortName            string
}

func (ds *DruidSpell) IsReady(sim *core.Simulation) bool {
	if ds == nil {
		return false
	}
	return ds.Spell.IsReady(sim)
}

func (ds *DruidSpell) CanCast(sim *core.Simulation, target *core.Unit) bool {
	if ds == nil {
		return false
	}
	return ds.Spell.CanCast(sim, target)
}

func (ds *DruidSpell) IsEqual(s *core.Spell) bool {
	if ds == nil || s == nil {
		return false
	}
	return ds.Spell == s
}

func (druid *Druid) UpdateBleedPower(bleedSpell *DruidSpell, sim *core.Simulation, target *core.Unit, updateCurrent bool, updateNew bool) {
	snapshotPower := bleedSpell.ExpectedTickDamage(sim, target)

	if updateCurrent {
		bleedSpell.CurrentSnapshotPower = snapshotPower

		if sim.Log != nil {
			druid.Log(sim, "%s Snapshot Power: %.1f", bleedSpell.ShortName, snapshotPower)
		}
	}

	if updateNew {
		bleedSpell.NewSnapshotPower = snapshotPower

		if (sim.Log != nil) && !updateCurrent {
			druid.Log(sim, "%s Projected Power: %.1f", bleedSpell.ShortName, snapshotPower)
		}
	}
}

// Agent is a generic way to access underlying druid on any of the agents (for example balance druid.)
type DruidAgent interface {
	GetDruid() *Druid
}
