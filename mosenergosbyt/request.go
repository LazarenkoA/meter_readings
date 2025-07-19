package mosenergosbyt

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// SetReadings передает показания счетчиков
func (m *Mosenergosbyt) SetReadings(T1, T2, T3 int) error {
	if m.session == "" {
		return errors.New("you are not logged in")
	}

	values := url.Values{
		"action":  {"sql"},
		"query":   {"bytProxy"},
		"session": {m.session},
	}

	valuesBody := url.Values{
		"nn_phone":      {"+7 (499) 550-9-550"},
		"plugin":        {"bytProxy"},
		"pr_flat_meter": {"0"},
		"proxyquery":    {"CalcCharge"},
		"vl_provider":   {`{"id_kng": 173359, "nm_abn": 82}`},
		"vl_t1":         {strconv.Itoa(T1)},
		"vl_t2":         {strconv.Itoa(T2)},
		"vl_t3":         {strconv.Itoa(T3)},
	}

	request, err := http.NewRequestWithContext(m.ctx, http.MethodPost, baseURL+"?"+values.Encode(), strings.NewReader(valuesBody.Encode()))
	if err != nil {
		return err
	}

	b, err := m.request(request)
	if err != nil {
		return errors.Wrap(err, "http request error")
	}

	var data response
	err = json.Unmarshal(b, &data)
	if err != nil {
		return errors.Wrap(err, "json unmarshal error")
	}

	if data.ErrorMessage != "" {
		return errors.New(data.ErrorMessage)
	}

	if len(data.Data) > 0 {
		if data.Data[0].NmResult != "" {
			return m.confirm(T1, T2, T3)
		}
	}

	return nil
}

func (m *Mosenergosbyt) confirm(T1, T2, T3 int) error {
	if m.session == "" {
		return errors.New("you are not logged in")
	}

	values := url.Values{
		"action":  {"sql"},
		"query":   {"SaveIndications"},
		"session": {m.session},
	}

	valuesBody := url.Values{
		"plugin":        {"propagateMesInd"},
		"pr_flat_meter": {"0"},
		"vl_provider":   {`{"id_kng": 173359, "nm_abn": 82}`},
		"vl_t1":         {strconv.Itoa(T1)},
		"vl_t2":         {strconv.Itoa(T2)},
		"vl_t3":         {strconv.Itoa(T3)},
	}

	request, err := http.NewRequestWithContext(m.ctx, http.MethodPost, baseURL+"?"+values.Encode(), strings.NewReader(valuesBody.Encode()))
	if err != nil {
		return err
	}

	b, err := m.request(request)
	if err != nil {
		return errors.Wrap(err, "http request error")
	}

	var data response
	err = json.Unmarshal(b, &data)
	if err != nil {
		return errors.Wrap(err, "json unmarshal error")
	}

	if data.ErrorMessage != "" {
		return errors.New(data.ErrorMessage)
	}

	if len(data.Data) > 0 {
		if data.Data[0].NmResult == "Показания успешно переданы" {
			return nil
		} else {
			log.Printf("nm_result = %s\n", data.Data[0].NmResult)
			return errors.New("undefined error")
		}
	}

	return nil
}

func (m *Mosenergosbyt) Auth() error {
	if m.session != "" {
		return nil
	}

	values := url.Values{
		"action": {"auth"},
		"query":  {"login"},
	}

	valuesBody := url.Values{
		"login":          {m.login},
		"psw":            {m.pass},
		"remember":       {"true"},
		"vl_device_info": {deviceInfo},
	}

	request, err := http.NewRequestWithContext(m.ctx, http.MethodPost, baseURL+"?"+values.Encode(), strings.NewReader(valuesBody.Encode()))
	if err != nil {
		return err
	}

	b, err := m.request(request)
	if err != nil {
		return errors.Wrap(err, "http request error")
	}

	var data response
	err = json.Unmarshal(b, &data)
	if err != nil {
		return errors.Wrap(err, "json unmarshal error")
	}

	if data.ErrorMessage != "" {
		m.session = ""
		return errors.New(data.ErrorMessage)
	} else {
		m.session = data.Data[0].Session
	}

	return nil
}

func (m *Mosenergosbyt) request(request *http.Request) ([]byte, error) {
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := m.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("response status isn't 200 OK returns: %d", resp.StatusCode))
	}

	return io.ReadAll(resp.Body)
}
