package handler

import (
	"expertisetest/adnetwork"
	"expertisetest/config"
	"fmt"
	"io/ioutil"
	"net/url"
	"sort"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/sirupsen/logrus"
)

var (
	instance *Handler
	once     sync.Once
)

// GetInstance always returns the same instance of Handler.
// Also ensuring the filter configs only get loaded once, instead of each api call.
func GetInstance() *Handler {
	once.Do(func() {
		var err error
		instance, err = New()
		if err != nil {
			logrus.Fatal(err)
		}
	})
	return instance
}

// Prefilter is a definition of a filter running on load.
type Prefilter struct {
	FilterType filterType          `json:"type"`
	Args       map[string][]string `json:"args"`
}

// Postfilter is a definition of a filter running on api call.
type Postfilter struct {
	OsVersion OsVersion `json:"osVersion"`
	Device    Device    `json:"device"`
}

// Handler handles loading and filtering of data.
type Handler struct {
	log                *logrus.Entry
	PrefilterMappings  []Prefilter `json:"prefilterMappings"`
	PostfilterMappings Postfilter  `json:"postfilterMappings"`
}

// New returns a new Handler.
func New() (*Handler, error) {
	h := &Handler{
		log: logrus.WithField("package", "handler"),
	}

	h.log.Debug("init")

	if err := h.LoadPrefilter(); err != nil {
		return h, errors.Wrap(err, "failed load prefilter")
	}

	if err := h.LoadPostfilter(); err != nil {
		return h, errors.Wrap(err, "failed load postfilter")
	}

	return h, nil
}

// Get fetches from redis
func (h *Handler) Get(key string) (*adnetwork.AdNetwork, error) {
	h.log.WithFields(logrus.Fields{
		"type": "get",
		"key":  key,
	}).Debug("init")
	an := &adnetwork.AdNetwork{}
	rd := config.GetInstance().RedisClient

	status, err := rd.Exists(key).Result()
	if err != nil {
		return nil, errors.Wrap(err, "failed to validate key")
	}
	if status == 0 {
		return nil, nil
	}
	cmd := rd.Get(key)

	if err := cmd.Scan(an); err != nil {
		return nil, fmt.Errorf("failed to scan key %q with error %v", key, err)
	}

	return an, nil
}

// GetRandom fetches a random value from redis
func (h *Handler) GetRandom() (*adnetwork.AdNetwork, error) {
	h.log.WithFields(logrus.Fields{
		"type": "random fetch",
	}).Debug("init")
	an := &adnetwork.AdNetwork{}
	rd := config.GetInstance().RedisClient

	key := rd.RandomKey().Val()
	cmd := rd.Get(key)
	if err := cmd.Scan(an); err != nil {
		return nil, fmt.Errorf("failed to scan key %q with error %v", key, err)
	}

	return an, nil
}

// Load is the main method to simulate fetching data from pipeline.
func (h *Handler) Load() (map[string]*adnetwork.AdNetwork, error) {
	h.log.WithField("filename", config.GetInstance().Pipefile).Debug("load pipefile")
	b, err := ioutil.ReadFile(config.GetInstance().Pipefile)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read from pipeline")
	}
	load := &LoadObject{}
	if err = ffjson.Unmarshal(b, load); err != nil {
		return nil, errors.Wrap(err, "failed to unmrashal pipeline data")
	}

	filtered := h.Prefilter(load.AdNetwork)

	if len(filtered) == 0 {
		return nil, errors.New("nil list returned from filtering")
	}

	return ToCountryMap(filtered)
}

// Store the prefiltered data to redis. dropDB will drop the database before refilling it back up,
// otherwise non-overwritten old records will remain in database.
func (h *Handler) Store(mappings map[string]*adnetwork.AdNetwork, dropDB bool) error {
	h.log.WithField("type", "store").Debug("init")
	rd := config.GetInstance().RedisClient
	// Not removing old data because it's better to have non-optimal list rather than an empty one.
	// TODO-DONE: Is it better to have old data or returning a random adNetwork on apiCall?
	// Possible solution, implement a config to change this behavior. At api call fetching
	// data with resulting nil set would in case of complete wipe on store return a random latest key anyway.
	// Possible downside to this is longer response times when missing, since both filtering processes have
	// to happen at api call in case of a random hit.
	// Keeping old data might cause hitting old random sets when original countries do not exist with small sets.
	// (searching for a not existing set (exp. SI), and hitting a not updated set for some other country (exp. GER))
	pipe := rd.TxPipeline()

	if dropDB {
		pipe.FlushDB()
	}

	for country, adNetwork := range mappings {
		// no TTL because it's better to have non-optimal list to an empty one.

		if err := pipe.Set(country, adNetwork, 0).Err(); err != nil {
			// log errors and continue
			h.log.Error(errors.Wrapf(err, "failed to set %q with error", country))
		}

	}

	if _, err := pipe.Exec(); err != nil {
		return errors.Wrap(err, "failed to exec transaction")
	}

	return nil
}

// LoadPrefilter loads prefilter settings and mappings from config file.
func (h *Handler) LoadPrefilter() error {
	h.log.WithField("filename", config.GetInstance().Prefilter).Debug("load prefilter")

	b, err := ioutil.ReadFile(config.GetInstance().Prefilter)
	if err != nil {
		return errors.Wrap(err, "failed to read from prefilter config")
	}

	if err = ffjson.Unmarshal(b, h); err != nil {
		return errors.Wrap(err, "failed to load unmarshal prefilters")
	}

	return nil
}

// LoadPostfilter loads postfilter settings and mappings from config file.
func (h *Handler) LoadPostfilter() error {
	h.log.WithField("filename", config.GetInstance().Postfilter).Debug("load postfiler")

	b, err := ioutil.ReadFile(config.GetInstance().Postfilter)
	if err != nil {
		return errors.Wrap(err, "failed to read from prefilter config")
	}

	if err = ffjson.Unmarshal(b, h); err != nil {
		return errors.Wrap(err, "failed to load unmarshal postfilter")
	}

	return nil
}

// OsVersion implements filtering by operating system and its version.
func (h *Handler) OsVersion(os, version string, an *adnetwork.AdNetwork) *adnetwork.AdNetwork {
	h.log.WithFields(logrus.Fields{
		"type":    "postfilter",
		"name":    "os_version",
		"country": an.Country,
	}).Debug("init")

	for _, osFilter := range h.PostfilterMappings.OsVersion.Args {
		if strings.ToLower(os) == osFilter.Os && containsString(osFilter.Versions, version) {
			return h.Exclude(an, osFilter.Exclude)

		}
	}

	return an
}

// DeviceFilter implements filtering on device type.
func (h *Handler) DeviceFilter(deviceType string, an *adnetwork.AdNetwork) *adnetwork.AdNetwork {
	h.log.WithFields(logrus.Fields{
		"type":    "postfilter",
		"name":    "device_type",
		"country": an.Country,
	}).Debug("init")

	for _, filter := range h.PostfilterMappings.Device.Args {
		if strings.ToLower(deviceType) == filter.Type {
			return h.Exclude(an, filter.Exclude)

		}
	}

	return an
}

// Postfilter is executed at api call type.
func (h *Handler) Postfilter(queryVals url.Values, an *adnetwork.AdNetwork) *adnetwork.AdNetwork {
	h.log.WithFields(logrus.Fields{
		"type": "postfilter",
	}).Debug("init")

	an = h.OsVersion(queryVals["platform"][0], queryVals["osVersion"][0], an)
	return h.DeviceFilter(queryVals["device"][0], an)
}

// Exclude removes all providers in the list from a specified network.
func (h *Handler) Exclude(an *adnetwork.AdNetwork, providers []string) *adnetwork.AdNetwork {
	h.log.WithFields(logrus.Fields{
		"type":      "exclude",
		"country":   an.Country,
		"providers": fmt.Sprintf("[%s]", strings.Join(providers, ", ")),
	}).Debug("init")

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

	wg.Wait()
	return an
}

// MutualPriority removes all providers except the first one found, list has to be sorted by priority.
func (h *Handler) MutualPriority(an *adnetwork.AdNetwork, providers []string) *adnetwork.AdNetwork {
	h.log.WithFields(logrus.Fields{
		"type":    "prefilter",
		"name":    "mutual_priority",
		"country": an.Country,
	}).Debug("init")
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
	h.log.WithFields(logrus.Fields{
		"type": "prefilter",
	}).Debug("init")

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

// SetLogger allows to override the default logger.
func (h *Handler) SetLogger(entry *logrus.Entry) {
	h.log = entry
}

// ToCountryMap returns an array of ad networks as a map[country]network
func ToCountryMap(an []*adnetwork.AdNetwork) (map[string]*adnetwork.AdNetwork, error) {
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
	out := []*adnetwork.SDK{}
	if arr == nil || len(arr) == 0 {
		return []*adnetwork.SDK{}
	}

	indexes := []int{}
	for i, item := range arr {
		for _, provider := range providers {
			if provider == item.Provider {
				indexes = append(indexes, i)
			}
		}
	}

	for i, item := range arr {
		if !containsInt(indexes, i) {
			out = append(out, item)
		}
	}

	return out
}

func containsString(arr []string, target string) bool {
	for _, item := range arr {
		if item == target {
			return true
		}
	}

	return false
}

func containsInt(arr []int, target int) bool {
	for _, item := range arr {
		if item == target {
			return true
		}
	}

	return false
}
