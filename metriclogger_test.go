package beeinstant_go_sdk

import (
    "testing"
    "time"
    "fmt"
    "net/http/httptest"
    "bytes"
    "net/http"
    "strings"
    "regexp"
    "strconv"
    "github.com/stretchr/testify/assert"
)

func TestMetricLoggerNoOp(t *testing.T) {
    metrics := GetMetricLogger().ExtendDimensions("location=Dublin")
    metrics.Record("ProcessingTime", 1, MILLI_SECOND)
    metrics.IncCounter("MyCounter", 1)
}

func TestMetricLoggerReal(t *testing.T) {
    myCounterRoot := 0
    myTimerRoot := 0
    myCounter := 0
    myTimer := 0

    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        buf := new(bytes.Buffer)
        buf.ReadFrom(r.Body)
        bodyStr := buf.String()

        metricLines := strings.Split(bodyStr, "\n")

        for _, metricLine := range metricLines {

            if strings.HasPrefix(metricLine, "d.api=PublishMetrics,d.env=Test,d.location=Dublin,d.service=GolangMonitoring") {

                re := regexp.MustCompile("m.MyTimer=(.+)ms")
                matches := re.FindStringSubmatch(metricLine)
                if len(matches) >= 2 {
                    for _, num := range strings.Split(matches[1], "+") {
                        n, _ := strconv.ParseFloat(num, 32)
                        myTimer += int(n)
                    }
                }

                re = regexp.MustCompile("m.MyCounter=(\\d+\\.\\d+)")
                matches = re.FindStringSubmatch(metricLine)
                if len(matches) >= 2 {
                    n, _ := strconv.ParseFloat(matches[1], 32)
                    myCounter += int(n)
                }

            } else if strings.HasPrefix(metricLine, "d.env=Test,d.service=GolangMonitoring") {

                re := regexp.MustCompile("m.MyTimerRoot=(.+)ms")
                matches := re.FindStringSubmatch(metricLine)
                if len(matches) >= 2 {
                    for _, num := range strings.Split(matches[1], "+") {
                        n, _ := strconv.ParseFloat(num, 32)
                        myTimerRoot += int(n)
                    }
                }

                re = regexp.MustCompile("m.MyCounterRoot=(\\d+\\.\\d+)")
                matches = re.FindStringSubmatch(metricLine)
                if len(matches) >= 2 {
                    n, _ := strconv.ParseFloat(matches[1], 32)
                    myCounterRoot += int(n)
                }
            }
        }

        fmt.Fprintln(w, "OK")
    }))

    InitMetricLogger("GolangMonitoring",
        "Test",
        "PUBLIC_KEY",
        "SECRET_KEY",
        ts.URL,
        2)

    for i := 0; i < 10; i++ {
        GetMetricLogger().IncCounter("MyCounterRoot", 1)
        GetMetricLogger().Record("MyTimerRoot", 100, MILLI_SECOND)
        GetMetricLogger().
            ExtendDimensions("api=PublishMetrics,location=Dublin").
            IncCounter("MyCounter", 2)
        GetMetricLogger().
            ExtendDimensions("api=PublishMetrics,location=Dublin").
            Record("MyTimer", 200, MILLI_SECOND)
        time.Sleep(500 * time.Millisecond)
    }

    time.Sleep(3 * time.Second)

    assert.Equal(t, 20, myCounter)
    assert.Equal(t, 10, myCounterRoot)
    assert.Equal(t, 2000, myTimer)
    assert.Equal(t, 1000, myTimerRoot)
}
