package deepgram

import (
	"context"
	api "github.com/deepgram/deepgram-go-sdk/v3/pkg/api/listen/v1/rest"
	interfaces "github.com/deepgram/deepgram-go-sdk/v3/pkg/client/interfaces/v1"
	client "github.com/deepgram/deepgram-go-sdk/v3/pkg/client/listen"
	"github.com/pkg/errors"
)

type DeepgramClient struct {
	key string
}

const (
	host = "https://api.deepgram.com"
)

func NewDeepgram(key string) *DeepgramClient {
	client.Init(client.InitLib{
		LogLevel: client.LogLevelStandard, // LogLevelStandard / LogLevelFull / LogLevelTrace / LogLevelVerbose
	})

	return &DeepgramClient{
		key: key,
	}
}

func (d *DeepgramClient) STT(ctx context.Context, filePath string) (string, error) {
	options := &interfaces.PreRecordedTranscriptionOptions{
		Model:       "nova-2-general",
		SmartFormat: true,
		Language:    "ru",
	}

	// create a Deepgram client
	c := client.NewREST(d.key, &interfaces.ClientOptions{Host: host})
	dg := api.New(c)

	// send/process file to Deepgram
	res, err := dg.FromFile(ctx, filePath, options)
	if err != nil {
		return "", errors.Wrap(err, "deepgram FromFile")
	}

	if res != nil && res.Results != nil && len(res.Results.Channels) > 0 && len(res.Results.Channels[0].Alternatives) > 0 {
		return res.Results.Channels[0].Alternatives[0].Transcript, nil
	}

	return "", errors.New("deepgram undefined error")
}
