package paladin

type sealRankMap []seal
type sealRankFactory func(seal)

func (ranks sealRankMap) RegisterAll(factory sealRankFactory) {
	for rank := 1; rank < len(ranks); rank++ {
		rankConfig := ranks[rank]
		rankConfig.rank = int32(rank)
		factory(rankConfig)
	}
}
