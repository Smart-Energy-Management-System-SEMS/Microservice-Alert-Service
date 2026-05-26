package valueobjects

type Operator string

const (
    OperatorGreaterThan        Operator = ">"
    OperatorGreaterThanOrEqual Operator = ">="
    OperatorLessThan           Operator = "<"
    OperatorLessThanOrEqual    Operator = "<="
    OperatorEqual              Operator = "=="
)
