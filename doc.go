// Metronome is a distributed and fault-tolerant event scheduler
//
// Agents
//
// Metronome as four agents:
//  - api: HTTP api to manage the tasks
//  - scheduler: plan tasks executions
//  - aggregator: maintain the databasse
//  - worker: perform HTTP POST to trigger remote systems according to the tasks schedules
package main
