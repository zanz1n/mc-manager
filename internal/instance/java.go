package instance

import (
	"fmt"

	"github.com/docker/docker/api/types/strslice"
	"github.com/zanz1n/mc-manager/internal/pb"
)

type JavaVariant interface {
	GetImage(pb.JavaVersion) (string, error)
}

var _ JavaVariant = (*temurinJre)(nil)

type temurinJre struct {
	// noble, alpine, etc...
	distro string
}

func NewTemurinJre(distro string) JavaVariant {
	return &temurinJre{distro: distro}
}

// GetImage implements JavaVariant.
func (t *temurinJre) GetImage(v pb.JavaVersion) (string, error) {
	return fmt.Sprintf("eclipse-temurin:%d-jre-%s", v, t.distro), nil
}

func makeJavaCommand(jvmArgs []string, jarName string, ram uint64) strslice.StrSlice {
	cmd := make(strslice.StrSlice, 0, 8+len(jvmArgs))

	cmd = append(cmd,
		"java",
		"-Xms128M",
		fmt.Sprintf("-Xmx%dM", ram/MiB),
	)

	if len(jvmArgs) != 0 {
		cmd = append(cmd, jvmArgs...)
	}

	cmd = append(cmd,
		"-Dterminal.jline=false",
		"-Dterminal.ansi=true",
		"-jar",
		jarName,
		"nogui",
	)

	return cmd
}
