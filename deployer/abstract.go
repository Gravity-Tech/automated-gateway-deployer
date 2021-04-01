package deployer

import "github.com/Gravity-Tech/gateway-deployer/ethereum/deployer"

type CrossChainGatewayDeployer interface {
	DeployOriginGateway() (deployer.GatewayPort, error)
	DeployDestinationGateway() (deployer.GatewayPort, error)
}
