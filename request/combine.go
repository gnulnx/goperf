package request

import (
	"time"
)

// Combine the slice of IterateReqRespAll structs into a single IterateReqRespAll
func Combine(results []IterateReqRespAll) *IterateReqRespAll {

	totalReqs := 0
	baseStatus := []int{}
	baseRespTimes := []time.Duration{}
	baseBytes := 0
	jsResps := map[string][]IterateReqResp{}
	cssResps := map[string][]IterateReqResp{}
	imgResps := map[string][]IterateReqResp{}

	var totalAvglRespTimes int64
	var totalAvgLinearlRespTimes int64
	var count int64
	for _, resp := range results {
		totalReqs += len(resp.BaseURL.Status)
		baseStatus = append(baseStatus, resp.BaseURL.Status...)
		baseRespTimes = append(baseRespTimes, resp.BaseURL.RespTimes...)
		baseBytes += resp.BaseURL.Bytes
		totalAvglRespTimes += int64(resp.AvgTotalRespTime)
		totalAvgLinearlRespTimes += int64(resp.AvgTotalLinearRespTime)
		count++

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
	avgTotalLinearRespTimes := time.Duration(totalAvgLinearlRespTimes / count)

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
		AvgTotalRespTime:       avgTotalRespTimes,
		AvgTotalLinearRespTime: avgTotalLinearRespTimes,
		BaseURL: IterateReqResp{
			Url:         results[0].BaseURL.Url,
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
