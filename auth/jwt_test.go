package auth_test

import (
	"testing"

	"github.com/alokic/gopkg/auth"
	"github.com/alokic/gopkg/typeutils"
)

func TestJWT_Generate(t *testing.T) {
	type fields struct {
		key string
	}
	type args struct {
		claims map[string]interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name:   "TestGenerateFailEmptyToken",
			fields: fields{key: ""},
			args: args{
				claims: map[string]interface{}{
					"id": 1234, "svc": "test",
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name:   "TestGenerateSuccess",
			fields: fields{key: "my_test_key"},
			args: args{
				claims: map[string]interface{}{
					"id": 1234, "svc": "test",
				},
			},
			want:    "some token",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := auth.NewJWT(tt.fields.key)
			_, err := j.Generate(tt.args.claims)
			if (err != nil) != tt.wantErr {
				t.Errorf("JWT.Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestJWT_VerifyToken(t *testing.T) {
	type fields struct {
		key string
	}
	type args struct {
		jwttoken string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]interface{}
		wantErr bool
	}{
		{
			name:   "TestVerifyTokenFail",
			fields: fields{key: "my_test_key"},
			args: args{
				jwttoken: "bad token",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:   "TestVerifyTokenSuccess",
			fields: fields{key: "my_test_key"},
			args: args{
				jwttoken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJfdHMiOiIxNTMzMTgyNDA0NzU0IiwiaWQiOiIxMjM0Iiwic3ZjIjoidGVzdCJ9.99tfiCckTJQ4EJcKSwmEInwR0_C-mASXCui3e4w89sE",
			},
			want: map[string]interface{}{
				"id": 1234, "svc": "test",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := auth.NewJWT(tt.fields.key)
			got, err := j.VerifyToken(tt.args.jwttoken)
			if (err != nil) != tt.wantErr {
				t.Errorf("JWT.VerifyToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.want == nil && got != nil {
				t.Errorf("JWT.VerifyToken() = %v, want nil", got)
			}

			if tt.want != nil && got != nil && typeutils.ToInt(got["id"]) != tt.want["id"] {
				t.Errorf("JWT.VerifyToken() = %v, want %v", got["id"], tt.want["id"])
			}

			if tt.want != nil && got == nil {
				t.Errorf("JWT.VerifyToken() = nil, want %v", tt.want)
			}
		})
	}
}
