package main

import (
	"context"
	"flag"
	"io"
	"log"
	"strconv"
	"time"

	"google.golang.org/grpc"

	pb "github.com/CamusEnergy/kinney/orchestrator"
)

var orchestrator = flag.String("orchestrator", "", "Address of the orchestrator server, in the form of [ip]:port.  [required]")

func main() {
	flag.Parse()

	if *orchestrator == "" {
		log.Fatal("--orchestrator is required")
	}

	conn, err := grpc.DialContext(context.Background(), *orchestrator, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("Error connecting to the orchestrator: %s", err)
	}
	defer conn.Close()

	client := pb.NewOrchestratorClient(conn)

	stream, err := client.Charger(context.Background())
	if err != nil {
		log.Fatalf("Error initializing stream: %s", err)
	}

	go scrape(stream)
	handleCommands(stream)
}

var (
	nextSessionID      = 0
	activeSessions     = map[int]time.Time{}
	maxSessions        = 4
	maxSessionDuration = 10 * time.Second
	scrapeInterval     = time.Second
)

func scrape(stream pb.Orchestrator_ChargerClient) {
	for {
		now := time.Now()

		// Remove sessions older than `maxSessionDuration`.
		oldestSessionStart := now.Add(-1 * maxSessionDuration)
		for sessionID, sessionStart := range activeSessions {
			if sessionStart.Before(oldestSessionStart) {
				// This session has expired.
				delete(activeSessions, sessionID)
			}
		}

		// For each available slot, add a new session.
		for i := 0; i < maxSessions-len(activeSessions); i++ {
			activeSessions[nextSessionID] = now
			nextSessionID++
		}

		// Send a `ChargerSession` message for each active session.
		for sessionID, _ := range activeSessions {
			msg := &pb.ChargerSession{
				Vehicle: strconv.Itoa(sessionID),
				Watts:   20,
			}
			if err := stream.Send(msg); err != nil {
				log.Fatalf("Error sending message: %s", err)
			}
		}

		time.Sleep(scrapeInterval)
	}
}

func handleCommands(stream pb.Orchestrator_ChargerClient) {
	for {
		chargerCommand, err := stream.Recv()
		if err == io.EOF {
			log.Print("handleCommands: stream closed")
			return
		} else if err != nil {
			log.Fatalf("Error receiving command: %s", err)
		}
		log.Printf("Received command: %s", chargerCommand)
	}
}
