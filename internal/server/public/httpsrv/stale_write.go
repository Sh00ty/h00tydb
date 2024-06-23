package httpsrv

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (s *Srv) StaleWrite(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ok, encoding := parseHttpEncoding(r)
	if !ok {
		writeErr(w, "invalid encoding type", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	var (
		key string
		val string
	)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeErr(w, fmt.Sprintf("can't read request body: %v", err), http.StatusBadRequest)
		return
	}

	switch encoding {
	case Json:
		type req struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		}
		v := req{}

		err = json.Unmarshal(body, &v)
		if err != nil {
			writeErr(w, err.Error(), http.StatusBadRequest)
			return
		}
		if v.Key == "" {
			writeErr(w, "key can't be empty", http.StatusBadRequest)
			return
		}
		if v.Value == "" {
			writeErr(w, "value can't be empty", http.StatusBadRequest)
			return
		}
		key = v.Key
		val = v.Value
	}

	_ = val
	_ = key
	_ = ctx

	writeErr(w, "unimplemented", http.StatusMethodNotAllowed)
}
