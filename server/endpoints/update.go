package endpoints

import (
	"expertisetest/handler"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pquerna/ffjson/ffjson"
)

var Update = func(w http.ResponseWriter, r *http.Request) {
	if !authorize(r.Context(), w, "admin") {
		return
	}

	if r.Body == nil {
		writeResponse(w, 400, fmt.Sprintf("invalid empty request"), nil)
	}

	in := &handler.LoadObject{}
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		writeResponse(w, 500, fmt.Sprintf("internal system error"), nil)
	}

	if err = ffjson.Unmarshal(b, in); err != nil {
		writeResponse(w, 500, fmt.Sprintf("internal system error"), nil)
	}

	h := handler.GetInstance()
	an := h.Prefilter(in.AdNetwork)
	m, err := handler.ToCountryMap(an)
	if err != nil {
		writeResponse(w, 500, fmt.Sprintf("internal system error"), nil)
	}

	if err = h.Store(m, true); err != nil {
		writeResponse(w, 500, fmt.Sprintf("internal system error"), nil)
	}

	writeResponse(w, 200, "", nil)
}
