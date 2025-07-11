package models

// BacktestState はバックテストの状態を表す
type BacktestState int

const (
	BacktestStateIdle BacktestState = iota
	BacktestStateRunning
	BacktestStatePaused
	BacktestStateStopped
	BacktestStateCompleted
	BacktestStateError
)

// String はBacktestStateの文字列表現を返します
func (bs BacktestState) String() string {
	switch bs {
	case BacktestStateIdle:
		return "Idle"
	case BacktestStateRunning:
		return "Running"
	case BacktestStatePaused:
		return "Paused"
	case BacktestStateStopped:
		return "Stopped"
	case BacktestStateCompleted:
		return "Completed"
	case BacktestStateError:
		return "Error"
	default:
		return "Unknown"
	}
}

// BacktestControlState はバックテスト制御の状態を表す
type BacktestControlState struct {
	IsPlaying bool         `json:"is_playing"`
	Speed     float64      `json:"speed"`
	State     BacktestState `json:"state"`
}

// BacktestController はバックテストの制御を管理するインターフェース
type BacktestController interface {
	Play(speed float64) error
	Pause() error
	SetSpeed(speed float64) error
	GetState() BacktestControlState
	IsRunning() bool
}