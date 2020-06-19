package main

import (
	"flag"
	"io"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"

	pb "github.com/CamusEnergy/kinney/orchestrator"
)

var address = flag.String("address", ":0", "The address to listen on, in the form of [ip]:port.")

func main() {
	flag.Parse()

	listener, err := net.Listen("tcp", *address)
	if err != nil {
		log.Fatalf("Error starting listener: %s", err)
	}
	log.Printf("Listening on %s", listener.Addr())

	grpcServer := grpc.NewServer()
	pb.RegisterOrchestratorServer(grpcServer, &server{})
	grpcServer.Serve(listener)
}

type server struct {
	pb.UnimplementedOrchestratorServer
}

func (s *server) Charger(stream pb.Orchestrator_ChargerServer) error {
	go sendCommands(stream)
	handleSessions(stream)
	return nil
}

func sendCommands(stream pb.Orchestrator_ChargerServer) {
	for {
		msg := &pb.ChargerCommand{
			Limit: 1234,
		}
		if err := stream.Send(msg); err != nil {
			log.Fatalf("Error sending command: %s", err)
		}

		time.Sleep(12 * time.Second)
	}
}

func handleSessions(stream pb.Orchestrator_ChargerServer) {
	for {
		chargerSession, err := stream.Recv()
		if err == io.EOF {
			log.Print("handleSessions: stream closed")
			return
		} else if err != nil {
			log.Fatalf("Error receiving message: %s", err)
		}

		log.Printf("Received: %s", chargerSession)
	}
}
