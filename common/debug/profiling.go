package debug

import (
	"net/http"
	_ "net/http/pprof"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/debug/log"
)

func StartAsyncProfiling() {
	go func() {
		if err := http.ListenAndServe("localhost:8082", nil); err != nil {
			log.Error(errors.Wrap(err, "Error", "Profiling stopped"))
		}
	}()
}
