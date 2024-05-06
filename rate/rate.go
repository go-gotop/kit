// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package rate provides a rate limiter.
package rate

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// Limit defines the maximum frequency of some events.
// Limit is represented as number of events per second.
// A zero Limit allows no events.
type Limit float64

// Inf is the infinite rate limit; it allows all events (even if burst is zero).
const Inf = Limit(math.MaxFloat64)

// Every converts a minimum time interval between events to a Limit.
func Every(interval time.Duration) Limit {
	if interval <= 0 {
		return Inf
	}
	return 1 / Limit(interval.Seconds())
}

// A Limiter controls how frequently events are allowed to happen.
// It implements a "token bucket" of size b, initially full and refilled
// at rate r tokens per second.
// Informally, in any large enough time interval, the Limiter limits the
// rate to r tokens per second, with a maximum burst size of b events.
// As a special case, if r == Inf (the infinite rate), b is ignored.
// See https://en.wikipedia.org/wiki/Token_bucket for more about token buckets.
//
// The zero value is a valid Limiter, but it will reject all events.
// Use NewLimiter to create non-zero Limiters.
//
// Limiter has three main methods, Allow, Reserve, and Wait.
// Most callers should use Wait.
//
// Each of the three methods consumes a single token.
// They differ in their behavior when no token is available.
// If no token is available, Allow returns false.
// If no token is available, Reserve returns a reservation for a future token
// and the amount of time the caller must wait before using it.
// If no token is available, Wait blocks until one can be obtained
// or its associated context.Context is canceled.
//
// The methods AllowN, ReserveN, and WaitN consume n tokens.
//
// Limiter is safe for simultaneous use by multiple goroutines.
type Limiter struct {
	mu     sync.Mutex
	limit  Limit
	burst  int
	tokens float64
	// last is the last time the limiter's tokens field was updated
	last time.Time
	// lastEvent is the latest time of a rate-limited event (past or future)
	lastEvent time.Time
	// 限制周期
	limitPeriod time.Duration
	// 第一次消费时间
	firstActTime time.Time
	// 单位时间内消费token数统计
	actTokensCount float64
	// redis client
	rdb redis.Client
}

// Limit returns the maximum overall event rate.
func (lim *Limiter) Limit() Limit {
	lim.mu.Lock()
	defer lim.mu.Unlock()
	return lim.limit
}

// Burst returns the maximum burst size. Burst is the maximum number of tokens
// that can be consumed in a single call to Allow, Reserve, or Wait, so higher
// Burst values allow more events to happen at once.
// A zero Burst allows no events, unless limit == Inf.
func (lim *Limiter) Burst() int {
	lim.mu.Lock()
	defer lim.mu.Unlock()
	return lim.burst
}

func (lim *Limiter) LimitPeriod() time.Duration {
	lim.mu.Lock()
	defer lim.mu.Unlock()
	return lim.limitPeriod
}

// TokensAt returns the number of tokens available at time t.
func (lim *Limiter) TokensAt(t time.Time) float64 {
	lim.mu.Lock()
	_, tokens := lim.advance(t) // does not mutate lim
	lim.mu.Unlock()
	return tokens
}

// Tokens returns the number of tokens available now.
func (lim *Limiter) Tokens() float64 {
	return lim.TokensAt(time.Now())
}

func (lim *Limiter) FirstActTime() time.Time {
	lim.mu.Lock()
	defer lim.mu.Unlock()
	return lim.firstActTime
}

func (lim *Limiter) ActTokensCount() float64 {
	lim.mu.Lock()
	defer lim.mu.Unlock()
	return lim.actTokensCount
}

// NewLimiter returns a new Limiter that allows events up to rate r and permits
// bursts of at most b tokens.
func NewLimiter(r Limit, b int) *Limiter {
	return &Limiter{
		limit: r,
		burst: b,
	}
}

// 有限时周期的限流器
func NewLimiterWithPeriod(rdb redis.Client, r Limit, b int, lp time.Duration) *Limiter {
	l := &Limiter{
		rdb:            rdb,
		limit:          r,
		burst:          b,
		limitPeriod:    lp,
		firstActTime:   time.Time{},
		actTokensCount: 0,
	}
	return l
}

// Allow reports whether an event may happen now.
func (lim *Limiter) Allow() bool {
	return lim.AllowN(time.Now(), 1)
}

func (lim *Limiter) AllowC(uniqId string) bool {
	return lim.AllowNC(uniqId, time.Now(), 1)
}

// AllowN reports whether n events may happen at time t.
// Use this method if you intend to drop / skip events that exceed the rate limit.
// Otherwise use Reserve or Wait.
func (lim *Limiter) AllowN(t time.Time, n int) bool {
	return lim.reserveN(t, n, 0).ok
}

func (lim *Limiter) AllowNC(uniqId string, t time.Time, n int) bool {
	return lim.reserveNC(uniqId, t, n, 0).ok
}

// A Reservation holds information about events that are permitted by a Limiter to happen after a delay.
// A Reservation may be canceled, which may enable the Limiter to permit additional events.
type Reservation struct {
	ok        bool
	lim       *Limiter
	tokens    int
	timeToAct time.Time
	// This is the Limit at reservation time, it can change later.
	limit Limit
}

// OK returns whether the limiter can provide the requested number of tokens
// within the maximum wait time.  If OK is false, Delay returns InfDuration, and
// Cancel does nothing.
func (r *Reservation) OK() bool {
	return r.ok
}

// Delay is shorthand for DelayFrom(time.Now()).
func (r *Reservation) Delay() time.Duration {
	return r.DelayFrom(time.Now())
}

// InfDuration is the duration returned by Delay when a Reservation is not OK.
const InfDuration = time.Duration(math.MaxInt64)

// DelayFrom returns the duration for which the reservation holder must wait
// before taking the reserved action.  Zero duration means act immediately.
// InfDuration means the limiter cannot grant the tokens requested in this
// Reservation within the maximum wait time.
func (r *Reservation) DelayFrom(t time.Time) time.Duration {
	if !r.ok {
		return InfDuration
	}
	delay := r.timeToAct.Sub(t)
	if delay < 0 {
		return 0
	}
	return delay
}

// Cancel is shorthand for CancelAt(time.Now()).
func (r *Reservation) Cancel() {
	r.CancelAt(time.Now())
}

// CancelAt indicates that the reservation holder will not perform the reserved action
// and reverses the effects of this Reservation on the rate limit as much as possible,
// considering that other reservations may have already been made.
func (r *Reservation) CancelAt(t time.Time) {
	if !r.ok {
		return
	}

	r.lim.mu.Lock()
	defer r.lim.mu.Unlock()

	if r.lim.limit == Inf || r.tokens == 0 || r.timeToAct.Before(t) {
		return
	}

	// calculate tokens to restore
	// The duration between lim.lastEvent and r.timeToAct tells us how many tokens were reserved
	// after r was obtained. These tokens should not be restored.
	restoreTokens := float64(r.tokens) - r.limit.tokensFromDuration(r.lim.lastEvent.Sub(r.timeToAct))
	if restoreTokens <= 0 {
		return
	}
	// advance time to now
	t, tokens := r.lim.advance(t)
	// calculate new number of tokens
	tokens += restoreTokens
	if burst := float64(r.lim.burst); tokens > burst {
		tokens = burst
	}
	// update state
	r.lim.last = t
	r.lim.tokens = tokens
	if r.timeToAct == r.lim.lastEvent {
		prevEvent := r.timeToAct.Add(r.limit.durationFromTokens(float64(-r.tokens)))
		if !prevEvent.Before(t) {
			r.lim.lastEvent = prevEvent
		}
	}
}

// Reserve is shorthand for ReserveN(time.Now(), 1).
func (lim *Limiter) Reserve() *Reservation {
	return lim.ReserveN(time.Now(), 1)
}

// ReserveN returns a Reservation that indicates how long the caller must wait before n events happen.
// The Limiter takes this Reservation into account when allowing future events.
// The returned Reservation’s OK() method returns false if n exceeds the Limiter's burst size.
// Usage example:
//
//	r := lim.ReserveN(time.Now(), 1)
//	if !r.OK() {
//	  // Not allowed to act! Did you remember to set lim.burst to be > 0 ?
//	  return
//	}
//	time.Sleep(r.Delay())
//	Act()
//
// Use this method if you wish to wait and slow down in accordance with the rate limit without dropping events.
// If you need to respect a deadline or cancel the delay, use Wait instead.
// To drop or skip events exceeding rate limit, use Allow instead.
func (lim *Limiter) ReserveN(t time.Time, n int) *Reservation {
	r := lim.reserveN(t, n, InfDuration)
	return &r
}

func (lim *Limiter) ReserveNC(uniqId string, t time.Time, n int) *Reservation {
	r := lim.reserveNC(uniqId, t, n, InfDuration)
	return &r
}

// Wait is shorthand for WaitN(ctx, 1).
func (lim *Limiter) Wait(ctx context.Context) (err error) {
	return lim.WaitN(ctx, 1)
}

// WaitN blocks until lim permits n events to happen.
// It returns an error if n exceeds the Limiter's burst size, the Context is
// canceled, or the expected wait time exceeds the Context's Deadline.
// The burst limit is ignored if the rate limit is Inf.
func (lim *Limiter) WaitN(ctx context.Context, n int) (err error) {
	// The test code calls lim.wait with a fake timer generator.
	// This is the real timer generator.
	newTimer := func(d time.Duration) (<-chan time.Time, func() bool, func()) {
		timer := time.NewTimer(d)
		return timer.C, timer.Stop, func() {}
	}

	return lim.wait(ctx, n, time.Now(), newTimer)
}

// wait is the internal implementation of WaitN.
func (lim *Limiter) wait(ctx context.Context, n int, t time.Time, newTimer func(d time.Duration) (<-chan time.Time, func() bool, func())) error {
	lim.mu.Lock()
	burst := lim.burst
	limit := lim.limit
	lim.mu.Unlock()

	if n > burst && limit != Inf {
		return fmt.Errorf("rate: Wait(n=%d) exceeds limiter's burst %d", n, burst)
	}
	// Check if ctx is already cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	// Determine wait limit
	waitLimit := InfDuration
	if deadline, ok := ctx.Deadline(); ok {
		waitLimit = deadline.Sub(t)
	}
	// Reserve
	r := lim.reserveN(t, n, waitLimit)
	if !r.ok {
		return fmt.Errorf("rate: Wait(n=%d) would exceed context deadline", n)
	}
	// Wait if necessary
	delay := r.DelayFrom(t)
	if delay == 0 {
		return nil
	}
	ch, stop, advance := newTimer(delay)
	defer stop()
	advance() // only has an effect when testing
	select {
	case <-ch:
		// We can proceed.
		return nil
	case <-ctx.Done():
		// Context was canceled before we could proceed.  Cancel the
		// reservation, which may permit other events to proceed sooner.
		r.Cancel()
		return ctx.Err()
	}
}

// SetLimit is shorthand for SetLimitAt(time.Now(), newLimit).
func (lim *Limiter) SetLimit(newLimit Limit) {
	lim.SetLimitAt(time.Now(), newLimit)
}

// SetLimitAt sets a new Limit for the limiter. The new Limit, and Burst, may be violated
// or underutilized by those which reserved (using Reserve or Wait) but did not yet act
// before SetLimitAt was called.
func (lim *Limiter) SetLimitAt(t time.Time, newLimit Limit) {
	lim.mu.Lock()
	defer lim.mu.Unlock()

	t, tokens := lim.advance(t)

	lim.last = t
	lim.tokens = tokens
	lim.limit = newLimit
}

// SetBurst is shorthand for SetBurstAt(time.Now(), newBurst).
func (lim *Limiter) SetBurst(newBurst int) {
	lim.SetBurstAt(time.Now(), newBurst)
}

// SetBurstAt sets a new burst size for the limiter.
func (lim *Limiter) SetBurstAt(t time.Time, newBurst int) {
	lim.mu.Lock()
	defer lim.mu.Unlock()

	t, tokens := lim.advance(t)

	lim.last = t
	lim.tokens = tokens
	lim.burst = newBurst
}

// reserveN is a helper method for AllowN, ReserveN, and WaitN.
// maxFutureReserve specifies the maximum reservation wait duration allowed.
// reserveN returns Reservation, not *Reservation, to avoid allocation in AllowN and WaitN.
func (lim *Limiter) reserveN(t time.Time, n int, maxFutureReserve time.Duration) Reservation {
	lim.mu.Lock()
	defer lim.mu.Unlock()

	if lim.limit == Inf {
		return Reservation{
			ok:        true,
			lim:       lim,
			tokens:    n,
			timeToAct: t,
		}
	} else if lim.limit == 0 {
		var ok bool
		if lim.burst >= n {
			ok = true
			lim.burst -= n
		}
		return Reservation{
			ok:        ok,
			lim:       lim,
			tokens:    lim.burst,
			timeToAct: t,
		}
	}
	t, tokens := lim.advance(t)

	// Calculate the remaining number of tokens resulting from the request.
	tokens -= float64(n)

	// Calculate the wait duration
	var waitDuration time.Duration
	if tokens < 0 {
		waitDuration = lim.limit.durationFromTokens(-tokens)
	}

	// Decide result
	ok := n <= lim.burst && waitDuration <= maxFutureReserve

	// Prepare reservation
	r := Reservation{
		ok:    ok,
		lim:   lim,
		limit: lim.limit,
	}
	if ok {
		r.tokens = n
		r.timeToAct = t.Add(waitDuration)

		// Update state
		lim.last = t
		lim.tokens = tokens
		lim.lastEvent = r.timeToAct
	}

	return r
}

// 自定义 reserveN 函数，支持超过单位时间内限制的话，再次消费需要等待
func (lim *Limiter) reserveNC(uniqId string, t time.Time, n int, maxFutureReserve time.Duration) Reservation {
	lim.mu.Lock()
	defer lim.mu.Unlock()

	// 判断缓存是否存在该 key，存在则赋值
	val, err := lim.rdb.Get(context.Background(), uniqId).Float64()
	if err != redis.Nil {
		lim.actTokensCount = val
	}
	fmt.Printf("key %v\n", val)

	// 判断缓存是否存在该 key，存在则赋值
	timeStr, err := lim.rdb.Get(context.Background(), uniqId+"_time").Result()
	if err != redis.Nil {
		lim.firstActTime, _ = time.Parse(time.RFC3339, timeStr)
	}

	// 如果是空时间，则赋值firstActTime
	if lim.firstActTime.IsZero() {
		lim.firstActTime = t
		lim.rdb.Set(context.Background(), uniqId+"_time", lim.firstActTime.Format(time.RFC3339), time.Hour*24)
	}

	// 如果t减去firstActTime 大于等于限制周期，则重置单位时间内消费token数统计
	if t.Sub(lim.firstActTime) >= lim.limitPeriod {
		lim.actTokensCount = 0
		lim.firstActTime = t
		err := lim.rdb.Set(context.Background(), uniqId, float64(0), time.Hour*24).Err()
		if err != nil {
			println("set actTokensCount error: ", err.Error())
		}
		err = lim.rdb.Set(context.Background(), uniqId+"_time", lim.firstActTime.Format(time.RFC3339), time.Hour*24).Err()
		if err != nil {
			println("set firstActTime error: ", err.Error())
		}

	}
	if lim.limit == Inf {
		return Reservation{
			ok:        true,
			lim:       lim,
			tokens:    n,
			timeToAct: t,
		}
	} else if lim.limit == 0 {
		var ok bool
		if lim.burst >= n {
			ok = true
			lim.burst -= n
		}
		return Reservation{
			ok:        ok,
			lim:       lim,
			tokens:    lim.burst,
			timeToAct: t,
		}
	}

	t, tokens := lim.advance(t)

	// Calculate the remaining number of tokens resulting from the request.
	tokens -= float64(n)

	// Calculate the wait duration
	var waitDuration time.Duration
	if tokens < 0 {
		waitDuration = lim.limit.durationFromTokens(-tokens)
	}

	// Decide result
	ok := n <= lim.burst && waitDuration <= maxFutureReserve
	// 如果tokens充足，但是单位时间内消费token数超过限制，则需要等待
	if ok && lim.actTokensCount+float64(n) > float64(lim.burst) {
		fmt.Printf("actTokensCount: %v, n: %v\n", lim.actTokensCount, n)
		fmt.Printf("总计（actTokensCount + n）: %v, burst: %v\n", lim.actTokensCount+float64(n), lim.burst)
		ok = false
	}
	// Prepare reservation
	r := Reservation{
		ok:    ok,
		lim:   lim,
		limit: lim.limit,
	}
	if ok {
		r.tokens = n
		r.timeToAct = t.Add(waitDuration)

		// Update state
		lim.last = t
		lim.tokens = tokens
		lim.lastEvent = r.timeToAct
		lim.actTokensCount += float64(n)
		lim.rdb.Set(context.Background(), uniqId, lim.actTokensCount, time.Hour*24)
	}
	return r
}

// advance calculates and returns an updated state for lim resulting from the passage of time.
// lim is not changed.
// advance requires that lim.mu is held.
func (lim *Limiter) advance(t time.Time) (newT time.Time, newTokens float64) {
	last := lim.last
	if t.Before(last) {
		last = t
	}

	// Calculate the new number of tokens, due to time that passed.
	elapsed := t.Sub(last)
	delta := lim.limit.tokensFromDuration(elapsed)
	tokens := lim.tokens + delta
	if burst := float64(lim.burst); tokens > burst {
		tokens = burst
	}
	return t, tokens
}

// durationFromTokens is a unit conversion function from the number of tokens to the duration
// of time it takes to accumulate them at a rate of limit tokens per second.
func (limit Limit) durationFromTokens(tokens float64) time.Duration {
	if limit <= 0 {
		return InfDuration
	}
	seconds := tokens / float64(limit)
	return time.Duration(float64(time.Second) * seconds)
}

// tokensFromDuration is a unit conversion function from a time duration to the number of tokens
// which could be accumulated during that duration at a rate of limit tokens per second.
func (limit Limit) tokensFromDuration(d time.Duration) float64 {
	if limit <= 0 {
		return 0
	}
	// 处理float64精度问题
	// return d.Seconds() * float64(limit) // 原代码
	sec := float64(d/time.Second) * float64(limit)
	nsec := float64(d%time.Second) * float64(limit)
	return sec + nsec/1e9
}
