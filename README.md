# BeeInstant Go SDK
BeeInstant Go SDK allows engineers to capture software performance metrics and business metrics directly from Go code.

## Sample Usage
**Initialize MetricLogger**
```go
InitMetricLogger("MyServiceName",
        "MyEnvironment",
        "PublicKey",
        "PrivateKey",
        "Endpoint",
        10) // flush once every 10 seconds
```

**Counter**
```go
GetMetricLogger().incCounter("MyCounter", 1)
```

**Timer**
```go
startTime := time.Now().UnixNano()
//
// processing works here
//
GetMetricLogger().record("MyProcessingTime",
    float64((time.Now().UnixNano()-startTime)/1000000),
    MILLI_SECOND)
```

**Arbitrary Metrics with Units**
```go
GetMetricLogger().record("MyPayload", 100, BYTE)
```

**Dimensions**

Add dimensions to bring more context to metrics.
```go
GetMetricLogger().
    extendDimensions("api=PublishMetrics,location=Dublin").
    incCounter("MyCounter", 1)
```

Multiple metrics per group of dimensions
```go
metrics := GetMetricLogger().
    extendDimensions("api=PublishMetrics,location=Dublin")
    
metrics.incCounter("MyCounter", 1)
metrics.record("MyTimer", 100, MILLI_SECOND)
```
