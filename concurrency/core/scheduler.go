package core

import "github.com/jiansoft/chopper/system"

type IScheduler interface {
    Schedule(firstInMs int64, taskFun interface{}, params ...interface{}) (d system.IDisposable)
    ScheduleOnInterval(firstInMs int64, regularInMs int64, taskFun interface{}, params ...interface{}) (d system.IDisposable)
    Start()
    Stop()
    Dispose()
}

type ISchedulerRegistry interface {
    Enqueue(taskFun interface{}, params ...interface{})
    EnqueueWithTask(task Task)
    Remove(d system.IDisposable)
}

type Scheduler struct {
    fiber       IExecutionContext
    running     bool
    isDispose   bool
    disposabler *Disposer
}

func (s *Scheduler) init(executionState IExecutionContext) *Scheduler {
    s.fiber = executionState
    s.running = true
    s.disposabler = NewDisposer()
    return s
}

func NewScheduler(executionState IExecutionContext) *Scheduler {
    return new(Scheduler).init(executionState)
}

func (s *Scheduler) Schedule(firstInMs int64, taskFun interface{}, params ...interface{}) (d system.IDisposable) {
    if firstInMs <= 0 {
        pendingAction := newPendingTask(Task{Func: taskFun, Params: params})
        s.Enqueue(pendingAction.Execute)
        return pendingAction
    }
    return s.ScheduleOnInterval(firstInMs, -1, taskFun, params...)
}

func (s *Scheduler) ScheduleOnInterval(firstInMs int64, regularInMs int64, taskFun interface{}, params ...interface{}) (d system.IDisposable) {
    pending := newTimerTask(s, Task{Func: taskFun, Params: params}, firstInMs, regularInMs)
    s.addPending(pending)
    return pending
}

//實作 ISchedulerRegistry.Enqueue
func (s *Scheduler) Enqueue(taskFun interface{}, params ...interface{}) {
    s.EnqueueWithTask(Task{Params: params, Func: taskFun})
}

func (s *Scheduler) EnqueueWithTask(task Task) {
    s.fiber.EnqueueWithTask(task)
}

//實作 ISchedulerRegistry.Remove
func (s *Scheduler) Remove(d system.IDisposable) {
    s.fiber.Enqueue(s.disposabler.Remove, d)
}

func (s *Scheduler) Start() {
    s.running = true
    s.isDispose = false
}

func (s *Scheduler) Stop() {
    s.running = false
}

func (s *Scheduler) Dispose() {
    s.Stop()
    s.isDispose = true
    s.disposabler.Dispose()
}

func (s *Scheduler) addPending(pending *timerTask) {
    if s.isDispose {
        return
    }
    s.disposabler.Add(pending)
    pending.schedule()
}
