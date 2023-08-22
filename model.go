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

func (p *pessoa) validateApelido() bool {
	return len(p.Apelido) > 0 && len(p.Apelido) <= 32
}

func (p *pessoa) validateNome() bool {
	return len(p.Apelido) > 0 && len(p.Apelido) <= 100
}

func (p *pessoa) validateNascimento() bool {
	re := regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)
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
	err := db.QueryRow(context.Background(), "INSERT INTO pessoas(Apelido, Nome, Nascimento, Stack) VALUES($1, $2, $3, $4) RETURNING id", p.Apelido, p.Nome, p.Nascimento, p.Stack).Scan(&p.Id)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			err = errors.New("apelido já cadastrado")
		}
	}
	return err
}

func (p *pessoa) getPerson(db *pgxpool.Pool, id string) error {
	return db.QueryRow(context.Background(), "SELECT ID, Apelido, Nome, Nascimento, Stack FROM pessoas WHERE id=$1", id).Scan(&p.Id, &p.Apelido, &p.Nome, &p.Nascimento, &p.Stack)
}

// Search persons using term on field busca using trgm extension
func (pessoas) searchPeople(db *pgxpool.Pool, term string) (pessoas, error) {
	pessoas := make(pessoas, 0)
	rows, err := db.Query(context.Background(), "SELECT ID, Apelido, Nome, Nascimento, Stack FROM pessoas WHERE busca %> lower(unaccent($1))", term)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var p pessoa
		err = rows.Scan(&p.Id, &p.Apelido, &p.Nome, &p.Nascimento, &p.Stack)
		if err != nil {
			return nil, err
		}
		pessoas = append(pessoas, p)
	}
	return pessoas, nil
}

func (pessoas) totalPeople(db *pgxpool.Pool) (int, error) {
	var total int
	err := db.QueryRow(context.Background(), "SELECT COUNT(id) FROM pessoas").Scan(&total)
	return total, err
}
