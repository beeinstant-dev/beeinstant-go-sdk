package beeinstant_go_sdk

import (
    "testing"
    "fmt"
    "strings"
    "github.com/stretchr/testify/assert"
    "net/http/httptest"
    "net/http"
    "bytes"
    "strconv"
    "time"
)

func generateTestData(index int, m map[string]map[string][]float64) {
    m["   LoCation  = Dublin" + fmt.Sprint(index) + " , Service = GoGo  , aPi = PublishMetric "] = make(map[string][]float64)
    m["   Location  = Dublin" + fmt.Sprint(index) + " , Service = GoGo  , api = PublishMetric "] = make(map[string][]float64)
    m["   LoCation  = Dublin" + fmt.Sprint(index) + " , aPi = PublishMetric, Service = GoGo  ,  "] = make(map[string][]float64)
    m["   Service = GoGo  ,LocaTIon  = Dublin" + fmt.Sprint(index) + " ,  api = PublishMetric "] = make(map[string][]float64)
    m["   SERvice = GoGo ,, LoCation  = Dublin" + fmt.Sprint(index) + " , aPi = PublishMetric,   ,  "] = make(map[string][]float64)
    m["   LoCation  = Dublin" + fmt.Sprint(index) + " , aPi = PublishMetric, Service = GoGo  ,  "]["  NumOfSuccess|c "] = []float64{20}
    m["   Location  = Dublin" + fmt.Sprint(index) + " , Service = GoGo  , api = PublishMetric "]["  Latency|ms "] = []float64{250, 800, 9}
    m["   Service = GoGo  ,LocaTIon  = Dublin" + fmt.Sprint(index) + " ,  api = PublishMetric "]["   Latency|ms "] = []float64{700, 1000}
    m["   Service = GoGo  ,LocaTIon  = Dublin" + fmt.Sprint(index) + " ,  api = PublishMetric "]["    Latency|ms "] = make([]float64, 0)
    m["   SERvice = GoGo ,, LoCation  = Dublin" + fmt.Sprint(index) + " , aPi = PublishMetric,   ,  "]["   NumOfSuccess|c "] = []float64{30}
    m["   SERvice = GoGo ,, LoCation  = Dublin" + fmt.Sprint(index) + " , aPi = PublishMetric,   ,  "]["  NumOfSuccess|c "] = make([]float64, 0)
}

func TestNormalizeAndSerializeMetricMap(t *testing.T) {

    m := make(map[string]map[string][]float64)

    generateTestData(0, m)
    generateTestData(1, m)

    normalizeMetricMap(&m)
    buf, _ := serializeMetricMap(m)

    result := string(buf.Bytes())
    buf.Reset()

    metricLines := strings.Split(result, "\n")

    assert.Equal(t, 3, len(metricLines))
    assert.Equal(t, "", metricLines[2])

    b1 := strings.Contains(metricLines[0], "Dublin0") && strings.Contains(metricLines[1], "Dublin1")
    b2 := strings.Contains(metricLines[0], "Dublin1") && strings.Contains(metricLines[1], "Dublin0")

    assert.True(t, b1 || b2)

    for _, metricLine := range metricLines[:2] {
        arr := strings.Split(metricLine, ",")
        assert.Equal(t, 5, len(arr))
        assert.Equal(t, "d.api=PublishMetric", arr[0])
        assert.True(t, "d.location=Dublin1" == arr[1] || "d.location=Dublin0" == arr[1])
        assert.Equal(t, "d.service=GoGo", arr[2])
        for i := 3; i <= 4; i++ {
            if strings.HasPrefix(arr[i], "m.NumOfSuccess=") {
                assert.Equal(t, arr[i], "m.NumOfSuccess=50.000000")
            } else {
                assert.True(t, "m.Latency=250.000000+800.000000+9.000000+700.000000+1000.000000ms" == arr[i] ||
                    "m.Latency=700.000000+1000.000000+250.000000+800.000000+9.000000ms" == arr[i])
            }
        }
    }
}

func TestSender(t *testing.T) {
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        buf := new(bytes.Buffer)
        buf.ReadFrom(r.Body)
        bodyStr := buf.String()

        assert.Equal(t,"d.api=PublishMetric,d.location=Dublin,d.service=GoGo,m.NumOfSuccess=20.000000\n", bodyStr)
        assert.Equal(t,"XWiMy4zU9zBdWAZj0vaD4OCVdOmnt4N4YjetxFBxJy4=", r.URL.Query().Get("signature"))
        assert.Equal(t, "PUBLIC_KEY", r.URL.Query().Get("publicKey"))
        epochNow := int(time.Now().Unix())
        timestamp, _ := strconv.Atoi(r.URL.Query().Get("timestamp"))
        assert.True(t, timestamp >= epochNow - 5 && timestamp <= epochNow + 5)

        fmt.Fprintln(w, "OK")
    }))

    defer ts.Close()

    m := make(map[string]map[string][]float64)
    m["   LoCation  = Dublin , Service = GoGo  , aPi = PublishMetric "] = make(map[string][]float64)
    m["   LoCation  = Dublin , Service = GoGo  , aPi = PublishMetric "]["  NumOfSuccess|c "] = []float64{20}

    sender := NewSender("PUBLIC_KEY", "SECRET_KEY", ts.URL)
    sender.Send(&m, false)
    assert.Equal(t, 0, len(m))
}
