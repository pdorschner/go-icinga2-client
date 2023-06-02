package icinga2

import (
	"fmt"
	"net/url"
)

type Service struct {
	Name               string   `json:"name,omitempty"`
	DisplayName        string   `json:"display_name"`
	HostName           string   `json:"host_name"`
	CheckCommand       string   `json:"check_command"`
	EnableActiveChecks bool     `json:"enable_active_checks"`
	Notes              string   `json:"notes"`
	NotesURL           string   `json:"notes_url"`
	ActionURL          string   `json:"action_url"`
	Vars               Vars     `json:"vars,omitempty"`
	Zone               string   `json:"zone,omitempty"`
	CheckInterval      float64  `json:"check_interval"`
	RetryInterval      float64  `json:"retry_interval"`
	MaxCheckAttempts   float64  `json:"max_check_attempts"`
	CheckPeriod        string   `json:"check_period,omitempty"`
	State              float64  `json:"state,omitempty"`
	LastStateChange    float64  `json:"last_state_change,omitempty"`
	Templates          []string `json:"templates,omitempty"`
}

type ServiceResults struct {
	Results []struct {
		Service Service `json:"attrs"`
	} `json:"results"`
}

type ServiceCreate struct {
	Templates []string `json:"templates"`
	Attrs     Service  `json:"attrs"`
}

func (s Service) GetCheckCommand() string {
	return s.CheckCommand
}

func (s Service) GetVars() Vars {
	return s.Vars
}

func (s Service) GetNotes() string {
	return s.Notes
}

func (s Service) GetNotesURL() string {
	return s.NotesURL
}

func (s *Service) FullName() string {
	return s.HostName + "!" + s.Name
}

func (s *WebClient) GetService(name string) (Service, error) {
	var serviceResults ServiceResults
	resp, err := s.napping.Get(s.URL+"/v1/objects/services/"+name, nil, &serviceResults, nil)
	if err != nil {
		return Service{}, err
	}
	if resp.HttpResponse().StatusCode != 200 {
		return Service{}, fmt.Errorf("Did not get 200 OK")
	}
	return serviceResults.Results[0].Service, nil
}

func (s *WebClient) CreateService(service Service) error {
	serviceCreate := ServiceCreate{Templates: service.Templates, Attrs: service}
	// Strip "name" from create payload
	serviceCreate.Attrs.Name = ""
	err := s.CreateObject("/services/"+service.FullName(), serviceCreate)

	return err
}

func (s *WebClient) ListServices(query QueryFilter) (services []Service, err error) {
	var serviceResults ServiceResults
	services = []Service{}

	resp, err := s.FilteredQuery(s.URL+"/v1/objects/services", query, &serviceResults, nil)
	if err != nil {
		return
	}
	if resp.HttpResponse().StatusCode != 200 {
		return []Service{}, fmt.Errorf("Did not get 200 OK")
	}
	for _, result := range serviceResults.Results {
		if s.Zone == "" || s.Zone == result.Service.Zone {
			services = append(services, result.Service)
		}
	}

	return
}

func (s *WebClient) DeleteService(name string) (err error) {
	_, err = s.napping.Delete(s.URL+"/v1/objects/services/"+name, &url.Values{"cascade": []string{"1"}}, nil, nil)
	return
}

func (s *WebClient) UpdateService(service Service) error {
	serviceUpdate := ServiceCreate{Attrs: service}
	// Strip "name" from update payload
	serviceUpdate.Attrs.Name = ""

	err := s.UpdateObject("/services/"+service.FullName(), serviceUpdate)
	return err
}

func (s *MockClient) GetService(name string) (Service, error) {
	if sv, ok := s.Services[name]; ok {
		return sv, nil
	} else {
		return Service{}, fmt.Errorf("service not found")
	}
}

func (s *MockClient) CreateService(service Service) error {
	s.mutex.Lock()
	s.Services[service.FullName()] = service
	s.mutex.Unlock()
	return nil
}

func (s *MockClient) ListServices(query QueryFilter) ([]Service, error) {
	services := []Service{}

	for _, x := range s.Services {
		// TODO: implement list filtering for MockClient
		services = append(services, x)
	}

	return services, nil
}

func (s *MockClient) DeleteService(name string) error {
	s.mutex.Lock()
	delete(s.Services, name)
	s.mutex.Unlock()
	return nil
}

func (s *MockClient) UpdateService(service Service) error {
	s.mutex.Lock()
	s.Services[service.FullName()] = service
	s.mutex.Unlock()
	return nil
}
