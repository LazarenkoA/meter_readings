package node_mos_ru

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"os"
	"os/exec"
)

type MosruAdapter struct {
	meters   []string
	login    string
	password string
}

func NewMosruAdapter(login, password string) *MosruAdapter {
	return &MosruAdapter{
		login:    login,
		password: password,
	}
}

func (m *MosruAdapter) GetMeters(ctx context.Context) ([]string, error) {
	pathToNode := os.Getenv("NODE_PATH")

	if len(m.meters) > 0 {
		return m.meters, nil
	}

	cmd := exec.CommandContext(ctx, pathToNode, "./node_mos_ru/main.js", "getMeters", m.login, m.password)
	_ = cmd.Wait()

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, errors.Wrap(err, "running browser script")
	}

	var data []string
	err = json.Unmarshal(output, &data)

	m.meters = data
	return data, err
}

func (m *MosruAdapter) SendReadingsWater(ctx context.Context) error {

	return nil
}
