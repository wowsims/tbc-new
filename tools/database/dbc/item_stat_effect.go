package dbc

type ItemStatEffect struct {
	ID              int
	EffectIsAura    bool
	EffectPointsMin []int
	EffectPointsMax []int
	EffectArg       []int
}
