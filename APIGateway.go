package main

import (
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"sync"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/julienschmidt/httprouter"
)

var instance *consulapi.Client
var once sync.Once

func GetClient() *consulapi.Client {
	if instance == nil {
		once.Do(func() {
			var config = consulapi.DefaultConfig()
			var client, err = consulapi.NewClient(config)

			if err != nil {
				log.Println("consul client error : ", err)
			}

			instance = client
		})
	}
	return instance
}

func LocalGateway(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	params := r.URL.Query()
	serviceName := params.Get(":service")
	serviceVersion := params.Get(":version")

	services, _, err := GetClient().Health().Service(serviceName, serviceVersion, true, nil)

	if err != nil {
		log.Println("consul query service error : ", err)
		fmt.Fprintf(w, "%s:%d/%s", "172.19.20.60", 888, "test")
		return
	}

	service := GetServiceRoundRobin(services)

	fmt.Fprintf(w, "%s:%d/%s", service.Node.Address, service.Service.Port, service.Service.Address)
}

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		t, err := template.ParseFiles("./views/index.html")
		if err != nil {
			log.Fatal(err)
		}

		data := struct {
			Title string
		}{
			Title: "Hello GAPIGateway!",
		}

		err = t.Execute(w, data)
		if err != nil {
			log.Println(err)
		}
}

func GetServiceRoundRobin(services []*consulapi.ServiceEntry) *consulapi.ServiceEntry {
	index := rand.Intn(len(services) - 1)
	return services[index]
}

func main() {
	router := httprouter.New()
	router.GET("/", Index)
	router.GET("/LocalGateway/:service/:version", LocalGateway)
	log.Println("ListenAndServe:8888")
	log.Fatal(http.ListenAndServe(":8888", router))
}
