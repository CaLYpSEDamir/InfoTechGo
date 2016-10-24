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

	subTypes := types[18:]
	// [:7] OK
	// [7:14] OK
	//[14:18] OK
	//[18:] OK

	// p(subTypes)

	endCh := make(chan bool)
	defer close(endCh)

	for _, info := range subTypes {
		subTypeCh := make(chan []string)
		researchCh1 := make(chan []string)

		// hacks
		researchCh2 := make(chan []string, 1000)
		researchCh3 := make(chan []string, 2000)

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
		p("endCh #1", <-endCh)
	}

	p(time.Since(t1))
	p("END")
}
