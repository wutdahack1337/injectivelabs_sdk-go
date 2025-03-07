package main

import (
	"fmt"
	"os"
	"time"

	"cosmossdk.io/math"
	"github.com/InjectiveLabs/sdk-go/client"
	chainclient "github.com/InjectiveLabs/sdk-go/client/chain"
	"github.com/InjectiveLabs/sdk-go/client/common"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func main() {
	fmt.Println("[!] Starting...")

	network := common.LoadNetwork("testnet", "lb")
	tmClient, err := rpchttp.New(network.TmEndpoint, "/websocket")
	if err != nil {
		panic(err)
	}

	senderAddress, cosmosKeyring, err := chainclient.InitCosmosKeyring(
		os.Getenv("HOME")+"./injective",
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

	//  Initialize grpc client
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

	// Prepare tx msg
	msg := &banktypes.MsgMultiSend{
		Inputs: []banktypes.Input{
			{
				Address: senderAddress.String(),
				Coins: []sdktypes.Coin{
					{
						Denom:  "inj",
						Amount: math.NewInt(1e18),
					},
					{
						Denom:  "peggy0x87aB3B4C8661e07D6372361211B96ed4Dc36B1B5",
						Amount: math.NewInt(1e6),
					},
				},
			},
		},
		Outputs: []banktypes.Output{
			{
				Address: "inj17apdl7cqsl4ca69axx23r0w00u3nphnmft9cnh",
				Coins: []sdktypes.Coin{
					{
						Denom:  "inj",
						Amount: math.NewInt(1e18),
					},
					{
						Denom:  "peggy0x87aB3B4C8661e07D6372361211B96ed4Dc36B1B5",
						Amount: math.NewInt(1e6),
					},
				},
			},
		},
	}

	err = chainClient.QueueBroadcastMsg(msg)
	if err != nil {
		fmt.Println(err)
	}

	time.Sleep(time.Second * 5)

	gasFee, err := chainClient.GetGasFee()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("[!] gas fee: ", gasFee, "INJ")
}
