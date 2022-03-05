package localuser

import (
	"github.com/matthewhartstonge/argon2"
)

type PwHasher struct {
	argon argon2.Config
}

func NewPwHasher(_salt []byte) PwHasher {
	argon := argon2.DefaultConfig()
	return PwHasher{argon}
}

func (p *PwHasher) Hash(pw string) ([]byte, error) {
	return p.argon.HashEncoded([]byte(pw))
}
func (p *PwHasher) Verify(hash []byte, pw string) (bool, error) {
	return argon2.VerifyEncoded([]byte(pw), hash)
}
