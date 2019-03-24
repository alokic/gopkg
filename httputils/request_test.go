package httputils

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"
)

func TestPost(t *testing.T) {
	type args struct {
		ctx     context.Context
		url     string
		body    interface{}
		headers []map[string]string
	}

	ctx := context.Background()

	tests := []struct {
		name    string
		args    args
		want    int
		want1   []byte
		wantErr bool
	}{
		{
			name: "TestPostInvalidUrl", want: http.StatusInternalServerError, want1: nil, wantErr: true,
			args: args{
				ctx: ctx, url: "https://www.google.i/invalid",
				body:    map[string]string{"msg": "hello"},
				headers: []map[string]string{{"content-type": "application/json"}},
			},
		},
		{
			name: "TestPostInvalidEndpoint", want: http.StatusNotFound, want1: nil, wantErr: false,
			args: args{
				ctx: ctx, url: "https://www.google.in/invalid",
				body:    map[string]string{"msg": "hello"},
				headers: []map[string]string{{"content-type": "application/json"}},
			},
		},
		{
			name: "TestPostSuccess", want: 200, want1: []byte(`{"msg":"hello"}`), wantErr: false,
			args: args{
				ctx: ctx, url: "https://postman-echo.com/post",
				body:    map[string]string{"msg": "hello"},
				headers: []map[string]string{{"content-type": "application/json"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := Post(tt.args.ctx, tt.args.url, tt.args.body, tt.args.headers...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Post() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Post() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(testDecode(got1, "data"), tt.want1) {
				t.Errorf("Post() got1 = %v, want %v", string(testDecode(got1, "data")), string(tt.want1))
			}
		})
	}
}

func TestGet(t *testing.T) {
	type args struct {
		ctx     context.Context
		url     string
		query   map[string]interface{}
		headers []map[string]string
	}

	ctx := context.Background()

	tests := []struct {
		name    string
		args    args
		want    int
		want1   []byte
		wantErr bool
	}{
		{
			name: "TestGetInvalidUrl", want: http.StatusInternalServerError, want1: nil, wantErr: true,
			args: args{
				ctx: ctx, url: "https://www.google.i/invalid",
				query:   nil,
				headers: []map[string]string{{"content-type": "application/json"}},
			},
		},
		{
			name: "TestGetInvalidEndpoint", want: http.StatusNotFound, want1: nil, wantErr: false,
			args: args{
				ctx: ctx, url: "https://www.google.in/invalid",
				query:   nil,
				headers: nil,
			},
		},
		{
			name: "TestGetSuccess", want: 200, want1: []byte(`{"msg":"hello"}`), wantErr: false,
			args: args{
				ctx: ctx, url: "https://postman-echo.com/get",
				query:   map[string]interface{}{"msg": "hello"},
				headers: []map[string]string{{"content-type": "application/json"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := Get(tt.args.ctx, tt.args.url, tt.args.query, tt.args.headers...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Get() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(testDecode(got1, "args"), tt.want1) {
				t.Errorf("Get() got1 = %v, want %v", string(testDecode(got1, "args")), string(tt.want1))
			}
		})
	}
}

// testDecode exctract dat from response
func testDecode(data []byte, key string) []byte {
	var m map[string]interface{}

	err := json.Unmarshal(data, &m)
	if err != nil {
		return nil
	}

	if _, ok := m[key]; !ok {
		return nil
	}

	b, err := json.Marshal(m[key])
	if err != nil {
		return nil
	}
	return b
}
