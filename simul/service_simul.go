package main

import (
	"github.com/BurntSushi/toml"
	"github.com/JoaoAndreSa/MedCo/lib"
	"github.com/JoaoAndreSa/MedCo/services"
	"github.com/JoaoAndreSa/MedCo/services/data"
	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/log"
	"gopkg.in/dedis/onet.v1/simul/monitor"
	"strconv"
	"sync"
)

//Defines the simulation for the service-medCo to be run with cothority/simul.
func init() {
	onet.SimulationRegister("ServiceMedCo", NewSimulationMedCo)
}

// SimulationMedCo the state of a simulation.
type SimulationMedCo struct {
	onet.SimulationBFTree

	NbrDPs             int     //number of clients (or in other words data holders)
	NbrResponses       int64   //number of survey entries (ClientClearResponse) per host
	NbrGroupsClear     int64   //number of non-sensitive (clear) grouping attributes
	NbrGroupsEnc       int64   //number of sensitive (encrypted) grouping attributes
	NbrGroupAttributes []int64 //number of different groups inside each grouping attribute
	NbrAggrAttributes  int64   //number of aggregating attributes
	RandomGroups       bool    //generate data randomly or num entries == num groups (deterministically)
	DataRepetitions    int     //repeat the number of entries x times (e.g. 1 no repetition; 1000 repetitions)
	Proofs             bool    //with proofs of correctness everywhere
}

// NewSimulationMedCo constructs a full MedCo service simulation.
func NewSimulationMedCo(config string) (onet.Simulation, error) {
	es := &SimulationMedCo{}
	_, err := toml.Decode(config, es)
	if err != nil {
		return nil, err
	}
	return es, nil
}

// Setup creates the tree used for that simulation
func (sim *SimulationMedCo) Setup(dir string, hosts []string) (*onet.SimulationConfig, error) {
	sc := &onet.SimulationConfig{}
	sim.CreateRoster(sc, hosts, 2000)
	err := sim.CreateTree(sc)

	if err != nil {
		return nil, err
	}

	log.Lvl1("Setup done")

	return sc, nil
}

// Run starts the simulation.
func (sim *SimulationMedCo) Run(config *onet.SimulationConfig) error {
	var start *monitor.TimeMeasure
	log.SetDebugVisible(1)
	// Setup Simulation
	nbrHosts := config.Tree.Size()
	log.Lvl1("Size:", nbrHosts, ", Rounds:", sim.Rounds)

	//TODO: PLEASE REMOVE THIS AFTERWARDS ... it's very ugly
	/*for i:=0; i< int(sim.NbrGroupsEnc)-2; i++ {
		sim.NbrGroupAttributes = append(sim.NbrGroupAttributes,int64(1))
	}
	log.LLvl1(sim.NbrGroupAttributes)*/

	// Does not make sense to have more servers than clients!!
	if nbrHosts > sim.NbrDPs {
		log.Fatal("hosts:", nbrHosts, "must be the same or lower as num_clients:", sim.NbrDPs)
		return nil
	}
	el := (*config.Tree).Roster

	for round := 0; round < sim.Rounds; round++ {
		log.Lvl1("Starting round", round, el)
		client := services.NewMedcoClient(el.List[0], strconv.Itoa(0))

		nbrDPs := make(map[string]int64)
		//how many data providers for each server
		for _, server := range el.List {
			nbrDPs[server.String()] = 1 // 1 DPs for each server
		}

		// Generate Survey Data
		surveyDesc := lib.SurveyDescription{GroupingAttributesClearCount: int32(sim.NbrGroupsClear), GroupingAttributesEncCount: int32(sim.NbrGroupsEnc), AggregatingAttributesCount: uint32(sim.NbrAggrAttributes)}
		surveyID, _, err := client.SendSurveyCreationQuery(el, lib.SurveyID(""), lib.SurveyID(""), surveyDesc, sim.Proofs, false, nil, nil, nil, nbrDPs, 0)

		if err != nil {
			log.Fatal("Service did not start.")
		}

		// RandomGroups (true/false) is to respectively generate random or non random entries
		testData := data.GenerateData(int64(sim.NbrDPs), sim.NbrResponses, sim.NbrGroupsClear, sim.NbrGroupsEnc, sim.NbrAggrAttributes, sim.NbrGroupAttributes, sim.RandomGroups)

		/// START SERVICE PROTOCOL
		if lib.TIME {
			start = monitor.NewTimeMeasure("SendingData")
		}

		log.Lvl1("Sending response data... ")
		dataHolder := make([]*services.API, sim.NbrDPs)
		wg := lib.StartParallelize(len(dataHolder))

		var mutex sync.Mutex
		for i, client := range dataHolder {
			start1 := lib.StartTimer(strconv.Itoa(i) + "_IndividualSendingData")
			if lib.PARALLELIZE {
				go func(i int, client *services.API) {
					mutex.Lock()
					data := testData[strconv.Itoa(i)]
					server := el.List[i%nbrHosts]
					mutex.Unlock()

					client = services.NewMedcoClient(server, strconv.Itoa(i+1))
					client.SendSurveyResponseQuery(*surveyID, data, el.Aggregate, sim.DataRepetitions)
					defer wg.Done()
				}(i, client)
			} else {
				start2 := lib.StartTimer(strconv.Itoa(i) + "_IndividualNewMedcoClient")

				client = services.NewMedcoClient(el.List[i%nbrHosts], strconv.Itoa(i+1))

				lib.EndTimer(start2)
				start3 := lib.StartTimer(strconv.Itoa(i) + "_IndividualSendSurveyResults")

				client.SendSurveyResponseQuery(*surveyID, testData[strconv.Itoa(i)], el.Aggregate, sim.DataRepetitions)

				lib.EndTimer(start3)

			}
			lib.EndTimer(start1)

		}
		lib.EndParallelize(wg)
		lib.EndTimer(start)

		start := lib.StartTimer("Simulation")

		grpClear, grp, aggr, err := client.SendGetSurveyResultsQuery(*surveyID)
		if err != nil {
			log.Fatal("Service could not output the results. ", err)
		}

		lib.EndTimer(start)
		// END SERVICE PROTOCOL

		// Print Output
		allData := make([]lib.ClientClearResponse, 0)
		log.Lvl1("Service output:")
		for i := range *grp {
			log.Lvl1(i, ")", (*grpClear)[i], ",", (*grp)[i], "->", (*aggr)[i])
			allData = append(allData, lib.ClientClearResponse{GroupingAttributesClear: (*grpClear)[i], GroupingAttributesEnc: (*grp)[i], AggregatingAttributes: (*aggr)[i]})
		}

		// Test Service Simulation
		if data.CompareClearResponses(data.ComputeExpectedResult(testData, sim.DataRepetitions), allData) {
			log.LLvl1("Result is right! :)")
		} else {
			log.LLvl1("Result is wrong! :(")
		}
	}

	return nil
}
