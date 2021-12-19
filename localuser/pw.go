package localuser

import (
	"github.com/tvdburgt/go-argon2"
)

type PwHasher struct {
	salt []byte
	ctx  *argon2.Context
}

func NewPwHasher(salt []byte) PwHasher {
	return PwHasher{salt, argon2.NewContext()}
}

func (p *PwHasher) Hash(pw string) ([]byte, error) {
	return argon2.Hash(p.ctx, []byte(pw), p.salt)
}
func (p *PwHasher) Verify(hash []byte, pw string) (bool, error) {
	return argon2.Verify(p.ctx, hash, []byte(pw), p.salt)
}
