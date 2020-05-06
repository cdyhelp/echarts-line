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
				data: {{.LegendData}}
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
				data: {{.XAxisData}}
			},
			yAxis: {
				type: 'value'
			},
			series: [
				{{range .Series}}
				{
					name: {{.Name}},
					type: 'line',
					data: {{.Data}},
					markPoint: {
						data: [
							{type: 'max', name: '最大值'},
							{type: 'min', name: '最小值'}
						]
					},
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
		LegendData []template.HTML
		XAxisData  []string
		Series     []struct {
			Name template.HTML
			Data []float64
		}
	}
	option.Page = template.HTML(inputData[0])
	option.Title = template.HTML(inputData[1])
	option.XAxisData = make([]string, len(inputData)-3)
	dataName := strings.Split(inputData[2], "\t")
	option.LegendData = make([]template.HTML, len(dataName)-1)
	option.Series = make([]struct {
		Name template.HTML
		Data []float64
	}, len(dataName)-1)
	for i := 1; i < len(dataName); i++ {
		option.LegendData[i-1] = template.HTML(dataName[i])
		option.Series[i-1].Name = template.HTML(dataName[i])
	}
	for i := 3; i < len(inputData); i++ {
		data := strings.Split(inputData[i], "\t")
		option.XAxisData[i-3] = data[0]
		for j := 0; j < len(dataName)-1; j++ {
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
