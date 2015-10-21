/*
===========================================================================
ORBIT VM PROTECTOR GPL Source Code
Copyright (C) 2015 Vasileios Anagnostopoulos.
This file is part of the ORBIT VM PROTECTOR Source Code (?ORBIT VM PROTECTOR Source Code?).  
ORBIT VM PROTECTOR Source Code is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.
ORBIT VM PROTECTOR Source Code is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.
You should have received a copy of the GNU General Public License
along with ORBIT VM PROTECTOR Source Code.  If not, see <http://www.gnu.org/licenses/>.
In addition, the ORBIT VM PROTECTOR Source Code is also subject to certain additional terms. You should have received a copy of these additional terms immediately following the terms and conditions of the GNU General Public License which accompanied the Doom 3 Source Code.  If not, please request a copy in writing from id Software at the address below.
If you have questions concerning this license or the applicable additional terms, you may contact in writing Vasileios Anagnostopoulos, Campani 3 Street, Athens Greece, POBOX 11252.
===========================================================================
*/
// orbit-vm-protector.go
package main

import (
	"github.com/emicklei/go-restful"
	"net/http"
	"container/list"
	"sync"
	"fmt"
	"log"
	"flag"
)


type VMFault struct {
	Vmid int
	Ovip string
}

type VMFaultReport struct{
	Vmfaultreport []VMFault
}

type DCFault struct{
	Dcid int
}

type DCFaultReport struct{
	Dcfaultreport []DCFault
}


var mu sync.Mutex
var vmlista *list.List = list.New()
var dclista *list.List = list.New()


func reporter_reportvmfaults(request *restful.Request, response *restful.Response) { //stop a stream
	params := new(VMFault)
	err := request.ReadEntity(params)
	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}
	mu.Lock()
	vmlista.PushBack(params.Vmid)
	mu.Unlock()	
}

func reporter_reportdcfaults(request *restful.Request, response *restful.Response) { //stop a stream
	params := new(DCFault)
	err := request.ReadEntity(params)
	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}
	mu.Lock()
	dclista.PushBack(params.Dcid)
	mu.Unlock()
}

func reporter_acquiredcfaults(request *restful.Request, response *restful.Response) { //stop a stream
	var get_dclista *list.List
	
	mu.Lock()
	get_dclista,dclista=dclista,list.New()
	mu.Unlock()
	
	nresp := new(DCFaultReport)
	nresp.Dcfaultreport=make([]DCFault,get_dclista.Len())
	index :=0
	for e := get_dclista.Front(); e != nil; e = e.Next() {
		s := e.Value.(*DCFault)
		nresp.Dcfaultreport[index] = *s
		index++
	}
	response.WriteEntity(nresp)
}

func reporter_acquirevmfaults(request *restful.Request, response *restful.Response) { //stop a stream
	var get_vmlista *list.List
	
	mu.Lock()
	get_vmlista,vmlista=vmlista,list.New()
	mu.Unlock()
	
	nresp := new(VMFaultReport)
	nresp.Vmfaultreport=make([]VMFault,get_vmlista.Len())
	index :=0
	for e := get_vmlista.Front(); e != nil; e = e.Next() {
		s:=e.Value.(*VMFault)
		nresp.Vmfaultreport[index] = *s
		index++
	}
	response.WriteEntity(nresp)
}

func reporter_describe(request *restful.Request, response *restful.Response) { //stop a stream
	fmt.Println("Inside orbit_describe")	
	response.WriteEntity("orbit-protection-aggregator")
}

func init() {
	log.Println("Inside init")
	aggregatorip = flag.String("aggregatorip", "", "the aggregator ip address")
	if aggregatorip == nil {
		panic("shitty aggregatorip")
	}
	aggregatorport = flag.String("aggregatorport", "", "the aggregator port")
	if aggregatorport == nil {
		panic("shitty aggregatorport")
	}
}


var aggregatorip *string
var aggregatorport *string

func main() {
	
	flag.Parse()
	fmt.Println("Registering")
	wsContainer := restful.NewContainer()
	ws := new(restful.WebService)
	ws.Path("/watchdog").Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)
	ws.Route(ws.GET("").To(reporter_describe))
	ws.Route(ws.POST("/vmfaults").To(reporter_reportvmfaults)).Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)
	ws.Route(ws.POST("/dcfaults").To(reporter_reportdcfaults)).Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)
	ws.Route(ws.GET("/vmfaults").To(reporter_acquirevmfaults)).Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)
	ws.Route(ws.GET("/dcfaults").To(reporter_acquiredcfaults)).Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)	
	//ws.Route(ws.POST("/vmmeasurements").To(reporter_reportvmmeasurements)).Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)	
	//ws.Route(ws.POST("/dcmeasurements").To(reporter_reportdcmeasurements)).Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)
	//ws.Route(ws.GET("/vmmeasurements").To(reporter_acquirevmmeasurements)).Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)	
	//ws.Route(ws.GET("/dcmeasurements").To(reporter_acquiredcmeasurements)).Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)	
	//wsContainer.Add(ws)

	// Add container filter to enable CORS
	/*
		cors := restful.CrossOriginResourceSharing{
			ExposeHeaders:  []string{"X-My-Header"},
			AllowedHeaders: []string{"Content-Type"},
			CookiesAllowed: false,
			Container:      wsContainer}
		wsContainer.Filter(cors.Filter)

		// Add container filter to respond to OPTIONS
		wsContainer.Filter(wsContainer.OPTIONSFilter)
	*/

	fmt.Printf("start listening on localhost:%s\n",*aggregatorport)
	server := &http.Server{Addr: *aggregatorip+":"+*aggregatorport, Handler: wsContainer}
	log.Fatal(server.ListenAndServe())
}