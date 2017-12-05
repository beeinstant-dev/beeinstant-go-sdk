package metriclogger

import (
    "../sender"
    "sync"
    "strings"
    "time"
    "math"
)

type MetricCollector interface {
    incCounter(metricName string, value int)
    record(metricName string, value float64, unit string)
}

type MetricLogger interface {
    MetricCollector
    extendDimensions(dimensions string) MetricCollector
}

type MetricMessage struct {
    dimensions string
    metric     string
    unit       string
    value      float64
}

type MetricsNoOp struct{}

type MetricsReal struct {
    dimensions   string
    metricLogger *MetricLoggerReal
}

type MetricLoggerNoOp struct{}

type MetricLoggerReal struct {
    rootDimensions string
    sender         *sender.Sender
    flushInterval  int64
    metricChannel  chan MetricMessage
    metrics        map[string]map[string][]float64
}

var (
    metricLogger MetricLogger = new(MetricLoggerNoOp)
    once         sync.Once
)

func InitMetricLogger(service, env, publicKey, secretKey, endpoint string, flushInterval int) MetricLogger {
    if flushInterval <= 0 {
        panic("Flush interval must be > 0")
    }

    once.Do(func() {
        rootDimensions := ""
        trimmedService := strings.TrimSpace(service)
        trimmedEnv := strings.TrimSpace(env)
        if len(trimmedService) > 0 {
            rootDimensions += "service=" + trimmedService + ","
        }
        if len(trimmedEnv) > 0 {
            rootDimensions += "env=" + trimmedEnv + ","
        }

        metricLoggerReal := &MetricLoggerReal{
            rootDimensions: rootDimensions,
            sender:         sender.NewSender(publicKey, secretKey, endpoint),
            flushInterval:  int64(flushInterval) * 1000,
            metricChannel:  make(chan MetricMessage, 100),
            metrics:        make(map[string]map[string][]float64),
        }
        metricLogger = metricLoggerReal

        go metricLoggerReal.processing()
    })
    return metricLogger
}

func GetMetricLogger() MetricLogger {
    return metricLogger
}

// MetricLogger real
func (metricLogger *MetricLoggerReal) processing()  {
    flushed := false
    t := time.Now().UnixNano() / 1000000

    for {
        select {
        case metricMessage := <-metricLogger.metricChannel:
            metricLogger.updateMetricMap(&metricMessage)
        case <-time.After(time.Millisecond * 500):
        }

        if int64(time.Now().UnixNano() / 1000000) % metricLogger.flushInterval <
            int64(math.Max(1, float64(metricLogger.flushInterval / 2))) {
            if flushed == false && time.Now().UnixNano() / 1000000 - t > 0 {
                metricLogger.sender.Send(&metricLogger.metrics, true)
                t = time.Now().UnixNano() / 1000000
                flushed = true
            }
        } else {
            flushed = false
        }
    }
}

func (metricLogger *MetricLoggerReal) updateMetricMap(message *MetricMessage) {
    dimensions := metricLogger.metrics[message.dimensions]
    if dimensions == nil {
        metricLogger.metrics[message.dimensions] = make(map[string][]float64)
        dimensions = metricLogger.metrics[message.dimensions]
    }
    metric := strings.TrimSpace(message.metric) + "|" + strings.ToLower(strings.TrimSpace(message.unit))
    m := dimensions[metric]
    if m == nil {
        dimensions[metric] = []float64{message.value}
    } else {
        if message.unit == "c" {
            m[0] += message.value
        } else {
            dimensions[metric] = append(m, message.value)
        }
    }
}

func (metricLogger *MetricLoggerReal) extendDimensions(dimensions string) MetricCollector {
    return &MetricsReal{dimensions: dimensions + ",", metricLogger: metricLogger}
}

func (metricLogger *MetricLoggerReal) incCounter(metricName string, value int) {
    if value >= 0 {
        metricLogger.metricChannel <- MetricMessage{
            dimensions: metricLogger.rootDimensions,
            metric:     metricName,
            unit:       "c",
            value:      float64(value),
        }
    }
}

func (metricLogger *MetricLoggerReal) record(metricName string, value float64, unit string) {
    if value >= 0 {
        metricLogger.metricChannel <- MetricMessage{
            dimensions: metricLogger.rootDimensions,
            metric:     metricName,
            unit:       unit,
            value:      value,
        }
    }
}

// Metrics real
func (metrics *MetricsReal) incCounter(metricName string, value int) {
    if value >= 0 {
        metrics.metricLogger.metricChannel <- MetricMessage{
            dimensions: metrics.metricLogger.rootDimensions + metrics.dimensions,
            metric:     metricName,
            unit:       "c",
            value:      float64(value),
        }
    }
}

func (metrics *MetricsReal) record(metricName string, value float64, unit string) {
    if value >= 0 {
        metrics.metricLogger.metricChannel <- MetricMessage{
            dimensions: metrics.metricLogger.rootDimensions + metrics.dimensions,
            metric:     metricName,
            unit:       unit,
            value:      value,
        }
    }
}

// MetricLogger no-op
func (metricLogger *MetricLoggerNoOp) extendDimensions(dimensions string) MetricCollector {
    return new(MetricsNoOp)
}
func (metricLogger *MetricLoggerNoOp) incCounter(metricName string, value int) {}
func (metricLogger *MetricLoggerNoOp) record(metricName string, value float64, unit string) {}

// Metrics no-op
func (metrics *MetricsNoOp) incCounter(metricName string, value int) {}
func (metrics *MetricsNoOp) record(metricName string, value float64, unit string) {}

// Units
const NANO_SECOND = "ns"
const MICRO_SECOND = "us"
const MILLI_SECOND = "ms"
const SECOND = "s"
const MINUTE = "m"
const HOUR = "h"
const BYTE = "b"
const KILO_BYTE = "kb"
const MEGA_BYTE = "mb"
const GIGA_BYTE = "gb"
const TERA_BYTE = "tb"
const BIT_PER_SEC = "bps"
const KILO_BIT_PER_SEC = "kbps"
const MEGA_BIT_PER_SEC = "mbps"
const GIGA_BIT_PER_SEC = "gbps"
const TERA_BIT_PER_SEC = "tbps"
const PERCENT = "p"
const NONE = ""
