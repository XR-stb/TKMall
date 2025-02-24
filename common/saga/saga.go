package saga

import (
	"context"
	"fmt"
)

type Step struct {
	Name       string
	Execute    func(context.Context) error
	Compensate func(context.Context) error
}

type Saga struct {
	steps    []Step
	executed []Step
}

func NewSaga() *Saga {
	return &Saga{
		steps:    make([]Step, 0),
		executed: make([]Step, 0),
	}
}

func (s *Saga) AddStep(step Step) {
	s.steps = append(s.steps, step)
}

func (s *Saga) Execute(ctx context.Context) error {
	for _, step := range s.steps {
		if err := step.Execute(ctx); err != nil {
			// 执行补偿
			return s.compensate(ctx, err)
		}
		s.executed = append(s.executed, step)
	}
	return nil
}

func (s *Saga) compensate(ctx context.Context, originalErr error) error {
	var compensationErr error

	// 反向执行补偿操作
	for i := len(s.executed) - 1; i >= 0; i-- {
		step := s.executed[i]
		if err := step.Compensate(ctx); err != nil {
			compensationErr = fmt.Errorf("compensation failed at step %s: %v", step.Name, err)
		}
	}

	if compensationErr != nil {
		return fmt.Errorf("original error: %v, compensation error: %v", originalErr, compensationErr)
	}

	return originalErr
}
