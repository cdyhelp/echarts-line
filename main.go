package main

import (
	"bufio"
	"flag"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	input  string
	output string
)

func init() {
	flag.StringVar(&input, "i", "", "input")
	flag.StringVar(&output, "o", "", "output")
	flag.Parse()
}

const (
	dataPage int = iota
	dataTitle
	dataAxisLabel
	dataLegendName
	dataValue
)

func main() {
	if input == "" {
		flag.PrintDefaults()
		return
	}
	if output == "" {
		output = strings.TrimSuffix(input, filepath.Ext(input))
	}

	inputFile, err := os.Open(input)
	defer inputFile.Close()
	if err != nil {
		log.Fatal(err)
		return
	}
	var inputData []string
	scanner := bufio.NewScanner(inputFile)
	for scanner.Scan() {
		inputData = append(inputData, scanner.Text())
	}

	const templ = `<!DOCTYPE html>
	<html>
	<head>
		<meta charset="utf-8">
		<title>{{.Page}}</title>
		<script src="https://cdn.jsdelivr.net/npm/echarts/dist/echarts.min.js"></script>
	</head>
	
	<body>
	<div class="select" style="margin-right:10px; margin-top:10px; position:fixed; right:10px;"></div>
	
	<div class="container">
		<div class="item" id="container"
			 style="width:900px;height:500px;"></div>
	</div>
	<script type="text/javascript">
		"use strict";
		let myChart = echarts.init(document.getElementById('container'), "white");
		let option = {
			title: {
				text: {{.Title}}
			},
			tooltip: {
				trigger: 'axis'
			},
			legend: {
				data: {{.LegendName}}
			},
			toolbox: {
				show: true,
				feature: {
					dataZoom: {
						yAxisIndex: 'none'
					},
					dataView: {readOnly: false},
					magicType: {type: ['line', 'bar']},
					restore: {},
					saveAsImage: {}
				}
			},
			xAxis: {
				type: 'category',
				boundaryGap: false,
				data: {{.XAxisData}},
				axisLabel: {
					formatter: '{value}{{.XAxisLabel}}'
				}
			},
			yAxis: {
				type: 'value',
				axisLabel: {
					formatter: '{value}{{.YAxisLabel}}'
				}
			},
			series: [
				{{range .Series}}
				{
					name: {{.Name}},
					type: 'line',
					label : {show : true},
					data: {{.Data}},
					markLine: {
						data: [
							{type: 'average', name: '平均值'}
						]
					}
				},
				{{end}}
			]
		};
		myChart.setOption(option);
	</script>
	
	<style>
		.container {margin-top:30px; display: flex;justify-content: center;align-items: center;}
		.item {margin: auto;}
	</style>
	</body>
	</html>`
	t := template.Must(template.New("escape").Parse(templ))
	var option struct {
		Page       template.HTML
		Title      template.HTML
		XAxisLabel template.HTML
		YAxisLabel template.HTML
		LegendName []template.HTML
		XAxisData  []string
		Series     []struct {
			Name template.HTML
			Data []float64
		}
	}
	option.Page = template.HTML(inputData[dataPage])
	option.Title = template.HTML(inputData[dataTitle])
	axisLabel := strings.Split(inputData[dataAxisLabel], "\t")
	option.XAxisLabel, option.YAxisLabel = template.HTML(axisLabel[0]), template.HTML(axisLabel[1])
	option.XAxisData = make([]string, len(inputData)-3)
	legendName := strings.Split(inputData[dataLegendName], "\t")
	option.LegendName = make([]template.HTML, len(legendName)-1)
	option.Series = make([]struct {
		Name template.HTML
		Data []float64
	}, len(legendName)-1)
	for i := 1; i < len(legendName); i++ {
		option.LegendName[i-1] = template.HTML(legendName[i])
		option.Series[i-1].Name = template.HTML(legendName[i])
	}
	for i := dataValue; i < len(inputData); i++ {
		data := strings.Split(inputData[i], "\t")
		option.XAxisData[i-3] = data[0]
		for j := 0; j < len(legendName)-1; j++ {
			number, err := strconv.ParseFloat(data[j+1], 64)
			if err != nil {
				log.Fatal(err)
				return
			}
			option.Series[j].Data = append(option.Series[j].Data, number)
		}
	}
	outputFile := output + ".html"
	f, err := os.Create(outputFile)
	defer f.Close()
	if err != nil {
		log.Fatal(err)
		return
	}
	if err := t.Execute(f, option); err != nil {
		log.Fatal(err)
		return
	}
	println("output file : " + outputFile)
}
