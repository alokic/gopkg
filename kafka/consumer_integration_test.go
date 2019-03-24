package kafka_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"testing"
)

func testConfig(t *testing.T) map[string]interface{} {
	testConfig := make(map[string]interface{})
	data, err := ioutil.ReadFile("testConfig.json")
	if err != nil {
		return testConfig
	}
	r := bytes.NewReader(data)
	json.NewDecoder(r).Decode(&testConfig)
	return testConfig
}

/*func testConsumerBuilder(t *testing.T) () {
	config := testConfig(t)
	if len(config) == 0 {
		t.Fatal("No testConfig found")
	}
	testBuilder := kafka.NewConsumerBuilder(config["groupName"].(string))
	testBuilder.SetTopic(testTopic)
	testBuilder.SetHost(testHosts)
	testBuilder.SetRetry(testRetryInterval)
	return testBuilder
}
*/
func TestConsumer_Next(t *testing.T) {

}
