package salutespeech

type UploadResponse struct {
	Status int `json:"status"`
	Result struct {
		RequestFileId string `json:"request_file_id"`
	} `json:"result"`
}

type RecognizeResponse struct {
	Status int `json:"status"`
	Result struct {
		TaskId    string `json:"id"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		Status    string `json:"status"`
	} `json:"result"`
}

type StatusResponse struct {
	Status int `json:"status"`
	Result struct {
		TaskId         string `json:"id"`
		CreatedAt      string `json:"created_at"`
		UpdatedAt      string `json:"updated_at"`
		Status         string `json:"status"`
		ResponseFileId string `json:"response_file_id"`
		Error          string `json:"error"`
	} `json:"result"`
}

type DownloadResponse struct {
	Status  int `json:"status"`
	Results []struct {
		Text           string `json:"text"`
		NormalizedText string `json:"normalized_text"`
		ResponseFileId string `json:"response_file_id"`
	} `json:"results"`
	Eou         bool `json:"eou"`
	Channel     int  `json:"channel"`
	SpeakerInfo struct {
		SpeakerId int `json:"speaker_id"`
	} `json:"speaker_info"`
}
