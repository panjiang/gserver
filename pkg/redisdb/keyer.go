package redisdb

import (
	"fmt"
	"strings"
)

// KeyPart .
type KeyPart string

// Keyer .
type Keyer struct {
	namespace string
}

// Gen 组合Key 当前服务的命名空间下
func (k *Keyer) Gen(key KeyPart, args ...interface{}) string {
	a := []string{k.namespace, string(key)}
	for _, arg := range args {
		a = append(a, fmt.Sprint(arg))
	}
	return strings.Join(a, ":")
}

// NewKeyer .
func NewKeyer(namespace string) *Keyer {
	return &Keyer{
		namespace: namespace,
	}
}

// GlobalKey 组合Key 全局
func GlobalKey(key KeyPart, args ...interface{}) string {
	a := []string{string(key)}
	for _, arg := range args {
		a = append(a, fmt.Sprint(arg))
	}
	return strings.Join(a, ":")
}
