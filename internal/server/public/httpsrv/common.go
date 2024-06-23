package httpsrv

import (
	"fmt"
	"net/http"
)

type contentType string

const (
	Json contentType = "application/json"
)

var AllowedContentTypes = map[contentType]struct{}{
	Json: {},
}

func parseHttpEncoding(r *http.Request) (bool, contentType) {
	encoding := r.Header["Content-Type"]
	for _, enc := range encoding {
		if _, ok := AllowedContentTypes[contentType(enc)]; ok {
			return true, contentType(enc)
		}
	}
	return false, ""
}

func writeErr(w http.ResponseWriter, message string, status int) error {
	w.WriteHeader(status)
	if message == "" {
		return nil
	}
	_, err := w.Write([]byte(fmt.Sprintf("\"error\":\"%s\"", message)))
	return err
}
