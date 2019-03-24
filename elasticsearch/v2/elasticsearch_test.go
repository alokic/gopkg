package v2_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"os"

	"github.com/alokic/gopkg/elasticsearch/v2"
)

var (
	testServer      = "http://localhost:9200"
	testEs          *v2.Elasticsearch
	testMappingFile = "test_mapping.json"
)

func init() {
	if v := os.Getenv("ES_HOST"); v != "" {
		fmt.Printf("Overriding ES host to: %v\n", v)
		testServer = v
	}
}

func TestElasticsearch_CreateIndex(t *testing.T) {
	type args struct {
		ctx       context.Context
		indexName string
		indexType string
		mapping   map[string]interface{}
		settings  map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    *v2.CreateIndexOutput
		wantErr bool
	}{
		{
			name:    "TestCreateIndexSuccess",
			args:    args{ctx: context.Background(), indexName: "test", indexType: "test", mapping: v2.DefaultMapping(testMappingFile), settings: map[string]interface{}{"number_of_shards": 2, "number_of_replicas": 2}},
			want:    &v2.CreateIndexOutput{Acknowledged: true},
			wantErr: false,
		},
		{
			name:    "TestCreateIndexDuplicateSuccess",
			args:    args{ctx: context.Background(), indexName: "test", indexType: "test", mapping: v2.DefaultMapping(testMappingFile), settings: map[string]interface{}{"number_of_shards": 2, "number_of_replicas": 2}},
			want:    &v2.CreateIndexOutput{Acknowledged: true},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, err := v2.New([]string{testServer}, nil)
			if err != nil {
				t.Errorf("v2.CreateIndex() error = %v", err)
				return
			}

			got, err := e.CreateIndex(tt.args.ctx, tt.args.indexName, tt.args.indexType, tt.args.mapping, tt.args.settings)
			if (err != nil) != tt.wantErr {
				t.Errorf("v2.CreateIndex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("v2.CreateIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestElasticsearch_Bulk(t *testing.T) {
	type args struct {
		ctx context.Context
		bs  []*v2.BulkInput
	}
	tests := []struct {
		name    string
		args    args
		want    *v2.BulkOutput
		wantErr bool
	}{
		{
			name: "TestBulkSuccess",
			args: args{
				ctx: context.Background(),
				bs: []*v2.BulkInput{
					{Cmd: "index", IndexName: "test", IndexType: "test", Id: 123, Data: testNewEntity(123, "Index")},
					{Cmd: "update", IndexName: "test", IndexType: "test", Id: 123, Data: testNewEntity(123, "Update")},
					{Cmd: "index", IndexName: "test", IndexType: "test", Id: 245, Data: testNewEntity(245, "Index")},
					{Cmd: "update", IndexName: "test", IndexType: "test", Id: 567, Data: testNewEntity(245, "Update")},
					{Cmd: "delete", IndexName: "test", IndexType: "test", Id: 567},
				},
			},
			want:    &v2.BulkOutput{SuccessIds: []string{"123", "123", "245"}},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, err := v2.New([]string{testServer}, nil)
			if err != nil {
				t.Errorf("v2.Bulk() error = %v", err)
				return
			}

			got, err := e.Bulk(tt.args.ctx, tt.args.bs)
			if (err != nil) != tt.wantErr {
				t.Errorf("v2.Bulk() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got.SuccessIds, tt.want.SuccessIds) {
				t.Errorf("v2.Bulk() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestElasticsearch_UpdateByQuery(t *testing.T) {
	type args struct {
		ctx context.Context
		in  *v2.UpdateByQueryInput
	}
	tests := []struct {
		name    string
		args    args
		want    *v2.UpdateByQueryOutput
		wantErr bool
	}{
		{
			name: "TestUpdateByQuerySuccess",
			args: args{
				ctx: context.Background(),
				in: &v2.UpdateByQueryInput{
					IndexName: "test",
					Query: `{
						"script": {
							"source": "ctx._source.amount=23.2",
							"lang":   "painless"
						},
						"query": {
							"bool": {
								"must": [{
									"term": {
										"amount": 223.4567
									}
								}]
							}
						}
					}`,
				},
			},
			want:    &v2.UpdateByQueryOutput{Updated: 1},
			wantErr: false,
		},
	}

	testEs.Bulk(context.Background(),
		[]*v2.BulkInput{
			{Cmd: "index", IndexName: "test", IndexType: "test", Id: 123, Data: testNewEntity(100, "Index")},
			{Cmd: "update", IndexName: "test", IndexType: "test", Id: 123, Data: testNewEntity(200, "Update")},
		})

	// let the data in ES flushed as it will give version conflicts
	testEs.Flush("test")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, err := v2.New([]string{testServer}, nil)
			if err != nil {
				t.Errorf("v2.UpdateByQuery() error = %v", err)
				return
			}

			got, err := e.UpdateByQuery(tt.args.ctx, tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("v2.UpdateByQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("v2.UpdateByQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestElasticsearch_DeleteyQuery(t *testing.T) {
	type args struct {
		ctx context.Context
		in  *v2.DeleteByQueryInput
	}
	tests := []struct {
		name    string
		args    args
		want    *v2.DeleteByQueryOutput
		wantErr bool
	}{
		{
			name: "TestUpdateByQuerySuccess",
			args: args{
				ctx: context.Background(),
				in: &v2.DeleteByQueryInput{
					IndexName: "test",
					Query: `{
						"query": {
							"term": {
								"comment": "123"
							}
						}
					}`,
				},
			},
			want:    &v2.DeleteByQueryOutput{Deleted: 1},
			wantErr: false,
		},
	}

	testEs.Bulk(context.Background(),
		[]*v2.BulkInput{
			{Cmd: "index", IndexName: "test", IndexType: "test", Id: 123, Data: testNewEntity(123, "Index")},
			{Cmd: "update", IndexName: "test", IndexType: "test", Id: 123, Data: testNewEntity(123, "Update")},
		})

	// let the data in ES flushed as it will give version conflicts
	testEs.Flush("test")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, err := v2.New([]string{testServer}, nil)
			if err != nil {
				t.Errorf("v2.DeleteyByQuery() error = %v", err)
				return
			}

			got, err := e.DeleteyByQuery(tt.args.ctx, tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("v2.DeleteyByQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("v2.DeleteyByQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestElasticsearch_Search(t *testing.T) {
	type args struct {
		ctx context.Context
		in  *v2.SearchInput
	}
	tests := []struct {
		name    string
		args    args
		want    *v2.SearchOutput
		wantErr bool
	}{
		{
			name: "TestSearchTermQuerySuccess",
			args: args{
				ctx: context.Background(),
				in: &v2.SearchInput{
					IndexName: "test",
					Query: `{
						"term": {
							"comment": "long"
						}
					}`,
					Offset:        1,
					Limit:         2,
					SortField:     "id",
					SortAscending: false,
				},
			},
			want:    &v2.SearchOutput{Count: 2},
			wantErr: false,
		},
		{
			name: "TestSearchBoolQuerySuccess",
			args: args{
				ctx: context.Background(),
				in: &v2.SearchInput{
					IndexName: "test",
					Query: `{
						"bool": {
							"must": [{
								"range": {
									"amount": {
										"gte": 100
									}
								}
							}]
						}
					}`,
					Offset:        0,
					Limit:         3,
					SortField:     "id",
					SortAscending: false,
					Factory: func() interface{} {
						return &testEntityData{
							testPbData: &testPbData{},
						}
					},
				},
			},
			want:    &v2.SearchOutput{Count: 3},
			wantErr: false,
		},
	}

	testEs.Bulk(context.Background(),
		[]*v2.BulkInput{
			{Cmd: "index", IndexName: "test", IndexType: "test", Id: 123, Data: testNewEntity(123, "Index")},
			{Cmd: "index", IndexName: "test", IndexType: "test", Id: 456, Data: testNewEntity(456, "Index")},
			{Cmd: "index", IndexName: "test", IndexType: "test", Id: 789, Data: testNewEntity(789, "Index")},
		})

	// let the data in ES flushed as it will give version conflicts
	testEs.Flush("test")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, err := v2.New([]string{testServer}, nil)
			if err != nil {
				t.Errorf("v2.Search() error = %v", err)
				return
			}

			got, err := e.Search(tt.args.ctx, tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("v2.Search() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// for _, v := range got.Results.([]*testEntityData) {
			// 	fmt.Println(v.testPbData)
			// }
			if got.Count != tt.want.Count {
				t.Errorf("v2.Search() = %v, want %v", got, tt.want)
			}
		})
	}
}

type testEntityData struct {
	*testPbData
}

type testPbData struct {
	Id          uint64  `json:"id,omitempty"`
	UserId      uint64  `json:"user_id,omitempty"`
	Name        string  `json:"name,omitempty"`
	Mobile      string  `json:"mobile,omitempty"`
	Amount      float64 `json:"amount,omitempty"`
	Subject     string  `json:"subject,omitempty"`
	Description string  `json:"description,omitempty"`
	Images      string  `json:"images,omitempty"`
	Audios      string  `json:"audios,omitempty"`
	Videos      string  `json:"videos,omitempty"`
	Attachments string  `json:"attachments,omitempty"`
	Comment     string  `json:"comment,omitempty"`
}

func testNewEntity(id uint64, prefix string) *testEntityData {
	switch prefix {
	case "Index":
		return &testEntityData{
			testPbData: &testPbData{
				Id:          id,
				Comment:     prefix + " " + "a good long long comment",
				Mobile:      prefix + " " + "+919999999999",
				Description: prefix + " " + "my very long long long description",
				Amount:      float64(id) + 23.4567,
			},
		}
	case "Update":
		return &testEntityData{
			testPbData: &testPbData{
				Id:      id,
				Comment: prefix + " " + "a good long long comment" + " " + fmt.Sprintf("%d", id),
				Amount:  float64(id) + 23.4567,
			},
		}
	}
	return nil
}

type testLogger struct {
}

func (t *testLogger) Println(v ...interface{}) {
	fmt.Println(v...)
}

type testConfig struct {
}

func (t *testConfig) Get(key string) interface{} {
	switch key {
	case "shard_num":
		return 2
	case "replica_num":
		return 1
	}
	return nil
}

func testSetup() {
	testEs, _ = v2.New([]string{testServer}, nil)
	testEs.CreateIndex(context.Background(), "test", "test", v2.DefaultMapping(testMappingFile), map[string]interface{}{"number_of_shards": 2, "number_of_replicas": 2})
}

func testTearDown() {
	testEs, _ = v2.New([]string{testServer}, nil)
}

func TestMain(m *testing.M) {
	testSetup()
	m.Run()
	testTearDown()
}
