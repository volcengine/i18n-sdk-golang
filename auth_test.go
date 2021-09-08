package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateAuthToken(t *testing.T) {
	for _, item := range []struct{
		key      string
		project  string
		operator string
	}{
		{"", "", ""},
		{"a", "project", "operator"},
		{"a", "project", "operator"},
		{"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9", "Starling", "admin"},
	} {
		token := CreateAuthToken(item.key, item.project, item.operator)
		t.Log(token)
		assert.NotEmpty(t, token)
	}
}
