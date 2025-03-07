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
		fmt.Print("[-] ")
		panic(err)
	}

	senderAddress, cosmosKeyring, err := chainclient.InitCosmosKeyring(
		os.Getenv("HOME")+"./injective",
		"injectived",
		"file",
		"inj-user",
		"1235678",
		"5d386fbdbf11f1141010f81a46b40f94887367562bd33b452bbaa6ce1cd1381e",
		false,
	)

	clientCtx, err := chainclient.NewClientContext(
		network.ChainId,
		senderAddress.String(),
		cosmosKeyring,
	)
	clientCtx = clientCtx.WithNodeURI(network.TmEndpoint).WithClient(tmClient)

	chainClient, err := chainclient.NewChainClient(
		clientCtx,
		network,
		common.OptionGasPrices(client.DefaultGasPriceWithDenom),
	)

	msg := &banktypes.MsgSend{
		FromAddress: senderAddress.String(),
		ToAddress:   "inj17apdl7cqsl4ca69axx23r0w00u3nphnmft9cnh",
		Amount: []sdktypes.Coin{
			{
				Denom:  "inj",
				Amount: math.NewInt(1000000000000000000),
			},
		},
	}

	err = chainClient.QueueBroadcastMsg(msg)

	time.Sleep(time.Second * 5)
	gasFee, err := chainClient.GetGasFee()
	if err != nil {
		fmt.Println("[-] ", err)
		return
	}
	fmt.Println("[!] gas fee: ", gasFee, "INJ")
}
