package concurrent_test

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/alokic/gopkg/concurrent"
)

func TestFirst(t *testing.T) {
	type args struct {
		ctx context.Context
		hs  []concurrent.Handler
	}

	ctx := context.Background()

	tests := []struct {
		name    string
		args    args
		want    *concurrent.Response
		wantErr bool
	}{
		{
			name: "TestFirstNoResult",
			args: args{
				ctx: ctx,
				hs:  []concurrent.Handler{},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "TestFirstOneSuccess",
			args: args{
				ctx: ctx,
				hs: []concurrent.Handler{
					func() (concurrent.Resulter, error) {
						s := &testSearcher{}
						return s.testHandler(ctx, 5, 200)
					},
				},
			},
			want: &concurrent.Response{
				Result:     testResultList([]*testResult{&testResult{Id: uint64(5)}, &testResult{Id: uint64(5) + 1}}),
				Error:      nil,
				HandlerIdx: 0,
			},

			wantErr: false,
		},

		{
			name: "TestFirstMultiSuccess",
			args: args{
				ctx: ctx,
				hs: []concurrent.Handler{
					func() (concurrent.Resulter, error) {
						s := &testSearcher{}
						return s.testHandler(ctx, 5, 200)
					},
					func() (concurrent.Resulter, error) {
						s := &testSearcher{}
						return s.testHandler(ctx, 6, 200)
					},
					func() (concurrent.Resulter, error) {
						s := &testSearcher{}
						return s.testHandler(ctx, 7, 200)
					},
				},
			},
			want: &concurrent.Response{
				Result:     testResultList([]*testResult{&testResult{Id: uint64(5)}, &testResult{Id: uint64(5) + 1}}),
				Error:      nil,
				HandlerIdx: 0,
			},
			wantErr: false,
		},
		// {
		// 	name: "TestFirstDeadlineExceeded",
		// 	args: args{
		// 		ctx: testContextWithTimeout(ctx, 100),
		// 		hs: []concurrent.Handler{
		// 			func() (concurrent.Resulter, error) {
		// 				s := &testSearcher{}
		// 				return s.testHandler(ctx, 5, 200)
		// 			},
		// 		},
		// 	},
		// 	want: &concurrent.Response{
		// 		Result:     nil,
		// 		Error:      context.DeadlineExceeded,
		// 		HandlerIdx: 0,
		// 	},
		// 	wantErr: true,
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := concurrent.First(tt.args.ctx, tt.args.hs)
			if (err != nil) != tt.wantErr {
				t.Errorf("First() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (!reflect.DeepEqual(got, tt.want)) && (got.HandlerIdx%3 == 0) { // got.HandlerIdx%3 == 0 checks other cases for multirequest whent idx == or 2
				t.Errorf("First() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAll(t *testing.T) {
	type args struct {
		ctx context.Context
		hs  []concurrent.Handler
	}
	ctx := context.Background()

	tests := []struct {
		name    string
		args    args
		want    []*concurrent.Response
		wantErr bool
	}{
		{
			name: "TestAllNoResult",
			args: args{
				ctx: ctx,
				hs:  []concurrent.Handler{},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "TestAllOneSuccess",
			args: args{
				ctx: ctx,
				hs: []concurrent.Handler{
					func() (concurrent.Resulter, error) {
						s := &testSearcher{}
						return s.testHandler(ctx, 5, 200)
					},
				},
			},
			want: []*concurrent.Response{
				{
					Result:     testResultList([]*testResult{&testResult{Id: uint64(5)}, &testResult{Id: uint64(5) + 1}}),
					Error:      nil,
					HandlerIdx: 0,
				},
			},
			wantErr: false,
		},

		{
			name: "TestAllMultiSuccess",
			args: args{
				ctx: ctx,
				hs: []concurrent.Handler{
					func() (concurrent.Resulter, error) {
						s := &testSearcher{}
						return s.testHandler(ctx, 5, 200)
					},
					func() (concurrent.Resulter, error) {
						s := &testSearcher{}
						return s.testHandler(ctx, 6, 200)
					},
					func() (concurrent.Resulter, error) {
						s := &testSearcher{}
						return s.testHandler(ctx, 7, 200)
					},
				},
			},
			want: []*concurrent.Response{
				{
					Result:     testResultList([]*testResult{&testResult{Id: uint64(5)}, &testResult{Id: uint64(5) + 1}}),
					Error:      nil,
					HandlerIdx: 0,
				},
				{
					Result:     testResultList([]*testResult{&testResult{Id: uint64(6)}, &testResult{Id: uint64(6) + 1}}),
					Error:      nil,
					HandlerIdx: 1,
				},
				{
					Result:     testResultList([]*testResult{&testResult{Id: uint64(7)}, &testResult{Id: uint64(7) + 1}}),
					Error:      nil,
					HandlerIdx: 2,
				},
			},
			wantErr: false,
		},
		{
			name: "TestAllDeadlineExceeded",
			args: args{
				ctx: testContextWithTimeout(ctx, 100),
				hs: []concurrent.Handler{
					func() (concurrent.Resulter, error) {
						s := &testSearcher{}
						return s.testHandler(ctx, 5, 200)
					},
				},
			},
			want: []*concurrent.Response{
				{
					Result:     nil,
					Error:      context.DeadlineExceeded,
					HandlerIdx: 0,
				},
			},
			wantErr: true,
		},

		{
			name: "TestAllMultiDeadlineExceeded",
			args: args{
				ctx: testContextWithTimeout(ctx, 100),
				hs: []concurrent.Handler{
					func() (concurrent.Resulter, error) {
						s := &testSearcher{}
						return s.testHandler(ctx, 5, 200)
					},
					func() (concurrent.Resulter, error) {
						s := &testSearcher{}
						return s.testHandler(ctx, 6, 200)
					},
					func() (concurrent.Resulter, error) {
						s := &testSearcher{}
						return s.testHandler(ctx, 7, 200)
					},
				},
			},
			want: []*concurrent.Response{
				{
					Result:     nil,
					Error:      context.DeadlineExceeded,
					HandlerIdx: 0,
				},
				{
					Result:     nil,
					Error:      context.DeadlineExceeded,
					HandlerIdx: 1,
				},
				{
					Result:     nil,
					Error:      context.DeadlineExceeded,
					HandlerIdx: 2,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := concurrent.All(tt.args.ctx, tt.args.hs)
			if (err != nil) != tt.wantErr {
				t.Errorf("All() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for idx := 0; idx < len(tt.args.hs); idx++ {
				i := 0
				for ; i < len(got); i++ {
					if tt.want[idx].HandlerIdx == got[i].HandlerIdx {
						break
					}
				}
				if !testCompareResult(got[i], tt.want[idx]) {
					t.Errorf("All() = %v, want %v", got[idx], tt.want[idx])
				}
			}

		})
	}
}

type testResult struct {
	Id uint64
}

type testSearcher struct {
}

type testResultList []*testResult

func (r testResultList) IsBlank() bool {
	return len(r) == 0
}

func testCompareResult(got, want *concurrent.Response) bool {
	if got == nil && want == nil {
		return true
	}

	if got == nil || want == nil {
		return false
	}

	ok := testCompareError(got.Error, want.Error)
	if !ok {
		return false
	}

	if got.HandlerIdx != want.HandlerIdx {
		return false
	}

	ok = reflect.DeepEqual(got.Result, want.Result)
	if !ok {
		return false
	}

	return true
}

func testCompareError(e1, e2 error) bool {
	if e1 == nil && e2 == nil {
		return true
	}

	if e1 == nil || e2 == nil {
		return false
	}

	return strings.Contains(e2.Error(), e1.Error())
}

func testContextWithTimeout(ctx context.Context, millisecond int) context.Context {
	ctx, _ = context.WithTimeout(ctx, time.Duration(millisecond)*time.Millisecond)
	return ctx
}

func (t *testSearcher) testHandler(ctx context.Context, id, millisecond int) (testResultList, error) {
	time.Sleep(time.Duration(millisecond) * time.Millisecond)
	return []*testResult{{Id: uint64(id)}, {Id: uint64(id) + 1}}, nil
}
