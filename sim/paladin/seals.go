package paladin

import (
	"strconv"
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

type proc struct {
	spellID int32
	value   float64
	scale   float64
	coeff   float64
}

type judge struct {
	spellID   int32
	minDamage float64
	maxDamage float64
	scale     float64
	coeff     float64
}

type seal struct {
	level      int32
	spellID    int32
	manaCost   float64
	scaleLevel int32
	proc       proc
	judge      judge
}

func (paladin *Paladin) registerSeals() {
	paladin.registerSealOfRighteousness()
	paladin.registerSealOfLight()
	paladin.registerSealOfWisdom()
	paladin.registerSealOfJustice()
	paladin.registerSealOfTheCrusader()
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
func (paladin *Paladin) applySeal(newSeal *core.Aura, judgement *core.Spell, sim *core.Simulation) {
	if paladin.CurrentSeal != nil {
		newSealLabel := newSeal.ActionID.SpellID
		currentSealLabel := paladin.CurrentSeal.ActionID.SpellID
		// If they are recasting the same seal, we just refresh the duration
		if newSealLabel == currentSealLabel {
			paladin.CurrentSeal.Refresh(sim)
			return
		}
	}

	// Twisting only occurs when current seal is Command or Righteousness
	if paladin.CurrentSeal.IsActive() && paladin.CurrentSeal.Tag == TwistTag {
		paladin.CurrentSeal.UpdateExpires(sim.CurrentTime + (time.Millisecond * 399)) // always update, even if it extends duration
		paladin.PreviousSeal = paladin.CurrentSeal
		paladin.PreviousJudgement = paladin.CurrentJudgement
	}

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
func (paladin *Paladin) registerSealOfRighteousness() {
	var ranks = []seal{
		{}, // Dummy to offset the index
		{level: 1, spellID: 21084, manaCost: 20, scaleLevel: 7, proc: proc{spellID: 25742, value: 108, scale: 18, coeff: 0.029}, judge: judge{spellID: 20187, minDamage: 15, maxDamage: 15, scale: 1.8, coeff: 0.209}},
		{level: 10, spellID: 20287, manaCost: 40, scaleLevel: 16, proc: proc{spellID: 25740, value: 216, scale: 17, coeff: 0.063}, judge: judge{spellID: 20280, minDamage: 25, maxDamage: 27, scale: 1.9, coeff: 0.455}},
		{level: 18, spellID: 20288, manaCost: 60, scaleLevel: 24, proc: proc{spellID: 25739, value: 352, scale: 23, coeff: 0.093}, judge: judge{spellID: 20281, minDamage: 39, maxDamage: 43, scale: 2.4, coeff: 0.674}},
		{level: 26, spellID: 20289, manaCost: 90, scaleLevel: 32, proc: proc{spellID: 25738, value: 541, scale: 31, coeff: 0.1}, judge: judge{spellID: 20282, minDamage: 57, maxDamage: 63, scale: 2.8, coeff: 0.728}},
		{level: 34, spellID: 20290, manaCost: 120, scaleLevel: 40, proc: proc{spellID: 25737, value: 785, scale: 37, coeff: 0.1}, judge: judge{spellID: 20283, minDamage: 78, maxDamage: 86, scale: 3.1, coeff: 0.728}},
		{level: 42, spellID: 20291, manaCost: 140, scaleLevel: 48, proc: proc{spellID: 25736, value: 1082, scale: 41, coeff: 0.1}, judge: judge{spellID: 20284, minDamage: 102, maxDamage: 112, scale: 3.8, coeff: 0.728}},
		{level: 50, spellID: 20292, manaCost: 170, scaleLevel: 56, proc: proc{spellID: 25735, value: 1407, scale: 47, coeff: 0.1}, judge: judge{spellID: 20285, minDamage: 131, maxDamage: 143, scale: 4.1, coeff: 0.728}},
		{level: 58, spellID: 20293, manaCost: 200, scaleLevel: 64, proc: proc{spellID: 25713, value: 1786, scale: 47, coeff: 0.1}, judge: judge{spellID: 20286, minDamage: 162, maxDamage: 178, scale: 4.1, coeff: 0.728}},
		{level: 66, spellID: 27155, manaCost: 260, scaleLevel: 70, proc: proc{spellID: 27156, value: 2112, scale: 53, coeff: 0.1}, judge: judge{spellID: 27157, minDamage: 208, maxDamage: 228, scale: 4.3, coeff: 0.728}}, // TODO: verify judgement scale
	}

	for rank := 1; rank < len(ranks); rank++ {
		if paladin.Level < ranks[rank].level {
			break
		}

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

		minDamage := ranks[rank].judge.minDamage + ranks[rank].judge.scale*float64(min(paladin.Level, ranks[rank].scaleLevel)-ranks[rank].level)
		maxDamage := ranks[rank].judge.maxDamage + ranks[rank].judge.scale*float64(min(paladin.Level, ranks[rank].scaleLevel)-ranks[rank].level)

		judgeSpell := paladin.RegisterSpell(core.SpellConfig{
			ActionID:       core.ActionID{SpellID: ranks[rank].judge.spellID},
			SpellSchool:    core.SpellSchoolHoly,
			ProcMask:       core.ProcMaskSpellDamage,
			Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagBinary, // | core.SpellFlagSuppressWeaponProcs | core.SpellFlagSuppressEquipProcs
			ClassSpellMask: SpellMaskJudgementOfRighteousness,

			BonusCoefficient: ranks[rank].judge.coeff,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				flags := spell.Flags
				baseDamage := sim.Roll(minDamage, maxDamage)
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

		baseDamage := float64(ranks[rank].proc.value) + float64(ranks[rank].proc.scale)*float64(min(paladin.Level, ranks[rank].scaleLevel)-ranks[rank].level)
		damage := 1.2*(float64(baseDamage)*1.2*1.03*paladin.MainHand().SwingSpeed/100) + 0.03*(float64(paladin.MainHand().WeaponDamageMax)+float64(paladin.MainHand().WeaponDamageMin))/2 + 1

		procSpell := paladin.RegisterSpell(core.SpellConfig{
			ActionID:       core.ActionID{SpellID: ranks[rank].proc.spellID},
			SpellSchool:    core.SpellSchoolHoly,
			ProcMask:       core.ProcMaskMeleeMHSpecial, //changed to ProcMaskMeleeMHSpecial, to allow procs from weapons/oils which do proc from SoR, -- TODO: Verify in TBC
			Flags:          core.SpellFlagMeleeMetrics,  // | core.SpellFlagSuppressEquipProcs | core.SpellFlagBatchStartAttackMacro, // but Wild Strikes does not proc, nor equip procs
			ClassSpellMask: SpellMaskSealOfRighteousness,

			BonusCoefficient: ranks[rank].proc.coeff,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				// effectively scales with coeff x 2, and damage dealt multipliers affect half the damage taken bonus -- TODO: Verify in TBC
				// x := spell.Unit.PseudoStats.BonusDamage + spell.Unit.GetStat(stats.HolyDamage) + spell.Unit.GetStat(stats.SpellDamage) + spell.Unit.PseudoStats.MobTypeSpellDamage
				// baseDamage := damage + spell.BonusCoefficient*(x+target.GetSchoolBonusDamageTaken(spell))

				result := spell.CalcDamage(sim, target, damage, spell.OutcomeMeleeSpecialCritOnly)

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
			Label:    "Seal of Righteousness" + paladin.Label + strconv.Itoa(rank),
			ActionID: core.ActionID{SpellID: ranks[rank].spellID},
			Duration: time.Second * 30,
			Tag:      TwistTag,

			OnSpellHitDealt: func(_ *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				if !result.Landed() {
					return
				}
				if spell.ProcMask.Matches(core.ProcMaskMeleeWhiteHit) {
					procSpell.Cast(sim, result.Target)
				}
			},
		})

		sealSpell := paladin.RegisterSpell(core.SpellConfig{
			ActionID:    aura.ActionID,
			SpellSchool: core.SpellSchoolHoly,
			Flags:       core.SpellFlagAPL,

			ManaCost: core.ManaCostOptions{
				FlatCost:        int32(ranks[rank].manaCost),
			},
			Cast: core.CastConfig{
				DefaultCast: core.Cast{
					GCD: core.GCDDefault,
				},
			},

			ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
				paladin.applySeal(aura, judgeSpell, sim)
			},
		})

		paladin.SealOfRighteousness = append(paladin.SealOfRighteousness, sealSpell)
		paladin.SealOfRighteousnessJudgements = append(paladin.SealOfRighteousnessJudgements, judgeSpell)
		paladin.SealOfRighteousnessAuras = append(paladin.SealOfRighteousnessAuras, aura)
	}
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
func (paladin *Paladin) registerSealOfLight() {
	var ranks = []seal{
		{}, // Dummy to offset the index
		{level: 30, spellID: 20165, manaCost: 110, scaleLevel: 0, proc: proc{spellID: 20167, value: 39, scale: 0, coeff: 0.0}, judge: judge{spellID: 20185, minDamage: 25, maxDamage: 25, scale: 0.0, coeff: 0.0}},
		{level: 40, spellID: 20347, manaCost: 140, scaleLevel: 0, proc: proc{spellID: 20333, value: 53, scale: 0, coeff: 0.0}, judge: judge{spellID: 20344, minDamage: 34, maxDamage: 34, scale: 0.0, coeff: 0.0}},
		{level: 50, spellID: 20348, manaCost: 180, scaleLevel: 0, proc: proc{spellID: 20334, value: 76, scale: 0, coeff: 0.0}, judge: judge{spellID: 20345, minDamage: 49, maxDamage: 49, scale: 0.0, coeff: 0.0}},
		{level: 60, spellID: 20349, manaCost: 210, scaleLevel: 0, proc: proc{spellID: 20340, value: 94, scale: 0, coeff: 0.0}, judge: judge{spellID: 20346, minDamage: 61, maxDamage: 61, scale: 0.0, coeff: 0.0}},
		{level: 69, spellID: 27160, manaCost: 280, scaleLevel: 0, proc: proc{spellID: 27161, value: 133, scale: 0, coeff: 0.0}, judge: judge{spellID: 27162, minDamage: 95, maxDamage: 95, scale: 0.0, coeff: 0.0}},
	}

	for rank := 1; rank < len(ranks); rank++ {
		if paladin.Level < ranks[rank].level {
			break
		}

		// 50% on hit to heal for judge.MinDamage
		registerJoLDebuff := func(target *core.Unit) *core.Aura {
			return target.GetOrRegisterAura(core.Aura{
				Label:    "Judgement of Light" + paladin.Label + strconv.Itoa(rank),
				ActionID: core.ActionID{SpellID: ranks[rank].judge.spellID},
				Tag:      JudgementAuraTag,
				Duration: time.Second * 20,

				OnSpellHitTaken: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
					if !result.Landed() || !spell.Unit.HasHealthBar() {
						return
					}

					if spell.ProcMask.Matches(core.ProcMaskMeleeOrRanged) {
						if sim.Proc(0.5, "Judgement of Light - Heal") {
							result := spell.CalcHealing(sim, spell.Unit, ranks[rank].judge.minDamage, spell.OutcomeAlwaysHit)
							spell.DealHealing(sim, result)
						}
					}

					if spell.ProcMask.Matches(core.ProcMaskMeleeWhiteHit) {
						aura.Refresh(sim)
					}
				},
			})
		}

		debuffs := paladin.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
			return registerJoLDebuff(target)
		})

		judgeSpell := paladin.RegisterSpell(core.SpellConfig{
			ActionID:       core.ActionID{SpellID: ranks[rank].judge.spellID},
			SpellSchool:    core.SpellSchoolHoly,
			ProcMask:       core.ProcMaskEmpty,
			Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagBinary, // | core.SpellFlagSuppressWeaponProcs | core.SpellFlagSuppressEquipProcs
			ClassSpellMask: SpellMaskJudgementOfLight,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.CalcAndDealOutcome(sim, target, spell.OutcomeAlwaysHit)
				debuffs.Get(target).Activate(sim)
			},
		})

		procSpell := paladin.RegisterSpell(core.SpellConfig{
			ActionID:       core.ActionID{SpellID: ranks[rank].proc.spellID},
			SpellSchool:    core.SpellSchoolHoly,
			ProcMask:       core.ProcMaskSpellHealing,
			Flags:          core.SpellFlagHelpful,
			ClassSpellMask: SpellMaskSealOfLight,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.CalcAndDealHealing(sim, target, ranks[rank].proc.value, spell.OutcomeAlwaysHit)
			},
		})

		aura := paladin.RegisterAura(core.Aura{
			Label:    "Seal of Light" + paladin.Label + strconv.Itoa(rank),
			ActionID: core.ActionID{SpellID: ranks[rank].spellID},
			Duration: time.Second * 30,

			OnSpellHitDealt: func(_ *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				if !result.Landed() {
					return
				}
				if spell.ProcMask.Matches(core.ProcMaskMeleeWhiteHit) {
					procSpell.Cast(sim, spell.Unit)
				}
			},
		})

		sealSpell := paladin.RegisterSpell(core.SpellConfig{
			ActionID:    aura.ActionID,
			SpellSchool: core.SpellSchoolHoly,
			Flags:       core.SpellFlagAPL,

			ManaCost: core.ManaCostOptions{
				FlatCost:        int32(ranks[rank].manaCost),
			},
			Cast: core.CastConfig{
				DefaultCast: core.Cast{
					GCD: core.GCDDefault,
				},
			},

			ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
				paladin.applySeal(aura, judgeSpell, sim)
			},
		})

		paladin.SealOfLight = append(paladin.SealOfLight, sealSpell)
		paladin.SealOfLightJudgements = append(paladin.SealOfLightJudgements, judgeSpell)
		paladin.SealOfLightAuras = append(paladin.SealOfLightAuras, aura)
	}
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
func (paladin *Paladin) registerSealOfWisdom() {
	var ranks = []seal{
		{}, // Dummy to offset the index
		{level: 38, spellID: 20166, manaCost: 135, scaleLevel: 0, proc: proc{spellID: 20168, value: 50, scale: 0, coeff: 0.0}, judge: judge{spellID: 20186, minDamage: 33, maxDamage: 33, scale: 0.0, coeff: 0.0}},
		{level: 48, spellID: 20356, manaCost: 170, scaleLevel: 0, proc: proc{spellID: 20350, value: 71, scale: 0, coeff: 0.0}, judge: judge{spellID: 20354, minDamage: 46, maxDamage: 46, scale: 0.0, coeff: 0.0}},
		{level: 58, spellID: 20357, manaCost: 200, scaleLevel: 0, proc: proc{spellID: 20351, value: 90, scale: 0, coeff: 0.0}, judge: judge{spellID: 20355, minDamage: 59, maxDamage: 59, scale: 0.0, coeff: 0.0}},
		{level: 67, spellID: 27166, manaCost: 270, scaleLevel: 0, proc: proc{spellID: 27167, value: 121, scale: 0, coeff: 0.0}, judge: judge{spellID: 27164, minDamage: 74, maxDamage: 74, scale: 0.0, coeff: 0.0}},
	}

	for rank := 1; rank < len(ranks); rank++ {
		if paladin.Level < ranks[rank].level {
			break
		}

		// 50% on hit to restore mana for judge.MinDamage
		judgeManaMetrics := paladin.Unit.NewManaMetrics(core.ActionID{SpellID: ranks[rank].judge.spellID})
		registerJoWDebuff := func(target *core.Unit) *core.Aura {
			return target.GetOrRegisterAura(core.Aura{
				Label:    "Judgement of Wisdom" + paladin.Label + strconv.Itoa(rank),
				ActionID: core.ActionID{SpellID: ranks[rank].judge.spellID},
				Tag:      JudgementAuraTag,
				Duration: time.Second * 20,

				OnSpellHitTaken: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
					if !result.Landed() || !spell.Unit.HasManaBar() {
						return
					}

					if spell.ProcMask.Matches(core.ProcMaskMeleeOrRanged) {
						if sim.Proc(0.5, "Judgement of Wisdom - Mana") {
							spell.Unit.AddMana(sim, ranks[rank].judge.minDamage, judgeManaMetrics)
						}
					}

					if spell.ProcMask.Matches(core.ProcMaskMeleeWhiteHit) {
						aura.Refresh(sim)
					}
				},
			})
		}

		debuffs := paladin.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
			return registerJoWDebuff(target)
		})

		judgeSpell := paladin.RegisterSpell(core.SpellConfig{
			ActionID:       core.ActionID{SpellID: ranks[rank].judge.spellID},
			SpellSchool:    core.SpellSchoolHoly,
			ProcMask:       core.ProcMaskEmpty,
			Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagBinary,
			ClassSpellMask: SpellMaskJudgementOfWisdom,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.CalcAndDealOutcome(sim, target, spell.OutcomeAlwaysHit)
				debuffs.Get(target).Activate(sim)
			},
		})

		sealManaMetrics := paladin.Unit.NewManaMetrics(core.ActionID{SpellID: ranks[rank].proc.spellID})
		procSpell := paladin.RegisterSpell(core.SpellConfig{
			ActionID:       core.ActionID{SpellID: ranks[rank].proc.spellID},
			SpellSchool:    core.SpellSchoolHoly,
			ProcMask:       core.ProcMaskEmpty,
			Flags:          core.SpellFlagHelpful,
			ClassSpellMask: SpellMaskSealOfWisdom,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				if spell.Unit.HasManaBar() {
					spell.Unit.AddMana(sim, ranks[rank].proc.value, sealManaMetrics)
				}
			},
		})

		aura := paladin.RegisterAura(core.Aura{
			Label:    "Seal of Wisdom" + paladin.Label + strconv.Itoa(rank),
			ActionID: core.ActionID{SpellID: ranks[rank].spellID},
			Duration: time.Second * 30,

			OnSpellHitDealt: func(_ *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				if !result.Landed() {
					return
				}
				if spell.ProcMask.Matches(core.ProcMaskMeleeWhiteHit) {
					procSpell.Cast(sim, spell.Unit)
				}
			},
		})

		sealSpell := paladin.RegisterSpell(core.SpellConfig{
			ActionID:    aura.ActionID,
			SpellSchool: core.SpellSchoolHoly,
			Flags:       core.SpellFlagAPL,

			ManaCost: core.ManaCostOptions{
				FlatCost:        int32(ranks[rank].manaCost),
			},
			Cast: core.CastConfig{
				DefaultCast: core.Cast{
					GCD: core.GCDDefault,
				},
			},

			ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
				paladin.applySeal(aura, judgeSpell, sim)
			},
		})

		paladin.SealOfWisdom = append(paladin.SealOfWisdom, sealSpell)
		paladin.SealOfWisdomJudgements = append(paladin.SealOfWisdomJudgements, judgeSpell)
		paladin.SealOfWisdomAuras = append(paladin.SealOfWisdomAuras, aura)
	}
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
func (paladin *Paladin) registerSealOfJustice() {
	var ranks = []seal{
		{}, // Dummy to offset the index
		{level: 22, spellID: 20164, manaCost: 10, scaleLevel: 0, proc: proc{spellID: 20170, value: 0, scale: 0, coeff: 0.0}, judge: judge{spellID: 20184, minDamage: 0, maxDamage: 0, scale: 0.0, coeff: 0.0}},
		{level: 48, spellID: 31895, manaCost: 10, scaleLevel: 0, proc: proc{spellID: 20170, value: 0, scale: 0, coeff: 0.0}, judge: judge{spellID: 31896, minDamage: 0, maxDamage: 0, scale: 0.0, coeff: 0.0}},
	}

	for rank := 1; rank < len(ranks); rank++ {
		if paladin.Level < ranks[rank].level {
			break
		}

		registerJoJDebuff := func(target *core.Unit) *core.Aura {
			return target.GetOrRegisterAura(core.Aura{
				Label:    "Judgement of Justice" + paladin.Label + strconv.Itoa(rank),
				ActionID: core.ActionID{SpellID: ranks[rank].judge.spellID},
				Tag:      JudgementAuraTag,
				Duration: time.Second * 20,

				OnSpellHitTaken: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
					if spell.ProcMask.Matches(core.ProcMaskMeleeWhiteHit) {
						aura.Refresh(sim)
					}
				},
			})
		}

		debuffs := paladin.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
			return registerJoJDebuff(target)
		})

		judgeSpell := paladin.RegisterSpell(core.SpellConfig{
			ActionID:       core.ActionID{SpellID: ranks[rank].judge.spellID},
			SpellSchool:    core.SpellSchoolHoly,
			ProcMask:       core.ProcMaskEmpty,
			Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagBinary,
			ClassSpellMask: SpellMaskJudgementOfJustice,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.CalcAndDealOutcome(sim, target, spell.OutcomeAlwaysHit)
				debuffs.Get(target).Activate(sim)
			},
		})

		procSpell := paladin.RegisterSpell(core.SpellConfig{
			ActionID:       core.ActionID{SpellID: ranks[rank].proc.spellID},
			SpellSchool:    core.SpellSchoolHoly,
			ProcMask:       core.ProcMaskEmpty,
			Flags:          core.SpellFlagMeleeMetrics,
			ClassSpellMask: SpellMaskSealOfJustice,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.CalcAndDealOutcome(sim, target, spell.OutcomeAlwaysHit)
			},
		})

		aura := paladin.RegisterAura(core.Aura{
			Label:    "Seal of Justice" + paladin.Label + strconv.Itoa(rank),
			ActionID: core.ActionID{SpellID: ranks[rank].spellID},
			Duration: time.Second * 30,

			OnSpellHitDealt: func(_ *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				if !result.Landed() {
					return
				}
				if spell.ProcMask.Matches(core.ProcMaskMeleeWhiteHit) {
					procSpell.Cast(sim, spell.Unit)
				}
			},
		})

		sealSpell := paladin.RegisterSpell(core.SpellConfig{
			ActionID:    aura.ActionID,
			SpellSchool: core.SpellSchoolHoly,
			Flags:       core.SpellFlagAPL,

			ManaCost: core.ManaCostOptions{
				BaseCostPercent: ranks[rank].manaCost,
			},
			Cast: core.CastConfig{
				DefaultCast: core.Cast{
					GCD: core.GCDDefault,
				},
			},

			ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
				paladin.applySeal(aura, judgeSpell, sim)
			},
		})

		paladin.SealOfJustice = append(paladin.SealOfJustice, sealSpell)
		paladin.SealOfJusticeJudgements = append(paladin.SealOfJusticeJudgements, judgeSpell)
		paladin.SealOfJusticeAuras = append(paladin.SealOfJusticeAuras, aura)
	}
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
func (paladin *Paladin) registerSealOfTheCrusader() {
	var ranks = []seal{
		{}, // Dummy to offset the index
		{level: 6, spellID: 21082, manaCost: 25, scaleLevel: 12, proc: proc{spellID: 21082, value: 36, scale: 0.7}, judge: judge{spellID: 21183, minDamage: 23}},
		{level: 12, spellID: 20162, manaCost: 40, scaleLevel: 20, proc: proc{spellID: 21082, value: 59, scale: 1.1}, judge: judge{spellID: 20188, minDamage: 35}},
		{level: 22, spellID: 20305, manaCost: 65, scaleLevel: 30, proc: proc{spellID: 21082, value: 108, scale: 1.7}, judge: judge{spellID: 20300, minDamage: 58}},
		{level: 32, spellID: 20306, manaCost: 90, scaleLevel: 40, proc: proc{spellID: 21082, value: 167, scale: 2.0}, judge: judge{spellID: 20301, minDamage: 92}},
		{level: 42, spellID: 20307, manaCost: 125, scaleLevel: 50, proc: proc{spellID: 21082, value: 254, scale: 2.2}, judge: judge{spellID: 20302, minDamage: 127}},
		{level: 52, spellID: 20308, manaCost: 160, scaleLevel: 60, proc: proc{spellID: 21082, value: 352, scale: 2.4}, judge: judge{spellID: 20303, minDamage: 161}},
		{level: 61, spellID: 27158, manaCost: 210, scaleLevel: 69, proc: proc{spellID: 21082, value: 474, scale: 2.6}, judge: judge{spellID: 27159, minDamage: 219}},
	}

	paladin.SealOfTheCrusader = make([]*core.Spell, len(ranks)+1)
	paladin.SealOfTheCrusaderJudgements = make([]*core.Spell, len(ranks)+1)
	paladin.SealOfTheCrusaderAuras = make([]*core.Aura, len(ranks)+1)

	for rank := 1; rank < len(ranks); rank++ {
		if paladin.Level < ranks[rank].level {
			break
		}

		registerJotCDebuff := func(target *core.Unit) *core.Aura {
			return target.GetOrRegisterAura(core.Aura{
				Label:    "Judgement of the Crusader" + paladin.Label + strconv.Itoa(rank),
				ActionID: core.ActionID{SpellID: ranks[rank].judge.spellID},
				Tag:      JudgementAuraTag,
				Duration: time.Second * 20,

				// TODO: Implement when Holy damage taken is added
				OnGain: func(aura *core.Aura, sim *core.Simulation) {
					// aura.Unit.PseudoStats.SchoolBonusDamageTaken[stats.SchoolIndexHoly] += bonus
				},

				OnExpire: func(aura *core.Aura, sim *core.Simulation) {
					// aura.Unit.PseudoStats.SchoolBonusDamageTaken[stats.SchoolIndexHoly] -= bonus
				},

				OnSpellHitTaken: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
					if result.Landed() && spell.ProcMask.Matches(core.ProcMaskMeleeWhiteHit) {
						aura.Refresh(sim)
					}
				},
			})
		}

		debuffs := paladin.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
			return registerJotCDebuff(target)
		})

		judgeSpell := paladin.RegisterSpell(core.SpellConfig{
			ActionID:       core.ActionID{SpellID: ranks[rank].judge.spellID},
			SpellSchool:    core.SpellSchoolHoly,
			ProcMask:       core.ProcMaskEmpty,
			Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagBinary,
			ClassSpellMask: SpellMaskJudgementOfTheCrusader,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.CalcAndDealOutcome(sim, target, spell.OutcomeAlwaysHit)
				debuffs.Get(target).Activate(sim)
			},
		})

		baseAp := float64(ranks[rank].proc.value)
		scalingAp := float64(ranks[rank].proc.scale)
		maximumScalingLevel := ranks[rank].scaleLevel
		minimumScalingLevel := ranks[rank].level
		meleeAp := baseAp + scalingAp*float64(min(paladin.Level, maximumScalingLevel)-minimumScalingLevel)

		aura := paladin.RegisterAura(core.Aura{
			Label:    "Seal of the Crusader" + paladin.Label + strconv.Itoa(rank),
			ActionID: core.ActionID{SpellID: ranks[rank].spellID},
			Duration: time.Second * 30,
		}).AttachMultiplyMeleeSpeed(1.4).
		AttachMultiplicativePseudoStatBuff(&paladin.AutoAttacks.MHAuto().DamageMultiplier, 1/1.4).
		AttachStatBuff(stats.AttackPower, meleeAp)

		paladin.SealOfTheCrusaderAuras = append(paladin.SealOfTheCrusaderAuras, aura)

		paladin.SealOfTheCrusader[rank] = paladin.RegisterSpell(core.SpellConfig{
			ActionID:    aura.ActionID,
			SpellSchool: core.SpellSchoolHoly,
			Flags:       core.SpellFlagAPL,

			ManaCost: core.ManaCostOptions{
				BaseCostPercent: ranks[rank].manaCost,
			},
			Cast: core.CastConfig{
				DefaultCast: core.Cast{
					GCD: core.GCDDefault,
				},
			},

			ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
				paladin.applySeal(aura, judgeSpell, sim)
			},
		})

		paladin.SealOfTheCrusaderJudgements[rank] = judgeSpell
	}
}

// Seal of Blood (Horde only)
// https://www.wowhead.com/tbc/spell=31892
//
// All melee attacks deal additional Holy damage, but the Paladin loses
// health equal to 10% of the total damage inflicted. Lasts 30 sec.
// Only one Seal can be active on the Paladin at any one time.
//
// Unleashing this Seal's energy will judge an enemy, instantly causing
// Holy damage at the cost of health equal to 33% of the damage caused.
func (paladin *Paladin) registerSealOfBlood() {
	var ranks = []seal{
		{}, // Dummy to offset the index
		{level: 64, spellID: 31892, manaCost: 210, proc: proc{spellID: 31893, value: 0.35}, judge: judge{spellID: 31898, minDamage: 295, maxDamage: 325, coeff: 0.429}},
	}

	for rank := 1; rank < len(ranks); rank++ {
		if paladin.Level < ranks[rank].level {
			break
		}

		judgeOfBloodHealthMetric := paladin.NewHealthMetrics(core.ActionID{SpellID: ranks[rank].judge.spellID})
		judgeSpell := paladin.RegisterSpell(core.SpellConfig{
			ActionID:       core.ActionID{SpellID: ranks[rank].judge.spellID},
			SpellSchool:    core.SpellSchoolHoly,
			ProcMask:       core.ProcMaskEmpty,
			Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagBinary,
			ClassSpellMask: SpellMaskJudgementOfBlood,

			BonusCoefficient: ranks[rank].judge.coeff,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				flags := spell.Flags
				baseDamage := sim.Roll(ranks[rank].judge.minDamage, ranks[rank].judge.maxDamage)
				result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialCritOnly)

				action := core.NewDelayedAction(core.DelayedActionOptions{
					DoAt:     sim.CurrentTime + core.SpellBatchWindow,
					Priority: core.ActionPriorityLow,
					OnAction: func(sim *core.Simulation) {
						currentFlags := spell.Flags
						spell.Flags = flags
						spell.DealDamage(sim, result)
						paladin.GainHealth(sim, -result.Damage*0.33, judgeOfBloodHealthMetric)
						spell.Flags = currentFlags
					},
				})

				sim.AddPendingAction(action)
			},
		})

		sealOfBloodHealthMetric := paladin.NewHealthMetrics(core.ActionID{SpellID: ranks[rank].proc.spellID})
		procSpell := paladin.RegisterSpell(core.SpellConfig{
			ActionID:       core.ActionID{SpellID: ranks[rank].proc.spellID},
			SpellSchool:    core.SpellSchoolHoly,
			ProcMask:       core.ProcMaskMeleeMHSpecial, //changed to ProcMaskMeleeMHSpecial, to allow procs from weapons/oils which do proc from SoR, -- TODO: Verify in TBC
			Flags:          core.SpellFlagMeleeMetrics,  // | core.SpellFlagSuppressEquipProcs | core.SpellFlagBatchStartAttackMacro, // but Wild Strikes does not proc, nor equip procs
			ClassSpellMask: SpellMaskSealOfBlood,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				baseDamage := spell.Unit.MHWeaponDamage(sim, spell.MeleeAttackPower()) * float64(ranks[rank].proc.value)
				result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)

				action := core.NewDelayedAction(core.DelayedActionOptions{
					DoAt:     sim.CurrentTime + core.SpellBatchWindow,
					Priority: core.ActionPriorityLow,
					OnAction: func(sim *core.Simulation) {
						spell.DealDamage(sim, result)
						paladin.GainHealth(sim, -result.Damage*0.1, sealOfBloodHealthMetric)
					},
				})

				sim.AddPendingAction(action)
			},
		})

		aura := paladin.RegisterAura(core.Aura{
			Label:    "Seal of Blood" + paladin.Label + strconv.Itoa(rank),
			ActionID: core.ActionID{SpellID: ranks[rank].spellID},
			Duration: time.Second * 30,

			OnSpellHitDealt: func(_ *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				if !result.Landed() {
					return
				}
				if spell.ProcMask.Matches(core.ProcMaskMeleeWhiteHit) {
					procSpell.Cast(sim, result.Target)
				}
			},
		})

		sealSpell := paladin.RegisterSpell(core.SpellConfig{
			ActionID:    aura.ActionID,
			SpellSchool: core.SpellSchoolHoly,
			Flags:       core.SpellFlagAPL,

			ManaCost: core.ManaCostOptions{
				FlatCost:        int32(ranks[rank].manaCost),
			},
			Cast: core.CastConfig{
				DefaultCast: core.Cast{
					GCD: core.GCDDefault,
				},
			},

			ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
				paladin.applySeal(aura, judgeSpell, sim)
			},
		})

		paladin.SealOfBlood = append(paladin.SealOfBlood, sealSpell)
		paladin.SealOfBloodJudgements = append(paladin.SealOfBloodJudgements, judgeSpell)
		paladin.SealOfBloodAuras = append(paladin.SealOfBloodAuras, aura)
	}
}

// Seal of Vengeance (Alliance only)
// https://www.wowhead.com/tbc/spell=31801
//
// Fills the Paladin with holy power, causing attacks to apply a Holy DoT
// effect for 15 seconds. The DoT can stack up to 5 times. Once stacked to
// 5 times, each of the Paladin's attacks also deals additional Holy damage.
// Lasts 30 sec. Only one Seal can be active on the Paladin at any one time.
//
// Unleashing this Seal's energy will judge an enemy, instantly causing
// Holy damage per application of Holy Vengeance on the target.
func (paladin *Paladin) registerSealOfVengeance() {
	var ranks = []seal{
		{}, // Dummy to offset the index
		{level: 64, spellID: 31801, manaCost: 250, proc: proc{spellID: 31803, value: 30, coeff: 0.034}, judge: judge{spellID: 31804, minDamage: 120, coeff: 0.429}},
	}

	for rank := 1; rank < len(ranks); rank++ {
		if paladin.Level < ranks[rank].level {
			break
		}

		holyVengeanceTag := "Holy Vengeance"

		judgeSpell := paladin.RegisterSpell(core.SpellConfig{
			ActionID:       core.ActionID{SpellID: ranks[rank].judge.spellID},
			SpellSchool:    core.SpellSchoolHoly,
			ProcMask:       core.ProcMaskEmpty,
			Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagBinary,
			ClassSpellMask: SpellMaskJudgementOfVengeance,

			BonusCoefficient: ranks[rank].judge.coeff,

			ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
				return target.GetActiveAuraWithTag(holyVengeanceTag) != nil
			},

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				damage := float64(target.GetActiveAuraWithTag(holyVengeanceTag).GetStacks()) * ranks[rank].judge.minDamage
				result := spell.CalcDamage(sim, target, damage, spell.OutcomeMeleeSpecialCritOnly)

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
			ActionID:       core.ActionID{SpellID: ranks[rank].proc.spellID},
			SpellSchool:    core.SpellSchoolHoly,
			ProcMask:       core.ProcMaskEmpty,
			Flags:          core.SpellFlagHelpful,
			ClassSpellMask: SpellMaskSealOfVengeance,

			Dot: core.DotConfig{
				Aura: core.Aura{
					Label:     "Holy Vengeance" + paladin.Label + strconv.Itoa(rank),
					Tag:       holyVengeanceTag,
					ActionID:  core.ActionID{SpellID: ranks[rank].proc.spellID},
					Duration:  time.Second * 15,
					MaxStacks: 5,
				},
				NumberOfTicks:    5,
				TickLength:       time.Second * 3,
				BonusCoefficient: ranks[rank].proc.coeff,
				OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
					dot.Snapshot(target, ranks[rank].proc.value)
				},
				OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
					dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
				},
			},

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.Dot(target).Apply(sim)
			},
		})

		aura := paladin.RegisterAura(core.Aura{
			Label:    "Seal of Vengeance" + paladin.Label + strconv.Itoa(rank),
			ActionID: core.ActionID{SpellID: ranks[rank].spellID},
			Duration: time.Second * 30,

			OnSpellHitDealt: func(_ *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				if !result.Landed() {
					return
				}
				if spell.ProcMask.Matches(core.ProcMaskMeleeWhiteHit) {
					procSpell.Cast(sim, spell.Unit)
				}
			},
		})

		sealSpell := paladin.RegisterSpell(core.SpellConfig{
			ActionID:    aura.ActionID,
			SpellSchool: core.SpellSchoolHoly,
			Flags:       core.SpellFlagAPL,

			ManaCost: core.ManaCostOptions{
				FlatCost:        int32(ranks[rank].manaCost),
			},
			Cast: core.CastConfig{
				DefaultCast: core.Cast{
					GCD: core.GCDDefault,
				},
			},

			ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
				paladin.applySeal(aura, judgeSpell, sim)
			},
		})

		paladin.SealOfVengeance = append(paladin.SealOfVengeance, sealSpell)
		paladin.SealOfVengeanceJudgements = append(paladin.SealOfVengeanceJudgements, judgeSpell)
		paladin.SealOfVengeanceAuras = append(paladin.SealOfVengeanceAuras, aura)
	}
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
func (paladin *Paladin) registerSealOfCommand() {
	var ranks = []seal{
		{}, // Dummy to offset the index
		{level: 20, spellID: 20375, scaleLevel: 28, manaCost: 65, proc: proc{spellID: 20424, value: 0.70, coeff: 0.29}, judge: judge{spellID: 20425, minDamage: 46, maxDamage: 50, scale: 2.8, coeff: 0.429}},
		{level: 30, spellID: 20915, scaleLevel: 38, manaCost: 110, proc: proc{spellID: 20424, value: 0.70, coeff: 0.29}, judge: judge{spellID: 20962, minDamage: 73, maxDamage: 80, scale: 3.05, coeff: 0.429}},
		{level: 40, spellID: 20918, scaleLevel: 48, manaCost: 140, proc: proc{spellID: 20424, value: 0.70, coeff: 0.29}, judge: judge{spellID: 20961, minDamage: 102, maxDamage: 112, scale: 2.8, coeff: 0.429}},
		{level: 50, spellID: 20919, scaleLevel: 58, manaCost: 180, proc: proc{spellID: 20424, value: 0.70, coeff: 0.29}, judge: judge{spellID: 20967, minDamage: 130, maxDamage: 143, scale: 3.05, coeff: 0.429}},
		{level: 60, spellID: 20920, scaleLevel: 68, manaCost: 210, proc: proc{spellID: 20424, value: 0.70, coeff: 0.29}, judge: judge{spellID: 20968, minDamage: 169, maxDamage: 186, scale: 3.05, coeff: 0.429}},
		{level: 70, spellID: 27170, scaleLevel: 78, manaCost: 280, proc: proc{spellID: 20424, value: 0.70, coeff: 0.29}, judge: judge{spellID: 27172, minDamage: 228, maxDamage: 252, scale: 3.05, coeff: 0.429}},
	}

	for rank := 1; rank < len(ranks); rank++ {
		if paladin.Level < ranks[rank].level {
			break
		}

		minDamage := ranks[rank].judge.minDamage + ranks[rank].judge.scale*float64(min(paladin.Level, ranks[rank].scaleLevel)-ranks[rank].level)
		maxDamage := ranks[rank].judge.maxDamage + ranks[rank].judge.scale*float64(min(paladin.Level, ranks[rank].scaleLevel)-ranks[rank].level)

		judgeSpell := paladin.RegisterSpell(core.SpellConfig{
			ActionID:       core.ActionID{SpellID: ranks[rank].judge.spellID},
			SpellSchool:    core.SpellSchoolHoly,
			ProcMask:       core.ProcMaskEmpty,
			Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagBinary,
			ClassSpellMask: SpellMaskJudgementOfCommand,

			BonusCoefficient: ranks[rank].judge.coeff,

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
			ActionID:       core.ActionID{SpellID: ranks[rank].proc.spellID},
			SpellSchool:    core.SpellSchoolHoly,
			ProcMask:       core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeProc,
			Flags:          core.SpellFlagMeleeMetrics,
			ClassSpellMask: SpellMaskSealOfCommand,

			BonusCoefficient: ranks[rank].proc.coeff,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				baseDamage := spell.Unit.MHWeaponDamage(sim, spell.MeleeAttackPower()) * ranks[rank].proc.value
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
			Label:    "Seal of Command" + paladin.Label + strconv.Itoa(rank),
			ActionID: core.ActionID{SpellID: ranks[rank].spellID},
			Duration: time.Second * 30,
			Tag:      TwistTag,
		}).AttachProcTrigger(core.ProcTrigger{
			ProcMask: core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeProc,
			ICD:      time.Second * 1,
			DPM:      paladin.NewLegacyPPMManager(7, core.ProcMaskMeleeWhiteHit),
			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				procSpell.Cast(sim, result.Target)
			},
		})

		sealSpell := paladin.RegisterSpell(core.SpellConfig{
			ActionID:    aura.ActionID,
			SpellSchool: core.SpellSchoolHoly,
			ProcMask:    core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeProc,
			Flags:       core.SpellFlagAPL,

			ManaCost: core.ManaCostOptions{
				FlatCost:        int32(ranks[rank].manaCost),
			},
			Cast: core.CastConfig{
				DefaultCast: core.Cast{
					GCD: core.GCDDefault,
				},
			},

			ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
				paladin.applySeal(aura, judgeSpell, sim)
			},
		})

		paladin.SealOfCommand = append(paladin.SealOfCommand, sealSpell)
		paladin.SealOfCommandJudgements = append(paladin.SealOfCommandJudgements, judgeSpell)
		paladin.SealOfCommandAuras = append(paladin.SealOfCommandAuras, aura)
	}
}
