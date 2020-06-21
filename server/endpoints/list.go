package endpoints

import (
	"expertisetest/adnetwork"
	"expertisetest/handler"
	"fmt"
	"net/http"

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
	if !authorize(r.Context(), w, "any") {
		return
	}

	if r.Method != http.MethodGet {
		writeResponse(w, 400, fmt.Sprintf("invalid method"), nil)
	}

	vals := r.URL.Query()
	h := handler.GetInstance()

	for _, item := range required {
		rec := vals[item]
		if rec == nil || len(rec) > 1 || len(rec) == 0 || rec[0] == "" {
			writeResponse(w, http.StatusBadRequest, fmt.Sprintf("missing required argument %q", item), nil)
			return
		}
	}

	out, err := h.Get(vals["countryCode"][0])
	if err != nil {
		writeResponse(w, http.StatusInternalServerError, errors.Wrap(err, "internal system error").Error(), nil)
		return
	}

	if out == nil {
		out, err = h.GetRandom()
		if err != nil {
			writeResponse(w, http.StatusInternalServerError, errors.Wrap(err, "internal system error").Error(), nil)
			return
		}

		out.Country = vals["countryCode"][0]
		arr := h.Prefilter([]*adnetwork.AdNetwork{out})
		if len(arr) != 1 {
			writeResponse(w, http.StatusInternalServerError, errors.Wrap(err, "internal system error").Error(), nil)
			return
		}
		out = arr[0]

		out = h.Postfilter(vals, out)
		writeResponse(w, http.StatusOK, "", out)
		return
	}
	out = h.Postfilter(vals, out)
	writeResponse(w, http.StatusOK, "", out)
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
