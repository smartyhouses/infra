package utils

const DefaultRequestLimit = 2

func CreateRequestLimitLock(limit int) func() func() {
	sem := make(chan struct{}, limit)
	acquire := func() { sem <- struct{}{} }
	release := func() { <-sem }

	lock := func() func() {
		acquire()
		return release
	}

	return lock
}