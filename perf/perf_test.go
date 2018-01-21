package perf

import (
	"github.com/gnulnx/goperf/request"
	"testing"
)

func BenchmarkGatherAllStatsGo(b *testing.B) {
	jsMap := map[string]*request.IterateReqResp{}
	cssMap := map[string]*request.IterateReqResp{}
	imgMap := map[string]*request.IterateReqResp{}
	fetchAllResp := request.FetchAll("https://qa.teaquinox.com", false)

	for i := 0; i < b.N; i++ {
		gatherAllStatsGo(fetchAllResp, jsMap, cssMap, imgMap)
	}
}

func BenchmarkGatherAllStats(b *testing.B) {
	jsMap := map[string]*request.IterateReqResp{}
	cssMap := map[string]*request.IterateReqResp{}
	imgMap := map[string]*request.IterateReqResp{}
	fetchAllResp := request.FetchAll("https://qa.teaquinox.com", false)

	for i := 0; i < b.N; i++ {
		gatherAllStats(fetchAllResp, jsMap, cssMap, imgMap)
	}
}
