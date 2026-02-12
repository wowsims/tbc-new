package affliction

import (
	"fmt"
	"math"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

func (warlock *AfflictionWarlock) NewAPLValue(rot *core.APLRotation, config *proto.APLValue) core.APLValue {
	switch config.Value.(type) {
	case *proto.APLValue_AfflictionCurrentSnapshot:
		return warlock.newAfflictionCurrentSnapshot(rot, config.GetAfflictionCurrentSnapshot(), config.Uuid)
	default:
		return warlock.Warlock.NewAPLValue(rot, config)
	}
}

// modified snapshot tracker, designed to be affliction specific.
// checks the snapshotted magnitude of existing dots relative to baseline.
// ignores crit and haste factors, since malefic effect ignores these
type APLValueAfflictionCurrentSnapshot struct {
	core.DefaultAPLValueImpl
	warlock            *AfflictionWarlock
	spell              *core.Spell
	sbssDotRefs        []**core.Spell
	targetRef          core.UnitReference
	baseValue          float64
	baseValueDummyAura *core.Aura // Used to get the base value at encounter start
}

func (warlock *AfflictionWarlock) newAfflictionCurrentSnapshot(rot *core.APLRotation, config *proto.APLValueAfflictionCurrentSnapshot, _ *proto.UUID) *APLValueAfflictionCurrentSnapshot {
	spell := rot.GetAPLSpell(config.SpellId)
	if spell == nil {
		return nil
	}

	targetRef := rot.GetTargetUnit(config.TargetUnit)

	baseValueDummyAura := core.MakePermanent(warlock.GetOrRegisterAura(core.Aura{
		Label:    "Dummy Aura - APL Current Snapshot Base Value",
		Duration: core.NeverExpires,
	}))

	return &APLValueAfflictionCurrentSnapshot{
		warlock:            warlock,
		spell:              spell,
		targetRef:          targetRef,
		baseValueDummyAura: baseValueDummyAura,
	}
}

func (value *APLValueAfflictionCurrentSnapshot) Finalize(rot *core.APLRotation) {
	value.sbssDotRefs = []**core.Spell{&value.warlock.Agony, &value.warlock.Corruption, &value.warlock.UnstableAffliction}
	if value.baseValueDummyAura != nil {
		value.baseValueDummyAura.ApplyOnInit(func(aura *core.Aura, sim *core.Simulation) {
			// Soulburn: Soul Swap
			if value.spell.ActionID.SpellID == 86121 && value.spell.ActionID.Tag == 1 {
				total := 0.0
				target := value.targetRef.Get()

				for _, spellRef := range value.sbssDotRefs {
					spell := (*spellRef)
					total += (spell.ExpectedTickDamage(sim, target) * spell.Dot(target).CalcTickPeriod().Seconds()) / (1 + (spell.SpellCritChance(target) * (spell.CritDamageMultiplier() - 1)))
				}
				value.baseValue = total

			} else {
				target := value.targetRef.Get()
				value.baseValue = value.spell.ExpectedTickDamage(sim, target) * value.spell.Dot(target).CalcTickPeriod().Seconds()
				value.baseValue /= (1 + (value.spell.SpellCritChance(target) * (value.spell.CritDamageMultiplier() - 1)))
			}
		})
	}
}

func (value *APLValueAfflictionCurrentSnapshot) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeFloat
}

func (value *APLValueAfflictionCurrentSnapshot) String() string {
	return fmt.Sprintf("Current Snapshot on %s", value.spell.ActionID)
}

func (value *APLValueAfflictionCurrentSnapshot) GetFloat(sim *core.Simulation) float64 {
	target := value.targetRef.Get()
	snapshotDamage := 0.0
	//Soulburn: Soul Swap
	if value.spell.ActionID.SpellID == 86121 && value.spell.ActionID.Tag == 1 {
		target := value.targetRef.Get()

		for _, spellRef := range value.sbssDotRefs {
			dot := (*spellRef).Dot(target)

			snapshotDamage += (dot.Spell.ExpectedTickDamageFromCurrentSnapshot(sim, target) * dot.TickPeriod().Seconds()) / (1 + (dot.SnapshotCritChance * (dot.Spell.CritDamageMultiplier() - 1)))
		}
	} else {
		dot := value.spell.Dot(target)
		snapshotDamage = (value.spell.ExpectedTickDamageFromCurrentSnapshot(sim, target) * dot.TickPeriod().Seconds()) / (1 + (dot.SnapshotCritChance * (dot.Spell.CritDamageMultiplier() - 1)))
	}

	if snapshotDamage == 0 {
		return -1
	}

	// Rounding this to effectively 3 decimal places as a percentage to avoid floating point errors
	return math.Round((snapshotDamage/value.baseValue)*100000)/100000 - 1
}
