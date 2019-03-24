package redis

import (
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"runtime"

	"github.com/garyburd/redigo/redis"
)

var (
	idBytes                = "id"
	sortKeyBytes           = "sort_key"
	expireAtBytes          = "ttl"
	runAtBytes             = "run_at"
	visibilityTimeoutBytes = "vis_time"
	retryCountBytes        = "retry_count"
	recordBytes            = "record"
)

// readScript : Get the script object
func readScript(name string) (*redis.Script, error) {
	_, filename, _, ok := runtime.Caller(4)
	if !ok {
		return nil, fmt.Errorf("%s.lua is not found", name)
	}
	absFilePath := path.Join(path.Dir(filename), "/scripts/", name)

	return redis.NewScript(0, readFile(absFilePath)), nil
}

// readFile : Read the script file as string
func readFile(fname string) string {
	b, err := ioutil.ReadFile(fname)

	if err != nil {
		log.Println(err)
		return ""
	}

	return string(b)
}

// fmtScriptResponse : helper function for Debug
func fmtScriptResponse(reply interface{}, err error) {
	if arr, ok := reply.([]interface{}); ok {
		for _, r := range arr {
			b, _ := r.([]byte)
			log.Println(string(b), b)
		}
	} else {
		b, _ := reply.([]byte)
		log.Println(string(b))
	}

	// log.Println(err)
	// log.Println(args)
}
