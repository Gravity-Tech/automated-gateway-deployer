package deployer

import (
	cfg "github.com/Gravity-Tech/automated-gateway-deployer/config"
	"github.com/Gravity-Tech/gateway-deployer/ethereum/deployer"

	"github.com/Gravity-Tech/gravity-core/common/helpers"

	"context"
	"os"
	"time"

	wavesCrypto "github.com/wavesplatform/go-lib-crypto"

	"github.com/Gravity-Tech/gateway-deployer/waves/contracts"
	wavesdeployer "github.com/Gravity-Tech/gateway-deployer/waves/deployer"

	"github.com/wavesplatform/gowaves/pkg/proto"

	"github.com/wavesplatform/gowaves/pkg/crypto"

	wavesClient "github.com/wavesplatform/gowaves/pkg/client"
)

func DeployGatewayOnWaves(privKey string, portType deployer.PortType, commonCfg *cfg.CommonInputConfig, chainConfig *cfg.CrossChainTokenConfig) (*deployer.GatewayPort, error) {
	const Wavelet = 1e8

	var testConfig waves.DeploymentConfig
	testConfig.Ctx = context.Background()

	wClient, err := wavesClient.NewClient(wavesClient.Options{ApiKey: "", BaseUrl: chainConfig.NodeURL })
	if err != nil {
		return nil, err
	}
	testConfig.Client = wClient
	testConfig.Helper = helpers.NewClientHelper(testConfig.Client)

	testConfig.Consuls = make([]string, 5, 5)
	for i := 0; i < 5; i++ {
		if i < len(chainConfig.ConsulsList) {
			testConfig.Consuls[i] = chainConfig.ConsulsList[i]
		} else {
			testConfig.Consuls[i] = "1"
		}
	}

	testConfig.Nebula, err = waves.GenerateAddressFromSeed('W', chainConfig.NebulaContractSeed)
	if err != nil {
		return nil, err
	}

	testConfig.Sub, err = waves.GenerateAddressFromSeed(cfg.ChainId, cfg.SubscriberContractSeed)
	if err != nil {
		return nil, err
	}

	nebulaScript, err := waves.ScriptFromFile(cfg.NebulaScriptFile)
	if err != nil {
		return nil, err
	}

	subScript, err := waves.ScriptFromFile(cfg.SubMockScriptFile)
	if err != nil {
		return nil, err
	}

	wCrypto := wavesCrypto.NewWavesCrypto()
	distributorPrivKey := os.Getenv("DEPLOYER_PRIV_KEY")
	distributionSeed, err := crypto.NewSecretKeyFromBase58(string(wCrypto.PrivateKey(wavesCrypto.Seed(distributorPrivKey))))
	if err != nil {
		return nil, err
	}

	nebulaAddressRecipient, err := proto.NewRecipientFromString(testConfig.Nebula.Address)
	if err != nil {
		return nil, err
	}
	subAddressRecipient, err := proto.NewRecipientFromString(testConfig.Sub.Address)
	if err != nil {
		return nil, err
	}

	massTx := &proto.MassTransferWithProofs{
		Type:      proto.MassTransferTransaction,
		Version:   1,
		SenderPK:  crypto.GeneratePublicKey(distributionSeed),
		Fee:       0.005 * Wavelet,
		Timestamp: wavesClient.NewTimestampFromTime(time.Now()),
		Transfers: []proto.MassTransferEntry{
			{
				Amount:    0.5 * Wavelet,
				Recipient: nebulaAddressRecipient,
			},
			{
				Amount:    0.5 * Wavelet,
				Recipient: subAddressRecipient,
			},
		},
		Attachment: proto.Attachment{},
	}

	err = massTx.Sign(cfg.ChainId, distributionSeed)
	if err != nil {
		return nil, err
	}
	_, err = testConfig.Client.Transactions.Broadcast(testConfig.Ctx, massTx)
	if err != nil {
		return nil, err
	}
	err = <-testConfig.Helper.WaitTx(massTx.ID.String(), testConfig.Ctx)
	if err != nil {
		return nil, err
	}

	var consulsString []string
	for _, v := range testConfig.Consuls {
		consulsString = append(consulsString, v)
	}

	err = wavesdeployer.DeploySubWaves(testConfig.Client, testConfig.Helper, subScript, nebulaAddressRecipient.String(), cfg.AssetID, cfg.ChainId, testConfig.Sub.Secret, testConfig.Ctx)
	if err != nil {
		return nil, err
	}

	oraclesString := consulsString[:]
	err = wavesdeployer.DeployNebulaWaves(testConfig.Client, testConfig.Helper, nebulaScript, cfg.ExistingGravityAddress,
		testConfig.Sub.Address, oraclesString, cfg.BftValue, contracts.BytesType, cfg.ChainId, testConfig.Nebula.Secret, testConfig.Ctx)
	if err != nil {
		return nil, err
	}

	return &testConfig, nil
}
