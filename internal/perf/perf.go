package perf

import (
	"context"
	"flag"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/influxdata/tdigest"
	"gitlab.com/Sh00ty/hootydb/internal"
)

type Ammo struct {
	Body    io.Reader
	Headers http.Header
	Path    string
	Method  string
}

type AmmoGen interface {
	GetAmmo() Ammo
}

var (
	concurrency int
	rps         int
	count       int
	urlsStr     string
)

const (
	bufSz = 10e4
)

func parseFlags() {
	flag.IntVar(&concurrency, "c", 4, "workers to load tests")
	flag.IntVar(&rps, "rps", 10, "rps to load")
	flag.IntVar(&count, "count", math.MaxInt, "maximum count of request")
	flag.StringVar(&urlsStr, "urls", "http://127.0.0.1:8000;http://127.0.0.1:8001;http://127.0.0.1:8002", "urls of db in format: url1;url2;url3")
	flag.Parse()
}

func NewPerfTesting(ctx context.Context, gen AmmoGen, log internal.Logger) {
	parseFlags()
	urls, err := parseUrls(urlsStr)
	if err != nil {
		log.Fatalf(ctx, err.Error())
	}

	var (
		td = guardedTdigest{
			td: *tdigest.New(),
			mu: sync.Mutex{},
		}
		reqStats = map[int]*atomic.Uint32{
			0:   {},
			200: {},
			400: {},
			500: {},
		}
		ch = make(chan *http.Request, bufSz)
		wg = sync.WaitGroup{}
	)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			clnt := http.Client{
				Timeout: time.Second,
			}
			for req := range ch {
				func() {
					ts := time.Now()
					resp, err := clnt.Do(req)
					td.Add(float64(time.Since(ts).Milliseconds()))
					if err != nil {
						c := reqStats[0]
						c.Add(1)
						log.Errorf(ctx, err, "request error")
						return
					}
					defer resp.Body.Close()

					status := resp.StatusCode
					status = status - status%100
					c := reqStats[status]
					c.Add(1)
					if status != 200 {
						res, err := io.ReadAll(resp.Body)
						if err != nil {
							log.Errorf(ctx, err, "failed to read body")
							return
						}
						log.ErrorMessage(ctx, "got not ok resp: %s", string(res))
					}
				}()
			}
		}()
	}

	var (
		ticker = time.NewTicker(time.Second / time.Duration(rps))
		k      = 0
	)
	for range ticker.C {
		if k >= count {
			ticker.Stop()
			break
		}
		ammo := gen.GetAmmo()
		url := urls.getNext()
		url = url.JoinPath(ammo.Path)
		req, err := http.NewRequest(ammo.Method, url.String(), ammo.Body)
		if err != nil {
			log.Fatalf(ctx, "failed to generate request from ammo: %v", err)
		}
		req.Header = ammo.Headers.Clone()
		ch <- req
		k++
	}
	close(ch)
	wg.Wait()

	reqStatsMsg := "request stats:\n"
	percetiles := []float64{0.5, 0.75, 0.9, 0.99, 0.999}
	tdRes := td.Percentile(percetiles)
	for i, p := range tdRes {
		reqStatsMsg += fmt.Sprintf("%.2f ---> %d\n", percetiles[i], int(p))
	}

	reqStatsMsg += "\nstatus codes stats:\n"
	sum := 0
	for _, st := range reqStats {
		sum += int(st.Load())
	}
	for code, st := range reqStats {
		reqStatsMsg += fmt.Sprintf("\n%d ---> %.2f", code, float64(st.Load())/float64(sum))
	}

	log.ErrorMessage(ctx, "work finished!!\n%s", reqStatsMsg)
}

type urlsBalancer struct {
	urls []*url.URL
	ptr  atomic.Uint64
}

func (b *urlsBalancer) getNext() *url.URL {
	return b.urls[int(b.ptr.Add(1))%len(b.urls)]
}

func parseUrls(urlsStr string) (*urlsBalancer, error) {
	urlsStr = strings.TrimSpace(urlsStr)
	urlsStrs := strings.Split(urlsStr, ";")

	urls := make([]*url.URL, 0, len(urlsStrs))
	for _, uStr := range urlsStrs {
		u, err := url.Parse(uStr)
		if err != nil {
			return nil, fmt.Errorf("url %s parse error: %v", u, err)
		}
		urls = append(urls, u)
	}
	return &urlsBalancer{urls: urls}, nil
}
