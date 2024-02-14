package database

import (
	"api/model"
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/bytedance/sonic"
	"github.com/redis/rueidis"
)

var ctx = context.Background()

type Cache struct {
	client rueidis.Client
}

func (p *Cache) Get(key string) (*model.Pessoa, error) {
	getCmd := p.client.
		B().
		Get().
		Key("pessoa:" + key).
		Build()

	pessoaBytes, err := p.client.Do(ctx, getCmd).AsBytes()
	if err != nil {
		return nil, err
	}

	var pessoa model.Pessoa
	err = sonic.Unmarshal(pessoaBytes, &pessoa)
	if err != nil {
		return nil, err
	}

	return &pessoa, nil
}

func (p *Cache) GetApelido(apelido string) (bool, error) {
	getApelidoCmd := p.client.
		B().
		Getbit().
		Key("apelido:" + apelido).
		Offset(0).
		Build()

	return p.client.Do(ctx, getApelidoCmd).AsBool()
}

func (p *Cache) Set(pessoa *model.Pessoa) error {
	item, err := sonic.MarshalString(pessoa)
	if err != nil {
		return err
	}

	setPessoaCmd := p.client.
		B().
		Set().
		Key("pessoa:" + pessoa.ID).
		Value(item).
		Ex(time.Minute).
		Build()

	setApelidoCmd := p.client.
		B().
		Setbit().
		Key("apelido:" + pessoa.Apelido).
		Offset(0).
		Value(1).
		Build()

	cmds := make(rueidis.Commands, 0, 2)
	cmds = append(cmds, setPessoaCmd)
	cmds = append(cmds, setApelidoCmd)

	for _, res := range p.client.DoMulti(ctx, cmds...) {
		err := res.Error()

		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Cache) SetSearch(term string, result []model.Pessoa) error {
	item, err := sonic.MarshalString(result)
	if err != nil {
		return err
	}

	setSearchCmd := p.client.
		B().
		Set().
		Key("termo_busca:" + term).
		Value(item).
		Ex(1.5 * 60000 * time.Millisecond).
		Build()

	return p.client.Do(ctx, setSearchCmd).Error()
}

func (p *Cache) GetSearch(term string) ([]model.Pessoa, error) {
	getSearchCmd := p.client.
		B().
		Get().
		Key("termo_busca:" + term).
		Build()

	resultBytes, err := p.client.
		Do(
			ctx,
			getSearchCmd,
		).
		AsBytes()

	if err != nil {
		return nil, err
	}

	var result []model.Pessoa
	err = sonic.Unmarshal(resultBytes, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func TryFindItemInCache(term string, cache *Cache) ([]model.Pessoa, error) {

	cachedResult, err := cache.GetSearch(term)
	if err != nil && !rueidis.IsRedisNil(err) {
		fmt.Printf("Error getting search from cache: %v", err)
		return nil, err
	}

	if len(cachedResult) > 0 {
		return cachedResult, nil
	}

	return nil, errors.New("PESSOA NÃO ENCONTRADA")
}

func CreateCache() *Cache {
	address := fmt.Sprintf(
		"%s:%s",
		os.Getenv("CACHE_HOST"),
		os.Getenv("CACHE_PORT"),
	)

	opts := rueidis.ClientOption{
		InitAddress:      []string{address},
		AlwaysPipelining: true,
	}
	client, err := rueidis.NewClient(opts)
	if err != nil {
		panic(err)
	}

	fmt.Println("Cache started")

	return &Cache{client: client}
}
