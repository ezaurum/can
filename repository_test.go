package can

import (
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type TestSession struct {
	ID               string        `json:"id"`
	ExpiresInSeconds time.Duration `json:"expires_in"`
}

func (ts TestSession) ExpiresIn() time.Duration {
	return ts.ExpiresInSeconds
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
	ts := TestSession{
		ID: "test",
	}
	assert.NoError(t, repository.Save(ts))
}

func TestGet(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:13789",
	})
	repository := New(client, &TestMarshaler{})
	ts := TestSession{
		ID:               "test",
		ExpiresInSeconds: time.Minute,
	}
	assert.NoError(t, repository.Save(ts))
	load, err := repository.Load("test")
	assert.NoError(t, err)
	assert.NotNil(t, load)
	assert.Equal(t, "test", load.Key())
	assert.Equal(t, time.Minute, load.ExpiresIn())
}
