package can

import (
	"context"
	"github.com/go-redis/redis/v8"
)

type Repository interface {
	Save(session Session) error
	Load(id string) (Session, error)
	SetMarshaler(marshaler SessionMarshaler)
}

type defaultRepository struct {
	redisClient *redis.Client
	marshaler   SessionMarshaler
}

func (d *defaultRepository) SetMarshaler(marshaler SessionMarshaler) {
	d.marshaler = marshaler
}

func (d *defaultRepository) Save(session Session) error {
	marshal, err := d.marshaler.Marshal(session)
	if nil != err {
		return err
	}
	// 기본 만료 시간 정해서 설정
	set := d.redisClient.Set(context.Background(), session.Key(), marshal, session.ExpiresIn())
	if set.Err() != nil {
		return set.Err()
	}
	return nil
}

func (d *defaultRepository) Load(id string) (Session, error) {
	get := d.redisClient.Get(context.Background(), id)
	if nil != get.Err() {
		return nil, get.Err()
	}
	if session, err := d.marshaler.Unmarshal([]byte(get.Val())); nil != err {
		return nil, err
	} else {
		return session, nil
	}
}

func New(redisClient *redis.Client, m SessionMarshaler) Repository {
	return &defaultRepository{
		redisClient: redisClient,
		marshaler:   m,
	}
}
