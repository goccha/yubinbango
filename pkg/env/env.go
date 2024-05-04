package env

import (
	"fmt"
	"strings"

	"github.com/goccha/envar"
)

func init() {
	_ = envar.Load()
	_env = &Env{}
	if err := envar.Bind(_env); err != nil {
		panic(err)
	}
}

func Setup(ver, rev string) {
	version = fmt.Sprintf("%s-%s", strings.ReplaceAll(ver, "/", "_"), rev)
}

var version string

func Version() string {
	return version
}

var _env *Env

type Env struct {
	DebugLog bool `env:"DEBUG_LOG"`
}

func DebugLog() bool {
	return _env.DebugLog
}
