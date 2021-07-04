package main

import "time"

type SubjectResponse struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type RumorResponse struct {
	Title  string     `json:"title"`
	Date   *time.Time `json:"date"`
	Source string     `json:"source"`
	Body   string     `json:"body"`
}
