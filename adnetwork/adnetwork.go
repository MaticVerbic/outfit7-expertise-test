package adnetwork

import "github.com/pquerna/ffjson/ffjson"

// SDK represents an ad provider.
type SDK struct {
	Provider string  `json:"provider"`
	Score    float64 `json:"score"`
	// To add other necessary fields add them here
}

// ScoreSorter implements sorter interface.
type ScoreSorter []*SDK

func (s ScoreSorter) Len() int      { return len(s) }
func (s ScoreSorter) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// Reverse the condition in Less to 'greater' to reverse order of sorting(left < right => left > right).
func (s ScoreSorter) Less(i, j int) bool { return s[i].Score > s[j].Score }

// AdNetwork represents a network od SDKs.
// Each AdNetwork is assigned to a country and is divided by each type of the ad.
type AdNetwork struct {
	Banner       []*SDK `json:"banner"`
	Interstitial []*SDK `json:"interstitial"`
	Video        []*SDK `json:"video"`
	Country      string `json:"country"`
}

// MarshalBinary satisfies encoding.BinaryMarshaler interface.
func (an *AdNetwork) MarshalBinary() (data []byte, err error) {
	return ffjson.Marshal(an)
}

// UnmarshalBinary satisfies encoding.BinaryUnmarshaler interface.
func (an *AdNetwork) UnmarshalBinary(data []byte) (err error) {
	return ffjson.Unmarshal(data, an)
}

// ContainsAllProviders returns true if all providers are present in specified slice.
func (an *AdNetwork) ContainsAllProviders(adType string, providers []string) bool {
	switch adType {
	case "banner":
		return containsAll(an.Banner, providers)
	case "interstitial":
		return containsAll(an.Interstitial, providers)
	case "video":
		return containsAll(an.Video, providers)
	}

	return false
}

// ContainsAnyProviders returns true if at least one provider is present in slice.
func (an *AdNetwork) ContainsAnyProviders(adType string, providers []string) []string {
	switch adType {
	case "banner":
		return containsAny(an.Banner, providers)
	case "interstitial":
		return containsAny(an.Interstitial, providers)
	case "video":
		return containsAny(an.Video, providers)
	}

	return []string{}
}

func containsAll(arr []*SDK, providers []string) bool {
	for _, key := range providers {
		found := false
		for _, sdk := range arr {
			if sdk.Provider == key {
				found = true
				break
			}
		}

		if !found {
			return false
		}
	}
	return true
}

// ContainsAny returns any provides from providers input contained in the list, in order of providers argument.
func containsAny(arr []*SDK, providers []string) []string {
	out := []string{}
	for _, key := range providers {
		for _, sdk := range arr {
			if sdk.Provider == key {
				out = append(out, key)
			}
		}
	}

	return out
}
