package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type client struct {
	base, token string
	http        *http.Client
}

func (c client) call(method, path string, input, output any) (int, error) {
	var body io.Reader
	if input != nil {
		data, err := json.Marshal(input)
		if err != nil {
			return 0, err
		}
		body = bytes.NewReader(data)
	}
	request, err := http.NewRequest(method, c.base+path, body)
	if err != nil {
		return 0, err
	}
	request.Header.Set("Authorization", "Bearer "+c.token)
	request.Header.Set("Content-Type", "application/json")
	response, err := c.http.Do(request)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()
	if output != nil {
		if err := json.NewDecoder(response.Body).Decode(output); err != nil {
			return response.StatusCode, err
		}
	} else {
		_, _ = io.Copy(io.Discard, response.Body)
	}
	return response.StatusCode, nil
}

func require(name string, got, expected int, output any) {
	if got != expected {
		fmt.Fprintf(os.Stderr, "%s: expected %d, got %d: %v\n", name, expected, got, output)
		os.Exit(1)
	}
	fmt.Printf("%s: %d\n", name, got)
}

func main() {
	base := flag.String("base", "http://127.0.0.1:28088", "service base URL")
	token := flag.String("token", "dev-token", "API token")
	flag.Parse()
	c := client{base: strings.TrimSuffix(*base, "/"), token: *token, http: &http.Client{Timeout: 10 * time.Second}}

	features := []map[string]any{
		{"chromosome": "1", "start": 90, "end": 150, "kind": "exon", "id": "exon-1", "gene": "GENE1", "transcript": "TX1"},
		{"chromosome": "1", "start": 100, "end": 100, "kind": "frequency", "id": "freq-1", "attributes": map[string]string{"population": "global", "frequency": "0.01"}},
		{"chromosome": "1", "start": 100, "end": 100, "kind": "clinical", "id": "clinical-1", "attributes": map[string]string{"label": "likely_pathogenic"}},
	}
	create := map[string]any{"name": "demo-reference", "build": "GRCh38", "features": features}
	var dataset map[string]any
	status, err := c.call(http.MethodPost, "/v1/reference-datasets", create, &dataset)
	if err != nil {
		panic(err)
	}
	require("create dataset", status, http.StatusCreated, dataset)
	datasetID := dataset["id"].(string)

	var published map[string]any
	status, err = c.call(http.MethodPost, "/v1/reference-datasets/"+datasetID+"/publish", nil, &published)
	if err != nil {
		panic(err)
	}
	require("publish dataset", status, http.StatusOK, published)

	preview := map[string]any{"chromosome": "chr1", "position": 100, "reference": "A", "alternate": "G,T", "filters": []string{"PASS"}}
	var previewOutput map[string]any
	status, err = c.call(http.MethodPost, "/v1/normalization/preview", preview, &previewOutput)
	if err != nil {
		panic(err)
	}
	require("normalization preview", status, http.StatusOK, previewOutput)

	jobRequest := map[string]any{"dataset_id": datasetID, "variants": []map[string]any{{"chromosome": "1", "position": 100, "reference": "A", "alternate": "G", "filters": []string{"PASS"}}, {"chromosome": "1", "position": 500, "reference": "A", "alternate": "T"}}}
	var accepted map[string]any
	status, err = c.call(http.MethodPost, "/v1/annotation-jobs", jobRequest, &accepted)
	if err != nil {
		panic(err)
	}
	require("submit annotation job", status, http.StatusAccepted, accepted)
	jobID := accepted["id"].(string)

	var jobOutput map[string]any
	for attempt := 0; attempt < 50; attempt++ {
		status, err = c.call(http.MethodGet, "/v1/annotation-jobs/"+jobID, nil, &jobOutput)
		if err != nil {
			panic(err)
		}
		require("query annotation job", status, http.StatusOK, jobOutput)
		state, _ := jobOutput["status"].(string)
		if state == "completed" || state == "partial" || state == "failed" {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if jobOutput["status"] != "completed" {
		fmt.Fprintf(os.Stderr, "unexpected job result: %v\n", jobOutput)
		os.Exit(1)
	}

	var region map[string]any
	status, err = c.call(http.MethodGet, "/v1/regions/1/90/150?dataset_id="+datasetID, nil, &region)
	if err != nil {
		panic(err)
	}
	require("region query", status, http.StatusOK, region)
	fmt.Println("simulation completed")
}
