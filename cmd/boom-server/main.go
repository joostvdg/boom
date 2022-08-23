package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/joostvdg/boom/api"
	"github.com/joostvdg/boom/internal/server"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"

	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

// TODO make env variables
const (
	service     = "boom-server"
	environment = "local"
)

// tracerProvider returns an OpenTelemetry TracerProvider configured to use
// the Jaeger exporter that will send spans to the provided url. The returned
// TracerProvider will also use a Resource configured with all the information
// about the application.
func tracerProvider(url string, name string) (*tracesdk.TracerProvider, error) {
	// Create the Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		return nil, err
	}
	tp := tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in a Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(service),
			attribute.String("environment", environment),
			attribute.String("name", name),
		)),
	)
	return tp, nil
}

func main() {
	helloPortOverride := flag.String("helloPort", api.HelloPort, fmt.Sprintf("PortSelf number for listening for Hello messages, default %s", api.HelloPort))
	helloName := flag.String("helloName", "MySelf", "Name of this Boom server")
	flag.Parse()

	tp, err := tracerProvider("http://localhost:14268/api/traces", *helloName)
	if err != nil {
		log.Fatal(err)
	}

	// Register our TracerProvider as the global so any imported
	// instrumentation in the future will default to using it.
	otel.SetTracerProvider(tp)

	myAddress := determineAddress()
	helloMessage := api.ConstructHelloMessage(*helloName, myAddress.String(), *helloPortOverride)
	goodbyeMessage := api.ConstructGoodbyeMessage(*helloName, myAddress.String(), *helloPortOverride)
	myself := createMyself(*helloName, myAddress)
	myIdentity := myself.Identifier()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	membershipServiceContext := &server.MembershipServiceContext{
		Context:        ctx,
		TracerProvider: tp,
		SelfAddress:    myAddress,
		Self:           myself,
		Identity:       myIdentity,
		HelloMessage:   helloMessage,
		GoodbyeMessage: goodbyeMessage,
		ServerPort:     *helloPortOverride,
	}
	membershipServices := []server.MembershipService{
		server.ListenForMulticast,
		server.MulticastExistence,
		server.StartMembershipServer,
		server.HandleMember,
		server.CleanupMembers,
	}

	var wg sync.WaitGroup
	for _, membershipService := range membershipServices {
		wg.Add(1)
		go func(service server.MembershipService) {
			service(membershipServiceContext)
			defer wg.Done()
		}(membershipService)
	}
	wg.Wait()

	fmt.Printf("Shutting down!\n")
	server.CloseChannels()
	server.NotifyMembersOfLeaving(goodbyeMessage)
	tp.ForceFlush(ctx)
}

func createMyself(name string, address net.Addr) api.Member {
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	ip, _ := api.NewIP4Address(address.String())

	member := api.Member{
		MemberName: name,
		Hostname:   hostname,
		IPSelf:     &ip,
	}
	return member
}

func determineAddress() net.Addr {
	connection, err := net.ListenUDP("udp4", nil)
	if err != nil {
		fmt.Println(err)
	}

	defer connection.Close()
	address := connection.LocalAddr()
	return address
}

// newResource returns a resource describing this application.
func newResource() *resource.Resource {
	r, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("boom-server"),
			semconv.ServiceVersionKey.String("v0.1.0"),
			attribute.String("environment", "demo"),
		),
	)
	return r
}
