package paladin

import (
	"fmt"
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

type proc struct {
	spellID int32
	value   float64
	coeff   float64
}

type judge struct {
	spellID   int32
	minDamage float64
	maxDamage float64
	coeff     float64
}

type seal struct {
	rank     int32
	level    int32
	spellID  int32
	manaCost float64
	proc     proc
	judge    judge
}

func (seal seal) GetRankLabel() string {
	return fmt.Sprintf("Rank %d", seal.rank)
}

var SealOfRighteousnessRanks = sealRankMap{
	{},
	{rank: 1, spellID: 21084, manaCost: 20, proc: proc{spellID: 25742, value: 216, coeff: 0.029}, judge: judge{spellID: 20187, minDamage: 26, maxDamage: 26, coeff: 0.209}},
	{rank: 2, spellID: 20287, manaCost: 40, proc: proc{spellID: 25740, value: 318, coeff: 0.063}, judge: judge{spellID: 20280, minDamage: 36, maxDamage: 39, coeff: 0.455}},
	{rank: 3, spellID: 20288, manaCost: 60, proc: proc{spellID: 25739, value: 490, coeff: 0.093}, judge: judge{spellID: 20281, minDamage: 53, maxDamage: 58, coeff: 0.674}},
	{rank: 4, spellID: 20289, manaCost: 90, proc: proc{spellID: 25738, value: 727, coeff: 0.1}, judge: judge{spellID: 20282, minDamage: 73, maxDamage: 80, coeff: 0.728}},
	{rank: 5, spellID: 20290, manaCost: 120, proc: proc{spellID: 25737, value: 1007, coeff: 0.1}, judge: judge{spellID: 20283, minDamage: 96, maxDamage: 105, coeff: 0.728}},
	{rank: 6, spellID: 20291, manaCost: 140, proc: proc{spellID: 25736, value: 1328, coeff: 0.1}, judge: judge{spellID: 20284, minDamage: 124, maxDamage: 135, coeff: 0.728}},
	{rank: 7, spellID: 20292, manaCost: 170, proc: proc{spellID: 25735, value: 1689, coeff: 0.1}, judge: judge{spellID: 20285, minDamage: 155, maxDamage: 168, coeff: 0.728}},
	{rank: 8, spellID: 20293, manaCost: 200, proc: proc{spellID: 25713, value: 2068, coeff: 0.1}, judge: judge{spellID: 20286, minDamage: 186, maxDamage: 203, coeff: 0.728}},
	{rank: 9, spellID: 27155, manaCost: 260, proc: proc{spellID: 27156, value: 2324, coeff: 0.1}, judge: judge{spellID: 27157, minDamage: 225, maxDamage: 246, coeff: 0.728}},
}

var SealOfLightRanks = sealRankMap{
	{},
	{rank: 1, spellID: 20165, manaCost: 110, proc: proc{spellID: 20167, value: 39, coeff: 0.0}, judge: judge{spellID: 20185, minDamage: 25, maxDamage: 25, coeff: 0.0}},
	{rank: 2, spellID: 20347, manaCost: 140, proc: proc{spellID: 20333, value: 53, coeff: 0.0}, judge: judge{spellID: 20344, minDamage: 34, maxDamage: 34, coeff: 0.0}},
	{rank: 3, spellID: 20348, manaCost: 180, proc: proc{spellID: 20334, value: 76, coeff: 0.0}, judge: judge{spellID: 20345, minDamage: 49, maxDamage: 49, coeff: 0.0}},
	{rank: 4, spellID: 20349, manaCost: 210, proc: proc{spellID: 20340, value: 94, coeff: 0.0}, judge: judge{spellID: 20346, minDamage: 61, maxDamage: 61, coeff: 0.0}},
	{rank: 5, spellID: 27160, manaCost: 280, proc: proc{spellID: 27161, value: 133, coeff: 0.0}, judge: judge{spellID: 27162, minDamage: 95, maxDamage: 95, coeff: 0.0}},
}

var SealOfWisdomRanks = sealRankMap{
	{},
	{rank: 1, spellID: 20166, manaCost: 135, proc: proc{spellID: 20168, value: 50, coeff: 0.0}, judge: judge{spellID: 20186, minDamage: 33, maxDamage: 33, coeff: 0.0}},
	{rank: 2, spellID: 20356, manaCost: 170, proc: proc{spellID: 20350, value: 71, coeff: 0.0}, judge: judge{spellID: 20354, minDamage: 46, maxDamage: 46, coeff: 0.0}},
	{rank: 3, spellID: 20357, manaCost: 200, proc: proc{spellID: 20351, value: 90, coeff: 0.0}, judge: judge{spellID: 20355, minDamage: 59, maxDamage: 59, coeff: 0.0}},
	{rank: 4, spellID: 27166, manaCost: 270, proc: proc{spellID: 27167, value: 121, coeff: 0.0}, judge: judge{spellID: 27164, minDamage: 74, maxDamage: 74, coeff: 0.0}},
}

var SealOfJusticeRanks = sealRankMap{
	{},
	{rank: 1, spellID: 20164, manaCost: 10, proc: proc{spellID: 20170, value: 0, coeff: 0.0}, judge: judge{spellID: 20184, minDamage: 0, maxDamage: 0, coeff: 0.0}},
	{rank: 2, spellID: 31895, manaCost: 10, proc: proc{spellID: 20170, value: 0, coeff: 0.0}, judge: judge{spellID: 31896, minDamage: 0, maxDamage: 0, coeff: 0.0}},
}

var SealOfTheCrusaderRanks = sealRankMap{
	{},
	{rank: 1, spellID: 21082, manaCost: 25, proc: proc{spellID: 21082, value: 41}, judge: judge{spellID: 21183, minDamage: 23}},
	{rank: 2, spellID: 20162, manaCost: 40, proc: proc{spellID: 21082, value: 68}, judge: judge{spellID: 20188, minDamage: 35}},
	{rank: 3, spellID: 20305, manaCost: 65, proc: proc{spellID: 21082, value: 122}, judge: judge{spellID: 20300, minDamage: 58}},
	{rank: 4, spellID: 20306, manaCost: 90, proc: proc{spellID: 21082, value: 183}, judge: judge{spellID: 20301, minDamage: 92}},
	{rank: 5, spellID: 20307, manaCost: 125, proc: proc{spellID: 21082, value: 272}, judge: judge{spellID: 20302, minDamage: 127}},
	{rank: 6, spellID: 20308, manaCost: 160, proc: proc{spellID: 21082, value: 372}, judge: judge{spellID: 20303, minDamage: 161}},
	{rank: 7, spellID: 27158, manaCost: 210, proc: proc{spellID: 21082, value: 495}, judge: judge{spellID: 27159, minDamage: 219}},
}

var SealOfCommandRanks = sealRankMap{
	{},
	{rank: 1, spellID: 20375, manaCost: 65, proc: proc{spellID: 20424, value: 0.70, coeff: 0.29}, judge: judge{spellID: 20425, minDamage: 68, maxDamage: 73, coeff: 0.429}},
	{rank: 2, spellID: 20915, manaCost: 110, proc: proc{spellID: 20424, value: 0.70, coeff: 0.29}, judge: judge{spellID: 20962, minDamage: 97, maxDamage: 105, coeff: 0.429}},
	{rank: 3, spellID: 20918, manaCost: 140, proc: proc{spellID: 20424, value: 0.70, coeff: 0.29}, judge: judge{spellID: 20961, minDamage: 124, maxDamage: 135, coeff: 0.429}},
	{rank: 4, spellID: 20919, manaCost: 180, proc: proc{spellID: 20424, value: 0.70, coeff: 0.29}, judge: judge{spellID: 20967, minDamage: 154, maxDamage: 168, coeff: 0.429}},
	{rank: 5, spellID: 20920, manaCost: 210, proc: proc{spellID: 20424, value: 0.70, coeff: 0.29}, judge: judge{spellID: 20968, minDamage: 193, maxDamage: 211, coeff: 0.429}},
	{rank: 6, spellID: 27170, manaCost: 280, proc: proc{spellID: 20424, value: 0.70, coeff: 0.29}, judge: judge{spellID: 27172, minDamage: 228, maxDamage: 252, coeff: 0.429}},
}

func (paladin *Paladin) registerSeals() {
	SealOfRighteousnessRanks.RegisterAll(paladin.registerSealOfRighteousness)
	SealOfLightRanks.RegisterAll(paladin.registerSealOfLight)
	SealOfWisdomRanks.RegisterAll(paladin.registerSealOfWisdom)
	SealOfJusticeRanks.RegisterAll(paladin.registerSealOfJustice)
	SealOfTheCrusaderRanks.RegisterAll(paladin.registerSealOfTheCrusader)
	paladin.registerSealOfBlood()
	paladin.registerSealOfVengeance()
}

// Seal Twist
const TwistTag = "Twistable"

// Command -> Blood
// Command -> Righteousness
// Command -> Wisdom
// Command -> Light
// Command -> Justice

// Blood -> X

// Righteous -> Command
// Righteous -> Blood
// Righteous -> Wisdom
// Righteous -> Light
// Righteous -> Justice

// Wisdom -> X

// Light -> X

// Justice -> X
func (paladin *Paladin) applySeal(newSeal *core.Aura, sealSpell *core.Spell, judgement *core.Spell, sim *core.Simulation) {
	if paladin.CurrentSeal != nil {
		newSealLabel := newSeal.ActionID.SpellID
		if newSealLabel == 0 {
			newSealLabel = newSeal.ActionIDForProc.SpellID
		}

		currentSealLabel := paladin.CurrentSeal.ActionID.SpellID
		if currentSealLabel == 0 {
			currentSealLabel = paladin.CurrentSeal.ActionIDForProc.SpellID
		}
		// If they are recasting the same seal, reactivate or refresh
		if newSealLabel == currentSealLabel {
			paladin.CurrentSeal.Activate(sim)
			return
		}
	}

	// Twisting only occurs when current seal is Command or Righteousness
	if paladin.CurrentSeal.IsActive() {
		if paladin.CurrentSeal.Tag == TwistTag {
			paladin.PreviousSealSpell = sealSpell
			paladin.PreviousSeal = paladin.CurrentSeal
			paladin.PreviousJudgement = paladin.CurrentJudgement
			pendingAction := core.NewDelayedAction(core.DelayedActionOptions{
				DoAt:     sim.CurrentTime + (time.Millisecond * 399),
				Priority: core.ActionPriorityLow,
				OnAction: func(sim *core.Simulation) {
					paladin.PreviousSeal.Deactivate(sim)
				},
			})
			sim.AddPendingAction(pendingAction)
		} else {
			paladin.CurrentSeal.Deactivate(sim)
		}
	}

	paladin.CurrentSealSpell = sealSpell
	paladin.CurrentSeal = newSeal
	paladin.CurrentJudgement = judgement
	paladin.CurrentSeal.Activate(sim)
}

// Seal of Righteousness
// https://www.wowhead.com/tbc/spell=21084
//
// Fills the Paladin with divine spirit for 30 sec, granting each melee attack
// additional Holy damage. Only one Seal can be active on the Paladin at any one time.
//
// Unleashing this Seal's energy will judge an enemy, instantly causing Holy damage.
func (paladin *Paladin) registerSealOfRighteousness(seal seal) {
	// ~~~~~~~~~ SEASON OF DISCOVERY DESCRIPTION, INFO SHOULD BE VERIFIED ~~~~~~~~~

	/*
	* Seal of Righteousness is a Spell/Aura that when active makes the paladin capable of procing
	* two different SpellIDs depending on a paladin's casted spell or melee swing.
	*
	* (Judgement of Righteousness):
	*   - Deals flat damage that is affected by Improved SoR talent, and
	*     has a spellpower scaling that is unaffected by that talent.
	*   - Targets magic defense and rolls to hit and crit.
	*
	* (Seal of Righteousness):
	*   - Procs from white hits.
	*   - Cannot miss or be dodged/parried/blocked if the underlying white hit lands.
	*   - Deals damage that is a function of weapon speed, and spellpower.
	*   - Calculates damage including spellpower scaling but ignoring damage multipliers,
	*      then feeds that value as base damage into the proc spell.
	 */

	judgeSpell := paladin.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: seal.judge.spellID},
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagBinary, // | core.SpellFlagSuppressWeaponProcs | core.SpellFlagSuppressEquipProcs
		ClassSpellMask: SpellMaskJudgementOfRighteousness,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		CritMultiplier:   paladin.DefaultSpellCritMultiplier(),
		BonusCoefficient: seal.judge.coeff,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			flags := spell.Flags
			baseDamage := sim.Roll(seal.judge.minDamage, seal.judge.maxDamage)
			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)

			action := core.NewDelayedAction(core.DelayedActionOptions{
				DoAt:     sim.CurrentTime + core.SpellBatchWindow,
				Priority: core.ActionPriorityLow,
				OnAction: func(sim *core.Simulation) {
					currentFlags := spell.Flags
					spell.Flags = flags
					spell.DealDamage(sim, result)
					spell.Flags = currentFlags
				},
			})

			sim.AddPendingAction(action)
		},
	})

	// Canonical Seal of Righteousness proc formula (Maintankadin / EJ, matches in-game testing):
	//   1H: damage = (0.85 * SoRcoef * Speed) - (QualityModifier * Speed * 0.03) + (0.03 * AvgWeaponDmg) + (0.092 * Speed * SP)
	//   2H: damage = (1.20 * SoRcoef * Speed) - (QualityModifier * Speed * 0.03) + (0.03 * AvgWeaponDmg) + (0.108 * Speed * SP)
	sorCoef := seal.proc.value * 1.2 * 1.03 / 100
	procSpell := paladin.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: seal.proc.spellID},
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskMeleeMHSpecial,
		Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagPassiveSpell,
		ClassSpellMask: SpellMaskSealOfRighteousness,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			mh := paladin.MainHand()

			baseCoef := 0.85
			spCoef := 0.092
			if mh.HandType == proto.HandType_HandTypeTwoHand {
				baseCoef = 1.2
				spCoef = 0.108
			}

			speed := mh.SwingSpeed
			spell.BonusCoefficient = spCoef * speed

			avgWeaponDmg := paladin.AutoAttacks.MH().AverageDamage()
			flatDamage := baseCoef*sorCoef*speed - mh.QualityModifier*speed*0.03 + 0.03*avgWeaponDmg
			result := spell.CalcDamage(sim, target, flatDamage, spell.OutcomeAlwaysHit)

			action := core.NewDelayedAction(core.DelayedActionOptions{
				DoAt:     sim.CurrentTime + core.SpellBatchWindow,
				Priority: core.ActionPriorityLow,
				OnAction: func(sim *core.Simulation) {
					spell.DealDamage(sim, result)
				},
			})

			sim.AddPendingAction(action)
		},
	})

	aura := paladin.MakeProcTriggerAura(core.ProcTrigger{
		Name:            "Seal of Righteousness" + paladin.Label + " " + seal.GetRankLabel(),
		ActionID:        core.ActionID{SpellID: seal.spellID},
		MetricsActionID: core.ActionID{SpellID: seal.spellID},
		Duration:        time.Second * 30,
		Outcome:         core.OutcomeLanded,
		Callback:        core.CallbackOnSpellHitDealt,
		ProcMask:        core.ProcMaskMeleeWhiteHit,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			procSpell.Cast(sim, result.Target)
		},
	})
	aura.Tag = TwistTag

	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: seal.spellID},
		ClassSpellMask: SpellMaskSealOfRighteousness,
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagAPL,
		Rank:           seal.rank,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		ManaCost: core.ManaCostOptions{
			FlatCost: int32(seal.manaCost),
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			paladin.applySeal(aura, spell, judgeSpell, sim)
		},
	})
}

// Seal of Light
// https://www.wowhead.com/tbc/spell=20165
//
// Fills the Paladin with divine light for 30 sec, giving each melee attack
// a chance to heal the Paladin. Only one Seal can be active on the Paladin
// at any one time.
//
// Unleashing this Seal's energy will judge an enemy for 20 sec, granting
// attacks against the judged enemy a chance to heal the attacker.
func (paladin *Paladin) registerSealOfLight(seal seal) {
	judgementOfLightAuras := paladin.NewEnemyAuraArray(core.JudgementOfLightAura)
	paladin.JudgementAuras = append(paladin.JudgementAuras, judgementOfLightAuras)

	judgeSpell := paladin.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{SpellID: seal.judge.spellID},
		SpellSchool:      core.SpellSchoolHoly,
		ProcMask:         core.ProcMaskEmpty,
		Flags:            core.SpellFlagMeleeMetrics | core.SpellFlagBinary,
		ClassSpellMask:   SpellMaskJudgementOfLight,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealOutcome(sim, target, spell.OutcomeAlwaysHit)
			judgementOfLightAuras.Get(target).Activate(sim)
		},
	})

	procSpell := paladin.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{SpellID: seal.proc.spellID},
		ClassSpellMask:   SpellMaskSealOfLight,
		SpellSchool:      core.SpellSchoolHoly,
		ProcMask:         core.ProcMaskSpellHealing,
		Flags:            core.SpellFlagHelpful | core.SpellFlagPassiveSpell,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealHealing(sim, target, seal.proc.value, spell.OutcomeAlwaysHit)
		},
	})

	aura := paladin.MakeProcTriggerAura(core.ProcTrigger{
		Name:            "Seal of Light" + paladin.Label + " " + seal.GetRankLabel(),
		ActionID:        core.ActionID{SpellID: seal.spellID},
		MetricsActionID: core.ActionID{SpellID: seal.spellID},
		Duration:        time.Second * 30,
		Outcome:         core.OutcomeLanded,
		Callback:        core.CallbackOnSpellHitDealt,
		ProcMask:        core.ProcMaskMeleeWhiteHit,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			procSpell.Cast(sim, result.Target)
		},
	})

	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: seal.spellID},
		ClassSpellMask: SpellMaskSealOfLight,
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagAPL,
		Rank:           seal.rank,
		ManaCost: core.ManaCostOptions{
			FlatCost: int32(seal.manaCost),
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{GCD: core.GCDDefault},
		},
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			paladin.applySeal(aura, spell, judgeSpell, sim)
		},
	})
}

// Seal of Wisdom
// https://www.wowhead.com/tbc/spell=20166
//
// Fills the Paladin with divine wisdom for 30 sec, giving each melee attack
// a chance to restore mana to the Paladin. Only one Seal can be active on
// the Paladin at any one time.
//
// Unleashing this Seal's energy will judge an enemy for 20 sec, granting
// attacks against the judged enemy a chance to restore mana to the attacker.
func (paladin *Paladin) registerSealOfWisdom(seal seal) {
	judgementOfWisdomAuras := paladin.NewEnemyAuraArray(core.JudgementOfWisdomAura)
	paladin.JudgementAuras = append(paladin.JudgementAuras, judgementOfWisdomAuras)

	judgeSpell := paladin.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{SpellID: seal.judge.spellID},
		SpellSchool:      core.SpellSchoolHoly,
		ProcMask:         core.ProcMaskEmpty,
		Flags:            core.SpellFlagMeleeMetrics | core.SpellFlagBinary,
		ClassSpellMask:   SpellMaskJudgementOfWisdom,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealOutcome(sim, target, spell.OutcomeAlwaysHit)
			judgementOfWisdomAuras.Get(target).Activate(sim)
		},
	})
	sealManaMetrics := paladin.Unit.NewManaMetrics(core.ActionID{SpellID: seal.proc.spellID})
	procSpell := paladin.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{SpellID: seal.proc.spellID},
		ClassSpellMask:   SpellMaskSealOfWisdom,
		SpellSchool:      core.SpellSchoolHoly,
		ProcMask:         core.ProcMaskEmpty,
		Flags:            core.SpellFlagHelpful | core.SpellFlagPassiveSpell,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			if spell.Unit.HasManaBar() {
				spell.Unit.AddMana(sim, seal.proc.value, sealManaMetrics)
			}
		},
	})
	aura := paladin.MakeProcTriggerAura(core.ProcTrigger{
		Name:            "Seal of Wisdom" + paladin.Label + " " + seal.GetRankLabel(),
		ActionID:        core.ActionID{SpellID: seal.spellID},
		MetricsActionID: core.ActionID{SpellID: seal.spellID},
		Duration:        time.Second * 30,
		Outcome:         core.OutcomeLanded,
		Callback:        core.CallbackOnSpellHitDealt,
		ProcMask:        core.ProcMaskMeleeWhiteHit,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			procSpell.Cast(sim, result.Target)
		},
	})
	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: seal.spellID},
		ClassSpellMask: SpellMaskSealOfWisdom,
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagAPL,
		Rank:           seal.rank,
		ManaCost: core.ManaCostOptions{
			FlatCost: int32(seal.manaCost),
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{GCD: core.GCDDefault},
		},
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			paladin.applySeal(aura, spell, judgeSpell, sim)
		},
	})
}

// Seal of Justice
// https://www.wowhead.com/tbc/spell=20164
//
// Fills the Paladin with the spirit of justice for 30 sec, giving each melee
// attack a chance to stun the target for 2 sec. Only one Seal can be active
// on the Paladin at any one time.
//
// Unleashing this Seal's energy will judge an enemy for 20 sec, preventing
// them from fleeing.
func (paladin *Paladin) registerSealOfJustice(seal seal) {
	registerJoJDebuff := func(target *core.Unit) *core.Aura {
		return target.GetOrRegisterAura(core.Aura{
			Label:    "Judgement of Justice",
			ActionID: core.ActionID{SpellID: seal.judge.spellID},
			Tag:      JudgementAuraTag,
			Duration: time.Second * 20,
			OnSpellHitTaken: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				if spell.ProcMask.Matches(core.ProcMaskMeleeWhiteHit) {
					aura.Refresh(sim)
				}
			},
		})
	}

	judgementOfJusticeAuras := paladin.NewEnemyAuraArray(registerJoJDebuff)
	paladin.JudgementAuras = append(paladin.JudgementAuras, judgementOfJusticeAuras)

	judgeSpell := paladin.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{SpellID: seal.judge.spellID},
		SpellSchool:      core.SpellSchoolHoly,
		ProcMask:         core.ProcMaskEmpty,
		Flags:            core.SpellFlagMeleeMetrics | core.SpellFlagBinary,
		ClassSpellMask:   SpellMaskJudgementOfJustice,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealOutcome(sim, target, spell.OutcomeAlwaysHit)
			judgementOfJusticeAuras.Get(target).Activate(sim)
		},
	})
	procSpell := paladin.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{SpellID: seal.proc.spellID},
		ClassSpellMask:   SpellMaskSealOfJustice,
		SpellSchool:      core.SpellSchoolHoly,
		ProcMask:         core.ProcMaskEmpty,
		Flags:            core.SpellFlagMeleeMetrics | core.SpellFlagPassiveSpell,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealOutcome(sim, target, spell.OutcomeAlwaysHit)
		},
	})
	aura := paladin.MakeProcTriggerAura(core.ProcTrigger{
		Name:            "Seal of Justice" + paladin.Label + " " + seal.GetRankLabel(),
		ActionID:        core.ActionID{SpellID: seal.spellID},
		MetricsActionID: core.ActionID{SpellID: seal.spellID},
		Duration:        time.Second * 30,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			procSpell.Cast(sim, result.Target)
		},
	})
	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: seal.spellID},
		ClassSpellMask: SpellMaskSealOfJustice,
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagAPL,
		Rank:           seal.rank,
		ManaCost: core.ManaCostOptions{
			BaseCostPercent: seal.manaCost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{GCD: core.GCDDefault},
		},
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			paladin.applySeal(aura, spell, judgeSpell, sim)
		},
	})
}

// Seal of the Crusader
// https://www.wowhead.com/tbc/spell=21082
//
// Fills the Paladin with the spirit of a crusader for 30 sec, increasing
// attack speed but reducing damage caused by each weapon hit. The
// Paladin also causes additional threat. Only one Seal can be active on
// the Paladin at any one time.
//
// Unleashing this Seal's energy will judge an enemy for 20 sec, increasing
// Holy damage taken from all sources.
func (paladin *Paladin) registerSealOfTheCrusader(seal seal) {
	percentBonus := core.Ternary(paladin.CouldHaveSetBonus(ItemSetJusticarBattlegear, 2), 1.15, 1.0)
	flatBonus := 0.0
	if paladin.Ranged().ID == 23203 { //https://www.wowhead.com/tbc/item=23203/libram-of-fervor
		flatBonus += 33.0
	} else if paladin.Ranged().ID == 27949 || paladin.Ranged().ID == 27983 { //https://www.wowhead.com/tbc/item=27949/libram-of-zeal
		flatBonus += 47.0
	}

	judgementOfTheCrusaderAuras := paladin.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return core.ImprovedSealOfTheCrusaderAura(target, 1, paladin.Talents.ImprovedSealOfTheCrusader, flatBonus, percentBonus)
	})

	paladin.JudgementAuras = append(paladin.JudgementAuras, judgementOfTheCrusaderAuras)

	judgeSpell := paladin.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{SpellID: seal.judge.spellID},
		SpellSchool:      core.SpellSchoolHoly,
		ProcMask:         core.ProcMaskEmpty,
		Flags:            core.SpellFlagMeleeMetrics | core.SpellFlagBinary,
		ClassSpellMask:   SpellMaskJudgementOfTheCrusader,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		CritMultiplier:   1,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealOutcome(sim, target, spell.OutcomeAlwaysHit)
			judgementOfTheCrusaderAuras.Get(target).Activate(sim)
		},
	})

	aura := paladin.RegisterAura(core.Aura{
		Label:    "Seal of the Crusader" + paladin.Label + " " + seal.GetRankLabel(),
		ActionID: core.ActionID{SpellID: seal.spellID},
		Duration: time.Second * 30,
	}).
		AttachMultiplyMeleeSpeed(1.4).
		AttachSpellMod(core.SpellModConfig{
			ProcMask:   core.ProcMaskMeleeMHAuto,
			Kind:       core.SpellMod_DamageDone_Flat,
			FloatValue: -0.4,
		}).
		AttachStatBuff(stats.AttackPower, seal.proc.value)

	paladin.RegisterSpell(core.SpellConfig{
		ActionID:         aura.ActionID,
		ClassSpellMask:   SpellMaskSealOfTheCrusader,
		SpellSchool:      core.SpellSchoolHoly,
		ProcMask:         core.ProcMaskEmpty,
		Flags:            core.SpellFlagAPL,
		Rank:             seal.rank,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		ManaCost: core.ManaCostOptions{
			FlatCost: int32(seal.manaCost),
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{GCD: core.GCDDefault},
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			paladin.applySeal(aura, spell, judgeSpell, sim)
		},
		RelatedSelfBuff: aura,
	})
}

// Seal of Blood
// https://www.wowhead.com/tbc/spell=31892
//
// All melee attacks deal additional Holy damage equal to 35% of normal weapon damage, but the Paladin loses health equal to 10% of the total damage inflicted.
//
// Unleashing this Seal's energy will judge an enemy, instantly causing 295 to 325 Holy damage at the cost of health equal to 33% of the damage caused.
func (paladin *Paladin) registerSealOfBlood() {
	judgeSpell := paladin.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{SpellID: 31898},
		SpellSchool:      core.SpellSchoolHoly,
		ProcMask:         core.ProcMaskMeleeMHSpecial,
		Flags:            core.SpellFlagMeleeMetrics,
		ClassSpellMask:   SpellMaskJudgementOfBlood,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		CritMultiplier:   paladin.DefaultMeleeCritMultiplier(),
		BonusCoefficient: 0.429,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			flags := spell.Flags
			baseDamage := sim.Roll(295, 325)
			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialCritOnly)
			action := core.NewDelayedAction(core.DelayedActionOptions{
				DoAt:     sim.CurrentTime + core.SpellBatchWindow,
				Priority: core.ActionPriorityLow,
				OnAction: func(sim *core.Simulation) {
					currentFlags := spell.Flags
					spell.Flags = flags
					spell.DealDamage(sim, result)
					spell.Flags = currentFlags
				},
			})
			sim.AddPendingAction(action)
		},
	})
	procSpell := paladin.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{SpellID: 31893},
		ClassSpellMask:   SpellMaskSealOfBlood,
		SpellSchool:      core.SpellSchoolHoly,
		ProcMask:         core.ProcMaskMeleeProc,
		Flags:            core.SpellFlagMeleeMetrics | core.SpellFlagPassiveSpell,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		CritMultiplier:   paladin.DefaultMeleeCritMultiplier(),
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := spell.Unit.MHWeaponDamage(sim, spell.MeleeAttackPower(target)) * 0.35
			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
			action := core.NewDelayedAction(core.DelayedActionOptions{
				DoAt:     sim.CurrentTime + core.SpellBatchWindow,
				Priority: core.ActionPriorityLow,
				OnAction: func(sim *core.Simulation) {
					spell.DealDamage(sim, result)
				},
			})
			sim.AddPendingAction(action)
		},
	})
	aura := paladin.MakeProcTriggerAura(core.ProcTrigger{
		Name:            "Seal of Blood" + paladin.Label,
		ActionID:        core.ActionID{SpellID: 31892},
		MetricsActionID: core.ActionID{SpellID: 31892},
		Duration:        time.Second * 30,
		Callback:        core.CallbackOnSpellHitDealt,
		Outcome:         core.OutcomeLanded,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !spell.ProcMask.Matches(core.ProcMaskMeleeWhiteHit) && !spell.Matches(SpellMaskSealOfCommand) {
				return
			}

			procSpell.Cast(sim, result.Target)
		},
	})
	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 31892},
		ClassSpellMask: SpellMaskSealOfBlood,
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagAPL,
		ManaCost: core.ManaCostOptions{
			FlatCost: 210,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			paladin.applySeal(aura, spell, judgeSpell, sim)
		},
	})
}

// Seal of Vengeance
// https://www.wowhead.com/tbc/spell=31801
//
// Fills the Paladin with holy power, granting each melee attack a chance to cause 150 Holy damage over 15 sec.
// This effect can stack up to 5 times.
// Only one Seal can be active on the Paladin at any one time.
// Lasts 30 sec.
//
// Unleashing this Seal's energy will judge an enemy, instantly causing 120 Holy damage per application of Holy Vengeance.
func (paladin *Paladin) registerSealOfVengeance() {
	holyVengeanceTag := "Holy Vengeance"
	judgeSpell := paladin.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{SpellID: 31804},
		SpellSchool:      core.SpellSchoolHoly,
		ProcMask:         core.ProcMaskEmpty,
		Flags:            core.SpellFlagMeleeMetrics,
		ClassSpellMask:   SpellMaskJudgementOfVengeance,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		CritMultiplier:   paladin.DefaultSpellCritMultiplier(),
		BonusCoefficient: 0.429,
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return target.GetActiveAuraWithTag(holyVengeanceTag) != nil
		},
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			damage := 120 * float64(target.GetActiveAuraWithTag(holyVengeanceTag).GetStacks())
			result := spell.CalcDamage(sim, target, damage, spell.OutcomeMagicHitAndCrit)
			action := core.NewDelayedAction(core.DelayedActionOptions{
				DoAt:     sim.CurrentTime + core.SpellBatchWindow,
				Priority: core.ActionPriorityLow,
				OnAction: func(sim *core.Simulation) {
					spell.DealDamage(sim, result)
				},
			})
			sim.AddPendingAction(action)
		},
	})
	procSpell := paladin.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{SpellID: 42463},
		ClassSpellMask:   SpellMaskSealOfVengeance,
		SpellSchool:      core.SpellSchoolHoly,
		ProcMask:         core.ProcMaskEmpty,
		Flags:            core.SpellFlagPassiveSpell,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			attackTable := spell.Unit.AttackTables[target.UnitIndex]
			damage := (10 + spell.BonusDamage(attackTable)*0.034/3) * paladin.MainHand().SwingSpeed
			spell.CalcAndDealDamage(sim, target, damage, spell.OutcomeMagicHit)
		},
	})
	holyVengeanceDot := paladin.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{SpellID: 31803},
		ClassSpellMask:   SpellMaskSealOfVengeance,
		SpellSchool:      core.SpellSchoolHoly,
		ProcMask:         core.ProcMaskEmpty,
		Flags:            core.SpellFlagPassiveSpell | core.SpellFlagMeleeMetrics,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		Dot: core.DotConfig{
			Aura: core.Aura{
				Label:     "Holy Vengeance" + paladin.Label,
				Tag:       holyVengeanceTag,
				ActionID:  core.ActionID{SpellID: 31803},
				MaxStacks: 5,
			},
			NumberOfTicks: 5,
			TickLength:    time.Second * 3,
			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				attackTable := dot.Spell.Unit.AttackTables[target.UnitIndex]
				dot.Snapshot(target, 30+dot.Spell.BonusDamage(attackTable)*0.034*float64(dot.GetStacks()))
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			hitResult := spell.CalcOutcome(sim, target, spell.OutcomeMagicHit)
			if !hitResult.Landed() {
				spell.DealOutcome(sim, hitResult)
				return
			}

			dot := spell.Dot(target)
			if dot.IsActive() {
				dot.AddStack(sim)
				dot.TakeSnapshot(sim)
				dot.Refresh(sim)
			} else {
				dot.Apply(sim)
				dot.SetStacks(sim, 1)
				dot.TakeSnapshot(sim)
			}
		},
	})
	aura := paladin.MakeProcTriggerAura(core.ProcTrigger{
		Name:            "Seal of Vengeance" + paladin.Label,
		ActionID:        core.ActionID{SpellID: 31801},
		MetricsActionID: core.ActionID{SpellID: 31801},
		Duration:        time.Second * 30,
		Callback:        core.CallbackOnSpellHitDealt,
		ProcMask:        core.ProcMaskMeleeWhiteHit,
		Outcome:         core.OutcomeLanded,
		DPM:             paladin.NewStaticLegacyPPMManager(15, core.ProcMaskMeleeWhiteHit),
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			dot := holyVengeanceDot.Dot(result.Target)
			if dot.IsActive() && dot.GetStacks() == 5 {
				procSpell.Cast(sim, result.Target)
			}

			holyVengeanceDot.Cast(sim, result.Target)
		},
	})
	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 31801},
		ClassSpellMask: SpellMaskSealOfVengeance,
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagAPL,
		ManaCost: core.ManaCostOptions{
			FlatCost: 250,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			paladin.applySeal(aura, spell, judgeSpell, sim)
		},
	})
}

// Seal of Command
// https://www.wowhead.com/tbc/spell=20375
//
// Gives the Paladin a chance to deal additional Holy damage equal to 70%
// of normal weapon damage. Only one Seal can be active on the Paladin at
// any one time. Lasts 30 sec.
//
// Unleashing this Seal's energy will judge an enemy, instantly causing
// 228 to 252 Holy damage, 456 to 504 if the target is stunned or incapacitated. (stunned is just damage x2)
func (paladin *Paladin) registerSealOfCommandRank(seal seal) {
	minDamage := seal.judge.minDamage
	maxDamage := seal.judge.maxDamage
	judgeSpell := paladin.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{SpellID: seal.judge.spellID},
		SpellSchool:      core.SpellSchoolHoly,
		ProcMask:         core.ProcMaskMeleeMHSpecial,
		Flags:            core.SpellFlagMeleeMetrics,
		ClassSpellMask:   SpellMaskJudgementOfCommand,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		CritMultiplier:   paladin.DefaultMeleeCritMultiplier(),
		BonusCoefficient: seal.judge.coeff,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := sim.Roll(minDamage, maxDamage)
			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialCritOnly)
			action := core.NewDelayedAction(core.DelayedActionOptions{
				DoAt:     sim.CurrentTime + core.SpellBatchWindow,
				Priority: core.ActionPriorityLow,
				OnAction: func(sim *core.Simulation) {
					spell.DealDamage(sim, result)
				},
			})
			sim.AddPendingAction(action)
		},
	})

	procSpell := paladin.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{SpellID: seal.proc.spellID},
		ClassSpellMask:   SpellMaskSealOfCommand,
		SpellSchool:      core.SpellSchoolHoly,
		ProcMask:         core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeProc,
		Flags:            core.SpellFlagMeleeMetrics | core.SpellFlagPassiveSpell,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		CritMultiplier:   paladin.DefaultMeleeCritMultiplier(),
		BonusCoefficient: seal.proc.coeff,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := spell.Unit.MHWeaponDamage(sim, spell.MeleeAttackPower(target)) * seal.proc.value
			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
			action := core.NewDelayedAction(core.DelayedActionOptions{
				DoAt:     sim.CurrentTime + core.SpellBatchWindow,
				Priority: core.ActionPriorityLow,
				OnAction: func(sim *core.Simulation) {
					spell.DealDamage(sim, result)
				},
			})
			sim.AddPendingAction(action)
		},
	})

	aura := paladin.RegisterAura(core.Aura{
		Label:    "Seal of Command" + paladin.Label + " " + seal.GetRankLabel(),
		ActionID: core.ActionID{SpellID: seal.spellID},
		Duration: time.Second * 30,
		Tag:      TwistTag,
	}).AttachProcTrigger(core.ProcTrigger{
		Outcome:  core.OutcomeLanded,
		ProcMask: core.ProcMaskMeleeWhiteHit,
		Callback: core.CallbackOnSpellHitDealt,
		ICD:      time.Second * 1,
		DPM:      paladin.NewLegacyPPMManager(7, core.ProcMaskMeleeWhiteHit),
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			procSpell.Cast(sim, result.Target)
		},
	})

	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       aura.ActionID,
		ClassSpellMask: SpellMaskSealOfCommand,
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagAPL,
		Rank:           seal.rank,
		ManaCost: core.ManaCostOptions{
			FlatCost: int32(seal.manaCost),
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{GCD: core.GCDDefault},
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			paladin.applySeal(aura, spell, judgeSpell, sim)
		},
	})
}
