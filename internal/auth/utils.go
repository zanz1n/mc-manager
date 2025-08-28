package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"

	"github.com/zanz1n/mc-manager/internal/dto"
)

const refreshTokenLen = 64

var (
	base64d = base64.StdEncoding
	// Size of the refresh token string.
	RefreshTokenLen = base64d.EncodedLen(refreshTokenLen)
)

func decodeRefreshToken(ts string) ([]byte, error) {
	token, err := base64d.DecodeString(ts)
	if err != nil || len(token) != refreshTokenLen {
		err = ErrInvalidRefreshToken
	}

	return token, err
}

func generateRefreshToken(userId dto.Snowflake) string {
	token := make([]byte, refreshTokenLen)
	rand.Read(token)

	binary.LittleEndian.PutUint64(token, uint64(userId))

	return base64d.EncodeToString(token)
}

func getRefreshTokenUser(token []byte) dto.Snowflake {
	return dto.Snowflake(binary.LittleEndian.Uint64(token[0:8]))
}
