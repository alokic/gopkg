package concurrent

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
)

type Resulter interface {
	IsBlank() bool
}

type Handler func() (Resulter, error)

type Response struct {
	Result     Resulter
	Error      error
	HandlerIdx int
}

type Error struct {
	Errs []error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}

	var b []byte
	for _, err := range e.Errs {
		b = append(b, []byte(err.Error())...)
	}
	return string(b)
}

func (e *Error) append(err error) *Error {
	if e == nil {
		e = &Error{}
	}
	e.Errs = append(e.Errs, err)
	return e
}

func (e *Error) toError() error {
	if e == nil {
		return nil
	}
	return errors.New(e.Error())
}

// First returns first result from fastest handler
func First(ctx context.Context, hs []Handler) (*Response, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errs := []string{}
	for resp := range exec(ctx, hs) {

		if resp.Error == nil && !resp.Result.IsBlank() {
			return resp, nil
		}

		if resp.Error != nil {
			errs = append(errs, resp.Error.Error())
		}

	}

	var err error
	if len(errs) > 0 {
		err = fmt.Errorf(strings.Join(errs, ","))
	}

	return nil, err
}

// All returns results from all handlers together
func All(ctx context.Context, hs []Handler) ([]*Response, error) {
	var errs *Error

	var arr []*Response
	for resp := range exec(ctx, hs) {
		arr = append(arr, resp)
		if resp.Error != nil {
			errs = errs.append(resp.Error)
		}
	}

	return arr, errs.toError()
}

func exec(ctx context.Context, hs []Handler) <-chan *Response {
	var wg sync.WaitGroup

	out := make(chan *Response)

	wg.Add(len(hs))

	for idx, h := range hs {
		go func(idx int, h Handler) {
			defer wg.Done()
			out <- perform(ctx, h, idx)
		}(idx, h)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func perform(ctx context.Context, h Handler, hidx int) *Response {
	out := make(chan *Response)

	go func(ctx context.Context) {
		result, err := h()
		out <- &Response{Result: result, Error: err, HandlerIdx: hidx}
		defer close(out) // close channel where its written
	}(ctx)

	select {
	case <-ctx.Done():
		return &Response{Result: nil, Error: ctx.Err(), HandlerIdx: hidx}
	case r := <-out:
		return r
	}
}
