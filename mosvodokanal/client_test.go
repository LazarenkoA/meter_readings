package mosvodokanal

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Auth(t *testing.T) {
	m := NewClient(context.Background(), "laz-mail@mail.ru", "CitadelClose1")
	err := m.Auth()
	assert.NoError(t, err)
}
