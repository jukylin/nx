package value_object

const (
	// 开始
	TranStart = 0

	// 结束
	TranEnd = 1

	// 补偿
	TranCompensate = 2

	// 补偿完成
	TranCompensateFinish = 3
)

type State struct {
	val int
}

func NewState(val int) *State {
	s := &State{}
	s.val = val
	return s
}

func (s *State) Set(val int) {
	s.val = val
}

func (s *State) Value() int {
	return s.val
}

func (s *State) String() string {
	switch s.val {
	case TranStart:
		return "开始"
	case TranEnd:
		return "结束"
	case TranCompensate:
		return "补偿"
	case TranCompensateFinish:
		return "补偿完成"
	}

	return "未知状态"
}
