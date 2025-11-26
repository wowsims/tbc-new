package affliction

import (
	"fmt"
	"math"
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

func (warlock *AfflictionWarlock) NewAPLValue(rot *core.APLRotation, config *proto.APLValue) core.APLValue {
	switch config.Value.(type) {
	case *proto.APLValue_WarlockHauntInFlight:
		spellInFlight := proto.APLValueSpellInFlight{
			SpellId: core.Spell{ActionID: core.ActionID{SpellID: 48181}}.ToProto(),
		}
		return rot.NewValueSpellInFlight(&spellInFlight, nil)
	case *proto.APLValue_AfflictionCurrentSnapshot:
		return warlock.newAfflictionCurrentSnapshot(rot, config.GetAfflictionCurrentSnapshot(), config.Uuid)
	case *proto.APLValue_AfflictionExhaleWindow:
		return warlock.newValueExhaleWindow(config.GetAfflictionExhaleWindow(), config.Uuid)
	default:
		return warlock.Warlock.NewAPLValue(rot, config)
	}
}

func (warlock *AfflictionWarlock) NewAPLAction(rot *core.APLRotation, config *proto.APLAction) core.APLActionImpl {
	switch config.Action.(type) {
	case *proto.APLAction_WarlockNextExhaleTarget:
		return warlock.newActionNextExhaleTarget(config.GetWarlockNextExhaleTarget())
	default:
		return nil
	}
}

type APLActionNextExhaleTarget struct {
	warlock        *AfflictionWarlock
	lastExecutedAt time.Duration
}

// Execute implements core.APLActionImpl.
func (action *APLActionNextExhaleTarget) Execute(sim *core.Simulation) {
	action.lastExecutedAt = sim.CurrentTime
	if action.warlock.CurrentTarget != action.warlock.LastInhaleTarget {
		return
	}

	nextTarget := core.NewUnitReference(&proto.UnitReference{Type: proto.UnitReference_NextTarget}, &action.warlock.Unit).Get()
	if nextTarget == nil {
		return
	}

	if sim.Log != nil {
		action.warlock.Log(sim, "Changing target to %s", nextTarget.Label)
	}

	action.warlock.CurrentTarget = nextTarget
}

func (action *APLActionNextExhaleTarget) Finalize(*core.APLRotation)         {}
func (action *APLActionNextExhaleTarget) GetAPLValues() []core.APLValue      { return nil }
func (action *APLActionNextExhaleTarget) GetInnerActions() []*core.APLAction { return nil }
func (action *APLActionNextExhaleTarget) GetNextAction(sim *core.Simulation) *core.APLAction {
	return nil
}
func (action *APLActionNextExhaleTarget) PostFinalize(*core.APLRotation) {}
func (action *APLActionNextExhaleTarget) ReResolveVariableRefs(*core.APLRotation, map[string]*proto.APLValue) {
}

func (action *APLActionNextExhaleTarget) IsReady(sim *core.Simulation) bool {
	// Prevent infinite loops by only allowing this action to be performed once at each timestamp.
	return action.lastExecutedAt != sim.CurrentTime
}

// Reset implements core.APLActionImpl.
func (action *APLActionNextExhaleTarget) Reset(sim *core.Simulation) {
	action.lastExecutedAt = core.NeverExpires
}

// String implements core.APLActionImpl.
func (action *APLActionNextExhaleTarget) String() string {
	return "Changing to Next Exhale Target"
}

func (warlock *AfflictionWarlock) newActionNextExhaleTarget(_ *proto.APLActionWarlockNextExhaleTarget) core.APLActionImpl {
	return &APLActionNextExhaleTarget{
		warlock:        warlock,
		lastExecutedAt: core.NeverExpires,
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

type APLValueExhaleWindow struct {
	core.DefaultAPLValueImpl
	warlock *AfflictionWarlock
}

func (warlock *AfflictionWarlock) newValueExhaleWindow(_ *proto.APLValueAfflictionExhaleWindow, _ *proto.UUID) core.APLValue {
	return &APLValueExhaleWindow{
		warlock: warlock,
	}
}
func (value *APLValueExhaleWindow) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeDuration
}
func (value *APLValueExhaleWindow) GetDuration(sim *core.Simulation) time.Duration {
	return time.Duration(value.warlock.ExhaleWindow)
}
func (value *APLValueExhaleWindow) String() string {
	return "Exhale Window()"
}
