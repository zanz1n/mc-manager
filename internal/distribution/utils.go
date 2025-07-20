package distribution

import (
	"context"
	"encoding/json"
	"hash"
	"io"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/zanz1n/mc-manager/internal/pb/distropb"
)

var validate = validator.New()

func getreq(ctx context.Context, c *http.Client, url string, v any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	res, err := c.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(v)
	if err != nil {
		return err
	}

	return validate.StructCtx(ctx, v)
}

func normalizeJavaLts(v uint8) distropb.JavaVersion {
	switch {
	case v <= 8:
		return distropb.JavaVersion_JAVA8
	case v <= 11:
		return distropb.JavaVersion_JAVA11
	case v <= 17:
		return distropb.JavaVersion_JAVA17
	case v <= 21:
		return distropb.JavaVersion_JAVA21
	// TODO: java 25 yet to be released
	case v <= 24:
		return distropb.JavaVersion_JAVA24
	}

	return distropb.JavaVersion_JAVA8
}

var _ io.Writer = (*hashWriter)(nil)

type hashWriter struct {
	w io.Writer
	h hash.Hash
}

// Write implements io.WriteCloser.
func (h *hashWriter) Write(p []byte) (n int, err error) {
	n, err = h.w.Write(p)
	if _, err2 := h.h.Write(p); err2 != nil {
		err = err2
	}
	return
}

func (h *hashWriter) Sum() []byte {
	return h.h.Sum(nil)
}
