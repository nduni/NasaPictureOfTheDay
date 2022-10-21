package app

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	collectorModels "github.com/nduni/S3J5c3RpYW4gR29nb0FwcHMgTkFTQQ/url-collector/models/collector"
)

var correctUrl = url.URL{
	RawQuery: "from=2022-10-18&to=2022-10-19",
}

var (
	from, _      = time.ParseInLocation("2006-01-02", "2022-10-18", time.Local)
	to, _        = time.ParseInLocation("2006-01-02", "2022-10-19", time.Local)
	toEarlier, _ = time.ParseInLocation("2006-01-02", "2022-10-17", time.Local)
)

func Test_validatePicturesGet(t *testing.T) {
	tests := []struct {
		name    string
		req     *http.Request
		want    collectorModels.PicturesQueryParams
		wantErr bool
	}{
		{
			name: "correct request",

			req: &http.Request{
				URL: &url.URL{
					RawQuery: "from=2022-10-18&to=2022-10-19",
				},
				Body: http.NoBody,
			},

			want: collectorModels.PicturesQueryParams{From: from, To: to},

			wantErr: false,
		},
		{
			name: "correct request with same dates",

			req: &http.Request{
				URL: &url.URL{
					RawQuery: "from=2022-10-18&to=2022-10-18",
				},
				Body: http.NoBody,
			},

			want:    collectorModels.PicturesQueryParams{From: from, To: from},
			wantErr: false,
		},
		{
			name: "correct params, not empty body",

			req: &http.Request{
				URL:  &correctUrl,
				Body: ioutil.NopCloser(strings.NewReader("not_empty_body")),
			},

			want:    collectorModels.PicturesQueryParams{From: from, To: to},
			wantErr: true,
		},
		{
			name: "'to' param wrong format, empty body",
			req: &http.Request{
				URL: &url.URL{
					RawQuery: "from=2022-10-18&to=2022-10-100",
				},
				Body: http.NoBody,
			},

			want:    collectorModels.PicturesQueryParams{From: from},
			wantErr: true,
		},
		{
			name: "'to' date before 'from', not empty body",
			req: &http.Request{
				URL: &url.URL{
					RawQuery: "from=2022-10-18&to=2022-10-17",
				},
				Body: ioutil.NopCloser(strings.NewReader("not_empty_body")),
			},
			want:    collectorModels.PicturesQueryParams{From: from, To: toEarlier},
			wantErr: true,
		},
		{
			name:    "correct params, body produces error",
			req:     requestWithErrorBody(),
			want:    collectorModels.PicturesQueryParams{From: from, To: to},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := createContext(tt.req)

			got, err := validatePicturesGet(c)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePicturesGet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("validatePicturesGet() = %v, want %v", got, tt.want)
			}
		})
	}
}

func createContext(req *http.Request) *gin.Context {
	newRequest, err := http.NewRequest("", "/test", req.Body)
	if err != nil {
		log.Fatal(err)
	}
	newRequest.URL = req.URL
	w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
	c.Request = newRequest
	return c
}

func requestWithErrorBody() *http.Request {
	requestWithErrorBody := httptest.NewRequest("", "/test", errReader(false))
	requestWithErrorBody.URL = &correctUrl
	return requestWithErrorBody
}

type errReader bool

func (errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("body error")
}