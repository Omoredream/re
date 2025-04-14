package jitoHTTP

type apiRequest struct {
	JsonRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Method  string `json:"method"`
	Params  []any  `json:"params"`
}

type apiResponse[T any] struct {
	JsonRPC string            `json:"jsonrpc"`
	ID      int               `json:"id"`
	Result  *T                `json:"result,omitempty"`
	Error   *apiResponseError `json:"error,omitempty"`
}

type apiResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type apiResponseContext[T any] struct {
	Context struct {
		Slot uint64 `json:"slot"`
	} `json:"context"`
	Value T `json:"value"`
}
