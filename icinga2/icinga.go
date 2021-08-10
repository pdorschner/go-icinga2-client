package icinga2

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"gopkg.in/jmcvetta/napping.v3"
)

type QueryFilter struct {
	Filter string `json:"filter"`
}

type Client interface {
	GetHost(string) (Host, error)
	CreateHost(Host) error
	ListHosts(string) ([]Host, error)
	DeleteHost(string) error
	UpdateHost(Host) error

	GetHostGroup(string) (HostGroup, error)
	CreateHostGroup(HostGroup) error
	ListHostGroups(string) ([]HostGroup, error)
	DeleteHostGroup(string) error
	UpdateHostGroup(HostGroup) error

	ListDowntimes(QueryFilter) ([]Downtime, error)

	GetService(string) (Service, error)
	CreateService(Service) error
	ListServices(QueryFilter) ([]Service, error)
	DeleteService(string) error
	UpdateService(Service) error

	ProcessCheckResult(Service, Action) error
}

type WebClient struct {
	napping           napping.Session
	URL               string
	Username          string
	Password          string
	Debug             bool
	DisableKeepAlives bool
	Zone              string
	TLSConfig         *tls.Config
}

type MockClient struct {
	Hostgroups map[string]HostGroup
	Hosts      map[string]Host
	Services   map[string]Service
	Actions    map[string][]Action
	mutex      sync.Mutex
}

type Vars map[string]interface{}

type Checkable interface {
	GetCheckCommand() string
	GetVars() Vars
	GetNotes() string
	GetNotesURL() string
}

type Object interface {
	GetVars() Vars
}

func New(s WebClient) (*WebClient, error) {
	transport := &http.Transport{
		TLSClientConfig:   s.TLSConfig,
		DisableKeepAlives: s.DisableKeepAlives,
		ForceAttemptHTTP2: true,
		Proxy:             http.ProxyFromEnvironment,
	}
	client := &http.Client{Transport: transport}

	s.napping = napping.Session{
		Log:      s.Debug,
		Client:   client,
		Userinfo: url.UserPassword(s.Username, s.Password),
	}

	s.URL = strings.TrimRight(s.URL, "/")

	return &s, nil
}

func NewMockClient() (c *MockClient) {
	c = new(MockClient)
	c.Hostgroups = make(map[string]HostGroup)
	c.Hosts = make(map[string]Host)
	c.Services = make(map[string]Service)
	c.Actions = make(map[string][]Action)
	c.mutex = sync.Mutex{}
	return
}

type Results struct {
	Results []struct {
		Code   float64  `json:"code"`
		Errors []string `json:"errors,omitempty"`
		Status string   `json:"status,omitempty"`
		Name   string   `json:"name,omitempty"`
		Type   string   `json:"type,omitempty"`
	} `json:"results"`
}

func (s *WebClient) CreateObject(path string, create interface{}) error {
	var results, errmsg Results

	resp, err := s.napping.Put(s.URL+"/v1/objects"+path, create, &results, &errmsg)

	return s.handleResults("create", path, resp, &results, &errmsg, err)
}

func (s *WebClient) UpdateObject(path string, create interface{}) error {
	var results, errmsg Results

	resp, err := s.napping.Post(s.URL+"/v1/objects"+path, create, &results, &errmsg)
	return s.handleResults("update", path, resp, &results, &errmsg, err)
}

func (s *WebClient) FilteredQuery(url string, filter QueryFilter, result, errmsg interface{}) (*napping.Response, error) {
	header := http.Header{
		"Accept": []string{"application/json"},
	}
	req := napping.Request{
		Method:  "GET",
		Url:     url,
		Header:  &header,
		Payload: filter,
		Result:  result,
		Error:   errmsg,
	}
	return s.napping.Send(&req)
}

func (s *WebClient) handleResults(typ, path string, resp *napping.Response, results, errmsg *Results, oerr error) error {
	var resultReport string

	if oerr != nil {
		return oerr
	}

	for _, r := range results.Results {
		if r.Code >= 400.0 {
			resultReport += r.Status + " " + strings.Join(r.Errors, " ") + " "
		}
	}

	for _, r := range errmsg.Results {
		if r.Code >= 400.0 {
			resultReport += r.Status + " " + strings.Join(r.Errors, " ") + " "
		}
	}

	if resp.HttpResponse().StatusCode >= 400 {
		return fmt.Errorf("%s %s : %s - %s", typ, path, resp.HttpResponse().Status, resultReport)
	}

	if resultReport != "" {
		return fmt.Errorf("%s %s : %s\n", typ, path, resultReport)
	}

	return oerr

}
