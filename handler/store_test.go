// +build integration

package handler

import (
	"expertisetest/adnetwork"
	"expertisetest/config"
	"testing"

	"github.com/pquerna/ffjson/ffjson"
)

// Exclude integration testing since CI currently doesn't support multi container testing.

func TestStore(t *testing.T) {
	// Override TestMain Config to include redis connection.
	config.OverrideInstance(config.NewTestDB())
	config.GetInstance().Pipefile = "pipefile_test.json"
	config.GetInstance().Prefilter = "prefilter.json"
	config.GetInstance().Postfilter = "postfilter.json"

	rd := config.GetInstance().RedisClient

	h := GetInstance()

	networks, err := h.Load()
	if err != nil {
		t.Error(err)
	}

	if err := h.Store(networks, true); err != nil {
		t.Error(err)
	}

	for key := range networks {
		an := &adnetwork.AdNetwork{}
		cmd := rd.Get(key)
		if err := cmd.Scan(an); err != nil {
			t.Errorf("failed to scan key %q with error %v", key, err)
		}

		gotByte, err := ffjson.Marshal(an)
		if err != nil {
			t.Errorf("failed to unmrashal got with err: %v", err)
		}

		expectedByte, err := ffjson.Marshal(networks[key])
		if err != nil {
			t.Errorf("failed to unmrashal expected with err: %v", err)
		}

		if string(gotByte) != string(expectedByte) {
			t.Logf("Got: %s, \nExpected: %s\n", string(gotByte), string(expectedByte))
			t.Fail()
		}
	}
}
