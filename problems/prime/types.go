package prime

type Input struct {
	Method string   `json:"method"`
	Number *float64 `json:"number"`
}

type Output struct {
	Method  string `json:"method"`
	IsPrime bool   `json:"prime"`
}
