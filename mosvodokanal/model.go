package mosvodokanal

type response struct {
	Login         string `json:"login"`
	Fio           string `json:"fio"`
	Authenticated bool   `json:"authenticated"`
	UserType      string `json:"userType"`
	Email         string `json:"email"`
	Phone         string `json:"phone"`
}
