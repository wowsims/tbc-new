
func (war *Warrior) registerBerserkerRage() {
	actionID := core.ActionID{SpellID: 18499}
	rageMetrics := war.NewRageMetrics(actionID)
	instantRage := 5 * float64(war.Talents.ImprovedBerserkerRage)

	war.BerserkerRageAura = war.RegisterAura(core.Aura{
		Label:    "Berserker Rage",
		ActionID: actionID,
		Duration: time.Second * 10,
	})

	spell := war.RegisterSpell(core.SpellConfig{
		ActionID: actionID,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    war.NewTimer(),
				Duration: time.Second * 30,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return war.StanceMatches(BerserkerStance)
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			if instantRage > 0 {
				war.AddRage(sim, instantRage, rageMetrics)
			}
			war.BerserkerRageAura.Activate(sim)
		},
	})

	war.AddMajorCooldown(core.MajorCooldown{
		Spell: spell,
		Type:  core.CooldownTypeSurvival,
	})
}
