package elasticsearch

import (
	es "github.com/alokic/gopkg/elasticsearch/v2"
	"gopkg.in/DataDog/dd-trace-go.v1/contrib/olivere/elastic"
)

// NewElasticsearch returns tracing enabled elasticsearch.Elasticsearch client
func NewElasticsearch(servers []string, serviceName string, opts ...*es.ClientOptions) (*es.Elasticsearch, error) {
	c := elastic.NewHTTPClient(
		[]elastic.ClientOption{elastic.WithServiceName(serviceName)}...,
	)

	opt := new(es.ClientOptions)
	if len(opts) > 0 {
		opt = opts[0]
	}

	opt.HttpClient = c
	return es.New(servers, opt)
}
