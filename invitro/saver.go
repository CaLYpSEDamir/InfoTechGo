package invitro

import (
	"fmt"
	"strings"
	"sync"
	"time"

	// for multi insert
	// "github.com/lib/pq"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// Saver for saving data
type Saver struct {
	connStr string
}

func (s *Saver) checkConn() {
	db, err := gorm.Open("postgres", s.connStr)
	defer db.Close()

	if err != nil {
		panic("Connection is refused!")
	}
}

func (s *Saver) getConn() *gorm.DB {
	db, err := gorm.Open("postgres", s.connStr)

	if err != nil {
		panic("Connection is refused!")
	}
	return db
}

// Init c
func (s *Saver) Init(dbInfo map[string]string) {

	var connParts []string
	for k, v := range dbInfo {
		connParts = append(
			connParts, strings.Join([]string{k, v}, "="))
	}
	s.connStr = strings.Join(connParts, " ")
	s.checkConn()

	conn := s.getConn()
	defer conn.Close()

	// Creating ResearchType table
	if conn.HasTable("research_types") {
		conn.DropTable("research_types")
	}
	conn.CreateTable(&ResearchType{})

	// Creating Research table
	if conn.HasTable("researches") {
		conn.DropTable("researches")
	}
	conn.CreateTable(&Research{})

}

// SaveTypes c
func (s *Saver) SaveTypes(typesInfo [][]string) {

	conn := s.getConn()
	defer conn.Close()

	var wg sync.WaitGroup

	for _, info := range typesInfo {
		wg.Add(1)
		go func(info []string) {
			defer wg.Done()

			title := info[1]
			re := ResearchType{Name: title}
			conn.Create(&re)
		}(info)
	}

	wg.Wait()
}

// SaveSubTypes c
func (s *Saver) SaveSubTypes(subTypeIn chan []string) {

	conn := s.getConn()
	// defer conn.Close()

	var wg sync.WaitGroup

	for subT := range subTypeIn {
		wg.Add(1)
		go func(subT []string) {
			defer wg.Done()

			parent := ResearchType{}
			conn.Where("name = ?", subT[2]).First(&parent)
			conn.Create(&ResearchType{Name: subT[1], ParentID: parent.ID})
		}(subT)
	}

	wg.Wait()
	conn.Close()
}

// SaveResearch1 c
func (s *Saver) SaveResearch(
	researchChIn chan []string, researchChOut chan []string) {

	conn := s.getConn()
	// defer conn.Close()

	var wg sync.WaitGroup

	for subT := range researchChIn {
		wg.Add(1)
		go func(subT []string) {
			defer wg.Done()

			subTypeName, researchName, resHref := subT[0], subT[2], subT[3]

			typ := ResearchType{}
			conn.Where("name = ?", subTypeName).First(&typ)

			if typ.ID == 0 {
				parTypeName := subT[1]
				parent := ResearchType{}
				conn.Where("name = ?", parTypeName).First(&parent)
				typ = ResearchType{Name: subTypeName, ParentID: parent.ID}
				conn.Create(&typ)
			}

			res := Research{Name: researchName, TypeID: typ.ID}
			conn.Create(&res)
			researchChOut <- []string{fmt.Sprint(res.ID), resHref}

		}(subT)
	}

	wg.Wait()
	conn.Close()
	close(researchChOut)

	// wait for conn closeSecond
	time.Sleep(100 * time.Millisecond)
}

// SaveResearch1 c
func (s *Saver) SaveResearchFullInfo(researchCh chan []string, resCh chan bool) {

	conn := s.getConn()
	// defer conn.Close()

	var wg sync.WaitGroup

	for res := range researchCh {
		wg.Add(1)
		go func(r []string) {
			defer wg.Done()
			rID := r[0]
			research := Research{}
			conn.Where("id = ?", rID).First(&research)

			research.Description = r[1]
			research.Training = r[2]
			research.Indication = r[3]
			research.Result = r[4]

			conn.Save(&research)

		}(res)
	}
	wg.Wait()

	conn.Close()

	// wait for conn close
	time.Sleep(100 * time.Millisecond)

	resCh <- true

}
