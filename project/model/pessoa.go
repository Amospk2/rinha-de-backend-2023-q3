package model

type Pessoa struct {
	ID 			string 		`json:"id,omitempty"`
	Apelido 	string 		`json:"apelido" validate:"required,max=32"`
	Nome 		string		`json:"nome" validate:"required,max=100""`
	Nascimento 	string		`json:"nascimento" validate:"required,datetime=2006-01-02"`
	Stack 		[]string	`json:"stack,omitempty" validate:"dive,max=32"`
}

func NewPessoa(
	id string,
	Apelido string,
	Nome string,
	birthdate string,
	stack []string,
) *Pessoa {
	return &Pessoa{
		ID:        id,
		Apelido:  Apelido,
		Nome:      Nome,
		Nascimento: birthdate,
		Stack:     stack,
	}
}