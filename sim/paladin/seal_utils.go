package paladin

import "github.com/wowsims/tbc/sim/core"

type sealRankMap []seal
type sealRankFactory func(seal)

func (ranks sealRankMap) RegisterAll(factory sealRankFactory) {
	for rank := 1; rank < len(ranks); rank++ {
		rankConfig := ranks[rank]
		rankConfig.rank = int32(rank)
		factory(rankConfig)
	}
}

func (paladin *Paladin) addSealRank(
	seals *[]*core.Spell,
	judgements *[]*core.Spell,
	auras *[]*core.Aura,
	sealSpell *core.Spell,
	judgement *core.Spell,
	aura *core.Aura,
) {
	*seals = append(*seals, sealSpell)
	*judgements = append(*judgements, judgement)
	*auras = append(*auras, aura)
}
