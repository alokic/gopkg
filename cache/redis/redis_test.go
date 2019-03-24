package redis

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/alokic/gopkg/cache"
	farm "github.com/alokic/gopkg/redisfarm"
)

type TestConfig struct {
	server string
}

var (
	testConfig = TestConfig{
		server: "redis://localhost:6379",
	}
)

func testCacheFarm() *farm.Farm {
	cl, err := farm.
		NewClusterBuilder().
		SetMaxIdleConns(6).
		SetMaxActiveConns(6).
		SetServers([]string{testConfig.server}).
		Build()

	if err != nil {
		log.Println(err)
		return nil
	}

	f, err := farm.
		NewBuilder().
		SetCluster([]*farm.Cluster{cl}).
		Build()

	if err != nil {
		log.Println(err)
		return nil
	}

	return f
}

func testCacheCreate() cache.Cache {
	c, err := NewCacheBuilder().
		SetFarm(testCacheFarm()).
		SetApp("test").
		SetPrefix("prefix").
		Build()

	if err != nil {
		log.Println(err)
		return nil
	}
	return c
}

func testCacheClearData() {
	(testCacheCreate().(*redisCache)).farm.AllConn().Do("flushall")
}

func testCacheLoadData(c cache.Cache, n int, dontExpire ...bool) {
	testCacheClearData()
	for i := 1; i <= n; i++ {
		item := cache.Item{
			Key: fmt.Sprintf("%d", i),
			Val: []byte(fmt.Sprintf("%d", i)),
			TTL: 60,
		}
		if len(dontExpire) > 0 && dontExpire[0] {
			item.TTL = 0
		}
		c.Put(context.Background(), &item)
	}
}

func TestCachePut(t *testing.T) {
	c := testCacheCreate()

	item := cache.Item{
		Key: fmt.Sprintf("%d", 1),
		Val: []byte(fmt.Sprintf("%d", 1)),
		TTL: 60,
	}
	r, err := c.Put(context.Background(), &item)

	if err != nil {
		t.Errorf(err.Error())
		return
	}

	log.Println("Pass:TestCachePut: ", r)
}

func TestCachePutOldOperation(t *testing.T) {
	c := testCacheCreate()

	testCacheLoadData(c, 2)

	// create time lag
	SetTimeDiffForTesting(-2000)

	item := cache.Item{
		Key: fmt.Sprintf("%d", 1),
		Val: []byte(fmt.Sprintf("%d", 100)),
		TTL: 60,
	}
	_, err := c.Put(context.Background(), &item)

	if err != nil {
		t.Errorf(err.Error())
		return
	}

	r, err := c.Get(context.Background(), "1")

	if err != nil {
		t.Errorf(err.Error())
		return
	}

	if string(r.Val) == "100" {
		t.Errorf("cache.Item should not be updated from old operation")
		return
	}

	log.Println("Pass:TestCachePutOldOperation: ", r)
}

func TestCacheGet(t *testing.T) {
	numRec := 10

	c := testCacheCreate()

	testCacheLoadData(c, numRec)

	r, err := c.Get(context.Background(), "7")

	if err != nil {
		t.Errorf(err.Error())
		return
	}

	if r.Key != "7" {
		t.Errorf("cache.Item %d should be found in get operation", 7)
		return
	}

	log.Println("Pass:TestCacheGet: ", r.Key, r.Val, r.TTL)
}

func TestCacheNil(t *testing.T) {
	numRec := 10

	c := testCacheCreate()

	testCacheLoadData(c, numRec)

	r, err := c.Get(context.Background(), "notexist")

	if err != nil {
		t.Errorf(err.Error())
		return
	}

	if r != nil {
		t.Errorf("cache.Item %d should be nil %v", 7, r)
		return
	}

	log.Println("Pass:TestCacheGetNil: ")
}

func TestCacheDeleted(t *testing.T) {
	c := testCacheCreate()

	testCacheLoadData(c, 2)

	SetTimeDiffForTesting(10)
	err := c.Delete(context.Background(), "2")

	if err != nil {
		t.Errorf(err.Error())
		return
	}

	r, err := c.Get(context.Background(), "2")

	if err != nil {
		t.Errorf(err.Error())
		return
	}

	if r != nil {
		t.Errorf("cache.Item should be marked deleted after delet operation")
		return
	}

	log.Println("Pass:TestCacheDeleted: ", r)
}

func TestCacheDeletedOldOperation(t *testing.T) {
	c := testCacheCreate()

	testCacheLoadData(c, 2)

	SetTimeDiffForTesting(-10)
	err := c.Delete(context.Background(), "2")

	if err != nil {
		t.Errorf(err.Error())
		return
	}

	log.Println("Pass:TestCacheDeletedOldOperation: ")
}

func TestCacheMultiGet(t *testing.T) {
	numRec := 3
	c := testCacheCreate()

	testCacheLoadData(c, numRec)

	str := []string{}
	for i := 1; i <= numRec; i++ {
		str = append(str, fmt.Sprintf("%d", i))
	}
	r, err := c.MultiGet(context.Background(), str)

	if err != nil {
		t.Errorf(err.Error())
		return
	}

	if len(r) != numRec {
		t.Errorf("Items should be returned in multiGet operation")
		return
	}

	log.Println("Pass:TestCacheMultiGet: ", r)
}

func TestCacheMultiGetNoDeletedItem(t *testing.T) {
	numRec := 3
	c := testCacheCreate()

	testCacheLoadData(c, numRec)

	c.Delete(context.Background(), "2")
	str := []string{}
	for i := 1; i <= numRec; i++ {
		str = append(str, fmt.Sprintf("%d", i))
	}
	r, err := c.MultiGet(context.Background(), str)

	if err != nil {
		t.Errorf(err.Error())
		return
	}

	if len(r) != numRec-1 {
		t.Errorf("Items should be returned in multiGet operation")
		return
	}

	log.Println("Pass:TestCacheMultiGetNoDeletedItem: ")
}

func TestCacheMultiDelete(t *testing.T) {
	numRec := 3
	c := testCacheCreate()

	testCacheLoadData(c, numRec)

	str := []string{}
	for i := 2; i <= numRec; i++ {
		str = append(str, fmt.Sprintf("%d", i))
	}
	err := c.MultiDelete(context.Background(), str)

	if err != nil {
		t.Errorf(err.Error())
		return
	}

	item, err := c.Get(context.Background(), "1")

	if item.Key != "1" {
		t.Errorf("Error: TestCacheMultiDelete Got %s, Want %s", item.Key, "1")
		return
	}

	item, err = c.Get(context.Background(), "2")

	if item != nil {
		t.Errorf("Error: TestCacheMultiDelete Got %v, Want nil", item)
		return
	}
	log.Println("Pass:TestCacheMultiGetNoDeletedItem: ")
}

func TestMain(m *testing.M) {
	m.Run()
}
