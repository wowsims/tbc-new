package magtheridonslair

import (
	"fmt"
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

const magtheridonMeleeDamageSpread = 0.413
const magtheridonID int32 = 17257

func addMagtheridon(raidPrefix string) {
	createMagtheridonPreset(raidPrefix, 25, 4_818_380, 14_603)
}

func createMagtheridonPreset(raidPrefix string, raidSize int32, bossHealth float64, bossMinBaseDamage float64) {
	bossName := fmt.Sprintf("Magtheridon %d", raidSize)

	core.AddPresetTarget(&core.PresetTarget{
		PathPrefix: raidPrefix,

		Config: &proto.Target{
			Id:        magtheridonID,
			Name:      bossName,
			Level:     73,
			MobType:   proto.MobType_MobTypeDemon,
			TankIndex: 0,

			Stats: stats.Stats{
				stats.Health:      bossHealth,
				stats.Armor:       7685,
				stats.AttackPower: 320,
			}.ToProtoArray(),

			SpellSchool:   proto.SpellSchool_SpellSchoolPhysical,
			SwingSpeed:    2,
			MinBaseDamage: bossMinBaseDamage,
			DamageSpread:  magtheridonMeleeDamageSpread,

			ParryHaste: true,
		},

		AI: makeMagtheridonAI(),
	})

	core.AddPresetEncounter(bossName, []string{
		raidPrefix + "/" + bossName,
	})
}

func makeMagtheridonAI() core.AIFactory {
	return func() core.TargetAI {
		return &MagtheridonAI{}
	}
}

type MagtheridonAI struct {
	// Unit references
	Target     *core.Target
	BossUnit   *core.Unit
	MainTank   *core.Unit
	OffTank    *core.Unit
	ValidTanks []*core.Unit

	// Spell + aura references
	Cleave *core.Spell
}

func (ai *MagtheridonAI) Initialize(target *core.Target, config *proto.Target) {
	// Save unit references
	ai.Target = target
	ai.Target.AutoAttacks.MHConfig().ActionID.Tag = magtheridonID

	ai.BossUnit = &target.Unit

	ai.MainTank = ai.BossUnit.CurrentTarget

	ai.ValidTanks = core.FilterSlice([]*core.Unit{ai.MainTank, ai.OffTank}, func(unit *core.Unit) bool {
		return unit != nil
	})

	// Register relevant spells and auras
	ai.registerCleave()

	ai.BossUnit.AutoAttacks.SetReplaceMHSwing(ai.TryCleave)
}

func (ai *MagtheridonAI) registerCleave() {
	ai.Cleave = ai.BossUnit.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{SpellID: 30619},
		SpellSchool:      core.SpellSchoolPhysical,
		ProcMask:         core.ProcMaskMeleeMHSpecial,
		Flags:            core.SpellFlagMeleeMetrics,
		DamageMultiplier: 1.5,

		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    ai.BossUnit.NewTimer(),
				Duration: time.Second * 10,
			},

			IgnoreHaste: true,
		},

		ApplyEffects: func(sim *core.Simulation, tankTarget *core.Unit, spell *core.Spell) {
			baseDamage := spell.Unit.AutoAttacks.MH().EnemyWeaponDamage(sim, spell.MeleeAttackPower(), magtheridonMeleeDamageSpread)

			spell.CalcAndDealDamage(sim, tankTarget, baseDamage, spell.OutcomeEnemyMeleeWhite)
		},
	})

	ai.BossUnit.RegisterResetEffect(func(sim *core.Simulation) {
		ai.Cleave.CD.Set(core.DurationFromSeconds(sim.RandomFloat("Cleave Timing") * ai.Cleave.CD.Duration.Seconds()))
	})
}

func (ai *MagtheridonAI) Reset(sim *core.Simulation) {
	// Randomize GCD and swing timings to prevent fake APL-Haste couplings.
	ai.Target.Enable(sim)
}

func (ai *MagtheridonAI) ExecuteCustomRotation(sim *core.Simulation) {
	target := ai.Target.CurrentTarget
	if target == nil {
		// For individual non tank sims we still want abilities to work
		target = &ai.Target.Env.Raid.Parties[0].Players[0].GetCharacter().Unit
	}

	ai.Target.ExtendGCDUntil(sim, sim.CurrentTime+core.BossGCD)
}

func (ai *MagtheridonAI) TryCleave(sim *core.Simulation, mhSwingSpell *core.Spell) *core.Spell {
	if !ai.Cleave.CanCast(sim, ai.MainTank) {
		return mhSwingSpell
	}

	return ai.Cleave
}
