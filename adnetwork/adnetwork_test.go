package adnetwork

import (
	"encoding/json"
	"fmt"
	"testing"
)

var arr = []*SDK{
	{Provider: "Facebook"},
	{Provider: "AdMob"},
	{Provider: "AdMob-OptOut"},
	{Provider: "Adx"},
	{Provider: "Unity Ads"},
	{Provider: "Huawei Ads"},
	{Provider: "Twitter"},
	{Provider: "Instagram"},
}

var adNetwork = AdNetwork{
	Country:      "SI",
	Banner:       arr,
	Interstitial: arr,
	Video:        arr,
}

func TestContainsAllProviders(t *testing.T) {
	tests := []struct {
		in       []string
		expected bool
		adType   string
	}{
		{
			[]string{"Facebook", "Twitter", "Instagram"},
			true,
			"banner",
		},
		{
			[]string{"Adx", "Unity Ads", "Sony"},
			false,
			"interstitial",
		},
		{
			[]string{},
			true,
			"video",
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			if got := adNetwork.ContainsAllProviders(test.adType, test.in); got != test.expected {
				t.Logf("Got: %t Expected: %t", got, test.expected)
				t.Fail()
			}
		})
	}
}

func TestContainsAnyProviders(t *testing.T) {
	tests := []struct {
		in       []string
		expected []string
		adType   string
	}{
		{
			[]string{"Facebook", "Twitter", "Instagram"},
			[]string{"Facebook", "Twitter", "Instagram"},
			"banner",
		},
		{
			[]string{"Adx", "Unity Ads", "Sony"},
			[]string{"Adx", "Unity Ads"},
			"interstitial",
		},
		{
			[]string{},
			[]string{},
			"video",
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			got := adNetwork.ContainsAnyProviders(test.adType, test.in)
			byteGot, err := json.MarshalIndent(got, "", "  ")
			if err != nil {
				t.Errorf("failed to marshal got: %v", err)
			}

			byteExpected, err := json.MarshalIndent(test.expected, "", "  ")
			if err != nil {
				t.Errorf("failed to marshal expected: %v", err)
			}

			if string(byteGot) != string(byteExpected) {
				t.Logf("Got: %s \nExpected: %s\n", string(byteGot), string(byteExpected))
				t.Fail()
			}
		})
	}
}

func TestContainsAll(t *testing.T) {
	tests := []struct {
		in            []string
		expected      bool
		expectedEmpty bool
	}{
		{
			[]string{"Facebook", "Twitter", "Instagram"},
			true,
			false,
		},
		{
			[]string{"Adx", "Unity Ads", "Sony"},
			false,
			false,
		},
		{
			[]string{},
			true,
			true,
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			if got := containsAll(arr, test.in); got != test.expected {
				t.Logf("Got: %t Expected: %t", got, test.expected)
				t.Fail()
			}
		})
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", len(tests)+i), func(t *testing.T) {
			if got := containsAll([]*SDK{}, test.in); got != test.expectedEmpty {
				t.Logf("Got: %t Expected: %t", got, test.expected)
				t.Fail()
			}
		})
	}
}

func TestContainsAny(t *testing.T) {
	tests := []struct {
		in       []string
		expected []string
	}{
		{
			[]string{"Facebook", "Twitter", "Instagram"},
			[]string{"Facebook", "Twitter", "Instagram"},
		},
		{
			[]string{"Adx", "Unity Ads", "Sony"},
			[]string{"Adx", "Unity Ads"},
		},
		{
			[]string{},
			[]string{},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			got := containsAny(arr, test.in)
			byteGot, err := json.MarshalIndent(got, "", "  ")
			if err != nil {
				t.Errorf("failed to marshal got: %v", err)
			}

			byteExpected, err := json.MarshalIndent(test.expected, "", "  ")
			if err != nil {
				t.Errorf("failed to marshal expected: %v", err)
			}

			if string(byteGot) != string(byteExpected) {
				t.Logf("Got: %s \nExpected: %s\n", string(byteGot), string(byteExpected))
				t.Fail()
			}
		})
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", len(tests)+i), func(t *testing.T) {
			got := containsAny([]*SDK{}, test.in)

			byteGot, err := json.MarshalIndent(got, "", "  ")
			if err != nil {
				t.Errorf("failed to marshal got: %v", err)
			}

			if len(got) > 0 {
				t.Logf("Got: %s \nExpected empty list", string(byteGot))
				t.Fail()
			}
		})
	}
}
