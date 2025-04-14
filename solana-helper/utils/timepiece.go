package Utils

import (
	"strings"
	"time"

	"github.com/gogf/gf/v2/os/gtime"
)

type timepieceStep struct {
	name  string
	spent time.Duration
}

type Timepiece struct {
	start *gtime.Time
	steps []timepieceStep
	times map[string]*gtime.Time
}

func NewTimepiece() *Timepiece {
	return &Timepiece{
		start: gtime.Now(),
		steps: make([]timepieceStep, 0),
		times: make(map[string]*gtime.Time),
	}
}

func (t *Timepiece) StepStart(step string) {
	t.times[step] = gtime.Now()
}

func (t *Timepiece) StepFinish(step string) (spent time.Duration) {
	spent = gtime.Now().Sub(t.times[step])
	t.steps = append(t.steps, timepieceStep{step, spent})
	return
}

func (t *Timepiece) Report() string {
	log := strings.Builder{}
	log.WriteString(gtime.Now().Sub(t.start).String())
	switch len(t.steps) {
	case 0:
	case 1:
		log.WriteString(" (")

		log.WriteString(t.steps[0].name)
		log.WriteString(": ")
		log.WriteString(t.steps[0].spent.String())

		log.WriteString(")")
	default:
		log.WriteString(" (")

		log.WriteString(t.steps[0].name)
		log.WriteString(": ")
		log.WriteString(t.steps[0].spent.String())

		for _, step := range t.steps[1:] {
			log.WriteString(", ")
			log.WriteString(step.name)
			log.WriteString(": ")
			log.WriteString(step.spent.String())
		}

		log.WriteString(")")
	}
	return log.String()
}
