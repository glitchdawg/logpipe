package main

import "time"

type UserLog struct {
	Level   string    `json:"level"`
	Message string    `json:"message"`
	User    string    `json:"user"`
	Ts      time.Time `json:"ts"`
}

type Report struct {
	Levels     map[string]int
	Users      map[string]int
	Malformed  int
	Processed  int
}

