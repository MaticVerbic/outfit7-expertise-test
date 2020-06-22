package endpoints

import (
	"expertisetest/adnetwork"
	"expertisetest/config"
	"expertisetest/handler"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/sirupsen/logrus"
)

// Response is http response object that is returned to the client.
type Response struct {
	Network *adnetwork.AdNetwork `json:"network,omitempty"`
	Err     string               `json:"error,omitempty"`
}

var required = []string{
	"countryCode",
	"platform",
	"osVersion",
	"device",
}

// List handles /list endpoint functionality.
var List = func(w http.ResponseWriter, r *http.Request) {
	log, ok := r.Context().Value(config.LogKey).(*logrus.Entry)
	if !ok {
		log = logrus.NewEntry(logrus.New())
		log.Error("failed to fetch logger")
	}

	if !authorize(r.Context(), w, "any") {
		log.WithField("user", r.Context().Value(config.UserKey)).Debug("authorized")
		return
	}

	if r.Method != http.MethodGet {
		log.Error("invalid method")
		writeResponse(w, 400, fmt.Sprintf("invalid method"), nil)
	}

	vals := r.URL.Query()
	h := handler.GetInstance()
	h.SetLogger(log)

	for _, item := range required {
		rec := vals[item]
		if rec == nil || len(rec) > 1 || len(rec) == 0 || rec[0] == "" {
			log.Error(fmt.Sprintf("missing required argument %q", item))
			writeResponse(w, http.StatusBadRequest, fmt.Sprintf("missing required argument %q", item), nil)
			return
		}
	}

	out, err := h.Get(vals["countryCode"][0])
	if err != nil {
		log.Error(errors.Wrapf(err, "failed to fetch list for country %q", vals["countryCode"][0]))
		writeResponse(w, http.StatusInternalServerError, errors.Wrap(err, "internal system error").Error(), nil)
		return
	}

	if out == nil {
		log.WithField("countryCode", vals["countryCode"][0]).Warn("cache miss")
		out, err = h.GetRandom()
		if err != nil {
			log.Error(errors.Wrap(err, "failed to fetch random"))
			writeResponse(w, http.StatusInternalServerError, errors.Wrap(err, "internal system error").Error(), nil)
			return
		}

		out.Country = vals["countryCode"][0]
		arr := h.Prefilter([]*adnetwork.AdNetwork{out})
		if len(arr) != 1 {
			log.Error(errors.Wrap(err, "failed to prefilter"))
			writeResponse(w, http.StatusInternalServerError, errors.Wrap(err, "internal system error").Error(), nil)
			return
		}
		out = arr[0]

		out = h.Postfilter(vals, out)
		if testEmpty(out) {
			for i := 0; i < config.GetInstance().RetryAttempts; i++ {
				err = nil
				out, err = retry(vals)
				if err != nil {
					log.Error(errors.Wrap(err, "failed to retry"))
				}

				if !testEmpty(out) {
					log.WithField("countryCode", vals["countryCode"][0]).Warn("retry miss")
					break
				}
			}
		}

		// if last executed retry was an error, return error to client.
		if err != nil {
			writeResponse(w, http.StatusInternalServerError, errors.Wrap(err, "internal system error").Error(), nil)
			return
		}

		writeResponse(w, http.StatusOK, "", out)
		return
	}

	out = h.Postfilter(vals, out)

	if testEmpty(out) {
		for i := 0; i < config.GetInstance().RetryAttempts; i++ {
			err = nil
			out, err = retry(vals)
			if err != nil {
				log.Error(errors.Wrap(err, "failed to retry"))
			}

			if !testEmpty(out) {
				log.WithField("countryCode", vals["countryCode"][0]).Warn("retry miss")
				break
			}
		}
	}

	// if last executed retry was an error, return error to client.
	if err != nil {
		writeResponse(w, http.StatusInternalServerError, errors.Wrap(err, "internal system error").Error(), nil)
		return
	}

	writeResponse(w, http.StatusOK, "", out)
}

func testEmpty(an *adnetwork.AdNetwork) bool {
	return len(an.Banner) == 0 ||
		len(an.Interstitial) == 0 ||
		len(an.Video) == 0
}

func retry(vals url.Values) (*adnetwork.AdNetwork, error) {
	h := handler.GetInstance()

	out, err := h.GetRandom()
	if err != nil {
		return out, errors.Wrapf(err, "failed to fetch random key")
	}

	out.Country = vals["countryCode"][0]
	arr := h.Prefilter([]*adnetwork.AdNetwork{out})
	if len(arr) != 1 {
		return out, errors.Wrapf(err, "failed to fetch random key")
	}
	out = arr[0]

	out = h.Postfilter(vals, out)
	return out, nil
}

func writeResponse(w http.ResponseWriter, status int, errStr string, out *adnetwork.AdNetwork) {
	w.WriteHeader(status)

	body, err := ffjson.Marshal(&Response{
		Network: out,
		Err:     errStr,
	})
	if err != nil {
		logrus.Error(err)
		return
	}

	if _, err := w.Write(body); err != nil {
		logrus.Error(err)
	}

	return
}
