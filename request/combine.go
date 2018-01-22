package request

import (
	"time"
)

func Combine(results []IterateReqRespAll) *IterateReqRespAll {
	/*
		Combine the slice of IterateReqRespAll structs into a single IterateReqRespAll
	*/
	totalReqs := 0
	baseStatus := []int{}
	baseRespTimes := []time.Duration{}
	baseBytes := 0
	jsResps := map[string][]IterateReqResp{}
	cssResps := map[string][]IterateReqResp{}
	imgResps := map[string][]IterateReqResp{}

	var totalAvglRespTimes int64 = 0
	var count int64 = 0
	for _, resp := range results {
		totalReqs += len(resp.BaseUrl.Status)
		baseStatus = append(baseStatus, resp.BaseUrl.Status...)
		baseRespTimes = append(baseRespTimes, resp.BaseUrl.RespTimes...)
		baseBytes += resp.BaseUrl.Bytes
		totalAvglRespTimes += int64(resp.AvgTotalRespTime)
		count += 1

		for _, jsresp := range resp.JSResps {
			jsResps[jsresp.Url] = append(jsResps[jsresp.Url], jsresp)
		}
		for _, cssresp := range resp.CSSResps {
			cssResps[cssresp.Url] = append(cssResps[cssresp.Url], cssresp)
		}
		for _, imgresp := range resp.IMGResps {
			imgResps[imgresp.Url] = append(imgResps[imgresp.Url], imgresp)
		}
	}

	avgTotalRespTimes := time.Duration(totalAvglRespTimes / count)

	combine := func(resps map[string][]IterateReqResp) []IterateReqResp {
		allResps := []IterateReqResp{}
		for k, v := range resps {
			status := []int{}
			respTimes := []time.Duration{}
			bytes := 0
			for _, resp := range v {
				status = append(status, resp.Status...)
				respTimes = append(respTimes, resp.RespTimes...)
				bytes += resp.Bytes
			}
			allResps = append(allResps, IterateReqResp{
				Url:         k,
				Status:      status,
				RespTimes:   respTimes,
				NumRequests: len(status),
				Bytes:       bytes,
			})
		}
		return allResps
	}

	return &IterateReqRespAll{
		AvgTotalRespTime: avgTotalRespTimes,
		BaseUrl: IterateReqResp{
			Url:         results[0].BaseUrl.Url,
			Status:      baseStatus,
			RespTimes:   baseRespTimes,
			NumRequests: len(baseStatus),
			Bytes:       baseBytes,
		},
		JSResps:  combine(jsResps),
		CSSResps: combine(cssResps),
		IMGResps: combine(imgResps),
	}
}
