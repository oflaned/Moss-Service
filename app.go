package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
)

type Program struct {
	Name     string `json:"name"`
	Language string `json:"language"`
	Code     string `json:"code"`
}

type MossResult struct {
	URL string `json:"url"`
}

func main() {
	http.HandleFunc("/moss", checkPlagiarismHandler)
	http.ListenAndServe(":8005", nil)

}
func checkPlagiarismHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Error reading request body")
		return
	}

	var programs []Program
	err = json.Unmarshal(body, &programs)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error parsing JSON")
		return
	}

	var mossFiles []string
	for _, program := range programs {
		fileName := program.Name + "." + program.Language
		err := ioutil.WriteFile(fileName, []byte(program.Code), 0644)
		defer os.Remove(fileName)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "Error writing program file")
			return
		}
		mossFiles = append(mossFiles, fileName)
	}

	resultURL, err := runMoss(mossFiles)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Error running Moss ", err)
		return
	}

	response := MossResult{
		URL: resultURL,
	}
	responseJSON, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Error creating response JSON")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)

	for _, fileName := range mossFiles {
		os.Remove(fileName)
	}
}

func runMoss(files []string) (string, error) {
	cmd := exec.Command("./moss", "-l", "cc")
	cmd.Args = append(cmd.Args, files...)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	err := cmd.Run()
	if err != nil {
		return "", err
	}

	output := stdout.String()

	resultURL := parseMossResultURL(output)

	return resultURL, nil
}

func parseMossResultURL(output string) string {
	startIndex := bytes.Index([]byte(output), []byte("http://moss.stanford.edu/results/"))
	if startIndex == -1 {
		return ""
	}

	endIndex := bytes.Index([]byte(output[startIndex:]), []byte("\n"))
	if endIndex == -1 {
		return ""
	}

	url := output[startIndex : startIndex+endIndex]
	return string(url)
}
