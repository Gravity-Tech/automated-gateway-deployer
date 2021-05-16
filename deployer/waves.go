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
	wavesHelper "github.com/Gravity-Tech/gateway-deployer/waves/helper"

	"github.com/wavesplatform/gowaves/pkg/proto"

	"github.com/wavesplatform/gowaves/pkg/crypto"

	wavesClient "github.com/wavesplatform/gowaves/pkg/client"
)

func DeployGatewayOnWaves(privKey string, portType deployer.PortType, commonCfg *cfg.CommonInputConfig, chainConfig *cfg.CrossChainTokenConfig) (*cfg.CrossChainDeploymentOutput, error) {
	const ChainID = 'W'
	const Wavelet = 1e8

	var testConfig wavesHelper.DeploymentConfig
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

	//testConfig.Nebula, err = wavesHelper.GenerateAddressFromSeed('W', chainConfig.NebulaContractSeed)
	testConfig.Nebula, err = wavesHelper.GenerateAddress(ChainID)
	if err != nil {
		return nil, err
	}

	//testConfig.Sub, err = wavesHelper.GenerateAddressFromSeed('W', chainConfig.SubscriberContractSeed)
	testConfig.Sub, err = wavesHelper.GenerateAddress(ChainID)
	if err != nil {
		return nil, err
	}

	nebulaScript, err := wavesHelper.ScriptFromFile(cfg.NebulaScriptFile)
	if err != nil {
		return nil, err
	}

	subScript, err := wavesHelper.ScriptFromFile(cfg.SubMockScriptFile)
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

	err = massTx.Sign(ChainID, distributionSeed)
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

	err = wavesdeployer.DeploySubWaves(testConfig.Client, testConfig.Helper, subScript, nebulaAddressRecipient.String(), chainConfig.AssetID, ChainID, testConfig.Sub.Secret, testConfig.Ctx)
	if err != nil {
		return nil, err
	}

	oraclesString := consulsString[:]
	err = wavesdeployer.DeployNebulaWaves(testConfig.Client, testConfig.Helper, nebulaScript, chainConfig.GravityAddress,
		testConfig.Sub.Address, oraclesString, int64(commonCfg.Bft), contracts.BytesType, ChainID, testConfig.Nebula.Secret, testConfig.Ctx)
	if err != nil {
		return nil, err
	}

	return &cfg.CrossChainDeploymentOutput{
		Gravity: cfg.Account{
			Address: chainConfig.GravityAddress,
		},
		Nebula:  cfg.Account{
			Address: testConfig.Nebula.Address,
			PrivKey: testConfig.Nebula.Secret.String(),
		},
		Port:    cfg.Account{
			Address: testConfig.Sub.Address,
			PrivKey: testConfig.Sub.Secret.String(),
		},
		Token:   chainConfig.AssetID,
	}, nil
}
