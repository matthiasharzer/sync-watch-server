package randomutil

import "time"

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func RandomString(length int) string {
	var seededRand = time.Now().UnixNano()
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand%int64(len(charset))]
		seededRand /= int64(len(charset))
	}
	return string(b)
}
