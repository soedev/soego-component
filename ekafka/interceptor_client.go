package ekafka

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"github.com/soedev/soego/core/eapp"
	"github.com/soedev/soego/core/elog"
	"github.com/soedev/soego/core/emetric"
	"github.com/soedev/soego/core/etrace"
	"github.com/soedev/soego/core/transport"
	"github.com/soedev/soego/core/util/xdebug"
	"github.com/soedev/soego/core/util/xstring"
	"github.com/spf13/cast"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"log"
	"strings"
	"time"
)

type ctxStartTimeKey struct{}

type clientProcessFn func(context.Context, Messages, *cmd) error

type cmd struct {
	name string
	res  interface{}
	msg  Message // 响应参数
}

type ClientInterceptor func(oldProcessFn clientProcessFn) (newProcessFn clientProcessFn)

func InterceptorClientChain(interceptors ...ClientInterceptor) ClientInterceptor {
	return func(p clientProcessFn) clientProcessFn {
		chain := p
		for i := len(interceptors) - 1; i >= 0; i-- {
			chain = buildInterceptor(interceptors[i], chain)
		}
		return chain
	}
}

func buildInterceptor(interceptor ClientInterceptor, oldProcess clientProcessFn) clientProcessFn {
	return interceptor(oldProcess)
}

func fixedClientInterceptor(_ string, _ *config) ClientInterceptor {
	return func(next clientProcessFn) clientProcessFn {
		return func(ctx context.Context, msgs Messages, cmd *cmd) error {
			start := time.Now()
			ctx = context.WithValue(ctx, ctxStartTimeKey{}, start)
			err := next(ctx, msgs, cmd)
			return err
		}
	}
}

func traceClientInterceptor(compName string, c *config) ClientInterceptor {
	tracer := etrace.NewTracer(trace.SpanKindProducer)
	return func(next clientProcessFn) clientProcessFn {
		return func(ctx context.Context, msgs Messages, cmd *cmd) error {
			carrier := propagation.MapCarrier{}
			ctx, span := tracer.Start(ctx, "kafka", carrier)
			defer span.End()

			headers := make([]kafka.Header, 0)
			for _, key := range carrier.Keys() {
				headers = append(headers, kafka.Header{
					Key:   key,
					Value: []byte(carrier.Get(key)),
				})
			}
			for _, value := range msgs {
				value.Headers = append(value.Headers, headers...)
				value.Time = time.Now()
			}
			err := next(ctx, msgs, cmd)

			span.SetAttributes(
				etrace.String("messaging.system", "kafka"),
				etrace.String("messaging.destination", cmd.msg.Topic),
			)

			return err
		}
	}
}

func accessClientInterceptor(compName string, c *config, logger *elog.Component) ClientInterceptor {
	return func(next clientProcessFn) clientProcessFn {
		return func(ctx context.Context, msgs Messages, cmd *cmd) error {
			loggerKeys := transport.CustomContextKeys()
			fields := make([]elog.Field, 0, 10+len(loggerKeys))

			if c.EnableAccessInterceptor {
				headers := make([]kafka.Header, 0)
				for _, key := range loggerKeys {
					if value := cast.ToString(transport.Value(ctx, key)); value != "" {
						fields = append(fields, elog.FieldCustomKeyValue(key, value))
						headers = append(headers, kafka.Header{
							Key:   key,
							Value: []byte(value),
						})
					}
				}
				for _, value := range msgs {
					value.Headers = append(value.Headers, headers...)
					value.Time = time.Now()
				}
			}

			err := next(ctx, msgs, cmd)
			cost := time.Since(ctx.Value(ctxStartTimeKey{}).(time.Time))
			if c.EnableAccessInterceptor {
				fields = append(fields,
					elog.FieldMethod(cmd.name),
					elog.FieldCost(cost),
				)

				// 开启了链路，那么就记录链路id
				if c.EnableTraceInterceptor && etrace.IsGlobalTracerRegistered() {
					fields = append(fields, elog.FieldTid(etrace.ExtractTraceID(ctx)))
				}
				if c.EnableAccessInterceptorReq {
					fields = append(fields, elog.Any("req", json.RawMessage(xstring.JSON(msgs.ToLog()))))
				}
				if c.EnableAccessInterceptorRes {
					fields = append(fields, elog.Any("res", json.RawMessage(xstring.JSON(cmd.res))))
				}
				logger.Info("access", fields...)
			}

			if !eapp.IsDevelopmentMode() {
				return err
			}
			if err != nil {
				log.Println("[ekafka.response]", xdebug.MakeReqResError(compName,
					fmt.Sprintf("%v", c.Brokers), cost, fmt.Sprintf("%s %v", cmd.name, xstring.JSON(msgs.ToLog())), err.Error()),
				)
			} else {
				log.Println("[ekafka.response]", xdebug.MakeReqResInfo(compName,
					fmt.Sprintf("%v", c.Brokers), cost, fmt.Sprintf("%s %v", cmd.name, xstring.JSON(msgs.ToLog())), fmt.Sprintf("%v", cmd.res)),
				)
			}
			return err
		}
	}
}

func metricClientInterceptor(compName string, config *config) ClientInterceptor {
	return func(next clientProcessFn) clientProcessFn {
		return func(ctx context.Context, msgs Messages, cmd *cmd) error {
			err := next(ctx, msgs, cmd)
			cost := time.Since(ctx.Value(ctxStartTimeKey{}).(time.Time))
			compNameTopic := fmt.Sprintf("%s.%s", compName, cmd.msg.Topic)
			emetric.ClientHandleHistogram.WithLabelValues("kafka", compNameTopic, cmd.name, strings.Join(config.Brokers, ",")).Observe(cost.Seconds())
			if err != nil {
				emetric.ClientHandleCounter.Inc("kafka", compNameTopic, cmd.name, strings.Join(config.Brokers, ","), "Error")
				return err
			}
			emetric.ClientHandleCounter.Inc("kafka", compNameTopic, cmd.name, strings.Join(config.Brokers, ","), "OK")
			return nil
		}
	}
}