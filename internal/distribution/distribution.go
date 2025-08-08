package distribution

import (
	"bytes"
	"context"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha3"
	"crypto/sha512"
	"errors"
	"hash"
	"io"
	"net/http"
	"os"

	"github.com/zanz1n/mc-manager/internal/pb"
)

type Distribution interface {
	GetLatest(ctx context.Context) (Version, error)
	GetVersion(ctx context.Context, semver string) (Version, error)

	GetAll(ctx context.Context) ([]string, error)
}

type Version struct {
	ID           string          `json:"id"`
	URL          string          `json:"url"`
	Hash         []byte          `json:"hash"`
	JVMArgs      []string        `json:"jvm_args"`
	HashType     pb.HashType     `json:"hash_type"`
	Distribution pb.Distribution `json:"distribution"`
	JavaVersion  pb.JavaVersion  `json:"java_version"`
}

// The returned hash can be nil in case the hash is not available.
func (v *Version) CreateHash() (h hash.Hash) {
	switch v.HashType {
	case pb.HashType_SHA1:
		h = sha1.New()
	case pb.HashType_SHA256:
		h = sha256.New()
	case pb.HashType_SHA224:
		h = sha3.New224()
	case pb.HashType_SHA384:
		h = sha3.New384()
	case pb.HashType_SHA512:
		h = sha512.New()
	}
	return
}

func (v *Version) Download(ctx context.Context, c *http.Client) (io.ReadCloser, error) {
	if c == nil {
		c = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.URL, nil)
	if err != nil {
		return nil, errors.Join(ErrHttp, err)
	}

	res, err := c.Do(req)
	if err != nil {
		return nil, errors.Join(ErrHttp, err)
	}
	return res.Body, nil
}

func (v *Version) DownloadTo(ctx context.Context, c *http.Client, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	body, err := v.Download(ctx, c)
	if err != nil {
		return err
	}
	defer body.Close()

	hashProto := v.CreateHash()
	if hashProto == nil {
		_, err = io.Copy(file, body)
		return err
	}

	hw := &hashWriter{
		w: file,
		h: hashProto,
	}

	_, err = io.Copy(hw, body)
	if err != nil {
		return err
	}

	gotHash := hw.Sum()
	if !bytes.Equal(v.Hash, gotHash) {
		return errors.Join(
			ErrHashFailed,
			errors.New("match failed"),
		)
	}
	return nil
}
