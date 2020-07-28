package service

import "github.com/chitoku-k/ejaculation-counter/reactor/repository"

type doublet struct {
	Repository repository.DoubletRepository
}

type Doublet interface {
	Get() []string
}

func NewDoublet(repository repository.DoubletRepository) Doublet {
	return &doublet{
		Repository: repository,
	}
}

func (ts *doublet) Get() []string {
	return ts.Repository.Get()
}
