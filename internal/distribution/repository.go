package distribution

import (
	"context"

	"github.com/zanz1n/mc-manager/internal/pb"
)

type Repository struct {
	m map[pb.Distribution]Distribution
}

func NewRepository() *Repository {
	return &Repository{
		m: make(map[pb.Distribution]Distribution),
	}
}

func (r *Repository) AddDistribution(distro pb.Distribution, repo Distribution) {
	r.m[distro] = repo
}

func (r *Repository) GetLatest(
	ctx context.Context,
	distro pb.Distribution,
) (Version, error) {
	d, ok := r.m[distro]
	if !ok {
		return Version{}, ErrInvalidDistribution
	}
	return d.GetLatest(ctx)
}

func (r *Repository) GetVersion(
	ctx context.Context,
	distro pb.Distribution,
	semver string,
) (Version, error) {
	d, ok := r.m[distro]
	if !ok {
		return Version{}, ErrInvalidDistribution
	}
	return d.GetVersion(ctx, semver)
}

func (r *Repository) GetAll(
	ctx context.Context,
	distro pb.Distribution,
) ([]string, error) {
	d, ok := r.m[distro]
	if !ok {
		return nil, ErrInvalidDistribution
	}
	return d.GetAll(ctx)
}
