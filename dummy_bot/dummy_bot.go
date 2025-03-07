package main

import (
	"context"
	"fmt"
	"time"

	"os"

	exchangetypes "github.com/InjectiveLabs/sdk-go/chain/exchange/types"
	"github.com/InjectiveLabs/sdk-go/client"
	chainclient "github.com/InjectiveLabs/sdk-go/client/chain"
	"github.com/InjectiveLabs/sdk-go/client/common"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func main() {
	fmt.Println("[!] Starting...")

	network := common.LoadNetwork("testnet", "lb")
	tmClient, err := rpchttp.New(network.TmEndpoint, "/websocket")
	if err != nil {
		panic(err)
	}

	senderAddress, cosmosKeyring, err := chainclient.InitCosmosKeyring(
		os.Getenv("HOME")+"/.injectived",
		"injectived",
		"file",
		"inj-user",
		"12345678",
		"5d386fbdbf11f1141010f81a46b40f94887367562bd33b452bbaa6ce1cd1381e",
		false,
	)
	if err != nil {
		panic(err)
	}

	clientCtx, err := chainclient.NewClientContext(
		network.ChainId,
		senderAddress.String(),
		cosmosKeyring,
	)
	if err != nil {
		panic(err)
	}

	clientCtx = clientCtx.WithNodeURI(network.TmEndpoint).WithClient(tmClient)

	chainClient, err := chainclient.NewChainClient(
		clientCtx,
		network,
		common.OptionGasPrices(client.DefaultGasPriceWithDenom),
	)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	marketsAssistant, err := chainclient.NewMarketsAssistant(ctx, chainClient)
	if err != nil {
		panic(err)
	}

	defaultSubaccountID := chainClient.DefaultSubaccount(senderAddress)

	// INJ/USDT
	marketId := "0x0611780ba69656949525013d947713300f56c37b6175e02f26bffa495c3208fe"

	for {
		fmt.Println("[!] Running the new section...")
		withMidPriceAndTob := true

		res, err := chainClient.FetchChainFullSpotMarket(ctx, marketId, withMidPriceAndTob)
		if err != nil {
			fmt.Println(err)
		}

		bestSellPrice := decimal.NewFromFloat(res.Market.MidPriceAndTob.BestSellPrice.MustFloat64() * 1e12)
		amount := decimal.NewFromFloat(2)
		fmt.Println("	[!] best buy price:", bestSellPrice, "USDT")
		fmt.Println("	[!] trying to BUY", amount, "INJ...")

		order := chainClient.CreateSpotOrder(
			defaultSubaccountID,
			&chainclient.SpotOrderData{
				OrderType:    exchangetypes.OrderType_BUY,
				Quantity:     amount,
				Price:        bestSellPrice,
				FeeRecipient: senderAddress.String(),
				MarketId:     marketId,
				Cid:          uuid.NewString(),
			},
			marketsAssistant)
		msg := new(exchangetypes.MsgBatchCreateSpotLimitOrders)
		msg.Sender = senderAddress.String()
		msg.Orders = []exchangetypes.SpotOrder{*order}

		simRes, err := chainClient.SimulateMsg(clientCtx, msg)
		if err != nil {
			fmt.Println(err)
			return
		}

		msgBatchCreateSpotLimitOrdersResponse := exchangetypes.MsgBatchCreateSpotLimitOrdersResponse{}
		err = msgBatchCreateSpotLimitOrdersResponse.Unmarshal(simRes.Result.MsgResponses[0].Value)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("	[+] simulated order hashes", msgBatchCreateSpotLimitOrdersResponse.OrderHashes)

		err = chainClient.QueueBroadcastMsg(msg)
		if err != nil {
			fmt.Println(err)
			return
		}

		time.Sleep(time.Second * 5)

		gasFee, err := chainClient.GetGasFee()
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("	[+] bought", amount, "INJ, gas fee:", gasFee, "INJ")
		time.Sleep(time.Second * 10)
		// #########################################################################

		bestBuyPrice := decimal.NewFromFloat(res.Market.MidPriceAndTob.BestBuyPrice.MustFloat64() * 1e12)
		amount = decimal.NewFromFloat(1)
		fmt.Println("	[!] best sell price:", bestBuyPrice, "USDT")
		fmt.Println("	[!] trying to SELL", amount, "INJ...")

		order = chainClient.CreateSpotOrder(
			defaultSubaccountID,
			&chainclient.SpotOrderData{
				OrderType:    exchangetypes.OrderType_SELL,
				Quantity:     amount,
				Price:        bestBuyPrice,
				FeeRecipient: senderAddress.String(),
				MarketId:     marketId,
				Cid:          uuid.NewString(),
			},
			marketsAssistant)
		msg = new(exchangetypes.MsgBatchCreateSpotLimitOrders)
		msg.Sender = senderAddress.String()
		msg.Orders = []exchangetypes.SpotOrder{*order}

		simRes, err = chainClient.SimulateMsg(clientCtx, msg)
		if err != nil {
			fmt.Println(err)
			return
		}

		msgBatchCreateSpotLimitOrdersResponse = exchangetypes.MsgBatchCreateSpotLimitOrdersResponse{}
		err = msgBatchCreateSpotLimitOrdersResponse.Unmarshal(simRes.Result.MsgResponses[0].Value)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("	[+] simulated order hashes", msgBatchCreateSpotLimitOrdersResponse.OrderHashes)

		err = chainClient.QueueBroadcastMsg(msg)
		if err != nil {
			fmt.Println(err)
			return
		}

		time.Sleep(time.Second * 5)

		gasFee, err = chainClient.GetGasFee()
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("	[+] sold", amount, "INJ, gas fee:", gasFee, "INJ")
		time.Sleep(time.Second * 10)

		fmt.Println("	[!] get ready for the next section...")
		time.Sleep(time.Second * 10)
	}

	fmt.Println("[!] Finished. Exiting...")
}
