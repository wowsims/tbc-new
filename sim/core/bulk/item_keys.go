package bulk

import (
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

type bulkSimCandidateOption struct {
	spec *proto.ItemSpec
	item core.Item
}

type itemSpecCacheKey struct {
	id           int32
	randomSuffix int32
}

// itemSpecFingerprintKey is a zero-allocation alternative to the string fingerprint.
// gemsHash is a position-weighted XOR of gem IDs; collision probability is negligible
// for the small, fixed set of gem IDs used in practice.
type itemSpecFingerprintKey struct {
	id           int32
	randomSuffix int32
	enchant      int32
	gemsHash     uint64
}

func buildItemSpecFingerprintKey(item *proto.ItemSpec) itemSpecFingerprintKey {
	if item == nil {
		return itemSpecFingerprintKey{}
	}
	var gemsHash uint64
	for i, gem := range item.GetGems() {
		// Knuth multiplicative hash per position to make order matter.
		gemsHash ^= uint64(uint32(gem)*2654435761) << (uint(i) & 63)
	}
	return itemSpecFingerprintKey{
		id:           item.GetId(),
		randomSuffix: item.GetRandomSuffix(),
		enchant:      item.GetEnchant(),
		gemsHash:     gemsHash,
	}
}

func dedupeCandidateOptions(options []bulkSimCandidateOption) []bulkSimCandidateOption {
	if len(options) <= 1 {
		return options
	}

	seen := make(map[itemSpecCacheKey]struct{}, len(options))
	deduped := make([]bulkSimCandidateOption, 0, len(options))
	for _, option := range options {
		key := buildItemSpecKey(option.spec)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		deduped = append(deduped, option)
	}

	return deduped
}

func optionsContainEquivalent(options []bulkSimCandidateOption, target bulkSimCandidateOption) bool {
	for _, option := range options {
		if candidateOptionsEqual(option, target) {
			return true
		}
	}
	return false
}

func candidateOptionsEqual(left bulkSimCandidateOption, right bulkSimCandidateOption) bool {
	return buildItemSpecKey(left.spec) == buildItemSpecKey(right.spec)
}

func candidateOptionEqualsItem(option bulkSimCandidateOption, item core.Item) bool {
	return buildItemSpecKey(option.spec) == buildItemSpecKey(item.ToItemSpecProto())
}

func candidateOptionEqualsItemPtr(option *bulkSimCandidateOption, item *core.Item) bool {
	if option == nil || item == nil {
		return option == nil && item == nil
	}
	return candidateOptionEqualsItem(*option, *item)
}

func buildItemSpecKey(itemSpec *proto.ItemSpec) itemSpecCacheKey {
	if itemSpec == nil {
		return itemSpecCacheKey{}
	}
	return itemSpecCacheKey{
		id:           itemSpec.GetId(),
		randomSuffix: itemSpec.GetRandomSuffix(),
	}
}
