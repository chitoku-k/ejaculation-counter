package service

import "github.com/chitoku-k/ejaculation-counter/reactor/repository"

type through struct {
	Repository repository.ThroughRepository
}

type Through interface {
	Get() []string
}

func NewThrough(repository repository.ThroughRepository) Through {
	return &through{
		Repository: repository,
	}
}

func (ts *through) Get() []string {
	return ts.Repository.Get()
}
