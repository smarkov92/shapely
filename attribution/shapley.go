package attribution

import (
	"github.com/ernestosuarez/itertools"
	"log"
	"shapleyTask/model"
	"sort"
	"strings"
	"sync"
)

func factorial(n float64) float64 {
	var fact float64
	fact = 1
	var num float64
	num = 2
	for ; num <= n+1; num++ {
		fact = fact * num
	}

	return fact
}

func iter(channels []string, i int, c1 chan []string, n *sync.WaitGroup, exit chan bool) {

	for j := range itertools.CombinationsStr(channels, i+1) {
		if len(j) != 0 {
			c1 <- j
		}
	}

	defer n.Done()
	log.Printf("%v of %v", i, len(channels))

	exit <- true
}

func powerSetOptimized(channels []string) [][]string {
	log.Println("Power Set Start")
	var powerSet [][]string
	var c1Sync sync.WaitGroup
	var exitSync sync.WaitGroup
	c1 := make(chan []string, 10000000)
	exit := make(chan bool)
	count := 0

	exitSync.Add(1)
	go func() {
		for j := range c1 {
			powerSet = append(powerSet, j)
		}
		exitSync.Done()
	}()

	exitSync.Add(1)
	go func() {
		for j := range exit {
			if j {
				count++
			}
		}
		exitSync.Done()
	}()

	log.Println("Power Set Start For")

	for i := 0; i < len(channels); {
		if i-count < 5 {
			log.Println("Start goroutine ", i)
			c1Sync.Add(1)
			go iter(channels, i, c1, &c1Sync, exit)
			i++
		}
	}

	c1Sync.Wait()
	close(c1)
	close(exit)
	exitSync.Wait()

	return powerSet
}

func subsets(s []string) []string {

	if len(s) == 1 {
		return s
	}
	var subChannels [][]string

	for i, _ := range s {
		for j := range itertools.CombinationsStr(s, i+1) {
			subChannels = append(subChannels, j)
		}

	}
	var res []string
	for _, sub := range subChannels {
		sort.Strings(sub)
		res = append(res, strings.Join(sub, ","))
	}

	return res

}

func vFunctionOrderliness(A []string, cValues map[string]uint64) uint64 {

	subsetsOfA := subsets(A)
	var worthOfA uint64
	for _, subset := range subsetsOfA {
		if _, ok := cValues[subset]; ok {
			worthOfA += cValues[subset]
		}
	}

	return worthOfA
}

func vFunctionEasy(A []string, cValues map[string]uint64) uint64 {

	var worthOfA uint64
	for _, subset := range A {
		if _, ok := cValues[subset]; ok {
			worthOfA += cValues[subset]
		}
	}

	return worthOfA
}

func CalculateShapleyVectorEasy(channelDict []model.DataForShap) map[string]float64 {
	var cValues = make(map[string]uint64, len(channelDict)) //c_values
	for i, _ := range channelDict {
		cValues[channelDict[i].GenerateKey()] = channelDict[i].Value
	}

	var shapVector = make(map[string]float64) //channels

	var channels []string
	var shapValues = make(map[string]float64)

	for i, _ := range channelDict {
		for _, source := range channelDict[i].Path {
			shapVector[source] += 1
		}
	}
	for key, _ := range shapVector {
		channels = append(channels, key)
	}

	vValues := make(map[string]uint64)
	log.Println("Start power set")
	pwSet := powerSetOptimized(channels)

	log.Println("Start vValue calc")
	for i, A := range pwSet {
		if i%1000000 == 0 {
			log.Printf("%v из %v", i, len(pwSet))
		}
		sort.Strings(A)
		vValues[(strings.Join(A, ","))] = vFunctionEasy(A, cValues)
	}

	N := float64(len(channels))

	log.Println("Start calculate  shapley value")

	//for _, channel := range channels {
	//	for A, _ := range vValues {
	//		if strings.Contains(A, channel) {
	//			continue
	//		} else {
	//			AWithChannel := strings.Split(A, ",")
	//			cardinalA := float64(len(AWithChannel))
	//			AWithChannel = append(AWithChannel, channel)
	//			sort.Strings(AWithChannel)
	//			AWithChannelKey := strings.Join(AWithChannel, ",")
	//			//weight := factorial(cardinalA) * factorial(N-cardinalA-1) / factorial(N)
	//
	//			weight := (factorial(cardinalA) * factorial(N-cardinalA-1)) / factorial(N)
	//			//weight := rationFactorial(cardinalA, N-cardinalA-1, N)
	//
	//			contrib := float64(vValues[AWithChannelKey]) - float64(vValues[A])
	//			shapValues[channel] += weight * contrib
	//
	//		}
	//	}
	//	shapValues[channel] += float64(vValues[channel]) / N
	//
	//}

	log.Println("Start calc weights optimized")
	for i, channel := range channels {
		for weight := range calcShapleyOptimized(channel, vValues, N) {
			shapValues[channel] += weight
		}
		shapValues[channel] += float64(vValues[channel]) / N
		log.Printf("Done %v of %v", i, len(channels))
	}

	log.Println("End Calc Value ")

	return shapValues

}

func calcShapleyOptimized(channel string, vValues map[string]uint64, N float64) chan float64 {

	countGor := 0

	//var goSync sync.WaitGroup
	var weightSync sync.WaitGroup

	//mapVvalues := make([]string, len(vValues))

	result := make(chan float64, len(vValues))
	exit := make(chan bool, len(vValues))
	defer close(result)

	//goSync.Add(1)
	//go func() {
	//	for j := range exit {
	//		if j {
	//			count++
	//		}
	//		//log.Println(count, " ", countGor)
	//	}
	//	goSync.Done()
	//}()

	for A, _ := range vValues {
		if !strings.Contains(A, channel) {
			weightSync.Add(1)
			countGor++
			go calcShapleyWeightOptimized(A, channel, vValues, N, result, &weightSync, exit)
		}
	}

	//for i := 0; i < len(mapVvalues); i++ {
	//	A := mapVvalues[i]
	//	if !strings.Contains(A, channel) {
	//		weightSync.Add(1)
	//		countGor++
	//		go calcShapleyWeightOptimized(A, channel, vValues, N, result, &weightSync, exit)
	//	}
	//}
	log.Println("Основной цикл закончился")
	log.Println("Ждем weight")
	weightSync.Wait()

	close(exit)
	log.Println("Ждем Закрыли exit")
	log.Println("Ждем Gosync weight")
	//goSync.Wait()

	return result
}

func calcShapleyWeightOptimized(A, channel string, vValues map[string]uint64, N float64, res chan float64, n *sync.WaitGroup, exit chan bool) {
	defer n.Done()
	AWithChannel := strings.Split(A, ",")
	cardinalA := float64(len(AWithChannel))
	AWithChannel = append(AWithChannel, channel)
	sort.Strings(AWithChannel)
	AWithChannelKey := strings.Join(AWithChannel, ",")
	//weight := factorial(cardinalA) * factorial(N-cardinalA-1) / factorial(N)

	weight := (factorial(cardinalA) * factorial(N-cardinalA-1)) / factorial(N)
	//weight := rationFactorial(cardinalA, N-cardinalA-1, N)

	contrib := float64(vValues[AWithChannelKey]) - float64(vValues[A])
	res <- weight * contrib
	exit <- true

}
