package responses

type UserTypeResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    bool   `json:"data"`
}
