package tbot

import (
	"fmt"
	"github.com/robfig/cron/v3"
	"time"
)

type Schedule struct {
	с *cron.Cron
}

func NewSchedule() *Schedule {
	return &Schedule{
		с: cron.New(),
	}
}

func (s *Schedule) Planning(spec string, key string, f func()) error {
	_, err := s.с.AddFunc(spec, func() { // "10 20 1 7 *"
		fmt.Println("Задача выполнена:", time.Now())
	})

	return err
}

func (s *Schedule) store() {

}

func (s *Schedule) ReStore() {

}
