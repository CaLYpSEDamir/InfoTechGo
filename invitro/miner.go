package invitro

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html/charset"
)

var p = fmt.Println

const host = "https://www.invitro.ru"

// Miner for obtaining data
type Miner struct{}

// GetDecodedDoc c
func (m *Miner) GetDecodedDoc(url string) *goquery.Document {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal("HTTP error:", err)
	}

	defer resp.Body.Close()

	utf8, err := charset.NewReader(resp.Body, resp.Header.Get("Content-Type"))
	if err != nil {
		log.Fatal("Encoding error:", err)
	}

	body, err := ioutil.ReadAll(utf8)
	if err != nil {
		log.Fatal("IO error:", err)
	}

	reader := bytes.NewReader(body)
	doc, err := goquery.NewDocumentFromReader(reader)

	return doc
}

// extractLinks method
func (m *Miner) extractLinks(
	aList *goquery.Selection) [][]string {

	var info [][]string

	aList.Each(
		func(_ int, s *goquery.Selection) {
			href, _ := s.Attr("href")
			title, _ := s.Attr("title")
			info = append(info, []string{href, title})
		})
	return info
}

// GetTypes method
func (m *Miner) GetTypes() [][]string {

	doc := m.GetDecodedDoc(host + "/analizes/for-doctors/")
	aList := doc.Find(".group-name-brd a")
	typesInfo := m.extractLinks(aList)

	return typesInfo
}

// GetSubTypes method
func (m *Miner) getSubTypes(doc *goquery.Document,
	typeName string, subTypeCh chan []string) {

	aList := doc.Find(".list-els a")
	subTypesInfo := m.extractLinks(aList)

	for _, subT := range subTypesInfo {
		subT = append(subT, typeName)
		subTypeCh <- subT
	}
	close(subTypeCh)
}

// GetResearches method
func (m *Miner) getResearches(
	doc *goquery.Document, mainType string, researchCh chan []string) {

	trList := doc.Find("#catalog-section-analiz tr:has(td)")

	var typ string

	for _, tr := range trList.Nodes {
		trDoc := goquery.NewDocumentFromNode(tr)
		if len(trDoc.Find(".sec_3, .sec_4, .sec_5").Nodes) > 0 {
			continue
		} else if len(trDoc.Find(".sec_2").Nodes) > 0 {
			span := trDoc.Find("span")
			typ = span.Text()
			continue
		} else {
			a := trDoc.Find("td.name>a")

			text := a.Text()
			href, exist := a.Attr("href")

			if !exist {
				continue
			}

			if typ == "" {
				researchCh <- []string{mainType, "nil", text, href}
			} else {
				researchCh <- []string{typ, mainType, text, href}
			}

		}
	}
	close(researchCh)
}

// ExtractTypeData method
func (m *Miner) ExtractTypeData(
	info []string, subTypeCh chan []string, researchCh chan []string) {

	url, typeName := info[0], info[1]
	doc := m.GetDecodedDoc(host + url)

	m.getSubTypes(doc, typeName, subTypeCh)

	m.getResearches(doc, typeName, researchCh)

}

// GetResearchFullInfo c
func (m *Miner) GetResearchFullInfo(researchChIn chan []string,
	researchChOut chan []string) {

	var wg sync.WaitGroup

	for res := range researchChIn {
		wg.Add(1)
		go func(r []string) {
			defer wg.Done()

			rID, url := r[0], r[1]
			// p(rID, url)
			// url := r[1]

			doc := m.GetDecodedDoc(host + url)

			script := doc.Find("script[language=\"JavaScript\"]:contains(\"arText\")")
			text := script.Text()

			splits := strings.Split(text, "var arTexts=")

			if len(splits) != 2 {
				// p("Not 2", r)
				return
			}
			split := splits[1]

			spl := strings.Split(split, ";")
			spl = spl[:len(spl)-1]
			newT := strings.Join(spl, ";")

			// bad custom ecranize
			newT = strings.Replace(newT, "'", `"`, -1)
			newT = strings.Replace(newT, "\n", "\\n", -1)
			newT = strings.Replace(newT, "\t", "\\t", -1)

			sh := SearchHelper{}
			json.Unmarshal([]byte(newT), &sh)

			info := []string{
				sh.Link_88, sh.Link_84, sh.Link_81, sh.Link_82,
			}
			newInfo := []string{rID}

			for _, link := range info {
				reader := bytes.NewReader([]byte(link))
				elem, _ := goquery.NewDocumentFromReader(reader)
				text = strings.Replace(elem.Children().Text(), "\n", " ", -1)
				newInfo = append(newInfo, text)
			}

			researchChOut <- newInfo

		}(res)
	}

	wg.Wait()

	close(researchChOut)
}
