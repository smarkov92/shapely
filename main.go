package main

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"shapleyTask/attribution"
	"shapleyTask/model"
	"strconv"
	"strings"
)

func loadShapleyDataFromCSV(path string) ([]model.DataForShap, error) {
	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	var data []model.DataForShap
	var singleData model.DataForShap
	csvReader := csv.NewReader(f)
	csvReader.Comma = ';'

	for {
		rec, err := csvReader.Read()
		log.Println(rec)
		if err == io.EOF {
			break
		}
		if len(rec) == 0 {
			continue
		}
		val, _ := strconv.Atoi(rec[1])
		singleData.Value = uint64(val)
		singleData.Path = strings.Split(rec[0], ",")
		data = append(data, singleData)
	}

	return data, nil

}

func runShapley() {

	data, _ := loadShapleyDataFromCSV("data_for_shap.csv")
	var totalValue uint64
	for _, item := range data {
		totalValue += item.Value
	}
	vector := attribution.CalculateShapleyVectorEasy(data)
	var res float64
	for _, value := range vector {
		res += value

	}
	log.Println(res)
}

func main() {
	runShapley()
}
