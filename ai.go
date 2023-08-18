package main

// This AI shell was developed by Jacky Kit using Go to test performance.
// JK Labs
// https://3jk.net
// mail@jackykit.com

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var modelsMap = map[string]string{
        "1":  "assistant",
        "2":  "gpt-3.5-turbo",
        "3":  "gpt-4",
        "4": "google-palm",
        "5": "llama_2_70b_chat",
}


func extractUrls(text string) []string {
	r := regexp.MustCompile(`http[s]?://(?:[a-zA-Z]|[0-9]|[$-_@.&+]|[!*\\(\\),]|(?:%[0-9a-fA-F][0-9a-fA-F]))+`)
	return r.FindAllString(text, -1)
}

func fetchURLContent(url string) string {
	res, err := http.Get(url)
	if err != nil {
		return fmt.Sprintf("Error fetching content from %s: %v", url, err)
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return fmt.Sprintf("Error parsing content from %s: %v", url, err)
	}

	doc.Find("script, style").Each(func(i int, s *goquery.Selection) {
		s.Remove()
	})

	return doc.Text()
}


func displayModelsHelp() {
    green := "\033[32m"
    jk := "\033[31m"
    reset := "\033[0m"

	cmd := exec.Command("clear")
	stdout, err := cmd.Output()
    if err != nil {
        fmt.Println(err.Error())
        return
    }

    fmt.Print(string(stdout))

	fmt.Println( jk+ "\n\nWelcome to JK Ai!" + reset + "\nPlease choose a model:")
	for i := 1; i <= len(modelsMap); i++ {
		key := fmt.Sprintf("%d", i)
		if model, exists := modelsMap[key]; exists {
			fmt.Printf(green + "%s: %s\n" + reset, key, model)
		}
	}
}


func main() {
	reader := bufio.NewReader(os.Stdin)

	// Display available models and ask for selection
	displayModelsHelp()
	fmt.Print("Choose a model number: ")
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)
	model, exists := modelsMap[choice]
	if !exists {
		fmt.Println("Invalid model choice. Exiting.")
		return
	}

	fmt.Print("Input: ")
	inputText, _ := reader.ReadString('\n')
	inputText = strings.TrimSpace(inputText)

	urls := extractUrls(inputText)
	var combinedContent string

	if len(urls) > 0 {
		for _, url := range urls {
			content := fetchURLContent(url)
			inputText = regexp.MustCompile(regexp.QuoteMeta(url)).ReplaceAllString(inputText, "")
			combinedContent = inputText + " " + content
		}
	} else {
		combinedContent = inputText
	}

	apiURL := "http://jklabs-ai001:8088/v1/chat/completions"
	data := map[string]interface{}{
		"model":    model,
		"messages": []map[string]string{{"role": "user", "content": combinedContent}},
	}

	jsonData, _ := json.Marshal(data)
	req, _ := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer LOGINLABS")

	client := &http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	var response map[string]interface{}
	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(body, &response)

	if choices, ok := response["choices"].([]interface{}); ok && len(choices) > 0 {
		fmt.Println("Response:", choices[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string))
	} else {
		fmt.Println("No response received.")
	}
}
