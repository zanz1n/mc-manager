package distribution

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/zanz1n/mc-manager/internal/pb/distropb"
)

type paperManifest struct {
	Versions []paperVersion `json:"versions"`
}

type paperVersion struct {
	Version struct {
		ID   string `json:"id" validate:"required"`
		Java struct {
			Flags struct {
				Recomended []string `json:"recommended"`
			} `json:"flags"`
			Version struct {
				Minimum uint8 `json:"minimum" validate:"required"`
			} `json:"version"`
		} `json:"java"`
	} `json:"version"`
}

type paperBuild struct {
	Downloads map[string]paperBuildDownload `json:"downloads"`
}

type paperBuildDownload struct {
	Name string `json:"name" validate:"required"`
	Size int    `json:"size" validate:"gt=0"`
	URL  string `json:"url" validate:"required,url"`

	Checksums struct {
		SHA256 string `json:"sha256" validate:"required,hexadecimal"`
	} `json:"checksums"`
}

type paper struct {
	c *http.Client
}

func NewPaper(c *http.Client) Distribution {
	if c == nil {
		c = http.DefaultClient
	}

	return &paper{c: c}
}

// GetLatest implements Distribution.
func (d *paper) GetLatest(ctx context.Context) (Version, error) {
	var data paperManifest

	err := getreq(
		ctx,
		d.c,
		"https://fill.papermc.io/v3/projects/paper/versions",
		&data,
	)
	if err != nil {
		return Version{}, errors.Join(ErrHttp, err)
	}

	if len(data.Versions) == 0 {
		return Version{}, ErrVersionNotFound
	}

	// TODO: implement cache of latest version
	return d.GetVersion(ctx, data.Versions[0].Version.ID)
}

// GetVersion implements Distribution.
func (d *paper) GetVersion(ctx context.Context, semver string) (Version, error) {
	fetchUrl := fmt.Sprintf(
		"https://fill.papermc.io/v3/projects/paper/versions/%s",
		semver,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fetchUrl, nil)
	if err != nil {
		return Version{}, errors.Join(ErrHttp, err)
	}

	res, err := d.c.Do(req)
	if err != nil {
		return Version{}, errors.Join(ErrHttp, err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return Version{}, ErrVersionNotFound
	}

	var data paperVersion

	err = json.NewDecoder(res.Body).Decode(&data)
	if err != nil {
		return Version{}, errors.Join(ErrHttp, err)
	}

	if err = validate.StructCtx(ctx, &data); err != nil {
		return Version{}, errors.Join(ErrHttp, err)
	}

	return d.getVersion(ctx, data)
}

// GetAll implements Distribution.
func (d *paper) GetAll(ctx context.Context) ([]string, error) {
	var data paperManifest

	err := getreq(
		ctx,
		d.c,
		"https://fill.papermc.io/v3/projects/paper/versions",
		&data,
	)
	if err != nil {
		return nil, errors.Join(ErrHttp, err)
	}

	res := make([]string, len(data.Versions))
	for i := range data.Versions {
		res[i] = data.Versions[i].Version.ID
	}

	return res, nil
}

func (d *paper) getVersion(ctx context.Context, version paperVersion) (Version, error) {
	fetchUrl := fmt.Sprintf(
		"https://fill.papermc.io/v3/projects/paper/versions/%s/builds/latest",
		version.Version.ID,
	)

	var data paperBuild

	err := getreq(ctx, d.c, fetchUrl, &data)
	if err != nil {
		return Version{}, errors.Join(ErrHttp, err)
	}

	var (
		ok       bool
		download paperBuildDownload
	)
	for _, v := range data.Downloads {
		download, ok = v, true
	}

	if !ok {
		return Version{}, ErrVersionNotFound
	}
	javaVersion := normalizeJavaLts(version.Version.Java.Version.Minimum)

	htype := distropb.HashType_SHA256
	hash, err := hex.DecodeString(download.Checksums.SHA256)
	if err != nil {
		htype, hash = distropb.HashType_HASH_NONE, nil
	}

	return Version{
		ID:           version.Version.ID,
		URL:          download.URL,
		Hash:         hash,
		JVMArgs:      version.Version.Java.Flags.Recomended,
		HashType:     htype,
		Distribution: distropb.Distribution_PAPER,
		JavaVersion:  javaVersion,
	}, nil
}
