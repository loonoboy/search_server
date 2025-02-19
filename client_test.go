package main

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
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
				Limit:      0,
				Offset:     0,
				Query:      "dolor",
				OrderField: "Name",
				OrderBy:    OrderByAsc,
			},
			Result: &SearchResponse{
				Users:    []User{},
				NextPage: true,
			},
			IsError: false,
		},
		TestCase{
			Req: SearchRequest{
				Limit:      2,
				Offset:     0,
				Query:      "dolor",
				OrderField: "",
				OrderBy:    OrderByDesc,
			},
			Result: &SearchResponse{
				Users: []User{{Id: 13, Name: "", Age: 40, About: "Consectetur dolore anim veniam aliqua deserunt officia eu. Et ullamco commodo ad officia duis ex incididunt proident consequat nostrud proident quis tempor. Sunt magna ad excepteur eu sint aliqua eiusmod deserunt proident. Do labore est dolore voluptate ullamco est dolore excepteur magna duis quis. Quis laborum deserunt ipsum velit occaecat est laborum enim aute. Officia dolore sit voluptate quis mollit veniam. Laborum nisi ullamco nisi sit nulla cillum et id nisi.\n", Gender: "male"},
					{Id: 33, Name: "", Age: 36, About: "Sint non sunt adipisicing sit laborum cillum magna nisi exercitation. Dolore officia esse dolore officia ea adipisicing amet ea nostrud elit cupidatat laboris. Proident culpa ullamco aute incididunt aute. Laboris et nulla incididunt consequat pariatur enim dolor incididunt adipisicing enim fugiat tempor ullamco. Amet est ullamco officia consectetur cupidatat non sunt laborum nisi in ex. Quis labore quis ipsum est nisi ex officia reprehenderit ad adipisicing fugiat. Labore fugiat ea dolore exercitation sint duis aliqua.\n", Gender: "female"}},
				NextPage: true,
			},
			IsError: false,
		},
		TestCase{
			Req: SearchRequest{
				Limit:      2,
				Offset:     0,
				Query:      "",
				OrderField: "Age",
				OrderBy:    OrderByDesc,
			},
			Result: &SearchResponse{
				Users: []User{{Id: 32, Name: "", Age: 40, About: "Incididunt culpa dolore laborum cupidatat consequat. Aliquip cupidatat pariatur sit consectetur laboris labore anim labore. Est sint ut ipsum dolor ipsum nisi tempor in tempor aliqua. Aliquip labore cillum est consequat anim officia non reprehenderit ex duis elit. Amet aliqua eu ad velit incididunt ad ut magna. Culpa dolore qui anim consequat commodo aute.\n", Gender: "female"},
					{Id: 13, Name: "", Age: 40, About: "Consectetur dolore anim veniam aliqua deserunt officia eu. Et ullamco commodo ad officia duis ex incididunt proident consequat nostrud proident quis tempor. Sunt magna ad excepteur eu sint aliqua eiusmod deserunt proident. Do labore est dolore voluptate ullamco est dolore excepteur magna duis quis. Quis laborum deserunt ipsum velit occaecat est laborum enim aute. Officia dolore sit voluptate quis mollit veniam. Laborum nisi ullamco nisi sit nulla cillum et id nisi.\n", Gender: "male"}},
				NextPage: true,
			},
			IsError: false,
		},
		TestCase{
			Req: SearchRequest{
				Limit:      2,
				Offset:     0,
				Query:      "",
				OrderField: "God",
				OrderBy:    OrderByDesc,
			},
			Result:  nil,
			IsError: true,
		},
		TestCase{
			Req: SearchRequest{
				Limit:      0,
				Offset:     0,
				Query:      "",
				OrderField: "Age",
				OrderBy:    OrderByAsc,
			},
			Result: &SearchResponse{
				Users:    []User{},
				NextPage: true,
			},
			IsError: false,
		},
		TestCase{
			Req: SearchRequest{
				Limit:      -1,
				Offset:     0,
				Query:      "",
				OrderField: "Age",
				OrderBy:    OrderByAsIs,
			},
			Result:  nil,
			IsError: true,
		},
		TestCase{
			Req: SearchRequest{
				Limit:      0,
				Offset:     -1,
				Query:      "",
				OrderField: "Age",
				OrderBy:    OrderByAsIs,
			},
			Result:  nil,
			IsError: true,
		},
		TestCase{
			Req: SearchRequest{
				Limit:      26,
				Offset:     1,
				Query:      "Boyd",
				OrderField: "ID",
				OrderBy:    OrderByAsIs,
			},
			Result: &SearchResponse{
				Users:    []User{},
				NextPage: false,
			},
			IsError: false,
		},
		TestCase{
			Req: SearchRequest{
				Limit:      1,
				Offset:     0,
				Query:      "",
				OrderField: "ID",
				OrderBy:    OrderByDesc,
			},
			Result: &SearchResponse{
				Users:    []User{{Id: 34, Name: "", Age: 34, About: "Lorem proident sint minim anim commodo cillum. Eiusmod velit culpa commodo anim consectetur consectetur sint sint labore. Mollit consequat consectetur magna nulla veniam commodo eu ut et. Ut adipisicing qui ex consectetur officia sint ut fugiat ex velit cupidatat fugiat nisi non. Dolor minim mollit aliquip veniam nostrud. Magna eu aliqua Lorem aliquip.\n", Gender: "male"}},
				NextPage: true,
			},
			IsError: false,
		},
		TestCase{
			Req: SearchRequest{
				Limit:      0,
				Offset:     0,
				Query:      "",
				OrderField: "ID",
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
			AccessToken: "valid token",
			URL:         ts.URL,
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

func TestSearchClient_FindUsers_Extended(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("AccessToken")
		if token != "valid_token" {
			http.Error(w, "Bad AccessToken", http.StatusUnauthorized)
			return
		}

		if r.URL.Query().Get("query") == "timeout" {
			time.Sleep(time.Second * 2)
			return
		}

		if r.URL.Query().Get("query") == "bad_json" {
			w.Write([]byte("{invalid_json}"))
			return
		}

		if r.URL.Query().Get("limit") == "25" {
			w.Write([]byte(`[{"Id":1,"Name":"Max","Age":30,"About":"Test user","Gender":"male"}]`))
			return
		}

		if r.URL.Query().Get("offset") == "1000" {
			w.Write([]byte(`[]`))
			return
		}

		w.Write([]byte(`[{"Id":1,"Name":"John Doe","Age":30,"About":"Test","Gender":"male"}]`))
	}))
	defer ts.Close()

	cases := []struct {
		name    string
		req     SearchRequest
		result  *SearchResponse
		isError bool
	}{
		{
			"Invalid token",
			SearchRequest{Limit: 1, Offset: 0},
			nil,
			true,
		},
		{
			"Timeout error",
			SearchRequest{Limit: 1, Offset: 0, Query: "timeout"},
			nil,
			true,
		},
		{
			"Invalid JSON response",
			SearchRequest{Limit: 1, Offset: 0, Query: "bad_json"},
			nil,
			true,
		},
		{
			"Max limit (25)",
			SearchRequest{Limit: 25, Offset: 0},
			&SearchResponse{Users: []User{{Id: 1, Name: "John Doe", Age: 30, About: "Test", Gender: "male"}}, NextPage: false},
			false,
		},
		{
			"Offset out of range",
			SearchRequest{Limit: 1, Offset: 1000},
			&SearchResponse{Users: []User{}, NextPage: false},
			false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := &SearchClient{AccessToken: "valid_token", URL: ts.URL}
			if tc.name == "Invalid token" {
				c.AccessToken = "invalid_token"
			}
			result, err := c.FindUsers(tc.req)

			if (err != nil) != tc.isError {
				t.Errorf("%s: expected error %v, got %v", tc.name, tc.isError, err)
			}
			if !reflect.DeepEqual(tc.result, result) {
				t.Errorf("%s: expected %#v, got %#v", tc.name, tc.result, result)
			}
		})
	}
}

func TestSearchClient_HTTPStatuses(t *testing.T) {
	cases := []struct {
		name         string
		statusCode   int
		responseBody string
		expectedErr  string
	}{
		{"Unauthorized", http.StatusUnauthorized, ``, "Bad AccessToken"},
		{"Internal Server Error", http.StatusInternalServerError, ``, "SearchServerMock fatal error"},
		{"Bad Request (OrderField invalid)", http.StatusBadRequest, `{"error": "ErrorBadOrderField"}`, "OrderFeld  invalid"},
		{"Bad Request (other error)", http.StatusBadRequest, `{"error": "Some other error"}`, "unknown bad request error: Some other error"},
		{"Invalid JSON Response", http.StatusOK, `{invalid_json}`, "cant unpack result json"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)
				w.Write([]byte(tc.responseBody))
			}))
			defer ts.Close()

			client := &SearchClient{
				AccessToken: "valid_token",
				URL:         ts.URL,
			}

			_, err := client.FindUsers(SearchRequest{Limit: 1})

			if err == nil || !strings.Contains(err.Error(), tc.expectedErr) {
				t.Errorf("expected error containing %q, got %v", tc.expectedErr, err)
			}
		})
	}
}
