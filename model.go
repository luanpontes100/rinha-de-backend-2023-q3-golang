package main

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type pessoa struct {
	Id         uuid.UUID `json:"id"`
	Apelido    string    `json:"Apelido"`
	Nome       string    `json:"Nome"`
	Nascimento string    `json:"Nascimento"`
	Stack      []string  `json:"Stack"`
}

type pessoas []pessoa

var re = regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)

func (p *pessoa) validateApelido() bool {
	return len(p.Apelido) > 0 && len(p.Apelido) <= 32
}

func (p *pessoa) validateNome() bool {
	return len(p.Nome) > 0 && len(p.Nome) <= 100
}

func (p *pessoa) validateNascimento() bool {
	return re.MatchString(p.Nascimento)
}

func (p *pessoa) validateStackArray() bool {
	if len(p.Stack) == 0 {
		return true
	}
	for i := 0; i < len(p.Stack); i++ {
		if len(p.Stack[i]) > 32 {
			return false
		}
	}
	return true
}

func (p *pessoa) validate() error {
	validated := p.validateApelido()
	if !validated {
		return errors.New("apelido inválido, deve ter entre 1 e 32 caracteres")
	}
	validated = p.validateNome()
	if !validated {
		return errors.New("nome inválido, deve ter entre 1 e 100 caracteres")
	}
	validated = p.validateNascimento()
	if !validated {
		return errors.New("nascimento inválido, deve estar no formato yyyy-mm-dd")
	}
	validated = p.validateStackArray()
	if !validated {
		return errors.New("stack inválido, deve ser um array de 32 caracteres")
	}
	return nil
}

func (p *pessoa) createPerson(db *pgxpool.Pool) error {
	err := db.QueryRow(context.Background(), "INSERT INTO pessoas(Apelido, Nome, Nascimento, Stack) VALUES($1, $2, $3, $4) RETURNING id", p.Apelido, p.Nome, p.Nascimento, strings.Join(p.Stack, " | ")).Scan(&p.Id)
	if err != nil {
		err = errors.New("apelido já cadastrado")
	}
	return err
}

func (p *pessoa) getPerson(db *pgxpool.Pool, id string) error {
	return db.QueryRow(context.Background(), "SELECT ID, Apelido, Nome, Nascimento, string_to_array(Stack, ' | ') as Stack FROM pessoas WHERE id=$1", id).Scan(&p.Id, &p.Apelido, &p.Nome, &p.Nascimento, &p.Stack)
}

// Search persons using term on field busca using trgm extension
func (pessoas) searchPeople(db *pgxpool.Pool, term string) (pessoas, error) {
	pessoas := make(pessoas, 50)
	rows, err := db.Query(context.Background(), "SELECT ID, Apelido, Nome, Nascimento, string_to_array(Stack, ' | ') as stack FROM pessoas WHERE busca ilike '%' || $1 || '%' limit 50", term)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var i int
	for rows.Next() {
		var p pessoa
		err = rows.Scan(&p.Id, &p.Apelido, &p.Nome, &p.Nascimento, &p.Stack)
		if err != nil {
			return nil, err
		}
		pessoas[i] = p
		i++
	}
	return pessoas, nil
}

func (pessoas) totalPeople(db *pgxpool.Pool) (int, error) {
	var total int
	err := db.QueryRow(context.Background(), "SELECT COUNT(id) FROM pessoas").Scan(&total)
	return total, err
}
