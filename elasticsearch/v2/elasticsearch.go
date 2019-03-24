package v2

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"time"

	"github.com/alokic/gopkg/typeutils"
	"github.com/olivere/elastic"
)

var (
	// ErrNoSearchQuery means no search query is provided during search.
	ErrNoSearchQuery = errors.New("no search query given")
	// ErrNoVersion means es-version information is not found during createIndex.
	ErrNoVersion = errors.New("no elasticsearch version found")

	underscoreDocVersion = "6.2.0"
	defaultSearchSize    = 100
)

// ObjectFactory is used to create objects in which search results will be unmarshalled.
type ObjectFactory func() interface{}

// QueryObject represents elasticsearch search query.
type QueryObject string

// Source implements elastic.Query interface.
func (o QueryObject) Source() (interface{}, error) {
	var m map[string]interface{}
	err := json.Unmarshal([]byte(o), &m)
	return m, err
}

type ClientOptions struct {
	HttpClient          *http.Client
	Retrier             elastic.Retrier
	SniffEnabled        *bool
	SniffInterval       time.Duration
	HealthCheckEnabled  *bool
	HealthCheckInterval time.Duration
	HealthCheckTimeout  time.Duration
	GzipEnabled         *bool //use pointer to distinguish from default false, if not set then we leave it to lib's default.
	InfoLogger          elastic.Logger
	ErrorLogger         elastic.Logger
	TraceLogger         elastic.Logger //prints http req and response
}

// Elasticsearch : struct.
type Elasticsearch struct {
	hosts []string
	*elastic.Client
	version    string
	httpClient *http.Client
}

// CreateIndexOutput is returned from CreateIndex method.
type CreateIndexOutput struct {
	Acknowledged bool
}

// DeleteIndexOutput is returned from DeleteIndex method.
type DeleteIndexOutput struct {
	Acknowledged bool
}

// BulkInput is input to Bulk method.
type BulkInput struct {
	Cmd       string
	IndexName string
	IndexType string
	Id        interface{}
	Data      interface{}
}

// Failed struct represent documents on which an bulk operation (index, update, delete) is failed.
type Failed struct {
	Id    string
	Error string
}

// BulkOutput is returned by Bulk method.
type BulkOutput struct {
	SuccessIds []string
	FailedIds  []*Failed
}

// UpdateByQueryInput is input to UpdateByQuery method.
type UpdateByQueryInput struct {
	Query     string
	IndexName string
}

// UpdateByQueryOutput is returned by  UpdateByQuery method.
type UpdateByQueryOutput struct {
	Updated int
}

// DeleteByQueryInput is input to DeleteByQuery method.
type DeleteByQueryInput struct {
	Query     string
	IndexName string
}

// DeleteByQueryOutput is returned by DeleteByQuery method.
type DeleteByQueryOutput struct {
	Deleted int
}

// SearchInput is input to search method.
type SearchInput struct {
	Query         QueryObject
	IndexName     string
	SortField     string
	SortAscending bool
	Offset        int
	Limit         int
	Factory       ObjectFactory
}

// SearchOutput is input to search method.
type SearchOutput struct {
	Count   int
	Results interface{}
}

// New initializes elasticsearch.
func New(servers []string, options *ClientOptions) (*Elasticsearch, error) {
	e := &Elasticsearch{
		hosts: servers,
	}

	clientOpts := []elastic.ClientOptionFunc{elastic.SetURL(servers...)}
	clientOpts = append(clientOpts, clientOptions(options)...)

	//this actually connects to the cluster.
	c, err := elastic.NewClient(clientOpts...)
	if err != nil {
		return nil, err
	}

	e.Client = c
	return e, nil
}

// DefaultMapping for index.
func DefaultMapping(file string) map[string]interface{} {
	data, err := ioutil.ReadFile(file)

	if err != nil {
		fmt.Println("gosh:yaml:kv:yaml file read failed", err)
		os.Exit(1)
	}
	var m map[string]interface{}
	json.Unmarshal(data, &m)
	return m
}

// CreateIndex creates elasticsearch index.
func (e *Elasticsearch) CreateIndex(ctx context.Context, indexName, indexType string, mapping map[string]interface{}, settings map[string]interface{}) (*CreateIndexOutput, error) {
	exists, err := e.Client.IndexExists(indexName).Do(ctx)
	if err != nil {
		return nil, err
	}

	if exists {
		return &CreateIndexOutput{Acknowledged: true}, nil
	}

	version := e.Version()
	if version == "" {
		return nil, ErrNoVersion
	}

	var indexMapping map[string]interface{}
	{
		indexMapping = map[string]interface{}{indexType: mapping}

		// if version < underscoreDocVersion {
		// 	indexMapping = map[string]interface{}{indexType: mapping}
		// } else {
		// 	indexMapping = map[string]interface{}{"_" + indexType: mapping}
		// }
	}

	// Create a new index

	m, _ := json.Marshal(mappingTemplate(settings, indexMapping))

	createIndex, err := e.Client.CreateIndex(indexName).BodyString(string(m)).Do(ctx)
	if err != nil {
		return nil, err
	}

	if !createIndex.Acknowledged {
		return &CreateIndexOutput{}, nil
	}

	return &CreateIndexOutput{Acknowledged: true}, nil
}

// DeleteIndex deletes elasticsearch index.
func (e *Elasticsearch) DeleteIndex(ctx context.Context, indexName string) (*DeleteIndexOutput, error) {
	deleteIndex, err := e.Client.DeleteIndex(indexName).Do(ctx)
	if err != nil {
		return nil, err
	}
	if !deleteIndex.Acknowledged {
		return &DeleteIndexOutput{}, nil
	}
	return &DeleteIndexOutput{Acknowledged: true}, nil
}

// Version of elasticsearch.
func (e *Elasticsearch) Version() string {
	if e.version != "" {
		return e.version
	}
	version, err := e.Client.ElasticsearchVersion(e.hosts[0])
	if err != nil {
		return ""
	}
	e.version = version
	return version
}

// Flush elasticsearch index.
func (e *Elasticsearch) Flush(indices ...string) {
	f := e.Client.Flush()
	f = f.Force(true)
	f.Do(context.Background())
}

// Bulk method can do multiple op (index,update, delete).
// It returns list of successful and failed ids.
// It expects an ID in the doc.
func (e *Elasticsearch) Bulk(ctx context.Context, in []*BulkInput) (*BulkOutput, error) {
	bulkRequest := e.Client.Bulk()

	for _, b := range in {
		id := typeutils.ToStr(b.Id)
		if id == "" {
			continue
		}

		switch b.Cmd {
		case "index":
			bulkRequest.Add(elastic.NewBulkIndexRequest().Index(b.IndexName).Type(b.IndexType).Id(id).Doc(b.Data))
		case "update":
			bulkRequest.Add(elastic.NewBulkUpdateRequest().DocAsUpsert(false).Index(b.IndexName).Type(b.IndexType).Id(id).Doc(b.Data))
		case "upsert":
			bulkRequest.Add(elastic.NewBulkUpdateRequest().DocAsUpsert(true).Index(b.IndexName).Type(b.IndexType).Id(id).Doc(b.Data))
		case "delete":
			bulkRequest.Add(elastic.NewBulkDeleteRequest().Index(b.IndexName).Type(b.IndexType).Id(id))
		}
	}

	bulkResponse, err := bulkRequest.Do(ctx)
	if err != nil {
		return nil, err
	}

	bo := &BulkOutput{}
	for _, b := range bulkResponse.Succeeded() {
		bo.SuccessIds = append(bo.SuccessIds, b.Id)
	}

	for _, b := range bulkResponse.Failed() {
		f := &Failed{Id: b.Id}
		if b.Error != nil {
			f.Error = b.Error.Reason
		}
		bo.FailedIds = append(bo.FailedIds, f)
	}

	return bo, nil
}

// UpdateByQuery updates based on query.
func (e *Elasticsearch) UpdateByQuery(ctx context.Context, in *UpdateByQueryInput) (*UpdateByQueryOutput, error) {
	us := e.Client.UpdateByQuery(in.IndexName).Conflicts("proceed")

	if in.Query != "" {
		us = us.Body(in.Query)
	}

	bs, err := us.Do(ctx)
	if err != nil {
		return nil, err
	}

	return &UpdateByQueryOutput{Updated: int(bs.Updated)}, nil
}

// DeleteyByQuery deletes based on query.
func (e *Elasticsearch) DeleteyByQuery(ctx context.Context, in *DeleteByQueryInput) (*DeleteByQueryOutput, error) {
	us := e.Client.DeleteByQuery(in.IndexName).Conflicts("proceed")

	if in.Query != "" {
		us = us.Body(in.Query)
	}

	bs, err := us.Do(ctx)
	if err != nil {
		return nil, err
	}

	return &DeleteByQueryOutput{Deleted: int(bs.Deleted)}, nil
}

// Search a query.
func (e *Elasticsearch) Search(ctx context.Context, in *SearchInput) (*SearchOutput, error) {
	s := e.Client.Search().Index(in.IndexName)

	if in.Query == "" {
		return nil, ErrNoSearchQuery
	}
	s = s.Query(in.Query)

	if in.SortField != "" {
		s = s.Sort(in.SortField, in.SortAscending)
	}

	sz := defaultSearchSize
	if in.Limit > 0 {
		sz = in.Limit
	}
	s = s.From(in.Offset).Size(sz)

	searchResult, err := s.Do(ctx)
	if err != nil {
		return nil, err
	}

	so := &SearchOutput{Count: len(searchResult.Hits.Hits)}

	if in.Factory == nil {
		in.Factory = func() interface{} {
			return map[string]interface{}{}
		}
	}

	sl := reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(in.Factory())), 0, len(searchResult.Hits.Hits))

	// Iterate through results
	for _, hit := range searchResult.Hits.Hits {
		t := in.Factory()

		err := json.Unmarshal(*hit.Source, &t)
		if err != nil {
			return nil, err
		}

		sl = reflect.Append(sl, reflect.ValueOf(t))
	}

	so.Results = sl.Interface()

	return so, err
}

func mappingTemplate(settings, indexMapping map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"settings": settings,
		"mappings": indexMapping,
	}
}

func clientOptions(options *ClientOptions) []elastic.ClientOptionFunc {
	opts := []elastic.ClientOptionFunc{}

	if options != nil {
		if options.HttpClient != nil {
			opts = append(opts, elastic.SetHttpClient(options.HttpClient))
		}
		if options.Retrier != nil {
			opts = append(opts, elastic.SetRetrier(options.Retrier))
		}
		if options.SniffEnabled != nil {
			opts = append(opts, elastic.SetSniff(*options.SniffEnabled))
			if options.SniffInterval != 0 {
				opts = append(opts, elastic.SetSnifferInterval(options.SniffInterval))
			}
		}
		if options.HealthCheckEnabled != nil {
			opts = append(opts, elastic.SetHealthcheck(*options.HealthCheckEnabled))
			if options.HealthCheckInterval != 0 {
				opts = append(opts, elastic.SetHealthcheckInterval(options.HealthCheckInterval))
			}
			if options.HealthCheckTimeout != 0 {
				opts = append(opts, elastic.SetHealthcheckTimeout(options.HealthCheckTimeout))
			}
		}
		if options.GzipEnabled != nil {
			opts = append(opts, elastic.SetGzip(*options.GzipEnabled))
		}
		if options.ErrorLogger != nil {
			opts = append(opts, elastic.SetErrorLog(options.ErrorLogger))
		}
		if options.InfoLogger != nil {
			opts = append(opts, elastic.SetInfoLog(options.InfoLogger))
		}
		if options.TraceLogger != nil {
			opts = append(opts, elastic.SetTraceLog(options.TraceLogger))
		}
	}

	return opts
}
