package config_test

import (
	"os"
	"reflect"
	"testing"

	"github.com/alokic/gopkg/config"
)

var (
	testYaml  = "./test/test.yml"
	test1Yaml = "./test/test1.yml"
)

func Test_client_Get(t *testing.T) {
	type args struct {
		key string
	}

	tests := []struct {
		name string
		args args
		want interface{}
	}{
		{
			name: "TestGetNil",
			args: args{key: "not_exist"},
			want: nil,
		},
		{
			name: "TestGetEnv",
			args: args{key: "test_concurrency"},
			want: "10",
		},
		{
			name: "TestGetCurrentNs",
			args: args{key: "concurrency"},
			want: 100,
		},
	}

	c, _ := config.New(testYaml, test1Yaml)
	c.SetNamespace("production")

	os.Setenv("TEST_CONCURRENCY", "10")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := c.Get(tt.args.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("client.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_client_Set(t *testing.T) {
	type args struct {
		key string
		val interface{}
	}
	tests := []struct {
		name string
		args args
		want interface{}
	}{
		{
			name: "TestGetCurrentNs",
			args: args{key: "set_concurrency", val: 75},
			want: 75,
		},
	}

	c, _ := config.New(testYaml)
	c.SetNamespace("production")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c.Set(tt.args.key, tt.args.val)
			if got := c.Get(tt.args.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("client.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}
