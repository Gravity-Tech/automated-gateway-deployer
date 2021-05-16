package deployer

import (
	dp "github.com/Gravity-Tech/gateway-deployer/ethereum/deployer"
)

type CrossChainGatewayDeployer interface {
	DeployOriginGateway() (dp.GatewayPort, error)
	DeployDestinationGateway() (dp.GatewayPort, error)
}
