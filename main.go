package main

import (
    "encoding/csv"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "math/rand"
    "net/http"
    "os"
)

type CSVData struct {
    Fullnames   []string
    Addresses   []string
    RandomTexts []string
}

func readCSV(filePath string) (CSVData, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return CSVData{}, err
    }
    defer file.Close()

    reader := csv.NewReader(file)
    reader.TrimLeadingSpace = true

    headers, err := reader.Read()
    if err != nil {
        return CSVData{}, err
    }

    fullnameIdx, addressIdx, randomTextIdx := -1, -1, -1

    for i, header := range headers {
        switch header {
        case "fullname":
            fullnameIdx = i
        case "address":
            addressIdx = i
        case "random-text":
            randomTextIdx = i
        }
    }

    if fullnameIdx == -1 || addressIdx == -1 || randomTextIdx == -1 {
        return CSVData{}, fmt.Errorf("CSV harus memiliki kolom 'fullname', 'address', dan 'random-text'")
    }

    var data CSVData

    for {
        record, err := reader.Read()
        if err == io.EOF {
            break
        }
        if err != nil {
            return CSVData{}, err
        }

        data.Fullnames = append(data.Fullnames, record[fullnameIdx])
        data.Addresses = append(data.Addresses, record[addressIdx])
        data.RandomTexts = append(data.RandomTexts, record[randomTextIdx])
    }

    return data, nil
}

func processPlaceholders(data interface{}, csvData CSVData) interface{} {
    switch v := data.(type) {
    case string:
        switch v {
        case "fullname":
            return csvData.Fullnames[rand.Intn(len(csvData.Fullnames))]
        case "address":
            return csvData.Addresses[rand.Intn(len(csvData.Addresses))]
        case "random-text":
            return csvData.RandomTexts[rand.Intn(len(csvData.RandomTexts))]
        default:
            return v
        }
    case []interface{}:
        newArr := make([]interface{}, len(v))
        for i, item := range v {
            newArr[i] = processPlaceholders(item, csvData)
        }
        return newArr
    case map[string]interface{}:
        newMap := make(map[string]interface{})
        for key, value := range v {
            newMap[key] = processPlaceholders(value, csvData)
        }
        return newMap
    default:
        return data
    }
}

func main() {
    rand.Seed(42) 

    csvData, err := readCSV("data.csv")
    if err != nil {
        log.Fatalf("Gagal membaca CSV: %v", err)
    }

    http.HandleFunc("/generate", func(w http.ResponseWriter, r *http.Request) {
        var requestData map[string]interface{}
        err := json.NewDecoder(r.Body).Decode(&requestData)
        if err != nil {
            http.Error(w, "Invalid request body", http.StatusBadRequest)
            return
        }

        processedData := processPlaceholders(requestData, csvData)

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(processedData)
    })

    port := 8080
    fmt.Printf("Server berjalan di port %d\n", port)
    log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
