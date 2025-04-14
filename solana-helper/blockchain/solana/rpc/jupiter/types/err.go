package jupiterTypes

type ErrorCode string

func (errorCode ErrorCode) Code() int {
	return 0
}
func (errorCode ErrorCode) Message() string {
	return string(errorCode)
}

func (errorCode ErrorCode) Detail() any {
	return nil
}

const (
	NotSupported                        ErrorCode = "NOT_SUPPORTED"
	CircularArbitrageIsDisabled         ErrorCode = "CIRCULAR_ARBITRAGE_IS_DISABLED"
	NoRoutesFound                       ErrorCode = "NO_ROUTES_FOUND"
	CouldNotFindAnyRoute                ErrorCode = "COULD_NOT_FIND_ANY_ROUTE"
	TokenNotTradable                    ErrorCode = "TOKEN_NOT_TRADABLE"
	RoutePlanDoesNotConsumeAllTheAmount ErrorCode = "ROUTE_PLAN_DOES_NOT_CONSUME_ALL_THE_AMOUNT"
)

type ErrorResponse struct {
	Error        string    `json:"error"`
	ErrorCode    ErrorCode `json:"errorCode"`
	ErrorCodeOld ErrorCode `json:"error_code"`
}
