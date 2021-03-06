package main

import (
	"log"
	"os"
	"runtime"
	"time"

	"github.com/KyberNetwork/server-go/fetcher"
	"github.com/KyberNetwork/server-go/http"
	persister "github.com/KyberNetwork/server-go/persister"
)

type fetcherFunc func(persister persister.Persister, fetcher *fetcher.Fetcher)

func main() {
	numCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPU)
	//set log for server
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	f, err := os.OpenFile("error.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	//clear error log file
	err = f.Truncate(0)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	log.SetOutput(f)

	persisterIns, _ := persister.NewPersister("ram")
	fertcherIns, err := fetcher.NewFetcher()
	if err != nil {
		log.Fatal(err)
	}

	//run fetch data
	runFetchData(persisterIns, fetchKyberEnabled, fertcherIns, 10)
	runFetchData(persisterIns, fetchMaxGasPrice, fertcherIns, 60)

	runFetchData(persisterIns, fetchGasPrice, fertcherIns, 30)

	runFetchData(persisterIns, fetchRateUSD, fertcherIns, 60)
	runFetchData(persisterIns, fetchBlockNumber, fertcherIns, 10)
	runFetchData(persisterIns, fetchRate, fertcherIns, 10)
	runFetchData(persisterIns, fetchEvent, fertcherIns, 30)
	//runFetchData(persisterIns, fetchKyberEnable, fertcherIns, 10)

	//run server
	server := http.NewHTTPServer(":3001", persisterIns)
	server.Run()

	//init fetch data

}

// func setLogServer() {
// 	log.SetFlags(log.LstdFlags | log.Lshortfile)
// 	f, err := os.OpenFile("error.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer f.Close()
// 	log.SetOutput(f)
// }

func runFetchData(persister persister.Persister, fn fetcherFunc, fertcherIns *fetcher.Fetcher, interval time.Duration) {
	fn(persister, fertcherIns)
	ticker := time.NewTicker(5 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				fn(persister, fertcherIns)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func fetchGasPrice(persister persister.Persister, fetcher *fetcher.Fetcher) {
	gasPrice, err := fetcher.GetGasPrice()
	if err != nil {
		log.Print(err)
		persister.SetNewGasPrice(false)
		return
	}
	persister.SaveGasPrice(gasPrice)
}

func fetchMaxGasPrice(persister persister.Persister, fetcher *fetcher.Fetcher) {
	gasPrice, err := fetcher.GetMaxGasPrice()
	if err != nil {
		log.Print(err)
		persister.SetNewMaxGasPrice(false)
		return
	}
	persister.SaveMaxGasPrice(gasPrice)
}

func fetchKyberEnabled(persister persister.Persister, fetcher *fetcher.Fetcher) {
	enabled, err := fetcher.CheckKyberEnable()
	if err != nil {
		log.Print(err)
		persister.SetNewKyberEnabled(false)
		return
	}
	persister.SaveKyberEnabled(enabled)
}

func fetchRateUSD(persister persister.Persister, fetcher *fetcher.Fetcher) {
	body, err := fetcher.GetRateUsd()
	if err != nil {
		log.Print(err)
		persister.SetNewRateUSD(false)
		return
	}
	err = persister.SaveRateUSD(body)
	if err != nil {
		log.Print(err)
		persister.SetNewRateUSD(false)
		return
	}
}

func fetchBlockNumber(persister persister.Persister, fetcher *fetcher.Fetcher) {
	blockNum, err := fetcher.GetLatestBlock()
	if err != nil {
		log.Print(err)
		persister.SetNewLatestBlock(false)
		return
	}
	err = persister.SaveLatestBlock(blockNum)
	if err != nil {
		persister.SetNewLatestBlock(false)
		log.Print(err)
		return
	}
}

func fetchRate(persister persister.Persister, fetcher *fetcher.Fetcher) {
	rates, err := fetcher.GetRate()
	if err != nil {
		log.Print(err)
		persister.SetNewRate(false)
		return
	}
	persister.SaveRate(rates)
	persister.SetNewRate(true)
}

func fetchEvent(persister persister.Persister, fetcher *fetcher.Fetcher) {
	if persister.GetIsNewLatestBlock() {
		blockNum := persister.GetLatestBlock()
		events, err := fetcher.GetEvents(blockNum)
		if err != nil {
			log.Print(err)
			persister.SetNewEvents(false)
			return
		}
		persister.SaveEvent(events)
		persister.SetNewEvents(true)
	} else {
		persister.SetNewEvents(false)
	}
}

// func fetchKyberEnable(persister persister.Persister, fetcher *fetcher.Fetcher) {
// 	enable, err := fetcher.GetKyberEnable()
// 	if err != nil {
// 		log.Print(err)
// 		persister.SetNewKyberEnable(false)
// 		return
// 	}
// 	persister.SaveKyberEnable(enable)
// 	persister.SetNewKyberEnable(true)
// }
