package main

import (
	"bytes"
	"fmt"
	"github.com/telegram-bot-api"
	"html"
	"html/template"
	"russiatravelapi"
	"strconv"
	"strings"
	"log"
)

// taken from https://github.com/kennygrant/sanitize
func HTML(s string) string {

	output := ""

	// Shortcut strings with no tags in them
	if !strings.ContainsAny(s, "<>") {
		output = s
	} else {

		// First remove line breaks etc as these have no meaning outside html tags (except pre)
		// this means pre sections will lose formatting... but will result in less uninentional paras.
		s = strings.Replace(s, "\n", "", -1)

		// Then replace line breaks with newlines, to preserve that formatting
		s = strings.Replace(s, "</p>", "\n", -1)
		s = strings.Replace(s, "<br>", "\n", -1)
		s = strings.Replace(s, "</br>", "\n", -1)
		s = strings.Replace(s, "<br/>", "\n", -1)

		// Walk through the string removing all tags
		b := bytes.NewBufferString("")
		inTag := false
		for _, r := range s {
			switch r {
			case '<':
				inTag = true
			case '>':
				inTag = false
			default:
				if !inTag {
					b.WriteRune(r)
				}
			}
		}
		output = b.String()
	}

	// Remove a few common harmless entities, to arrive at something more like plain text
	output = strings.Replace(output, "&#8216;", "'", -1)
	output = strings.Replace(output, "&#8217;", "'", -1)
	output = strings.Replace(output, "&#8220;", "\"", -1)
	output = strings.Replace(output, "&#8221;", "\"", -1)
	output = strings.Replace(output, "&nbsp;", " ", -1)
	output = strings.Replace(output, "&quot;", "\"", -1)
	output = strings.Replace(output, "&apos;", "'", -1)

	// Translate some entities into their plain text equivalent (for example accents, if encoded as entities)
	output = html.UnescapeString(output)

	// In case we have missed any tags above, escape the text - removes <, >, &, ' and ".
	output = template.HTMLEscapeString(output)

	// After processing, remove some harmless entities &, ' and " which are encoded by HTMLEscapeString
	output = strings.Replace(output, "&#34;", "\"", -1)
	output = strings.Replace(output, "&#39;", "'", -1)
	output = strings.Replace(output, "&amp; ", "& ", -1)     // NB space after
	output = strings.Replace(output, "&amp;amp; ", "& ", -1) // NB space after

	return output
}

func LocationToString(geo *tgbotapi.Location) string {
	return FloatToString(geo.Latitude) + "," + FloatToString(geo.Longitude)
}

func FloatToString(input float64) string {
	// to convert a float number to a string
	return strconv.FormatFloat(input, 'f', 6, 64)
}

func StringToLocation(coords string) russiatravelapi.Location {
	coords = strings.Replace(coords, "\"", "", -1)
	location := strings.Split(coords, ",")
	loc := russiatravelapi.Location{}
	one, err := strconv.ParseFloat(location[0], 64)
	two, err2 := strconv.ParseFloat(location[1], 64)

	if err != nil {
		fmt.Println(err)
	}

	if err2 != nil {
		fmt.Println(err)
	}
	loc.Latitude = one
	loc.Longitude = two
	return loc
}

func shortenDesc(desc string) string {
	log.Printf("%s", "Description shorted")
	for len(desc) > 2000 {
		desc = desc[:len(desc)-1]
		for string(desc[len(desc)-1]) != "." {
			desc = desc[:len(desc)-1]
		}
	}
	return desc
}
