package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	errTest = errors.New("testing")
	client  = &http.Client{Timeout: time.Second}
)

type User struct {
	Id     int
	Name   string
	Age    int
	About  string
	Gender string
}

type SearchResponse struct {
	Users    []User
	NextPage bool
}

type SearchErrorResponse struct {
	Error string
}

const (
	OrderByAsc  = -1
	OrderByAsIs = 0
	OrderByDesc = 1

	ErrorBadOrderField = `OrderField invalid`
)

type SearchRequest struct {
	Limit      int
	Offset     int    // Можно учесть после сортировки
	Query      string // подстрока в 1 из полей
	OrderField string
	OrderBy    int
}

type SearchClient struct {
	// токен, по которому происходит авторизация на внешней системе, уходит туда через хедер
	AccessToken string
	// урл внешней системы, куда идти
	URL string
}

// FindUsers отправляет запрос во внешнюю систему, которая непосредственно ищет пользоваталей
func (srv *SearchClient) FindUsers(req SearchRequest) (*SearchResponse, error) {

	searcherParams := url.Values{}

	if req.Limit < 0 {
		return nil, fmt.Errorf("limit must be > 0")
	}
	if req.Limit > 25 {
		req.Limit = 25
	}
	if req.Offset < 0 {
		return nil, fmt.Errorf("offset must be > 0")
	}

	//нужно для получения следующей записи, на основе которой мы скажем - можно показать переключатель следующей страницы или нет
	req.Limit++

	searcherParams.Add("limit", strconv.Itoa(req.Limit))
	searcherParams.Add("offset", strconv.Itoa(req.Offset))
	searcherParams.Add("query", req.Query)
	searcherParams.Add("order_field", req.OrderField)
	searcherParams.Add("order_by", strconv.Itoa(req.OrderBy))

	searcherReq, err := http.NewRequest("GET", srv.URL+"?"+searcherParams.Encode(), nil)
	searcherReq.Header.Add("AccessToken", srv.AccessToken)

	resp, err := client.Do(searcherReq)
	if err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			return nil, fmt.Errorf("timeout for %s", searcherParams.Encode())
		}
		return nil, fmt.Errorf("unknown error %s", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return nil, fmt.Errorf("Bad AccessToken")
	case http.StatusInternalServerError:
		return nil, fmt.Errorf("SearchServer fatal error")
	case http.StatusBadRequest:
		errResp := SearchErrorResponse{}
		err = json.Unmarshal(body, &errResp)
		if err != nil {
			return nil, fmt.Errorf("cant unpack error json: %s", err)
		}
		if errResp.Error == "ErrorBadOrderField" {
			return nil, fmt.Errorf("OrderFeld %s invalid", req.OrderField)
		}
		return nil, fmt.Errorf("unknown bad request error: %s", errResp.Error)
	}

	data := []User{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, fmt.Errorf("cant unpack result json: %s", err)
	}

	result := SearchResponse{}
	if len(data) == req.Limit {
		result.NextPage = true
		result.Users = data[0 : len(data)-1]
	} else {
		result.Users = data[0:len(data)]
	}

	return &result, err
}

type XMLUser struct {
	Id        int    `xml:"id"`
	Age       int    `xml:"age"`
	About     string `xml:"about"`
	Gender    string `xml:"gender"`
	FirstName string `xml:"first_name"`
	LastName  string `xml:"last_name"`
}

type XMLUsers struct {
	XMLName xml.Name  `xml:"root"`
	Rows    []XMLUser `xml:"row"`
}

//func parser() {
//	// read file dataset.xml
//	data, err := os.ReadFile("dataset.xml")
//	if err != nil {
//		fmt.Println("Error reading dataset.xml:", err)
//		return
//	}
//
//	// parsing XML
//	var users XMLUsers
//	err = xml.Unmarshal(data, &users)
//	if err != nil {
//		fmt.Println("Error parsing dataset.xml:", err)
//		return
//	}
//
//	//print data
//	for _, user := range users.Rows {
//		fmt.Printf("ID: %d, Name: %s, Age: %d, Gender: %s, About: %s\n", user.Id, user.Name, user.Age, user.Gender, user.About)
//	}
//}

func SortUsers(users []XMLUser, orderField string, orderBy int) error {
	validField := map[string]bool{
		"ID":   true,
		"Name": true,
		"Age":  true,
	}
	if orderField == "" {
		orderField = "Name"
	} else if !validField[orderField] {
		return errors.New("OrderField invalid")
	}

	if orderBy == OrderByAsIs {
		return nil
	}

	sort.Slice(users, func(i, j int) bool {
		switch orderField {
		case "ID":
			if orderBy == OrderByDesc {
				return users[i].Id > users[j].Id
			}
			return users[i].Id < users[j].Id
		case "Age":
			if orderBy == OrderByDesc {
				return users[i].Age > users[j].Age
			}
			return users[i].Age < users[j].Age
		case "Name":
			NameI := users[i].FirstName + " " + users[i].LastName
			NameJ := users[j].FirstName + " " + users[j].LastName
			if orderBy == OrderByDesc {
				return NameI > NameJ
			}
			return NameI < NameJ
		}
		return false
	})
	return nil
}

func main() {
	query := "dolor"
	orderField := "Name"
	orderBy := OrderByAsc
	limit := 2
	offset := 2

	// read file dataset.xml
	data, err := os.ReadFile("dataset.xml")
	if err != nil {
		fmt.Println("Error reading dataset.xml:", err)
		return
	}

	// parsing XML
	var users XMLUsers
	err = xml.Unmarshal(data, &users)
	if err != nil {
		fmt.Println("Error parsing dataset.xml:", err)
		return
	}

	// Фильтрация по query
	var filteredUsers []XMLUser

	// Если query пустое, возвращаем все записи без фильтрации
	if query == "" {
		filteredUsers = users.Rows
	} else {
		// Если query не пустое, ищем подстроку в Name и About
		for _, user := range users.Rows {
			fullName := strings.ToLower(user.FirstName + " " + user.LastName)
			about := strings.ToLower(user.About)

			// Добавляем в результат, если query встречается в имени или описании
			if strings.Contains(fullName, strings.ToLower(query)) || strings.Contains(about, strings.ToLower(query)) {
				filteredUsers = append(filteredUsers, user)
			}
		}
	}

	err = SortUsers(filteredUsers, orderField, orderBy)
	if err != nil {
		fmt.Println("Error sorting:", err)
		return
	}

	// Применяем offset и limit
	if offset > len(filteredUsers) {
		fmt.Println("[]") // Если offset выходит за границы — пустой массив
		return
	}

	end := offset + limit
	if end > len(filteredUsers) {
		end = len(filteredUsers)
	}

	//print data
	for _, user := range filteredUsers[offset:end] {
		fmt.Printf("ID: %d, Name: %s %s, Age: %d, Gender: %s, About: %s\n", user.Id, user.FirstName, user.LastName, user.Age, user.Gender, user.About)
	}
}
