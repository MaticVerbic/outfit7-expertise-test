package adnetwork

import (
	"fmt"
	"math/rand"
	"strconv"

	"github.com/pquerna/ffjson/ffjson"
)

// SDK represents an ad provider.
type SDK struct {
	Provider string  `json:"provider"`
	Score    float64 `json:"score"`
	// To add other necessary fields add them here
}

// used in simulating the pipeline ...
var sdks = []string{
	"AdMob", "AdMob-OptOut", "UnityAds", "Facebook",
	"Startapp", "AppLovin", "HuaweiAds", "Vungle", "Tapjoy",
	"AppNext", "Chartboost", "InMobi", "Facebook",
	"Adx", "Twitter", "MoPub", "Instagram",
}

var countries = []string{
	"EN", "ES", "HM", "DO", "GG", "SV", "TN", "IN", "PM", "AE", "SA", "LS", "VA", "ST", "TM", "KM",
	"CG", "SY", "CR", "MT", "SN", "AF", "BH", "AO", "CK", "IM", "CY", "AG", "AL", "UM", "SR", "TG",
	"TO", "TH", "NO", "FM", "BY", "HN", "ER", "SX", "IQ", "IE", "MO", "VN", "MH", "MZ", "SI", "DJ",
	"BW", "AR", "TK", "ML", "NZ", "LK", "PN", "BZ", "NE", "LB", "LR", "DM", "SG", "DK", "GB", "HK",
	"SK", "BJ", "NP", "KG", "PE", "KI", "MA", "NC", "CH", "PH", "AI", "EE", "CN", "VU", "CD", "SD",
	"SC", "MK", "TZ", "RE", "MP", "SJ", "BS", "CI", "MV", "SZ", "MY", "AX", "NF", "EG", "GQ", "IS",
	"AS", "LU", "KW", "LV", "VE", "ID", "AZ", "TT", "DZ", "SS", "AW", "AQ", "NG", "MC", "KE", "NU",
	"IR", "RS", "KP", "BB", "ZW", "UZ", "JM", "GS", "LT", "NI", "TL", "WF", "NL", "GN", "NA", "TR",
	"PL", "LA", "BE", "HU", "RU", "MG", "RO", "EH", "US", "UY", "LI", "BD", "MW", "EC", "OM", "KR",
	"MD", "ME", "IT", "NR", "TJ", "MM", "LY", "MN", "GT", "MS", "CZ", "AD", "SL", "FJ", "IL", "SE",
	"FI", "JP", "TW",
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

// GenerateList generates a list of random ad networks based on
// countries and sdks.
func GenerateList() []*AdNetwork {
	an := []*AdNetwork{}

	for _, country := range countries {
		an = append(an, &AdNetwork{
			Country:      country,
			Banner:       generateSDK(),
			Interstitial: generateSDK(),
			Video:        generateSDK(),
		})
	}

	return an
}

func generateSDK() []*SDK {
	rn := rand.Intn(len(sdks)-1) + 1
	included := []int{}

	out := []*SDK{}

	for i := 0; i < rn; i++ {
		sdkIndex := rand.Intn(len(sdks))
		if containsInt(included, sdkIndex) {
			continue
		}
		score := fmt.Sprintf("%.2f", rand.Float64()*10)
		fscore, _ := strconv.ParseFloat(score, 64)
		out = append(out, &SDK{
			Provider: sdks[sdkIndex],
			Score:    fscore,
		})

		included = append(included, sdkIndex)
	}

	return out
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

func containsInt(arr []int, target int) bool {
	for _, item := range arr {
		if item == target {
			return true
		}
	}

	return false
}
