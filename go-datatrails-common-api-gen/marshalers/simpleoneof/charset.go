package simpleoneof

import (
	"context"
	"net/http"

	"google.golang.org/protobuf/proto"
)

// AddEncodingHeader adds application/json; charset=utf-8 to response headers
func AddEncodingHeader(ctx context.Context, rw http.ResponseWriter, msg proto.Message) error {
	headers := rw.Header()
	headers.Set("content-type", "application/json; charset=utf-8")
	return nil
}
