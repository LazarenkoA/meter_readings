package node_mos_ru

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

type MosruAdapter struct {
	meters   []map[string]interface{}
	login    string
	password string
}

type ReadingsItem struct {
	Indication float64 `json:"indication"`
	DeviceId   string  `json:"device_id"`
	Period     string  `json:"period"`
}

type Readings struct {
	PayerCode string         `json:"payer_code"`
	Flat      string         `json:"flat"`
	Items     []ReadingsItem `json:"items"`
}

func NewMosruAdapter(login, password string) *MosruAdapter {
	return &MosruAdapter{
		login:    login,
		password: password,
	}
}

func (m *MosruAdapter) GetMeters(ctx context.Context) ([]map[string]interface{}, error) {
	if len(m.meters) > 0 {
		return m.meters, nil
	}

	outData, err := m.run(ctx, "getMeters")
	if err != nil {
		return nil, err
	}

	var data []map[string]interface{}
	err = json.Unmarshal(outData, &data)

	m.meters = data
	return data, err
}

func (m *MosruAdapter) SendReadingsWater(ctx context.Context, data *Readings) error {
	if len(data.Items) == 0 {
		return errors.New("no data transmitted")
	}

	bdata, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = m.run(ctx, "sendMeterReadings", string(bdata))
	return err
}

func (m *MosruAdapter) run(ctx context.Context, command string, addParams ...string) ([]byte, error) {
	pathToNode := os.Getenv("NODE_PATH")
	scriptRoot := os.Getenv("NODE_SCRIPT_ROOT")

	params := append(append([]string{}, filepath.Join(scriptRoot, "main.js"), command, m.login, m.password), addParams...)

	log.Println("run node script")
	cmd := exec.CommandContext(ctx, pathToNode, params...)
	_ = cmd.Wait()

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout, cmd.Stderr = &outBuf, &errBuf

	if err := cmd.Run(); err != nil {
		log.Printf("STDERR:\n%s", errBuf.String())
		return nil, errors.Wrap(err, "running browser script")
	}

	return outBuf.Bytes(), nil
}
