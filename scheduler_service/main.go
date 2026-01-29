package main

import (
	"context"
	"log"
	"net"
	"sync"

	"github.com/google/uuid"
	"google.golang.org/grpc"

	commonpb "taskforge/proto/common"
	schedpb "taskforge/proto/scheduler"
	workerpb "taskforge/proto/worker"
)

// -------- Worker State --------

type WorkerState struct {
	stream       workerpb.WorkerService_WorkerStreamServer
	runningTasks int32
}

// -------- gRPC Server --------

type SchedulerServer struct {
	schedpb.UnimplementedSchedulerServiceServer
	workerpb.UnimplementedWorkerServiceServer

	store   *Store
	workers map[string]*WorkerState
	mu      sync.Mutex
}

// -------- RPCs --------

func (s *SchedulerServer) RegisterFunction(
	ctx context.Context,
	req *schedpb.RegisterFunctionRequest,
) (*schedpb.RegisterFunctionResponse, error) {

	id := uuid.New().String()

	err := s.store.SaveFunction(Function{
		ID:       id,
		Name:     req.Name,
		Language: req.Language,
		Code:     req.Code,
	})
	if err != nil {
		return nil, err
	}

	log.Println("Function registered:", id)

	return &schedpb.RegisterFunctionResponse{
		FunctionId: id,
	}, nil
}

func (s *SchedulerServer) TriggerExecution(
	ctx context.Context,
	req *schedpb.TriggerExecutionRequest,
) (*schedpb.TriggerExecutionResponse, error) {

	execID := uuid.New().String()

	err := s.store.SaveExecution(Execution{
		ID:         execID,
		FunctionID: req.FunctionId,
		Status:     "PENDING",
	})
	if err != nil {
		return nil, err
	}

	worker := s.pickLeastLoadedWorker()
	if worker == nil {
		return nil, grpc.Errorf(grpc.Code(grpc.ErrClientConnClosing), "no workers available")
	}

	// mark RUNNING
	exec, _ := s.store.GetExecution(execID)
	exec.Status = "RUNNING"
	_ = s.store.SaveExecution(exec)

	worker.runningTasks++

	err = worker.stream.Send(&workerpb.SchedulerMessage{
		Message: &workerpb.SchedulerMessage_Execute{
			Execute: &commonpb.ExecutionRequest{
				ExecutionId: execID,
				FunctionId:  req.FunctionId,
				Payload:     req.Payload,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	log.Println("Execution dispatched:", execID)

	return &schedpb.TriggerExecutionResponse{
		ExecutionId: execID,
		Status:      "ACCEPTED",
	}, nil
}

func (s *SchedulerServer) GetExecutionStatus(
	ctx context.Context,
	req *schedpb.GetExecutionStatusRequest,
) (*schedpb.GetExecutionStatusResponse, error) {

	exec, ok := s.store.GetExecution(req.ExecutionId)
	if !ok {
		return &schedpb.GetExecutionStatusResponse{
			Status: "NOT_FOUND",
		}, nil
	}

	return &schedpb.GetExecutionStatusResponse{
		Status: exec.Status,
		Output: exec.Output,
		Error:  exec.Error,
	}, nil
}

// -------- Worker Stream --------

func (s *SchedulerServer) WorkerStream(
	stream workerpb.WorkerService_WorkerStreamServer,
) error {

	var workerID string

	for {
		msg, err := stream.Recv()
		if err != nil {
			log.Println("Worker disconnected:", workerID)
			s.mu.Lock()
			delete(s.workers, workerID)
			s.mu.Unlock()
			return err
		}

		switch m := msg.Message.(type) {

		case *workerpb.WorkerMessage_Hello:
			workerID = m.Hello.WorkerId
			s.mu.Lock()
			s.workers[workerID] = &WorkerState{
				stream:       stream,
				runningTasks: 0,
			}
			s.mu.Unlock()
			log.Println("Worker connected:", workerID)

		case *workerpb.WorkerMessage_Heartbeat:
			s.mu.Lock()
			if w, ok := s.workers[m.Heartbeat.WorkerId]; ok {
				w.runningTasks = m.Heartbeat.RunningTasks
			}
			s.mu.Unlock()

		case *workerpb.WorkerMessage_Result:
			exec, ok := s.store.GetExecution(m.Result.ExecutionId)
			if !ok {
				continue
			}

			exec.Status = "COMPLETED"
			exec.Output = m.Result.Output
			exec.Error = m.Result.Error
			_ = s.store.SaveExecution(exec)

			s.mu.Lock()
			if w, ok := s.workers[workerID]; ok && w.runningTasks > 0 {
				w.runningTasks--
			}
			s.mu.Unlock()

			log.Println("Execution completed:", exec.ID)
		}
	}
}

// -------- Worker Selection --------

func (s *SchedulerServer) pickLeastLoadedWorker() *WorkerState {
	s.mu.Lock()
	defer s.mu.Unlock()

	var selected *WorkerState
	for _, w := range s.workers {
		if selected == nil || w.runningTasks < selected.runningTasks {
			selected = w
		}
	}
	return selected
}

// -------- Main --------

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}

	store := NewStore()

	scheduler := &SchedulerServer{
		store:   store,
		workers: make(map[string]*WorkerState),
	}

	server := grpc.NewServer()

	schedpb.RegisterSchedulerServiceServer(server, scheduler)
	workerpb.RegisterWorkerServiceServer(server, scheduler) 

	log.Println("Scheduler gRPC server running on :50051")
	log.Println("Scheduler ready for concurrent workers")

	if err := server.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
