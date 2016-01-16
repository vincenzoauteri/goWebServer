package main

import (
   "runtime"
   "reflect"
   "time"
)

var profiling struct {
    Avg int
    ExecutionTime  time.Duration
    Executions int
    Samples int
}

func GetFunctionName(i interface{}) string {
    return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}
