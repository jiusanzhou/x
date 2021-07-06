package clientlb

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
	"unsafe"
)

// Balancer is a balancer for apiservers
type Balancer struct {
	endpoints []*Endpoint
	healthEndpoints []*Endpoint
	unhealthEndpoints []*Endpoint

	rl sync.Locker

	heartbeat time.Duration

	// base round tripper from client
	base http.RoundTripper

	healthChecker HealthChecker

	quitCh chan struct{}
}

var _ http.RoundTripper = (*Balancer)(nil)

// RoundTrip ...
func (br *Balancer) RoundTrip(req *http.Request) (*http.Response, error) {

	if br.base == nil {
		return nil, errors.New("internal error no next round trip")
	}

	// so if we need to resend the request, we must copy the request

	// var req2 *http.Request
	
	// req2 := CloneRequest(req)

	// var data bytes.Buffer

	// data.ReadFrom(req.Body)
	// req.Body = ioutil.NopCloser(&data)
	// req2.Body = ioutil.NopCloser(bytes.NewReader(data.Bytes()))

	// ok,let send the req with the client

	// let's choose a endpoint health to send the request
	ep, err := br.nextEndpoint()
	if err != nil {
		// TODO: or use the default roundtripper?
		return nil, err
	}

	log.Println("[DEBUG] balancer modfiy", req.URL.Host, "=>", ep.url.Host, ":", req.Method, req.URL.String())

	ep.Apply(req)

	resp, err := br.base.RoundTrip(req)

	// should we check ???

	if err != nil {
		// mark unhealth
		log.Println("[ERROR] while round tripper the request, we get error:", err)
		br.markHealth(ep, false)
	}

	return resp, err
}

func (br *Balancer) nextEndpoint() (*Endpoint, error) {
	br.rl.Lock()
	defer br.rl.Unlock()

	if len(br.healthEndpoints) == 0 {
		return nil, errors.New("no health endpoint")
	}

	// TODO: implement round algorithm, currenttly only support random way

	return br.healthEndpoints[rand.Intn(len(br.healthEndpoints))], nil
}

// markHealth 
func (br *Balancer) markHealth(ep *Endpoint, ok bool) {
	br.rl.Lock()
	defer func() {
		br.rl.Unlock()
		log.Println("[INFO] mark:", unsafe.Pointer(ep), "healthy:", ok, "Healths:", br.healthEndpoints, "Unhealths:", br.unhealthEndpoints)
	}()

	// IMPROVE: make the slcie swapper

	// add unhealthIndexs and healthIndexs
	// check if we exits

	// add to health, and remove from unhealth
	if ok {

		// add health
		has := false
		for _, ex := range br.healthEndpoints {
			if ex == ep {
				has = true
				break
			}
		}

		if !has {
			br.healthEndpoints = append(br.healthEndpoints, ep)
		}

		// remove unhealth
		for i, ex := range br.unhealthEndpoints {
			if ex == ep {
				// ok let's remove
				max := len(br.unhealthEndpoints)-1
				if i == max {
					br.unhealthEndpoints = br.unhealthEndpoints[0:i]
				} else if i == 0 {
					br.unhealthEndpoints = br.unhealthEndpoints[1:max]
				} else {
					br.unhealthEndpoints = append(br.unhealthEndpoints[0:i], br.unhealthEndpoints[i+1:max]...)
				}
				return
			}
		}
	} else {
		// add to unhealth
		has := false
		for _, ex := range br.unhealthEndpoints {
			if ex == ep {
				has = true
				break
			}
		}

		if !has {
			br.unhealthEndpoints = append(br.unhealthEndpoints, ep)
		}

		// remove health
		for i, ex := range br.healthEndpoints {
			if ex == ep {
				// ok let's remove
				max := len(br.healthEndpoints)
				if i == max - 1 {
					br.healthEndpoints = br.healthEndpoints[0:i]
				} else if i == 0 {
					br.healthEndpoints = br.healthEndpoints[1:max]
				} else {
					br.healthEndpoints = append(br.healthEndpoints[0:i], br.healthEndpoints[i+1:max]...)
				}
				return
			}
		}
	}
}

// recovery the unhealth to the normal
func (br *Balancer) recovery() {
	// check un health list

	// create a new req to send for health check

	for _, ep := range br.unhealthEndpoints {
		log.Println("[DEBUG] check health for", ep.raw)

		// create an new req
		req, err := http.NewRequest("GET", ep.raw, nil)
		if err != nil {
			fmt.Println("[ERROR] new request from endpoint error:", err)
			continue
		}

		// finish the request
		br.healthChecker(ep, req, nil)

		// use the base to send the request
		resp, err := br.base.RoundTrip(req)
		if err != nil {
			// ok we must not health
			log.Println("[ERROR] connect", ep.raw, "still error:", err)
			return
		}

		// let's check the resp
		if br.healthChecker(ep, req, resp) {
			// ok we are health, let's mark
			br.markHealth(ep, true)
			log.Println("[INFO] recovery healthy:", ep.raw)
			return
		}

		log.Println("[ERROR]", ep.raw, "still not health")
	}
}

func (br *Balancer) recoveryLoop() {

	tk := time.NewTicker(br.heartbeat)

	for {
		select {
		case <- tk.C:
			// let's do recovery
			br.recovery()
		case <- br.quitCh:
			log.Println("received quit signal ...")
			return
		}
	}
}

// NewBalancer ...
func NewBalancer(opts ...Option) (*Balancer, error) {
	br := &Balancer{
		heartbeat: time.Second * 5,
		rl: &sync.Mutex{},
	}

	for _, o := range opts {
		o(br)
	}

	// ok let's get the endpoints and do something

	// should we need to check and init at here???

	// init the healthIndex and indexMap
	for _, ep := range br.endpoints {
		// default mark all as health
		br.markHealth(ep, true)
	}

	go br.recoveryLoop()

	return br, nil
}

// Option ... 
type Option func(*Balancer)

// AddEndpoints ...
func AddEndpoints(endpoints ...string) Option {
	return func(br *Balancer) {
		for _, e := range endpoints {
			ep, err := NewEndpoint(e)
			if err != nil {
				log.Println("[ERROR] new endpoints from", e, "error:", err)
				continue
			}

			br.endpoints = append(br.endpoints, ep)
		}
	}
}

// SetRoundTripper ...
func SetRoundTripper(rt http.RoundTripper) Option {
	return func(br *Balancer) {
		br.base = rt
	}
}

// SetHealthChecker ...
func SetHealthChecker(check HealthChecker) Option {
	return func(br *Balancer) {
		br.healthChecker = check
	}
}

var balancerRegistry = map[string]*Balancer{}

// NewBalancerRoundTripper ...
func NewBalancerRoundTripper(endpoints ...string) func (rt http.RoundTripper) http.RoundTripper {
	return func (rt http.RoundTripper) http.RoundTripper {
		if len(endpoints) == 0 {
			log.Println("[WARN] without load balancer")
			return rt
		}

		key := strings.Join(endpoints, ",")

		var bl *Balancer

		// store the balancer in the global cache
		if bl, ok := balancerRegistry[key]; ok {
			log.Println("[INFO] reuse balancer with endpoints:", key)
			return bl
		}

		bl, err := NewBalancer(
			AddEndpoints(endpoints...),
			SetRoundTripper(rt),
			SetHealthChecker(NewSimpleHealthCheck("GET", "/healthz", "ok")),
		)

		if err != nil {
			log.Println("[ERROR] new balancer round tripper error:", err)
			return rt
		}

		// store to the cache
		balancerRegistry[key] = bl

		log.Println("[INFO] new balancer success:", endpoints)
		return bl
	}
}
