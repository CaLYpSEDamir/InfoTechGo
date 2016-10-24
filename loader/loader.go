package loader

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/CaLYpSEDamir/InfoTechGo/invitro"
)

var p = fmt.Println

// Configuration structure for config
type Configuration struct {
	Dbs map[string]map[string]string
}

// GetConfig comment
func GetConfig() *Configuration {

	// if _, err := os.Stat("../loader/local.json"); os.IsNotExist(err) {
	// 	file, err := os.Open("../loader/local.json")
	// } else if _, err := os.Stat("../loader/conf.json"); os.IsNotExist(err) {
	file, _ := os.Open("../loader/conf.json")
	// } else {
	// 	file := nil
	// 	log.Fatal("No conf file!")
	// }

	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	decodeErr := decoder.Decode(&configuration)

	if decodeErr != nil {
		log.Fatal("Invalid conf file!")
	}
	return &configuration
}

// Loader - main object processing extraction and saving data
type Loader struct {
	config *Configuration
	Miner  invitro.Miner
	// Preparer invitro.Preparer
	Saver invitro.Saver
}

// Init c
func (l *Loader) Init() {
	l.config = GetConfig()
	l.Saver.Init(l.config.Dbs["psql"])
}

// Load starts all process
func (l *Loader) Load() {
	t1 := time.Now()

	types := l.Miner.GetTypes()
	l.Saver.SaveTypes(types)

	//hack
	// postgres too many connection (setted max_conn=2000)
	// http too many connection
	rangLsit := [][]int{
		[]int{0, 7},
		[]int{7, 14},
		// []int{0, 14},
		[]int{14, 21},
		[]int{21, len(types)},
		// []int{14, len(types)},
	}

	for i, rang := range rangLsit {

		subTypes := types[rang[0]:rang[1]]

		endCh := make(chan bool)
		defer close(endCh)

		for _, info := range subTypes {
			subTypeCh := make(chan []string)
			researchCh1 := make(chan []string)

			// hacks
			researchCh2 := make(chan []string, 1000)
			researchCh3 := make(chan []string, 1000)

			go func(info []string) {
				l.Miner.ExtractTypeData(info, subTypeCh, researchCh1)
			}(info)

			go func() {
				l.Saver.SaveSubTypes(subTypeCh)
				l.Saver.SaveResearch(researchCh1, researchCh2)
			}()

			go l.Miner.GetResearchFullInfo(researchCh2, researchCh3)

			go l.Saver.SaveResearchFullInfo(researchCh3, endCh)
		}

		for range subTypes {
			p("endCh #", i, <-endCh)
		}

		p(time.Since(t1))
		p("END", i)
	}
}
