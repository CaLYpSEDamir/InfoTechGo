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

// GetConfig установка конфигураций
func GetConfig() *Configuration {

	file, err := os.Open("../loader/local.json")

	if err != nil {
		p(err)
		file, _ = os.Open("../loader/conf.json")
	}

	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	decodeErr := decoder.Decode(&configuration)

	if decodeErr != nil {
		log.Fatal("Invalid conf file!")
	}
	return &configuration
}

// Loader - processing extraction and saving data
type Loader struct {
	config *Configuration
	Miner  invitro.Miner
	Saver  invitro.Saver
}

// Init инициализация Loader
func (l *Loader) Init() {
	l.config = GetConfig()
	l.Saver.Init(l.config.Dbs["psql"])
}

// Load процесс достачи и загрузки
func (l *Loader) Load() {
	t1 := time.Now()

	types := l.Miner.GetTypes()
	l.Saver.SaveTypes(types)

	//hack
	// postgres too many connection (needs increase max_connections)
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

			// hacks buffered channels
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
