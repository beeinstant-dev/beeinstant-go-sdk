# BeeInstant Go SDK
BeeInstant Go SDK allows engineers to capture software **performance metrics** and **business metrics** directly from Go code. BeeInstant provides **multi-dimensional metrics** which can be aggregated with **full-spectrum of statistics** not only at host level, but also **at service level**. For example, engineers can put SLA on percentile 99th of Latency metric at service level across thousands of hosts/servers/containers. 

## Installation
```
go get -v github.com/beeinstant-dev/beeinstant-go-sdk
```

## Sample Usage

**Import**
```go
import (
    beeinstant "github.com/beeinstant-dev/beeinstant-go-sdk"
)
```

**Initialize MetricLogger**
```go
beeinstant.InitMetricLogger("MyServiceName",
        "MyEnvironment",
        "PublicKey",
        "PrivateKey",
        "Endpoint",
        10) // flush once every 10 seconds
```

**Counter**
```go
beeinstant.GetMetricLogger().IncCounter("MyCounter", 1)
```

**Timer**
```go
startTime := time.Now().UnixNano()
//
// processing works here
//
beeinstant.GetMetricLogger().Record("MyProcessingTime",
    float64((time.Now().UnixNano()-startTime)/1000000),
    MILLI_SECOND)
```

**Arbitrary Metrics with Units**
```go
beeinstant.GetMetricLogger().Record("MyPayload", 100, BYTE)
```

**Dimensions**

Add dimensions to bring more context to metrics.
```go
beeinstant.GetMetricLogger().
    ExtendDimensions("api=PublishMetrics,location=Dublin").
    IncCounter("MyCounter", 1)
```

Multiple metrics per group of dimensions
```go
metrics := beeinstant.GetMetricLogger().
    ExtendDimensions("api=PublishMetrics,location=Dublin")
    
metrics.IncCounter("MyCounter", 1)
metrics.Record("MyTimer", 100, MILLI_SECOND)
```
