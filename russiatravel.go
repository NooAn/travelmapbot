package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)
//---------------------------------CONST----------------------------------------
const MIN_COUNT_SYMBOL  = 2000
// ------------------------------- TYPES ----------------------------------------

// APIResponse is a response from the Russia.Travel API
type APIResponse struct {
	Items           []items `xml:"items"`
	ResponseCode    string  `xml:"responseCode"`
	ResponseMessage string  `xml:"responseMessage"`
}

// Item is returned by "get-objects-for-update" request. It's basically
// every place object that can be found and it contains almost everything.
type Item struct {
	XMLName         xml.Name        `xml:"item"`
	ItemID          uint32          `xml:"id,attr"`
	Image           string          `xml:"image,attr"`
	Geo             string          `xml:"geo,attr"`
	Types           []types         `xml:"types"`
	Name            []name          `xml:"name"`
	addressCountry  string          `xml:"addressCountry"`
	addressLocality string          `xml:"addressLocality"`
	addressArea     string          `xml:"addressArea"`
	addressRegion   []addressRegion `xml:"addressRegion"`
	streetAddress   string          `xml:"streetAddress"`
	Location        string          `xml:"geo"` // no actual need for it
	Review          []Review        `xml:"review"`
	Photos          []photos        `xml:"photos"`
	Url             string          `xml:"url"`
	Telephone       string          `xml:"telephone"`
	Email           string          `xml:"email"`
	Rating          string          `xml:"ratingValue"`
}

type photos struct {
	XMLName xml.Name `xml:"photos"`
	Photo   []photo  `xml:"photo"`
}
type photo struct {
	XMLName xml.Name `xml:"photo"`
	Link    string   `xml:"file"`
}

type types struct {
	XMLName xml.Name `xml:"types"`
	Type    []string `xml:"type"`
}

type name struct {
	XMLName xml.Name `xml:"name"`
	Text    string   `xml:"text"`
	Lang    string   `xml:"lang,attr"`
}
type Review struct {
	XMLName xml.Name `xml:"review"`
	//Text    string   `xml:"text"`
	//Lang    string   `xml:"lang,attr"`
	Data []ReviewData `xml:"text"`
}

// type review for request only
type RReview struct {
	XMLName xml.Name `xml:"review"`
	Text    string   `xml:"text"`
}

type ReviewData struct {
	Data string `xml:",chardata"`
	Lang string `xml:"lang,attr"`
}

type Location struct {
	Latitude  float64
	Longitude float64
}

type Request struct {
	XMLName       xml.Name        `xml:"request"`
	Action        string          `xml:"action,attr"`
	LastUpdate    string          `xml:"lastupdate,attr,omitempty"`
	Page          int             `xml:"page,attr,omitempty"`
	Items         []items         `xml:"items,omitempty"`
	addressRegion []addressRegion `xml:"addressRegion,omitempty"`
	Point         []Point         `xml:"point,omitempty"`
	objectType    []objectType    `xml:"objectType,omitempty"`
	Attributes    []Attributes    `xml:"attributes"`
}
type Attributes struct {
	XMLName xml.Name `xml:"attributes"`
	Review  *RReview `xml:"review"`
	// to be added if necsesary
	//Videos []videos `xml:"videos"`
	//objectType    []objectType    `xml:"objectType"`
	//Photos        []photos        `xml:"photos"`
	//addressRegion []addressRegion `xml:"addressRegion"`
}

type objectType struct {
	TypeID string `xml:"id"`
}

type addressRegion struct {
	RegionID uint32 `xml:"id"`
}
type items struct {
	XMLName    xml.Name `xml:"items"`
	Page       int      `xml:"page,attr"`
	TotalPages int      `xml:"totalPages,attr"`
	Item       []Item   `xml:"item"`
}

type Point struct {
	Geo    string `xml:",chardata"`
	Radius int    `xml:"radius,attr"`
}

// --------------------------- FUNCTIONS -------------------------------------------

func CreateRequestDependingOnRadius(radius int, geo string) []byte {
	v := &Request{Action: "get-objects-for-update"}
	newPoint := Point{Geo: geo, Radius: radius}
	v.Point = append(v.Point, newPoint)
	v.Attributes = append(v.Attributes, Attributes{Review: &RReview{Text: ""}})

	output, err := xml.MarshalIndent(v, "  ", "    ")
	if err != nil {
		fmt.Println("error: %v\n", err)
	}

	return output
}

func SendRequest(xmlbody string) []byte {
	form := url.Values{}
	form.Add("login", "view")
	form.Add("hash", "view")
	form.Add("xml", xmlbody)

	respi, _ := http.PostForm("http://api.russia.travel", form)
	defer respi.Body.Close()

	body, _ := ioutil.ReadAll(respi.Body)

	return body
}

func GetResponse(body []byte) APIResponse {
	response := APIResponse{}
	err := xml.Unmarshal(body, &response)
	if err != nil {
		fmt.Println("error: %v", err)
	}
	return response
}

func GetNames(items []Item) []string {
	var res []string
	for _, i := range items {
		res = append(res, i.Name[0].Text)
	}

	return res
}

func GetPhotoLinks(items []Item) []string {
	var res []string
	for _, i := range items {
		res = append(res, i.Image)
	}

	return res
}

func GetCoordinates(items []Item) []string {
	var res []string
	for _, i := range items {
		res = append(res, i.Geo)
	}

	return res
}

func GetReviews(items []Item) []string {
	var res []string
	for _, i := range items {
		res = append(res, i.Review[0].Data[0].Data)
	}

	return res
}

func Len(items []Item) int {
	lenght := 0
	for range items {
		lenght += 1
	}

	return lenght
}
