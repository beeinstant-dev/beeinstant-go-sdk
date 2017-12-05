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
        10)
```

**Counter**
```go
GetMetricLogger().incCounter("MyCounter", 1)
```

**Timer and Metrics with Units**
```go
GetMetricLogger().record("MyTimer", 100, MILLI_SECOND)
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
