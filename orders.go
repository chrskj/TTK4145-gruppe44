// -spawne en backup (phoenix)
//
// Holde styr på alle andres ordre-lagre til fil, holde alle orders oppdatert
//
// -be om cost function
// -lagre til fil, holde alle orders oppdatert
// -tildel ordre

//sende ut alle orders den har til komm

//ta imot alle andres ordre fra komm, og vurdere hva som er nye ordre, og hva som er ferdige ordre (og da slette)
package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"
)

var numElevators int = 3
var thisElevator int
var costChan chan ChannelPacket

type Order struct {
	elevator  int
	toFloor   int64
	direction int64
	timestamp uint64
}

type ChannelPacket struct{
	packetType string
	elevator int
	toFloor int64
	direction int64
	timestamp uint64
	cost float64
}

func init(ordersToCom, comToOrders,ordersToElevAlgo,elevAlgoToOrders) {
	var data []Order
	var costChan chan OcomPacket
	var tstamp uint64 = 1
	ticker := time.NewTicker(100 * time.Millisecond)
	go func() {
		for {
			<-ticker.C
			tstamp++
		}
	}()
	//data = readFile()
	//try to get data from others

	//get elevator ID

	go orderRoutine(data)
}

func orderRoutine(data []Order,ordersToCom chan ChannelPacket, comToOrders chan ChannelPacket,ordersToElevAlgo chan ChannelPacket,elevAlgoToOrders chan ChannelPacket){
	var costChan chan ChannelPacket
	for{
		select{
		case temp:= <- comToOrders:
			switch temp.packetType{
			case "compareCost":
				costChan <- temp
			case "orderComplete":
				toRemove := Order{
					elevator:  temp.elevator
					toFloor:   temp.toFloor
					direction: temp.direction
					timestamp: temp.timestamp
				}
				removeOrder(toRemove)
			case "addOrder":
				newOrder := Order{
					elevator:  temp.elevator
					toFloor:   temp.toFloor
					direction: temp.direction
					timestamp: temp.timestamp
				}
				addOrder(newOrder)
			case "getOrderList":
				packet := Order{
					packetType: "orderList"

				}
				ordersToCom <- 
			}
		case temp:= <- elevAlgoToOrders:
			switch temp.packetType{
			case "buttonPress":
				newOrder := Order{
					elevator: -1
					toFloor: temp.toFloor
					direction: temp.direction
					timestamp: tstamp
				}
				//check if order already exists
				for index, value := range data {
					if value.toFloor == newOrder.toFloor && value.direction == newOrder.direction {
						newOrder.timestamp = 0
						break
					}
				}
				//if not: start the cost compare
				if newOrder.timestamp>0{
					go costCompare(newOrder, ordersToCom)
				}
			}
		}
	}
}

func costCompare(newOrder Order,ordersToCom chan ChannelPacket){
	ordersToCom <- ChannelPacket{
		packetType: "requestCostFunc"
		elevator: thisElevator
	}
	costTicker := time.NewTicker(10*time.Millisecond)
	var ticks uint = 0
	var costs []ChannelPacket
	for recievedOrders :=0; recievedOrders<numElevators && ticks<200;{
		select{
		case temp <- costChan:
			unique := true
			for _,val := range costs{
				if val == temp{
					unique = false
				}
			}
			if unique {
				costs = append(costs, temp)
				recievedOrders++
			}
		case <- costTicker.C:
			ticks++
		}
	}
	max := 9999
	for _, val := range costs{
		if val.cost < max {
			max = [newOrder.elevator,val.cost]
		}
	}
	if newOrder.elevator != -1{
		data = go addOrder(data,newOrder)
	} else {
		//error, no costs received
	}
}

func readFile() []Order {
	var data []Order
	file, err := os.Open("orders.csv")
	checkError("Cannot create file", err)
	defer file.Close()

	reader := csv.NewReader(file)
	fmt.Println("before read")
	for {
		input, error := reader.Read()
		if error == io.EOF {
			break
		}
		for i := 0; i < numElevators; i++ {
			toFloorTemp, _ := strconv.ParseInt(input[0+3*i], 10, 64)
			directionTemp, _ := strconv.ParseInt(input[1+3*i], 10, 64)
			tstampTemp, _ := strconv.ParseUint(input[2+3*i], 10, 64)
			elevatorTemp := i + 1
			data = append(data, Order{
				elevator:  elevatorTemp,
				toFloor:   toFloorTemp,
				direction: directionTemp,
				timestamp: tstampTemp,
			})
		}
		toFloorTemp, _ := strconv.ParseInt(input[3*numElevators], 10, 64)
		tstampTemp, _ := strconv.ParseUint(input[3*numElevators+1], 10, 64)
		data = append(data, Order{
			elevator:  0,
			toFloor:   toFloorTemp,
			timestamp: tstampTemp,
		})
	}
	return data
}

func writeToFile(data []Order) {
	fmt.Println("before write")
	file, err := os.Create("orders.csv")
	checkError("Cannot create file", err)
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()

	for i := 0; i < (len(data) / (numElevators+1))); i++ {
		values := data[((numElevators+1) * i):((numElevators+1)*i + (numElevators+1))]
		var value []string
		for j := 0; j < 3; j++ {

			value = append(value, strconv.FormatInt(values[j].toFloor, 10))
			value = append(value, strconv.FormatInt(values[j].direction, 10))
			value = append(value, strconv.FormatUint(values[j].timestamp, 10))
		}
		value = append(value, strconv.FormatInt(values[numElevators].toFloor, 10))
		value = append(value, strconv.FormatUint(values[numElevators].timestamp, 10))
		var valueStr string
		for j := 0; j < 3*numElevators+1; j++ {
			valueStr = valueStr + value[j] + ","
		}
		valueStr = valueStr[:len(valueStr)-1]
		err := writer.Write(value)
		checkError("Cannot write to file", err)
	}
}

func addOrder(data []Order, newOrder Order) []Order {
	blankOrder := Order{
		elevator:  0,
		toFloor:   0,
		direction: 0,
		timestamp: 0,
	}
	for index, value := range data {
		if value.timestamp == 0 && value.elevator == newOrder.elevator {
			data[index] = newOrder
			newOrder.timestamp = 0
		}
	}
	if newOrder.timestamp != 0 {
		data = append(data, newOrder)
		for i := 0; i < numElevators; i++ {
			data = append(data, blankOrder)
		}
	}
	return data
}

func removeOrder(data []Order, toRemove Order) []Order {
	blankOrder := Order{
		elevator:  0,
		toFloor:   0,
		direction: 0,
		timestamp: 0,
	}
	for index, value := range data {
		if value == toRemove {
			data[index] = blankOrder
		}
	}
	return data
}

func checkError(message string, err error) {
	if err != nil {
		log.Fatal(message, err)
	}
}