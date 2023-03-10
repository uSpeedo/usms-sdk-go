package mock

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/uSpeedo/usms-sdk-go/um"
	"github.com/uSpeedo/usms-sdk-go/um/auth"
	"github.com/uSpeedo/usms-sdk-go/um/request"
	"github.com/uSpeedo/usms-sdk-go/um/response"

	proto "github.com/uSpeedo/usms-sdk-go/private/protocol/http"
)

func newTestClient() *um.Client {
	cfg := um.NewConfig()
	credential := auth.NewCredential()
	return um.NewClient(&cfg, &credential)
}

func newMockedHttpClient() *HttpClient {
	c := NewHttpClient()

	_ = c.MockHTTP(func(req *proto.HttpRequest, resp *proto.HttpResponse) error {
		if action := req.GetQuery("Action"); len(action) != 0 && action == "TestMockHTTP" {
			_ = resp.SetBody([]byte(`{"Action": "TestMockHTTPResponse"}`))
		}
		return nil
	})

	_ = c.MockData(func(reqData Request, respData Response) error {
		if action, ok := reqData["Action"]; ok && action == "TestMockData" {
			respData["Action"] = "TestMockDataResponse"
		}
		return nil
	})

	_ = c.MockData(func(reqData Request, respData Response) error {
		if action, ok := reqData["Action"]; ok && action == "TestMockError" {
			return http.ErrServerClosed
		}
		return nil
	})
	return c
}

func TestHttpClient_Send(t *testing.T) {
	httpClient := newMockedHttpClient()

	type args struct {
		Action string
	}

	tests := []struct {
		name    string
		client  *HttpClient
		args    args
		want    string
		wantErr bool
		err     error
	}{
		{"http", httpClient, args{"TestMockHTTP"}, `{"Action":"TestMockHTTPResponse"}`, false, nil},
		{"data", httpClient, args{"TestMockData"}, `{"Action":"TestMockDataResponse"}`, false, nil},
		{"error", httpClient, args{"TestMockError"}, "", true, http.ErrServerClosed},
	}

	// test for mocked http client
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := proto.NewHttpRequest()
			_ = req.SetQuery("Action", tt.args.Action)
			got, err := tt.client.Send(req)

			if tt.wantErr {
				// assert response error
				assert.Error(t, err)
				assert.Equal(t, tt.err, err)
			} else {
				// assert response body
				assert.NoError(t, err)
				assert.Equal(t, tt.want, string(got.GetBody()))
			}
		})
	}
}

func TestMockClient(t *testing.T) {
	client := newTestClient()
	httpClient := newMockedHttpClient()
	_ = client.SetHttpClient(httpClient)

	type args struct {
		Action string
	}

	tests := []struct {
		name    string
		client  *um.Client
		args    args
		want    string
		wantErr bool
	}{
		{"http", client, args{"TestMockHTTP"}, `TestMockHTTPResponse`, false},
		{"data", client, args{"TestMockData"}, `TestMockDataResponse`, false},
		{"error", client, args{"TestMockError"}, "", true},
	}

	// test for mocked um client
	for _, tt := range tests {
		req := &request.CommonBase{}
		resp := &response.CommonBase{}
		t.Run(tt.name, func(t *testing.T) {
			err := client.InvokeAction(tt.args.Action, req, resp)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tt.want, resp.Action)
			}
		})
	}
}
