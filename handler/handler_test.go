package handler

import (
	"encoding/json"
	"expertisetest/adnetwork"
	"expertisetest/config"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/sirupsen/logrus"
)

var arr = []*adnetwork.SDK{
	{Provider: "Facebook"},
	{Provider: "AdMob"},
	{Provider: "AdMob-OptOut"},
	{Provider: "Adx"},
	{Provider: "Unity Ads"},
	{Provider: "Huawei Ads"},
	{Provider: "Twitter"},
	{Provider: "Instagram"},
}

var an = map[string]*adnetwork.AdNetwork{
	"CN": {
		Banner: []*adnetwork.SDK{
			{
				Provider: "AdMob-OptOut",
				Score:    10,
			},
			{
				Provider: "Huawei Ads",
				Score:    8,
			},
		},
		Interstitial: []*adnetwork.SDK{
			{
				Provider: "AdMob",
				Score:    9.9,
			},
			{
				Provider: "Huawei Ads",
				Score:    2.1,
			},
		},
		Video:   []*adnetwork.SDK{},
		Country: "CN",
	},
	"US": {
		Banner: []*adnetwork.SDK{
			{
				Provider: "Facebook",
				Score:    8,
			},
			{
				Provider: "AdMob",
				Score:    3,
			},
		},
		Interstitial: []*adnetwork.SDK{},
		Video: []*adnetwork.SDK{
			{
				Provider: "Facebook",
				Score:    10,
			},
			{
				Provider: "AdMob",
				Score:    9.9,
			},
		},
		Country: "US",
	},
}

func TestLoad(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(ioutil.Discard)

	config.GetTestInstance().Pipefile = "pipefile.json"
	config.GetTestInstance().Prefilter = "prefilter.json"
	h, err := New(logrus.NewEntry(logger))
	if err != nil {
		t.Error(err)
	}

	m, err := h.Load()
	if err != nil {
		t.Error(err)
	}
	for country, network := range m {
		expected := an[country]
		if expected == nil {
			t.Logf("nil entry in map: %s", country)
			t.Fail()
		}

		byteGot, err := json.MarshalIndent(network, "", "  ")
		if err != nil {
			t.Errorf("failed to marshal network: %v", err)
		}

		byteExpected, err := json.MarshalIndent(an[country], "", "  ")
		if err != nil {
			t.Errorf("failed to marshal expected: %v", err)
		}

		if string(byteGot) != string(byteExpected) {
			t.Logf("Got: %s \nExpected: %s\n", string(byteGot), string(byteExpected))
			t.Fail()
		}

	}
}

func TestExcludeFromSDK(t *testing.T) {
	tests := []struct {
		in       []string
		expected []*adnetwork.SDK
	}{
		{
			[]string{"Facebook", "Twitter", "Instagram"},
			[]*adnetwork.SDK{
				{Provider: "AdMob"},
				{Provider: "AdMob-OptOut"},
				{Provider: "Adx"},
				{Provider: "Unity Ads"},
				{Provider: "Huawei Ads"},
			},
		},
		{
			[]string{"Facebook", "Twitter", "Instagram"},
			[]*adnetwork.SDK{
				{Provider: "AdMob"},
				{Provider: "AdMob-OptOut"},
				{Provider: "Adx"},
				{Provider: "Unity Ads"},
				{Provider: "Huawei Ads"},
			},
		},
		{
			[]string{},
			arr,
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			got := excludeFromSDK(arr, test.in)

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
