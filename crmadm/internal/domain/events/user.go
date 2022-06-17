package events

import (
	"github.com/crxfoz/teaserad/crmadm/internal/domain/entity"
	"golang.org/x/crypto/bcrypt"
)

type NewUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
	MyRole   string `json:"-"`
}

func (nu *NewUser) CanCreateUsers() bool {
	return nu.Role == entity.RoleAdmin
}

func (nu *NewUser) HashedPassword() string {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(nu.Password), bcrypt.DefaultCost)
	return string(hashedPassword)
}

func (nu *NewUser) IsValide() bool {
	switch nu.Role {
	case entity.RoleAdmin, entity.RoleModerator:
		return true
	}

	return false
}
