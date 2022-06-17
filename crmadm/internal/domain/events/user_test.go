package events

import (
	"testing"

	"github.com/crxfoz/teaserad/crmadm/internal/domain/entity"
	"github.com/stretchr/testify/assert"
)

func TestNewUser_HashedPassword(t *testing.T) {
	pwd := "admin123"

	newUser := &NewUser{
		Username: "",
		Password: pwd,
		Role:     "",
		MyRole:   "",
	}

	hashedPwd := newUser.HashedPassword()

	t.Log("Pwd:", hashedPwd)

	againstUser := entity.User{
		ID:        0,
		Username:  "",
		Password:  hashedPwd,
		Role:      "",
		CreatedAt: 0,
	}

	assert.Nil(t, againstUser.CheckPassword(pwd))
}
