package httpsrv

import (
	"fmt"

	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"gitlab.com/Sh00ty/hootydb/internal/kv"
)

func (s *Srv) Read(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := mux.Vars(r)
	key := params["key"]
	key = strings.TrimSpace(key)
	if key == "" {
		writeErr(w, "invalid key format", http.StatusBadRequest)
		return
	}
	val, err := s.sharder.Read(ctx, kv.Key(key))
	if err != nil {
		if kv.IsNotFound(err) {
			writeErr(w, "", http.StatusNotFound)
			return
		}
		s.log.Errorf(ctx, err, "failed to public read key %s", key)
		writeErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	v := val.Val.(string)
	w.Write([]byte(fmt.Sprintf("\"value\":\"%s\"", v)))

	s.log.Infof(ctx, "public read returns value %s with version %v", v, val.Ver)
}
