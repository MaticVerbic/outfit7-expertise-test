package endpoints

import (
	"expertisetest/config"
	"expertisetest/handler"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/sirupsen/logrus"
)

// Update handles /list endpoint functionality.
var Update = func(w http.ResponseWriter, r *http.Request) {
	// Fetch logger from logger middleware.
	log, ok := r.Context().Value(config.LogKey).(*logrus.Entry)
	if !ok {
		log = logrus.NewEntry(logrus.New())
		log.Error("failed to fetch logger")
	}

	// Authorize the client.
	if !authorize(r.Context(), w, "admin") {
		log.WithField("user", r.Context().Value(config.UserKey)).Debug("authorized")
		return
	}

	// Validate request
	// Method check
	if r.Method != http.MethodPost {
		log.Error("invalid http method on update")
		writeResponse(w, 400, fmt.Sprintf("invalid method"), nil)
	}

	// Validate body
	if r.Body == nil {
		log.Error("empty body on update")
		writeResponse(w, 400, fmt.Sprintf("invalid empty request"), nil)
	}

	// Handle arguments
	dropDB := false
	if r.URL.Query()["wipe"] != nil &&
		len(r.URL.Query()["wipe"]) == 1 &&
		r.URL.Query()["wipe"][0] == "true" {
		dropDB = true
	}

	// Parse body
	in := &handler.LoadObject{}
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error(errors.Wrap(err, "failed to read body"))
		writeResponse(w, 500, fmt.Sprintf("internal system error"), nil)
	}

	if err = r.Body.Close(); err != nil {
		log.Error(errors.Wrap(err, "failed to close body"))
	}

	if err = ffjson.Unmarshal(b, in); err != nil {
		log.Error(errors.Wrap(err, "failed to unmrashal json"))
		writeResponse(w, 500, fmt.Sprintf("internal system error"), nil)
	}

	// Parse data and store
	h := handler.GetInstance()
	h.SetLogger(log)

	an := h.Prefilter(in.AdNetwork)
	m, err := handler.ToCountryMap(an)
	if err != nil {
		writeResponse(w, 500, fmt.Sprintf("internal system error"), nil)
	}

	if err = h.Store(m, dropDB); err != nil {
		writeResponse(w, 500, fmt.Sprintf("internal system error"), nil)
	}

	writeResponse(w, 200, "", nil)
}
