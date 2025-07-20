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
	scriptRoot := os.Getenv("NODE_SCRIPT_ROOT")
	_ = scriptRoot
	if len(m.meters) > 0 {
		return m.meters, nil
	}

	cmd := exec.CommandContext(ctx, pathToNode, filepath.Join(scriptRoot, "main.js"), "getMeters", m.login, m.password)
	_ = cmd.Wait()

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout, cmd.Stderr = &outBuf, &errBuf

	if err := cmd.Run(); err != nil {
		log.Printf("STDERR:\n%s", errBuf.String())
		return nil, errors.Wrap(err, "running browser script")
	}

	var data []string
	err := json.Unmarshal(outBuf.Bytes(), &data)

	m.meters = data
	return data, err
}

func (m *MosruAdapter) SendReadingsWater(ctx context.Context) error {

	return nil
}
