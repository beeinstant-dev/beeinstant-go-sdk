package sender

import (
    "net/http"
    "time"
    "bytes"
    "strings"
    "fmt"
    "sort"
    "crypto/hmac"
    "crypto/sha256"
    "encoding/base64"
    "log"
    "io"
    "io/ioutil"
    "net/url"
)

type Sender struct {
    publicKey string
    secretKey string
    endpoint  string
    client    *http.Client
}

func NewSender(publicKey, secretKey, endpoint string) *Sender {
    s := new(Sender)
    s.publicKey = publicKey
    s.secretKey = secretKey
    s.endpoint = endpoint
    tr := &http.Transport{
        MaxIdleConnsPerHost: 1024,
        TLSHandshakeTimeout: 0 * time.Second,
    }
    s.client = &http.Client{Transport: tr}
    return s
}

func sign(message *bytes.Buffer, secret string) string {
    key := []byte(secret)
    h := hmac.New(sha256.New, key)
    h.Write(message.Bytes())
    return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func normalizeDimension(dimensions string) string {
    buf := new(bytes.Buffer)
    dArr := strings.Split(dimensions, ",")
    dMap := make(map[string]string)
    for _, d := range dArr {
        kv := strings.Split(d, "=")
        if len(kv) == 2 {
            key := strings.TrimSpace(strings.ToLower(kv[0]))
            value := strings.TrimSpace(kv[1])
            if len(key) > 0 && len(value) > 0 {
                dMap[key] = value
            }
        }
    }
    keys := make([]string, len(dMap))
    i := 0
    for key, _ := range dMap {
        keys[i] = key
        i += 1
    }
    sort.Strings(keys)

    for i, key := range keys {
        buf.WriteString("d." + key + "=" + dMap[key])
        if i < len(keys)-1 {
            buf.WriteByte(',')
        }
    }

    result := string(buf.Bytes())
    buf.Reset()

    return result
}

func normalizeMetricMap(metrics *map[string]map[string][]float64) {
    normalizedMap := make(map[string]map[string][]float64)

    for dimensions, names := range *metrics {
        nd := normalizeDimension(dimensions)
        if len(nd) > 0 {
            d := normalizedMap[nd]
            if d == nil {
                normalizedMap[nd] = make(map[string][]float64)
                d = normalizedMap[nd]
            }
            for name, value := range names {
                tn := strings.TrimSpace(name)
                if len(tn) > 0 {
                    _, ok := d[tn]
                    if ok == false {
                        d[tn] = make([]float64, 0)
                    }
                    if strings.HasSuffix(tn, "|c") && len(value) > 0 {
                        if len(d[tn]) == 0 {
                            d[tn] = append(d[tn], value[0])
                        } else {
                            d[tn][0] += value[0]
                        }
                    } else {
                        d[tn] = append(d[tn], value...)
                    }
                }
            }
        }
    }

    *metrics = normalizedMap
}

func serializeMetricMap(metrics map[string]map[string][]float64) (*bytes.Buffer, int) {
    numOfMetrics := 0
    buf := new(bytes.Buffer)
    for dimension, names := range metrics {
        buf.WriteString(dimension)
        for name, values := range names {
            buf.WriteString(",m.")
            nameAndUnit := strings.Split(name, "|")
            metricUnit := ""
            metricName := name
            if len(nameAndUnit) == 2 {
                metricName = nameAndUnit[0]
                metricUnit = nameAndUnit[1]
                if metricUnit == "c" {
                    metricUnit = ""
                }
            }
            buf.WriteString(metricName)
            numOfMetrics += 1
            buf.WriteByte('=')
            for idx, value := range values {
                buf.WriteString(fmt.Sprintf("%f", value))
                if idx == len(values)-1 {
                    buf.WriteString(metricUnit)
                } else {
                    buf.WriteByte('+')
                }
            }
        }
        buf.WriteByte('\n')
    }
    return buf, numOfMetrics
}

func (s *Sender) packAndPost(processingMap map[string]map[string][]float64) {
    normalizeMetricMap(&processingMap)
    serializedMapBuf, numOfMetrics := serializeMetricMap(processingMap)
    signature := sign(serializedMapBuf, s.secretKey)
    resp, err := s.client.Post(s.endpoint+
        "/PutMetric?signature="+ url.QueryEscape(signature)+
        "&publicKey="+ url.QueryEscape(s.publicKey)+
        "&timestamp="+ fmt.Sprint(time.Now().Unix()),
        "text/plain", serializedMapBuf)
    if err != nil {
        log.Println("err", err)
    } else {
        log.Println(fmt.Sprintf("Sent %d metrics to BeeInstant @ %s, publicKey %s", numOfMetrics, s.endpoint, s.publicKey))
        io.Copy(ioutil.Discard, resp.Body)
        resp.Body.Close()
    }
    serializedMapBuf.Reset()
}

func (s *Sender) Send(metrics *map[string]map[string][]float64, nonBlocking bool) {
    // swap maps
    processingMap := *metrics
    *metrics = make(map[string]map[string][]float64)

    if nonBlocking {
        go s.packAndPost(processingMap)
    } else {
        s.packAndPost(processingMap)
    }
}
