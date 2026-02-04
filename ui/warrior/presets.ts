import { Class, ConsumesSpec, Debuffs, IndividualBuffs, PartyBuffs, RaidBuffs, TristateEffect } from '../core/proto/common';
import { defaultRaidBuffMajorDamageCooldowns } from '../core/proto_utils/utils';

export const DefaultIndividualBuffs = IndividualBuffs.create({
	blessingOfKings: true,
	blessingOfMight: TristateEffect.TristateEffectImproved,
	unleashedRage: true,
});

export const DefaultPartyBuffs = PartyBuffs.create({
	ferociousInspiration: 2,
	braidedEterniumChain: true,
	graceOfAirTotem: TristateEffect.TristateEffectImproved,
	strengthOfEarthTotem: TristateEffect.TristateEffectImproved,
	windfuryTotem: TristateEffect.TristateEffectImproved,
	battleShout: TristateEffect.TristateEffectImproved,
});

export const DefaultRaidBuffs = RaidBuffs.create({
	...defaultRaidBuffMajorDamageCooldowns(Class.ClassWarrior),
	powerWordFortitude: TristateEffect.TristateEffectImproved,
	giftOfTheWild: TristateEffect.TristateEffectImproved,
});

export const DefaultDebuffs = Debuffs.create({
	improvedSealOfTheCrusader: true,
	bloodFrenzy: true,
	giftOfArthas: true,
	mangle: true,
	exposeArmor: TristateEffect.TristateEffectImproved,
	faerieFire: TristateEffect.TristateEffectImproved,
	sunderArmor: true,
	curseOfRecklessness: true,
	huntersMark: TristateEffect.TristateEffectImproved,
});

export const DefaultConsumables = ConsumesSpec.create({
	potId: 22838,
	flaskId: 22854,
	foodId: 27658,
	conjuredId: 5512,
	explosiveId: 30217,
	superSapper: true,
	ohImbueId: 29453,
	drumsId: 351355,
	scrollAgi: true,
	scrollStr: true,
});
