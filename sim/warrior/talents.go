package warrior

func (war *Warrior) ApplyTalents() {
	war.registerArmsTalents()
	war.registerFuryTalents()
	war.registerProtectionTalents()
}
