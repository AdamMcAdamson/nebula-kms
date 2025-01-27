package responses

type AllowedResponse struct {
	Status    int         `json:"status"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
	IsAllowed bool        `json:"is_allowed"`
}
