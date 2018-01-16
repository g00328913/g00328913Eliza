package main

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"encoding/json"
	"html/template"
	"net/http"
	"log"
	"os"
	"bufio"
	"strconv"
)

//struct for the users input
type UserInput struct {
	UserText string
}

//struct for Eliza output
type ElizaOutput struct {
	ElizaMessage string
}



// ElizaResponse
func ElizaResponse(input string) string {
	
	if matched, _ := regexp.MatchString(`(?i).*\bfather\b.*`, input); matched {
		return "Why don't you tell me more about your father?"
	}

	re := regexp.MustCompile(`(?i).*\bI'?\s*a?m \b([^.?!]*)[.?!]?`)
	if matched := re.MatchString(input); matched {
		subMatch := re.ReplaceAllString(input, "$1?")
		reflectedString := Reflect(subMatch)
		return "How do you know you are " + reflectedString
	}

	responses := []string{
		"I'm not sure what you're trying to say. Could you explain it to me?",
		"How does that make you feel?",
		"Why do you say that?",
	}

	randindex := rand.Intn(len(responses))

	return responses[randindex]
} // ElizaResponse
	
func Reflect(input string) string {
	// Split the input on word boundaries.
	// boundaries := regexp.MustCompile(`\b`)
	boundaries := regexp.MustCompile(`(?=\S*['-])([a-zA-Z'-]+)`)
	tokens := boundaries.Split(input, -1)

	// Some key prepositions
	prepositions := []string{
		"to",
		"by",
		"under",
		"about",
		"on",
		"according",
		"over",
		"of",
		"without",
	}

	// List the reflections.
	reflections := [][]string{
		{`was`, `were`},
		{`I`, `you`},
		{`I'm`, `you are`},
		{`I'd`, `you would`},
		{`I've`, `you have`},
		{`I'll`, `you will`},
		{`my`, `your`},
		{`you're`, `I am`},
		{`were`, `was`},
		{`you've`, `I have`},
		{`you'll`, `I will`},
		{`your`, `my`},
		{`yours`, `mine`},
		// {`you`, `me`},
		{`me`, `you`},
	}

	
	for i, token := range tokens {
		for _, reflection := range reflections {

			
			if token == "you" {

				// Loop through the prepositions
				for j, preposition := range prepositions {
					// Compare the previous word, that is 'token[i-2]' to the 'preposition'. 'token[i-1]' is the space character.
					if tokens[i-2] == preposition {
						// If 'you' is an object pronoun to a preposition, the swap it for 'me'
						tokens[i] = "me"
						break
					}

					
					if j == len(prepositions)-1 {
						tokens[i] = "I"
					}

				} // for j, prepostition
				// As for the rest of reflections, keep doing the normal substitution
			} else if matched, _ := regexp.MatchString(reflection[0], token); matched {
				tokens[i] = reflection[1]
				break
			} // if - else if

		} // for 'reflection'
	} // for 'i'

	
	return strings.Join(tokens, ``)

} // Reflect


// Replacer is a struct with two elements: a compiled regular expression and an array of strings containing possible replacements matching the regular expression.
type Replacer struct {
	original     *regexp.Regexp
	replacements []string
}

// ReadReplacersFromFile reads an array of Replacers from a text file.
// It takes a single argument: a string which is the path to the data file.
// The file should be a series of sections with the following format:
//   All lines that begin with a hash symbol are ignored.
//   Each section should begin with a regular expression on a single line.
//   Each subsequent line, until a blank line, should contain a possible
//   replacement for a string matching the regular expression.
//   Each section should end with at least one blank line.
// The idea is to create an array that can be traversed, looking for the first
// regular expression to match some input string. Once a match is found, a
// random replacement string is returned.
func ReadReplacersFromFile(path string) []Replacer {
	// Open the file, logging a fatal error if it fails, close on return.
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Create an empty array of Replacers.
	var replacers []Replacer

	// Read the file line by line.
	for scanner, readoriginal := bufio.NewScanner(file), false; scanner.Scan(); {
		// Read the next line and decide what to do.
		switch line := scanner.Text(); {
		// If the line starts with a # character then skip it.
		case strings.HasPrefix(line, "#"):
			// Do nothing
		// If we see a blank line, then make sure we indicate a new section.
		case len(line) == 0:
			readoriginal = false
		// If we haven't read the original, then append an element to the
		// replacers array, compiling the regular expression. The replacements
		// array is left blank for now.
		case readoriginal == false:
			replacers = append(replacers, Replacer{original: regexp.MustCompile(line)})
			readoriginal = true
		// Otherwise read a replacement and add it to the last replacer.
		default:
			replacers[len(replacers)-1].replacements = append(replacers[len(replacers)-1].replacements, line)
		}
	}
	// Return the replacers array.
	return replacers
}

// Eliza is a data structure representing a chatbot.
type Eliza struct {
	responses     []Replacer
	substitutions []Replacer
}

// ElizaFromFiles reads in text files containing responses and substitutions
func ElizaFromFiles(responsePath string, substitutionPath string) Eliza {
	eliza := Eliza{}

	eliza.responses = ReadReplacersFromFile(responsePath)
	eliza.substitutions = ReadReplacersFromFile(substitutionPath)

	return eliza
}


func (me *Eliza) RespondTo(input string) string {
	// Look for a possible response.
	for _, response := range me.responses {
		// Check if the user input matches the original, capturing any groups.
		if matches := response.original.FindStringSubmatch(input); matches != nil {
			// Select a random response.
			output := response.replacements[rand.Intn(len(response.replacements))]
			// We'll tokenise the captured groups using the following regular expression.
			boundaries := regexp.MustCompile(`[\s,.?!]+`)
			// Fill the response with each captured group from the input.
			// This is a bit complex, because we have to reflect the pronouns.
			for m, match := range matches[1:] {
				// First split the captured group into tokens.
				tokens := boundaries.Split(match, -1)
				// Loop through the tokens.
				for t, token := range tokens {
					// Loop through the potential substitutions.
					for _, substitution := range me.substitutions {
						// Check if the original of the current substitution matches the token.
						if substitution.original.MatchString(token) {
							// If it matches, replace the token with one of the replacements (at random).
							// Then break.
							tokens[t] = substitution.replacements[rand.Intn(len(substitution.replacements))]
							break
						}
					}
				}
				// Replace $1 with the first match, $2 with the second, etc.
				// Note that element 0 of matches is the original match, not a captured group.
				output = strings.Replace(output, "$"+strconv.Itoa(m+1), strings.Join(tokens, " "), -1)
			}
			// Send the filled answer back.
			return output
		}
	}
	// If there are no matches, then return this generic response.
	return "I don't know what to say."
}


// Redirect to '/index'
func redirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "http://localhost:8080/index", 301)
}

// Default Request Handler
func defaultHandler(w http.ResponseWriter, r *http.Request) {
	// fp := path.Join("templates", "ajax-json.html")
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

} // defaultHandler



func ajaxHandler(w http.ResponseWriter, r *http.Request) {
	
	var userInput UserInput
	err := json.NewDecoder(r.Body).Decode(&userInput)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	
	eliza := ElizaFromFiles("data/responses.txt", "data/substitutions.txt")

	var elizaOutput ElizaOutput
	
	elizaOutput.ElizaMessage = eliza.RespondTo(userInput.UserText)

	
	reply, err := json.Marshal(elizaOutput)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Write(reply)

} // ajaxHandler

//adapted with assistance from https://golang.org/pkg/net/http/
func main() {
    http.HandleFunc("/", redirect)
    http.HandleFunc("/index", defaultHandler)
    http.HandleFunc("/ajax", ajaxHandler)

    fmt.Println("Server running at port 8080...")

    err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}