package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateAuthToken(t *testing.T) {
	for _, item := range []struct {
		pid      int64
		nid      int64
		key      string
		operator string
	}{
		{},
		{1, 2, "key", "operator"},
		{3546, 37848, "f7b03d90ca7c11ebb3fa69285ff09173", "admin"},
	} {
		token := CreateAuthToken(item.pid, item.nid, item.key, item.operator)
		t.Log(token)
		if item.pid != 0 && item.nid != 0 && len(item.key) != 0 {
			assert.NotEmpty(t, token)
		}
	}
}
