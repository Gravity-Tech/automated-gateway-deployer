package deployer

import (
	"context"
	"fmt"
	cfg "github.com/Gravity-Tech/automated-gateway-deployer/config"
	"github.com/Gravity-Tech/gateway-deployer/ethereum/deployer"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func DeployGatewayOnEVM(privKey string, portType deployer.PortType, commonCfg *cfg.CommonInputConfig, evmConfig *cfg.CrossChainTokenConfig) (*cfg.CrossChainDeploymentOutput, error) {
	ctx, cancelCtx := context.WithCancel(context.Background())

	defer cancelCtx()

	fmt.Println("Deploy Ethereum contracts")
	fmt.Printf("Node url: %s\n", evmConfig.NodeURL)

	ethClient, err := ethclient.DialContext(ctx, evmConfig.NodeURL)
	if err != nil {
		return nil, err
	}

	privateKey, err := crypto.HexToECDSA(privKey)
	if err != nil {
		return nil, err
	}

	transactor := bind.NewKeyedTransactor(privateKey)

	if evmConfig.ChainType == "ftm" {
		transactor.GasLimit = 2e6
	}

	ethDeployer := deployer.NewEthDeployer(ethClient, transactor)

	fmt.Println("Deploy gravity contract")

	gravityAddress := evmConfig.GravityAddress

	fmt.Printf("Gravity address: %s\n", gravityAddress)

	var consulsList []common.Address
	for _, consul := range evmConfig.ConsulsList {
		consulsList = append(consulsList, common.HexToAddress(consul))
	}

	port, err := ethDeployer.DeployPort(
		gravityAddress,
		int(portType),
		evmConfig.AssetID,
		consulsList,
		commonCfg.Bft,
		portType,
		ctx,
	)

	if err != nil {
		return nil, err
	}

	return &cfg.CrossChainDeploymentOutput{
		Gravity: cfg.Account{
			Address: gravityAddress,
		},
		Nebula: cfg.Account{
			Address: port.NebulaAddress,
		},
		Port: cfg.Account{
			Address: port.PortAddress,
		},
		Token: port.ERC20Address,
	}, nil
}
