package can

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type TestSession struct {
	ID                string        `json:"id"`
	ExpiresInDuration time.Duration `json:"expires_in"`
}

func (ts TestSession) ExpiresIn() time.Duration {
	return ts.ExpiresInDuration
}

func (ts TestSession) Key() string {
	return ts.ID
}

type TestMarshaler struct {
}

func (t TestMarshaler) Marshal(session Session) ([]byte, error) {
	return json.Marshal(session)
}

func (t TestMarshaler) Unmarshal(bytes []byte) (Session, error) {
	var s TestSession
	if err := json.Unmarshal(bytes, &s); nil != err {
		return nil, err
	}
	return s, nil
}

var _ SessionMarshaler = &TestMarshaler{}

func TestConnect(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:13789",
	})
	repository := New(client, &TestMarshaler{})
	assert.NotNil(t, repository)
}

func TestSet(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:13789",
	})
	repository := New(client, &TestMarshaler{})
	id := "test1"
	ts := TestSession{
		ID: id,
	}
	assert.NoError(t, repository.Save(ts))
}

func TestGet(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:13789",
	})
	repository := New(client, &TestMarshaler{})
	id := "test-get"
	ts := TestSession{
		ID:                id,
		ExpiresInDuration: time.Minute,
	}
	assert.NoError(t, repository.Save(ts))
	load, err := repository.Load(id)
	assert.NoError(t, err)
	assert.NotNil(t, load)
	assert.Equal(t, id, load.Key())
	assert.Equal(t, time.Minute, load.ExpiresIn())
}

func TestGroupGet(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:13789",
	})
	repository := New(client, &TestMarshaler{})
	ts := TestSession{
		ID:                "test",
		ExpiresInDuration: time.Minute,
	}
	assert.NoError(t, repository.Save(ts))
	load, err := repository.Load("test")
	assert.NoError(t, err)
	assert.NotNil(t, load)
	assert.Equal(t, "test", load.Key())
	assert.Equal(t, time.Minute, load.ExpiresIn())
}

func TestExpire(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:13789",
	})
	repository := New(client, &TestMarshaler{})
	id := "test-expire"
	ts := TestSession{
		ID:                id,
		ExpiresInDuration: time.Second * 2,
	}
	expired := false
	repository.OnExpired(func(key string) {
		fmt.Println("id", key)
		if key == id {
			expired = true
		}
	})
	assert.NoError(t, repository.Save(ts))
	load, err := repository.Load(id)
	assert.NoError(t, err)
	assert.NotNil(t, load)
	assert.Equal(t, id, load.Key())
	assert.Equal(t, time.Second*2, load.ExpiresIn())

	time.Sleep(time.Second * 1)
	_ = repository.Save(TestSession{
		ID:                id,
		ExpiresInDuration: time.Second * 2,
	})
	time.Sleep(time.Second * 1)
	load0, err0 := repository.Load(id)
	assert.NoError(t, err0)
	assert.NotNil(t, load0)
	assert.Equal(t, id, load0.Key())
	assert.Equal(t, time.Second*2, load0.ExpiresIn())
	time.Sleep(time.Second * 3)

	_, err1 := repository.Load(id)
	assert.Error(t, err1)

	assert.Equal(t, true, expired)
	time.Sleep(time.Second * 1)
}
