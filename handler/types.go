package handler

import "expertisetest/adnetwork"

type filterType string

const (
	excludeCountry filterType = "excCtr"
	mutualPriority filterType = "mutPri"
)

// LoadObject simulates a json object returned by pipeline.
type LoadObject struct {
	AdNetwork []*adnetwork.AdNetwork `json:"data"`
}

// OsVersion postfilter.
type OsVersion struct {
	Args []OsVersionArgs `json:"args"`
}

// OsVersionArgs for postfilter.
type OsVersionArgs struct {
	Os       string   `json:"os"`
	Versions []string `json:"versions"`
	Exclude  []string `json:"exclude"`
}

// Device postfilter.
type Device struct {
	Args []DeviceArgs `json:"args"`
}

// DeviceArgs for postfilter
type DeviceArgs struct {
	Type    string   `json:"type"`
	Exclude []string `json:"exclude"`
}
