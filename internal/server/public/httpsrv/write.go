package httpsrv

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"gitlab.com/Sh00ty/hootydb/internal/kv"
)

func (s *Srv) Write(w http.ResponseWriter, r *http.Request) {
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

	err = s.sharder.Write(ctx, kv.Key(key), val)
	if err != nil {
		writeErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
