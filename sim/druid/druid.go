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

	IdolFerociousBiteBonus float64 // Idol of the Beast (25667): +14 per combo point
	IdolMangleCatBonus     float64 // Idol of the Wild (28064): +24 flat to Mangle Cat
	IdolMangleBearBonus    float64 // Idol of the Wild (28064): +52 flat to Mangle Bear
	IdolShredBonus         float64 // Everbloom Idol (29390): +88 flat to Shred
	IdolRipBonus           float64 // Idol of Feral Shadows (28372): +7 per combo point per tick
	IdolLacerateBonus      float64 // Idol of Ursoc (27744): +8 per tick per stack
	IdolMaulBonus          float64 // Idol of Brutality (23198): +50 flat to Maul
	IdolSwipeBonus         float64 // Idol of Brutality (23198): +10 flat to Swipe

	MHAutoSpell *core.Spell

	Barkskin             *DruidSpell
	Dash                 *DruidSpell
	DemoralizingRoar     *DruidSpell
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
	DemoralizingRoarAuras    core.AuraArray
	FaerieFireAuras          core.AuraArray
	MangleAuras              core.AuraArray
	MoonkinFormAura          *core.Aura
	ProwlAura                *core.Aura
	TigersFuryAura           *core.Aura

	form DruidForm

	IntensityEnrageRageBonus float64

	// Maul queue (fires on next auto-attack swing, like warrior Heroic Strike)
	maulQueueAura  *core.Aura
	maulQueueSpell *core.Spell
	maulRealismICD *core.Cooldown
}

const (
	DruidSpellFlagNone        int64 = 0
	DruidSpellEntanglingRoots int64 = 1 << iota
	DruidSpellDemoralizingRoar
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
	druid.registerInnervateCD()
}

func (druid *Druid) RegisterBalanceSpells() {
	druid.registerStarfireSpell()
	druid.registerMoonfireSpell()
	druid.registerWrathSpell()
	druid.registerHurricaneSpell()
	druid.registerFaerieFireSpell()
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
}

func (druid *Druid) RegisterFeralTankSpells() {
	druid.registerBearFormSpell()
	druid.registerBarkskin()
	druid.registerDemoralizingRoarSpell()
	druid.registerFaerieFireFeralSpell()
	druid.registerFrenziedRegenerationSpell()
	druid.registerLacerateSpell()
	druid.registerMangleBearSpell()
	druid.registerMaulSpell()
	druid.registerSwipeBearSpell()
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

	// TBC: Druids have a -1.87% base dodge correction to match in-game values.
	druid.PseudoStats.BaseDodgeChance -= 0.0187

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
