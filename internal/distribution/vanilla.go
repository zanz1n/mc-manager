package distribution

import (
	"context"
	"encoding/hex"
	"errors"
	"net/http"

	"github.com/zanz1n/mc-manager/internal/pb"
)

type vanillaManifest struct {
	Latest struct {
		Release  string `json:"release" validate:"required"`
		Snapshot string `json:"snapshot" validate:"required"`
	} `json:"latest"`

	Versions []vanillaManifestVersion `json:"versions" validate:"required"`
}

type vanillaManifestVersion struct {
	ID   string `json:"id" validate:"required"`
	Type string `json:"type" validate:"required"`
	URL  string `json:"url" validate:"required,url"`
}

type vanillaVersion struct {
	ID   string `json:"id" validate:"required"`
	Type string `json:"type" validate:"required"`

	Downloads struct {
		Server struct {
			SHA1 string `json:"sha1" validate:"required,hexadecimal"`
			Size int    `json:"size" validate:"gt=0"`
			URL  string `json:"url" validate:"required,url"`
		} `json:"server"`
	} `json:"downloads"`

	JavaVersion struct {
		Component    string `json:"component" validate:"required"`
		MajorVersion uint8  `json:"majorVersion" validate:"required"`
	} `json:"javaVersion"`
}

type vanilla struct {
	c *http.Client
}

func NewVanilla(c *http.Client) Distribution {
	if c == nil {
		c = http.DefaultClient
	}

	return &vanilla{c: c}
}

// GetLatest implements Distribution.
func (d *vanilla) GetLatest(ctx context.Context) (Version, error) {
	data, err := d.getManifest(ctx)
	if err != nil {
		return Version{}, err
	}

	var (
		version vanillaManifestVersion
		ok      bool
	)
	for _, v := range data.Versions {
		if v.ID == data.Latest.Release {
			version, ok = v, true
		}
	}

	if !ok {
		return Version{}, ErrVersionNotFound
	}
	return d.getVersion(ctx, version)
}

// GetLatestSubver implements Distribution.
func (d *vanilla) GetVersion(ctx context.Context, semver string) (Version, error) {
	data, err := d.getManifest(ctx)
	if err != nil {
		return Version{}, err
	}

	var (
		version vanillaManifestVersion
		ok      bool
	)
	for _, v := range data.Versions {
		if v.ID == semver {
			version, ok = v, true
		}
	}

	if !ok {
		return Version{}, ErrVersionNotFound
	}
	return d.getVersion(ctx, version)
}

// GetAll implements Distribution.
func (d *vanilla) GetAll(ctx context.Context) ([]string, error) {
	data, err := d.getManifest(ctx)
	if err != nil {
		return nil, err
	}

	res := make([]string, len(data.Versions))
	for i := range data.Versions {
		res[i] = data.Versions[i].ID
	}
	return res, nil
}

func (d *vanilla) getManifest(ctx context.Context) (vanillaManifest, error) {
	var data vanillaManifest

	err := getreq(
		ctx,
		d.c,
		"https://launchermeta.mojang.com/mc/game/version_manifest.json",
		&data,
	)
	if err != nil {
		err = errors.Join(ErrHttp, err)
	}

	return data, err
}

func (d *vanilla) getVersion(
	ctx context.Context,
	version vanillaManifestVersion,
) (Version, error) {
	var data vanillaVersion

	err := getreq(ctx, d.c, version.URL, &data)
	if err != nil {
		return Version{}, errors.Join(ErrHttp, err)
	}
	javaVersion := normalizeJavaLts(data.JavaVersion.MajorVersion)

	htype := pb.HashType_SHA1
	hash, err := hex.DecodeString(data.Downloads.Server.SHA1)
	if err != nil {
		htype, hash = pb.HashType_HASH_NONE, nil
	}

	return Version{
		ID:           data.ID,
		URL:          data.Downloads.Server.URL,
		Hash:         hash,
		HashType:     htype,
		Distribution: pb.Distribution_VANILLA,
		JavaVersion:  javaVersion,
	}, nil
}
