package database

import (
	"api/model"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	MaxWorker = 1
	MaxQueue  = 1
)

type JobQueue chan Job

// Job represents the job to be run
type Job struct {
	Payload *model.Pessoa
}

// A buffered channel that we can send work requests on.

func CreateJobQueue() JobQueue {
	return make(JobQueue, MaxQueue)
}

// Worker represents the worker that executes the job
type Worker struct {
	WorkerPool chan chan Job
	JobChannel chan Job
	quit       chan bool
	db         *pgxpool.Pool
}

func (w Worker) Start() {
	dataCh := make(chan Job)
	insertCh := make(chan []Job)

	go w.bootstrap(dataCh)

	go w.processData(dataCh, insertCh)
}

func (w Worker) Stop() {
	go func() {
		w.quit <- true
	}()
}

func (w Worker) bootstrap(dataCh chan Job) {
	for {
		w.WorkerPool <- w.JobChannel

		select {
		case job := <-w.JobChannel:
			dataCh <- job

		case <-w.quit:
			return
		}
	}
}

func (w Worker) processData(dataCh chan Job, insertCh chan []Job) {
	batch := make([]Job, 0)
	columns := []string{"id", "apelido", "nome", "nascimento", "stack"}
	identifier := pgx.Identifier{"pessoa"}

	for {
		select {
		case data := <-dataCh:
			batch = append(batch, data)
			_, err := w.db.CopyFrom(
				context.Background(),
				identifier,
				columns,
				pgx.CopyFromSlice(len(batch), w.CopyFromSlice(batch)),
			)

			if err != nil {
				fmt.Printf("Error on insert batch", err)
			}

			fmt.Println("Sucessfully insert")

			batch = make([]Job, 0)

		}
	}
}

func (Worker) CopyFromSlice(batch []Job) func(i int) ([]interface{}, error) {
	return func(i int) ([]interface{}, error) {
		return []interface{}{
			batch[i].Payload.ID,
			batch[i].Payload.Apelido,
			batch[i].Payload.Nome,
			batch[i].Payload.Nascimento,
			batch[i].Payload.Stack,
		}, nil
	}
}

func CreateWorker(workerPool chan chan Job, db *pgxpool.Pool) Worker {
	return Worker{
		WorkerPool: workerPool,
		JobChannel: make(chan Job),
		quit:       make(chan bool),
		db:         db,
	}
}

type Dispatcher struct {
	maxWorkers int
	// A pool of workers channels that are registered with the dispatcher
	WorkerPool chan chan Job
	jobQueue   chan Job
	db         *pgxpool.Pool
}

func CreateDispatcher(db *pgxpool.Pool, jobQueue JobQueue) *Dispatcher {
	maxWorkers := MaxWorker

	pool := make(chan chan Job, maxWorkers)

	return &Dispatcher{
		WorkerPool: pool,
		maxWorkers: maxWorkers,
		jobQueue:   jobQueue,
		db:         db,
	}
}

func (d *Dispatcher) Run() {
	// starting n number of workers
	for i := 0; i < d.maxWorkers; i++ {
		worker := CreateWorker(d.WorkerPool, d.db)
		worker.Start()
	}

	go d.dispatch()
}

func (d *Dispatcher) dispatch() {
	for {
		select {
		case job := <-d.jobQueue:
			// a job request has been received
			go func(job Job) {
				// try to obtain a worker job channel that is available.
				// this will block until a worker is idle
				jobChannel := <-d.WorkerPool

				// dispatch the job to the worker job channel
				jobChannel <- job
			}(job)
		}
	}
}
