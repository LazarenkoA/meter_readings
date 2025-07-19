package mosvodokanal

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// SetReadings передает показания счетчиков
func (m *Mosvodokanal) SetReadings(T1, T2, T3 int) error {

	return nil
}

func (m *Mosvodokanal) Auth() error {
	valuesBody := url.Values{
		"username": {m.login},
		"password": {m.pass},
		"captcha":  {},
	}

	request, err := http.NewRequestWithContext(m.ctx, http.MethodPost, baseURL+"/login", strings.NewReader(valuesBody.Encode()))
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

	return nil
}

func (m *Mosvodokanal) request(request *http.Request) ([]byte, error) {
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

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
