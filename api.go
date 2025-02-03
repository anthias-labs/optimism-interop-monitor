package main

import (
	"fmt"
	"math/big"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type BinStat struct {
	MessageCount     uint64
	TotalLatency     *big.Int
	MissingPart      uint64
	MissingReception uint64
	MissingRelay     uint64
}

type BlockPrettyStat struct {
	MessageCount     uint64  `json:"messageCount"`
	AvgLatency       float64 `json:"avgLatency"`
	MissingPart      uint64  `json:"missingMessages"`
	MissingReception uint64  `json:"missingReception"`
	MissingRelay     uint64  `json:"missingRelay"`
}

func prettyBlockStat(bs BlockStat) (bps BlockPrettyStat) {
	bps.MessageCount = bs.MessageCount
	if bps.MessageCount == 0 {
		bps.AvgLatency = 0
	} else {
		bps.AvgLatency = float64(bs.TotalLatency.Uint64()) / float64(bps.MessageCount)
	}
	bps.MissingPart = (bs.ReceivedMessages + bs.SentMesssages) - 2*(bs.MessageCount)
	bps.MissingReception = bs.SentMesssages - bs.MessageCount
	bps.MissingRelay = bs.ReceivedMessages - bs.MessageCount

	return
}

func binToPrettyStat(bs BinStat) (bps BlockPrettyStat) {
	bps.MessageCount = bs.MessageCount
	if bps.MessageCount == 0 {
		bps.AvgLatency = 0
	} else {
		bps.AvgLatency = float64(bs.TotalLatency.Uint64()) / float64(bps.MessageCount)
	}
	bps.MissingPart = bs.MissingPart

	return
}

func homeRoute(c echo.Context) error {
	return c.String(http.StatusOK, "Ok!")
}

func (agg *Aggregator) All(c echo.Context) error {
	prettyStats := make(map[uint64]BlockPrettyStat)

	fromBlock := c.QueryParam("from")
	var err error
	from := uint64(0)

	if fromBlock != "" {
		from, err = strconv.ParseUint(fromBlock, 10, 64)

		if err != nil {
			return c.String(http.StatusBadRequest, "Invalid `from` value")
		}
	}

	if c.QueryParam("bin") == "" {
		for key, val := range agg.BlockStats {
			if key >= from {
				prettyStats[key] = prettyBlockStat(val)
			}

		}

		return c.JSON(http.StatusOK, prettyStats)
	}

	binSize, err := strconv.ParseUint(c.QueryParam("bin"), 10, 64)
	if binSize <= 0 {
		err = fmt.Errorf("invalid bin size")
	}

	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid `bin` value")
	}

	aggStats := make(map[uint64]BinStat)

	for key, val := range agg.BlockStats {
		if key >= from {
			bin := key - (key % binSize)
			_, ex := aggStats[bin]

			if !ex {
				aggStats[bin] = BinStat{0, big.NewInt(0), 0, 0, 0}
			}

			newStats := BinStat{}
			newStats.MessageCount = aggStats[bin].MessageCount + val.MessageCount
			newStats.TotalLatency = big.NewInt(0).Add(aggStats[bin].TotalLatency, val.TotalLatency)
			newStats.MissingPart = aggStats[bin].MissingPart + (val.ReceivedMessages + val.SentMesssages) - 2*(val.MessageCount)
			newStats.MissingReception = aggStats[bin].MissingReception + val.SentMesssages - val.MessageCount
			newStats.MissingRelay = aggStats[bin].MissingRelay + val.ReceivedMessages - val.MessageCount
			aggStats[bin] = newStats
		}
	}

	for key, val := range aggStats {
		prettyStats[key] = binToPrettyStat(val)
	}

	return c.JSON(http.StatusOK, prettyStats)

}

func (agg *Aggregator) LatestBlockRoute(c echo.Context) error {
	blockCountParam := c.QueryParam("count")
	var blockCount uint64 = agg.config.AggregateBlockAmount
	var err error

	if blockCountParam != "" {
		blockCount, err = strconv.ParseUint(blockCountParam, 10, 64)

		if err != nil {
			return c.String(http.StatusBadRequest, "Invalid `count` value")
		}

		if blockCount > 2*agg.config.AggregateBlockAmount && agg.config.PurgeOldBlocks {
			return c.String(http.StatusBadRequest, "`count` too large; purgeOldBlocks is set to true")
		}
	}

	stats := agg.AggregateLatestBlocks(blockCount)

	return c.JSON(http.StatusOK, stats)
}

func StartApi(config *Config, agg *Aggregator) {
	e := echo.New()
	e.HideBanner = true
	e.Debug = true

	e.GET("/", homeRoute)
	e.GET("/all", agg.All)
	e.GET("/latest", agg.LatestBlockRoute)

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", config.APIPort)))
}
