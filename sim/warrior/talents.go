package warrior

func (war *Warrior) ApplyTalents() {
	war.registerArmsTalents()
	war.registerFuryTalents()
	war.registerProtectionTalents()
}

// func (war *Warrior) registerJuggernaut() {
// 	if !war.Talents.Juggernaut {
// 		return
// 	}

// 	war.AddStaticMod(core.SpellModConfig{
// 		ClassMask: SpellMaskCharge,
// 		Kind:      core.SpellMod_Cooldown_Flat,
// 		TimeValue: -8 * time.Second,
// 	})
// }

// func (war *Warrior) registerImpendingVictory() {
// 	if !war.Talents.ImpendingVictory {
// 		return
// 	}

// 	actionID := core.ActionID{SpellID: 103840}
// 	healthMetrics := war.NewHealthMetrics(actionID)

// 	war.RegisterSpell(core.SpellConfig{
// 		ActionID:       actionID,
// 		SpellSchool:    core.SpellSchoolPhysical,
// 		ProcMask:       core.ProcMaskMeleeMHSpecial,
// 		Flags:          core.SpellFlagAPL | core.SpellFlagMeleeMetrics,
// 		ClassSpellMask: SpellMaskImpendingVictory,

// 		RageCost: core.RageCostOptions{
// 			Cost:   10,
// 			Refund: 0.8,
// 		},
// 		Cast: core.CastConfig{
// 			DefaultCast: core.Cast{
// 				GCD: core.GCDDefault,
// 			},
// 			IgnoreHaste: true,
// 			CD: core.Cooldown{
// 				Timer:    war.NewTimer(),
// 				Duration: time.Second * 30,
// 			},
// 		},

// 		DamageMultiplier: 1,
// 		CritMultiplier:   war.DefaultCritMultiplier(),

// 		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
// 			war.VictoryRushAura.Deactivate(sim)

// 			baseDamage := 56 + spell.MeleeAttackPower()*0.56
// 			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)

// 			healthMultiplier := core.TernaryFloat64(war.T15Tank2P != nil && war.T15Tank2P.IsActive(), 0.4, 0.2)

// 			if result.Landed() {
// 				war.GainHealth(sim, war.MaxHealth()*healthMultiplier, healthMetrics)
// 			} else {
// 				spell.IssueRefund(sim)
// 			}
// 		},
// 	})
// }

// func (war *Warrior) registerBladestorm() {
// 	if !war.Talents.Bladestorm {
// 		return
// 	}

// 	actionID := core.ActionID{SpellID: 46924}

// 	damageMultiplier := 1.2
// 	if war.Spec == proto.Spec_SpecArmsWarrior {
// 		damageMultiplier += 0.6
// 	} else if war.Spec == proto.Spec_SpecProtectionWarrior {
// 		damageMultiplier *= 1.33
// 	}

// 	mhSpell := war.RegisterSpell(core.SpellConfig{
// 		ActionID:       actionID.WithTag(1), // Real Spell ID: 50622
// 		SpellSchool:    core.SpellSchoolPhysical,
// 		ClassSpellMask: SpellMaskBladestormMH,
// 		ProcMask:       core.ProcMaskMeleeMHSpecial,
// 		Flags:          core.SpellFlagPassiveSpell,

// 		DamageMultiplier: damageMultiplier,
// 		CritMultiplier:   war.DefaultCritMultiplier(),

// 		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
// 			results := spell.CalcAoeDamageWithVariance(sim, spell.OutcomeMeleeWeaponSpecialHitAndCrit, func(sim *core.Simulation, spell *core.Spell) float64 {
// 				return spell.Unit.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower())
// 			})

// 			war.CastNormalizedSweepingStrikesAttack(results, sim)
// 			spell.DealBatchedAoeDamage(sim)
// 		},
// 	})

// 	ohSpell := war.RegisterSpell(core.SpellConfig{
// 		ActionID:       actionID.WithTag(2), // Real Spell ID: 95738,
// 		SpellSchool:    core.SpellSchoolPhysical,
// 		ClassSpellMask: SpellMaskBladestormOH,
// 		ProcMask:       core.ProcMaskMeleeOHSpecial,
// 		Flags:          core.SpellFlagPassiveSpell,

// 		DamageMultiplier: damageMultiplier,
// 		CritMultiplier:   war.DefaultCritMultiplier(),

// 		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
// 			spell.CalcAndDealAoeDamageWithVariance(sim, spell.OutcomeMeleeWeaponSpecialHitAndCrit, func(sim *core.Simulation, spell *core.Spell) float64 {
// 				return spell.Unit.OHNormalizedWeaponDamage(sim, spell.MeleeAttackPower())
// 			})
// 		},
// 	})

// 	war.AddStaticMod(core.SpellModConfig{
// 		ClassMask: SpellMaskBattleShout | SpellMaskCommandingShout | SpellMaskRallyingCry | SpellMaskLastStand | SpellMaskDemoralizingShout | SpellMaskBerserkerRage,
// 		Kind:      core.SpellMod_AllowCastWhileChanneling,
// 	})

// 	flags := core.SpellFlagChanneled | core.SpellFlagMeleeMetrics | core.SpellFlagAPL | core.SpellFlagCastWhileChanneling
// 	if war.Spec != proto.Spec_SpecProtectionWarrior {
// 		flags |= core.SpellFlagReadinessTrinket
// 	}

// 	spell := war.RegisterSpell(core.SpellConfig{
// 		ActionID:       actionID.WithTag(0),
// 		SpellSchool:    core.SpellSchoolPhysical,
// 		ClassSpellMask: SpellMaskBladestorm,
// 		Flags:          flags,
// 		ProcMask:       core.ProcMaskEmpty,

// 		Cast: core.CastConfig{
// 			DefaultCast: core.Cast{
// 				GCD: core.GCDDefault,
// 			},
// 			IgnoreHaste: true,
// 			CD: core.Cooldown{
// 				Timer:    war.NewTimer(),
// 				Duration: time.Minute * 1,
// 			},
// 		},

// 		DamageMultiplier: 1.0,
// 		CritMultiplier:   war.DefaultCritMultiplier(),

// 		Dot: core.DotConfig{
// 			IsAOE: true,
// 			Aura: core.Aura{
// 				Label: "Bladestorm",
// 				OnExpire: func(aura *core.Aura, sim *core.Simulation) {
// 					war.ExtendGCDUntil(sim, sim.CurrentTime+war.ReactionTime)
// 				},
// 			},
// 			NumberOfTicks: 6,
// 			TickLength:    time.Second * 1,
// 			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
// 				mhSpell.Cast(sim, target)

// 				if war.OffHand() != nil && (war.OffHand().WeaponType != proto.WeaponType_WeaponTypeUnknown && war.OffHand().WeaponType != proto.WeaponType_WeaponTypeShield) {
// 					ohSpell.Cast(sim, target)
// 				}
// 			},
// 		},
// 		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
// 			dot := spell.AOEDot()
// 			dot.Apply(sim)
// 			dot.TickOnce(sim)
// 		},
// 	})

// 	war.AddMajorCooldown(core.MajorCooldown{
// 		Spell: spell,
// 		Type:  core.CooldownTypeDPS,
// 	})
// }
