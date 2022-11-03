package goconf

import "fmt"

type ErrKeyNotFound struct {
	err string
}

func NewErrKeyNotFound(key string) ErrKeyNotFound {
	return ErrKeyNotFound{
		err: fmt.Sprintf("key not found: %s", key),
	}
}

func (e ErrKeyNotFound) Error() string {
	return e.err
}
