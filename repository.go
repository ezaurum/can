package can

import (
	"context"
	"github.com/go-redis/redis/v8"
)

type Repository interface {
	Save(session Session) error
	Load(id string) (Session, error)
	SetMarshaler(marshaler SessionMarshaler)
	Close()
	OnExpired(handler SessionExpireHandler)
}

type SessionExpireHandler func(id string)

type defaultRepository struct {
	redisClient     *redis.Client
	marshaler       SessionMarshaler
	expireSubscribe *redis.PubSub
	expiredFunction func(string)
}

func (d *defaultRepository) OnExpired(handler SessionExpireHandler) {
	d.expiredFunction = handler
}

func (d *defaultRepository) Close() {
	if nil != d.expireSubscribe {
		_ = d.expireSubscribe.Close()
	}
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
	repository := defaultRepository{
		redisClient: redisClient,
		marshaler:   m,
	}
	expiredSub := repository.redisClient.PSubscribe(context.Background(), "__keyevent@*__:expired")
	ch := expiredSub.Channel()
	go func() {
		for {
			select {
			case msg := <-ch:
				if repository.expiredFunction != nil {
					repository.expiredFunction(msg.Payload)
				}
			}
		}
	}()
	return &repository
}
