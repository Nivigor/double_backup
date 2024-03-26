// writer
package main

import (
	"errors"
	"io"
	"slices"
)

type Answer struct {
	n   int
	err error
}

type multiWriter struct {
	n       int
	cmd     chan []byte
	answer  chan Answer
	writers []io.WriteCloser
}

func (w *multiWriter) Write(p []byte) (n int, err error) {
	var answer Answer
	var ns int = 0x7fffffff
	errs := make([]error, w.n)
	l := len(p)
	for i := 0; i < w.n; i++ {
		w.cmd <- p
	}
	for i := 0; i < w.n; i++ {
		answer = <-w.answer
		ns = min(ns, answer.n)
		errs[i] = answer.err
		if answer.err == nil && answer.n != l {
			errs[i] = io.ErrShortWrite
		}
	}
	return ns, errors.Join(errs...)
}

func (w *multiWriter) Close() error {
	errs := make([]error, w.n)
	close(w.cmd)
	for i := 0; i < w.n; i++ {
		answer := <-w.answer
		errs[i] = answer.err
	}
	return errors.Join(errs...)
}

func MultiWriter(writers ...io.WriteCloser) io.WriteCloser {
	w := multiWriter{}
	w.n = len(writers)
	w.writers = slices.Clone(writers)
	w.cmd = make(chan []byte, w.n)
	w.answer = make(chan Answer, w.n)

	write := func(ind int) {
		var answer Answer
		for {
			buf, opened := <-w.cmd
			if !opened {
				break
			}
			answer.n, answer.err = w.writers[ind].Write(buf)
			w.answer <- answer
		}
		answer.n, answer.err = 0, w.writers[ind].Close()
		w.answer <- answer
	}

	for i := 0; i < w.n; i++ {
		j := i
		go write(j)
	}
	return &w
}
