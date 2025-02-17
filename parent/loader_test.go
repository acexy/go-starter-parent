package parent

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

// redis module
type redis struct {
}

func (r redis) Setting() *Setting {
	return NewSetting("", 3, true, time.Second*3, nil)
}

func (r redis) Start() (interface{}, error) {
	time.Sleep(time.Second)
	return &redis{}, nil
}

func (r redis) Stop(maxWaitTime time.Duration) (gracefully bool, stopped bool, err error) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	go func() {
		defer cancelFunc()
		time.Sleep(time.Second)
	}()
	select {
	case <-time.After(maxWaitTime):
		return false, true, errors.New("timeout")
	case <-ctx.Done():
		return true, true, err
	}
}

// gorm module
type gorm struct {
}

func (g gorm) Setting() *Setting {
	return NewSetting("gorm", 20, true, time.Second, func(instance interface{}) {
		_, ok := instance.(*gorm)
		if ok {
			fmt.Println("init invoke")
		}
	})
}

func (g gorm) Start() (interface{}, error) {
	return &gorm{}, nil
}

func (g gorm) Stop(maxWaitTime time.Duration) (gracefully bool, stopped bool, err error) {
	time.Sleep(time.Second)
	return true, true, err
}

// gin module
type gin struct {
}

func (g gin) Setting() *Setting {
	return NewSetting("gin", 0, false, time.Second, nil)
}

func (g gin) Start() (interface{}, error) {
	return &gin{}, nil
}

func (g gin) Stop(maxWaitTime time.Duration) (gracefully bool, stopped bool, err error) {
	return false, false, errors.New("something error")
}

var starters []Starter

func init() {
	starters = []Starter{
		&redis{},
		&gorm{},
		&gin{},
	}
}

func showStopResult(result []*StopResult) {
	for _, v := range result {
		fmt.Printf("%+v\n", v)
	}
}

// Test

func TestStartAndStop(t *testing.T) {
	loader := NewStarterLoader(starters)
	err := loader.Start()
	if err != nil {
		println(err)
		return
	}
	err = loader.Start() // 重复启动

	result, err := loader.Stop(time.Second)
	if err != nil {
		println(err)
	}
	showStopResult(result)
}

func TestStartAndStopBySetting(t *testing.T) {
	loader := NewStarterLoader(starters)
	err := loader.Start()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("all started")
	time.Sleep(time.Second * 3)
	fmt.Println("stopping ...")
	result, err := loader.StopBySetting()
	if err != nil {
		fmt.Println(err)
		return
	}
	showStopResult(result)
}

func TestStartAndStopBySettingTimeout(t *testing.T) {
	loader := NewStarterLoader(starters)
	err := loader.Start()
	if err != nil {
		println(err)
		return
	}
	time.Sleep(time.Second * 3)
	result, err := loader.StopBySetting(time.Millisecond * 200)
	if err != nil {
		fmt.Println(err)
		return
	}
	showStopResult(result)
}

func TestStarterControl(t *testing.T) {
	loader := NewStarterLoader(starters)
	err := loader.Start()
	if err != nil {
		fmt.Println(err)
		return
	}
	result, err := loader.StopStarter("gorm", time.Second)
	if err != nil {
		println(err)
	}
	showStopResult([]*StopResult{result})
	fmt.Println(loader.StoppedStarters())
	_ = loader.Start()
	fmt.Println(loader.StoppedStarters())
}
