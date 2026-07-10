package core

type ProgressFunc func(step int, total int, label string)

func noopProgress(step int, total int, label string) {}
