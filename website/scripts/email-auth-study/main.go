// Scans SPF, DMARC, BIMI, MTA-STS and MX for the first N domains of a Tranco CSV.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Result struct {
	Rank   int      `json:"rank"`
	Domain string   `json:"domain"`
	MX     bool     `json:"mx"`
	NullMX bool     `json:"nullMx"`
	SPF    []string `json:"spf"`
	DMARC  []string `json:"dmarc"`
	BIMI   []string `json:"bimi"`
	STS    []string `json:"sts"`
	Errors []string `json:"errors,omitempty"`
}

var servers = []string{"1.1.1.1:53", "8.8.8.8:53", "9.9.9.9:53"}
var n atomic.Uint64

func resolver() *net.Resolver {
	return &net.Resolver{PreferGo: true, Dial: func(ctx context.Context, network, _ string) (net.Conn, error) {
		d := net.Dialer{Timeout: 4 * time.Second}
		return d.DialContext(ctx, network, servers[n.Add(1)%uint64(len(servers))])
	}}
}

func notFound(err error) bool {
	var de *net.DNSError
	return errors.As(err, &de) && de.IsNotFound
}

func txt(r *net.Resolver, name, prefix string) ([]string, error) {
	var out []string
	var err error
	for try := 0; try < 3; try++ {
		ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
		var recs []string
		recs, err = r.LookupTXT(ctx, name)
		cancel()
		if err == nil || notFound(err) {
			for _, t := range recs {
				if strings.HasPrefix(strings.ToLower(strings.TrimSpace(t)), prefix) {
					out = append(out, t)
				}
			}
			return out, nil
		}
		time.Sleep(time.Duration(300*(try+1)) * time.Millisecond)
	}
	return nil, err
}

func main() {
	limit, _ := strconv.Atoi(os.Args[2])
	f, _ := os.Open(os.Args[1])
	sc := bufio.NewScanner(f)
	type job struct {
		rank   int
		domain string
	}
	jobs := make(chan job)
	out := make(chan Result)
	var wg sync.WaitGroup
	workers, _ := strconv.Atoi(os.Getenv("WORKERS"))
	if workers == 0 {
		workers = 64
	}
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r := resolver()
			for j := range jobs {
				res := Result{Rank: j.rank, Domain: j.domain}
				var err error
				if res.SPF, err = txt(r, j.domain, "v=spf1"); err != nil {
					res.Errors = append(res.Errors, "spf")
				}
				if res.DMARC, err = txt(r, "_dmarc."+j.domain, "v=dmarc1"); err != nil {
					res.Errors = append(res.Errors, "dmarc")
				}
				if res.BIMI, err = txt(r, "default._bimi."+j.domain, "v=bimi1"); err != nil {
					res.Errors = append(res.Errors, "bimi")
				}
				if res.STS, err = txt(r, "_mta-sts."+j.domain, "v=stsv1"); err != nil {
					res.Errors = append(res.Errors, "sts")
				}
				var mx []*net.MX
				for try := 0; try < 3; try++ {
					ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
					mx, err = r.LookupMX(ctx, j.domain)
					cancel()
					if err == nil || notFound(err) {
						break
					}
					time.Sleep(time.Duration(300*(try+1)) * time.Millisecond)
				}
				if err != nil && !notFound(err) {
					res.Errors = append(res.Errors, "mx")
				}
				for _, m := range mx {
					if m.Host == "." || m.Host == "" {
						res.NullMX = true
					} else {
						res.MX = true
					}
				}
				out <- res
			}
		}()
	}
	go func() {
		for sc.Scan() {
			parts := strings.SplitN(sc.Text(), ",", 2)
			rank, _ := strconv.Atoi(parts[0])
			if rank > limit {
				break
			}
			jobs <- job{rank, strings.TrimSpace(parts[1])}
		}
		close(jobs)
		wg.Wait()
		close(out)
	}()
	enc := json.NewEncoder(os.Stdout)
	done := 0
	for r := range out {
		enc.Encode(r)
		done++
		if done%1000 == 0 {
			fmt.Fprintln(os.Stderr, done)
		}
	}
}
