package main

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type TestCase struct {
	Req     SearchRequest
	Result  *SearchResponse
	IsError bool
}

func TestSearchClient_FindUsers(t *testing.T) {
	cases := []TestCase{
		TestCase{
			Req: SearchRequest{
				Limit:      28,
				Offset:     0,
				Query:      "",
				OrderField: "Name",
				OrderBy:    OrderByAsc,
			},
			Result: &SearchResponse{
				Users:    []User{},
				NextPage: true,
			},
			IsError: false,
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))

	for caseNum, item := range cases {
		c := &SearchClient{
			URL: ts.URL,
		}
		result, err := c.FindUsers(item.Req)

		if err != nil && !item.IsError {
			t.Errorf("[%d] unexpected error: %#v", caseNum, err)
		}
		if err == nil && item.IsError {
			t.Errorf("[%d] expected error, got nil", caseNum)
		}
		if !reflect.DeepEqual(item.Result, result) {
			t.Errorf("[%d] wrong result, expected %#v, got %#v", caseNum, item.Result, result)
		}
	}
	ts.Close()
}
