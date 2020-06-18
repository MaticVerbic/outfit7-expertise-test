package handler

import (
	"encoding/json"
	"expertisetest/adnetwork"
	"expertisetest/config"
	"fmt"
	"io/ioutil"
	"sort"
	"sync"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type filterType string

const (
	excludeCountry filterType = "excCtr"
	mutualPriority filterType = "mutPri"
)

// Handler handles loading and filtering of data.
type Handler struct {
	log               *logrus.Entry
	PrefilterMappings []Prefilter `json:"prefilterMappings"`
}

// Prefilter is a definition of a filter running on load.
type Prefilter struct {
	FilterType filterType          `json:"type"`
	Args       map[string][]string `json:"args"`
}

// New returns a new Handler.
func New(log *logrus.Entry) (*Handler, error) {
	h := &Handler{
		log: log.WithField("package", "handler"),
	}

	if err := h.LoadPrefilter(); err != nil {
		return h, errors.Wrap(err, "failed load prefilter")
	}

	return h, nil
}

// LoadObject simulates a json object returned by pipeline.
type LoadObject struct {
	AdNetwork []*adnetwork.AdNetwork `json:"data"`
}

// LoadPrefilter loads prefilter settings and mappings from config file.
func (h *Handler) LoadPrefilter() error {
	b, err := ioutil.ReadFile(config.GetInstance().Prefilter)
	if err != nil {
		return errors.Wrap(err, "failed to read from prefilter config")
	}

	if err = json.Unmarshal(b, h); err != nil {
		return errors.Wrap(err, "failed to load unmarshal prefilters")
	}

	return nil
}

// Exclude removes all providers in the list from a specified network.
func (h *Handler) Exclude(an *adnetwork.AdNetwork, providers []string) *adnetwork.AdNetwork {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		an.Banner = excludeFromSDK(an.Banner, providers)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		an.Interstitial = excludeFromSDK(an.Interstitial, providers)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		an.Video = excludeFromSDK(an.Video, providers)
	}()

	return an
}

// MutualPriority removes all providers except the first one found, list has to be sorted by priority.
func (h *Handler) MutualPriority(an *adnetwork.AdNetwork, providers []string) *adnetwork.AdNetwork {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		if prov := an.ContainsAnyProviders("banner", providers); len(prov) > 1 {
			an.Banner = excludeFromSDK(an.Banner, prov[1:])
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if prov := an.ContainsAnyProviders("interstitial", providers); len(prov) > 1 {
			an.Interstitial = excludeFromSDK(an.Interstitial, prov[1:])
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if prov := an.ContainsAnyProviders("video", providers); len(prov) > 1 {
			an.Video = excludeFromSDK(an.Video, prov[1:])
		}
	}()

	wg.Wait()

	return an
}

// Prefilter ...
// Current implementation is such that no complex channel implementations are
// required in order to implement a concurrency model for faster prefiltering.
// The goal here is to handle a prefilter for each ad network concurrently as well as
// a list of each ad type concurrently. This allows to mitigate some load time due to
// high requirement for O(n) traversals over separate adType lists for each AdNetwork.
func (h *Handler) Prefilter(an []*adnetwork.AdNetwork) []*adnetwork.AdNetwork {
	var wg sync.WaitGroup
	ch := make(chan *adnetwork.AdNetwork, len(an))

	for _, network := range an {
		wg.Add(1)
		go func(an *adnetwork.AdNetwork, ch chan *adnetwork.AdNetwork) {
			defer wg.Done()
			for _, prefilter := range h.PrefilterMappings {
				switch prefilter.FilterType {
				case excludeCountry:
					if ct := prefilter.Args[an.Country]; ct != nil {
						an = h.Exclude(an, ct)
					}
				case mutualPriority:
					for _, args := range prefilter.Args {
						an = h.MutualPriority(an, args)
					}
				}
			}

			sort.Sort(adnetwork.ScoreSorter(an.Banner))
			sort.Sort(adnetwork.ScoreSorter(an.Interstitial))
			sort.Sort(adnetwork.ScoreSorter(an.Video))
			ch <- an

		}(network, ch)
	}

	wg.Wait()
	close(ch)
	arr := []*adnetwork.AdNetwork{}

	for network := range ch {
		arr = append(arr, network)
	}

	return arr
}

// Load is the main method to simulate fetching data from pipeline.
func (h *Handler) Load() (map[string]*adnetwork.AdNetwork, error) {
	b, err := ioutil.ReadFile(config.GetInstance().Pipefile)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read from pipeline")
	}
	load := &LoadObject{}
	if err = json.Unmarshal(b, load); err != nil {
		return nil, errors.Wrap(err, "failed to unmrashal pipeline data")
	}

	filtered := h.Prefilter(load.AdNetwork)

	return toCountryMap(filtered)
}

// returns an array of ad networks as a map[country]network
func toCountryMap(an []*adnetwork.AdNetwork) (map[string]*adnetwork.AdNetwork, error) {
	m := map[string]*adnetwork.AdNetwork{}

	for _, network := range an {
		if exists := m[network.Country]; exists != nil {
			return m, fmt.Errorf("key exists in map: %s", network.Country)
		}

		m[network.Country] = network
	}

	return m, nil
}

// removes sdks from network if they contain given providers.
func excludeFromSDK(arr []*adnetwork.SDK, providers []string) []*adnetwork.SDK {
	if len(arr) == 0 {
		return []*adnetwork.SDK{}
	}

	i := 0
	for {
		if i == len(arr) {
			break
		}

		found := false
		for _, provider := range providers {
			if arr[i].Provider == provider {
				arr = append(arr[:i], arr[i+1:]...)
				found = true
			}
		}

		if !found {
			i++
		}
	}

	return arr
}
