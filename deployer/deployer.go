package deployer

import (
	"fmt"
	cfg "github.com/Gravity-Tech/automated-gateway-deployer/config"
	evmcfg "github.com/Gravity-Tech/gateway-deployer/ethereum/config"
	"github.com/Gravity-Tech/gateway-deployer/ethereum/deployer"
	"os"
)

type WavesToEVMDeployerConfig struct {

}

type WavesToEVMCrossChainDeployer struct {
	Config WavesToEVMDeployerConfig
}

func Deploy(configPath string) (*cfg.Output, error) {
	var err error

	deploymentResult := new(cfg.Output)
	deploymentConfig := new(cfg.DeployInputConfig)
	err = evmcfg.ParseConfig(configPath, deploymentConfig)
	if err != nil {
		return nil, err
	}

	err = deploymentConfig.Validate()
	if err != nil {
		return nil, err
	}

	if deploymentConfig.OriginToken.ChainType != "waves" || deploymentConfig.DestToken.ChainType == "waves" {
		return nil, fmt.Errorf("evm based chains token wrapping is not supported yet")
	}

	evmConfig := deploymentConfig.DestToken
	wavesConfig := deploymentConfig.OriginToken

	evmPort, err := DeployGatewayOnEVM(os.Getenv(DestinationDeployer), deployer.IBPort, &deploymentConfig.CommonInputConfig, &evmConfig)
	if err != nil {
		return nil, err
	}

	wavesPort, err := DeployGatewayOnWaves(os.Getenv(OriginDeployer), deployer.LUPort, &deploymentConfig.CommonInputConfig, &wavesConfig)
	if err != nil {
		return nil, err
	}

	deploymentResult.Destination = *evmPort
	deploymentResult.Origin = *wavesPort

	return deploymentResult, nil
}