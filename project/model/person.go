package person

import (
	"encoding/json"
)


type Pessoa struct {
	ID 			string 		`json:"id,omitempty"`
	Apelido 	string 		`json:"apelido,omitempty"`
	Nome 		string		`json:"nome,omitempty"`
	Nascimento 	string		`json:"nascimento,omitempty"`
	Stack 		[]string	`json:"stack,omitempty"`
}