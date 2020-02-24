package joe

import (
	"bytes"
	"encoding/gob"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestStorage(t *testing.T) {
	logger := zaptest.NewLogger(t)
	store := NewStorage(logger)

	ok, err := store.Get("test", nil)
	assert.NoError(t, err)
	assert.False(t, ok)

	err = store.Set("test", "foo")
	assert.NoError(t, err)

	err = store.Set("test", "foo")
	assert.NoError(t, err, "setting a key more than once should not error")

	err = store.Set("test-2", "bar")
	assert.NoError(t, err)

	keys, err := store.Keys()
	assert.NoError(t, err)
	assert.Equal(t, []string{"test", "test-2"}, keys)

	ok, err = store.Get("test", nil)
	assert.NoError(t, err, "getting a key without a target to unmarshal to should not fail")
	assert.True(t, ok)

	var val string
	ok, err = store.Get("test", &val)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, "foo", val)

	ok, err = store.Delete("does-not-exist")
	assert.NoError(t, err)
	assert.False(t, ok)

	ok, err = store.Delete("test")
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = store.Get("test", nil)
	assert.NoError(t, err)
	assert.False(t, ok)

	assert.NoError(t, store.Close())
}

func TestStorage_Encoder(t *testing.T) {
	logger := zaptest.NewLogger(t)
	enc := new(gobEncoder)
	store := NewStorage(logger)
	store.SetMemoryEncoder(enc)

	val := []string{"foo", "bar"}
	err := store.Set("test", val)
	require.NoError(t, err)

	var actual []string
	ok, err := store.Get("test", &actual)
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, val, actual)
}

func TestStorage_EncoderErrors(t *testing.T) {
	logger := zaptest.NewLogger(t)
	enc := new(gobEncoder)

	store := NewStorage(logger)
	store.SetMemoryEncoder(enc)

	err := store.Set("test", "ok")
	require.NoError(t, err, "should insert the first value without an error")

	enc.encodeErr = errors.New("something went wrong")
	err = store.Set("test", "foo")
	assert.EqualError(t, err, "encode data: something went wrong")

	var actual []string
	enc.decodeErr = errors.New("this did not work")
	ok, err := store.Get("test", &actual)
	assert.EqualError(t, err, "decode data: this did not work")
	assert.False(t, ok)
}

// gobEncoder is an example of a different encoder. This is not part of joe to
// avoid the extra import in production code.
type gobEncoder struct {
	encodeErr error
	decodeErr error
}

func (e gobEncoder) Encode(value interface{}) ([]byte, error) {
	if err := e.encodeErr; err != nil {
		return nil, err
	}

	data := new(bytes.Buffer)
	enc := gob.NewEncoder(data)
	err := enc.Encode(value)
	return data.Bytes(), err
}

func (e gobEncoder) Decode(data []byte, target interface{}) error {
	if err := e.decodeErr; err != nil {
		return err
	}

	enc := gob.NewDecoder(bytes.NewBuffer(data))
	return enc.Decode(target)
}
