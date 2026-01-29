package main

import (
	"context"
	"encoding/json"
	"log"
	"sync/atomic"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	commonpb "taskforge/proto/common"
	workerpb "taskforge/proto/worker"
)

// ---------- Worker Pool ----------

type Job struct {
	req    *commonpb.ExecutionRequest
	stream workerpb.WorkerService_WorkerStreamClient
}

const WORKER_POOL_SIZE = 4

var runningTasks int32 // shared load counter

func workerLoop(id int, jobs <-chan Job) {
	log.Println("worker goroutine started:", id)

	for job := range jobs {
		atomic.AddInt32(&runningTasks, 1)

		req := job.req
		log.Println("executing:", req.ExecutionId, "on worker", id)

		output, err := executeFunction(req)

		result := &commonpb.ExecutionResult{
			ExecutionId: req.ExecutionId,
			Success:     err == nil,
			Output:      output,
		}
		if err != nil {
			result.Error = err.Error()
		}

		err = job.stream.Send(&workerpb.WorkerMessage{
			Message: &workerpb.WorkerMessage_Result{
				Result: result,
			},
		})
		if err != nil {
			log.Println("failed to send result:", err)
		}

		atomic.AddInt32(&runningTasks, -1)
	}
}

// ---------- Execution Logic (v1) ----------

func executeFunction(req *commonpb.ExecutionRequest) ([]byte, error) {
	var input map[string]interface{}
	_ = json.Unmarshal(req.Payload, &input)

	// real execution (v1)
	time.Sleep(1 * time.Second)

	result := map[string]interface{}{
		"worker":  "v1",
		"message": "function executed successfully",
		"input":   input,
	}

	return json.Marshal(result)
}

// ---------- Heartbeat ----------

func startHeartbeat(stream workerpb.WorkerService_WorkerStreamClient, workerID string) {
	ticker := time.NewTicker(2 * time.Second)

	for range ticker.C {
		load := atomic.LoadInt32(&runningTasks)

		err := stream.Send(&workerpb.WorkerMessage{
			Message: &workerpb.WorkerMessage_Heartbeat{
				Heartbeat: &workerpb.WorkerHeartbeat{
					WorkerId:     workerID,
					RunningTasks: load,
				},
			},
		})
		if err != nil {
			log.Println("heartbeat failed:", err)
			return
		}
	}
}

// ---------- Main ----------

func main() {
	conn, err := grpc.Dial(
		"localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := workerpb.NewWorkerServiceClient(conn)

	stream, err := client.WorkerStream(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	workerID := "worker-1"

	// send hello
	err = stream.Send(&workerpb.WorkerMessage{
		Message: &workerpb.WorkerMessage_Hello{
			Hello: &workerpb.WorkerHello{
				WorkerId: workerID,
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Println("worker connected to scheduler as", workerID)

	// start heartbeat
	go startHeartbeat(stream, workerID)

	// worker pool
	jobQueue := make(chan Job, 100)
	for i := 0; i < WORKER_POOL_SIZE; i++ {
		go workerLoop(i, jobQueue)
	}

	// receive executions
	for {
		msg, err := stream.Recv()
		if err != nil {
			log.Fatal("scheduler closed stream:", err)
		}

		switch m := msg.Message.(type) {
		case *workerpb.SchedulerMessage_Execute:
			jobQueue <- Job{
				req:    m.Execute,
				stream: stream,
			}
		}
	}
}
