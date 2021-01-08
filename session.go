package can

import "time"

type Session interface {
	Key() string
	ExpiresIn() time.Duration
}
