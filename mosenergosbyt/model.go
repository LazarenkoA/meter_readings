package mosenergosbyt

type response struct {
	Success      bool   `json:"success"`
	Total        int    `json:"total"`
	ErrorCode    int    `json:"err_code"`
	ErrorMessage string `json:"err_text"`
	ErrorID      string `json:"err_id"`
	Data         []struct {
		KdResult          int         `json:"kd_result"`
		NmResult          string      `json:"nm_result"`
		IdProfile         string      `json:"id_profile"`
		CntAuth           int         `json:"cnt_auth"`
		NewToken          interface{} `json:"new_token"`
		PrChangeMethodTfa interface{} `json:"pr_change_method_tfa"`
		MethodTfa         interface{} `json:"method_tfa"`
		VlTfaAuthToken    interface{} `json:"vl_tfa_auth_token"`
		VlTfaDeviceToken  interface{} `json:"vl_tfa_device_token"`
		PrShowCaptcha     interface{} `json:"pr_show_captcha"`
		Session           string      `json:"session"`
	} `json:"data"`
	MetaData struct {
		ResponseTime float64 `json:"responseTime"`
	} `json:"metaData"`
}
