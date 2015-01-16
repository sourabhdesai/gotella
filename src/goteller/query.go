package goteller

import (
	"net/http"
)

type Query struct {
	TTL         byte
	MinSpeed    uint16
	SearchQuery string
	onHit       func([]QueryResult, uint32, string) []QueryResult
	onResponse  func(error, uint32, string, *http.Response)
}

func (query *Query) OnHit(onhit func([]QueryResult, uint32, string) []QueryResult) {
	if onhit != nil {
		query.onHit = onhit
	}
}

func (query *Query) OnResponse(onres func(error, uint32, string, *http.Response)) {
	if onres != nil {
		query.onResponse = onres
	}
}
