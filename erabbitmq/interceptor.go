package erabbitmq

import (
	"encoding/json"
	"fmt"
	"github.com/soedev/soego/core/eapp"
	"github.com/soedev/soego/core/emetric"
	"github.com/soedev/soego/core/util/xdebug"
	"log"
	"time"
)

const (
	metricType = "rabbitmq"
)

type Interceptor func(oldProcessFn processFn) (newProcessFn processFn)

func InterceptorChain(interceptors ...Interceptor) func(oldProcess processFn) processFn {
	build := func(interceptor Interceptor, oldProcess processFn) processFn {
		return interceptor(oldProcess)
	}

	return func(oldProcess processFn) processFn {
		chain := oldProcess
		for i := len(interceptors) - 1; i >= 0; i-- {
			chain = build(interceptors[i], chain)
		}
		return chain
	}
}

func debugInterceptor(compName string, c *config) func(processFn) processFn {
	return func(oldProcess processFn) processFn {
		return func(cmd *cmd) error {
			if !eapp.IsDevelopmentMode() {
				return oldProcess(cmd)
			}
			beg := time.Now()
			err := oldProcess(cmd)
			cost := time.Since(beg)
			if err != nil {
				log.Println("[erabbitmq.response]", xdebug.MakeReqResError(compName,
					fmt.Sprintf("%v", c.Url), cost, fmt.Sprintf("%s %v", cmd.name, mustJsonMarshal(cmd.req)), err.Error()),
				)
			} else {
				log.Println("[erabbitmq.response]", xdebug.MakeReqResInfo(compName,
					fmt.Sprintf("%v", c.Url), cost, fmt.Sprintf("%s %v", cmd.name, mustJsonMarshal(cmd.req)), fmt.Sprintf("%v", cmd.res)),
				)
			}
			return err
		}
	}
}

func metricInterceptor(compName string, c *config) func(processFn) processFn {
	return func(oldProcess processFn) processFn {
		return func(cmd *cmd) error {
			beg := time.Now()
			err := oldProcess(cmd)
			cost := time.Since(beg)
			if err != nil {
				emetric.ClientHandleCounter.Inc(metricType, compName, cmd.name, c.Url, "Error")
			} else {
				emetric.ClientHandleCounter.Inc(metricType, compName, cmd.name, c.Url, "OK")
			}
			emetric.ClientHandleHistogram.WithLabelValues(metricType, compName, cmd.name, c.Url).Observe(cost.Seconds())
			return err
		}
	}
}

func mustJsonMarshal(val interface{}) string {
	res, _ := json.Marshal(val)
	return string(res)
}
