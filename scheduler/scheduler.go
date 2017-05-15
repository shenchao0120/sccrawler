package scheduler

import "net/http"

type GenHttpClient func() *http.Client

